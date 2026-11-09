package platform

// Jump-to-terminal support.
//
// A session knows its reporter PID, its parent PID, its project and its title,
// but not a window handle. Reporters are fire-and-forget hooks that must never
// pay for a Win32 lookup, and several of them deliberately use synthetic PIDs
// so the store can keep one stable session across thousands of short-lived
// hook processes. So the overlay resolves the terminal itself:
//
//   1. snapshot the process table (pid -> parent pid, exe name)
//   2. walk up from the session's parentPid / pid to find ancestor processes
//   3. enumerate visible top-level windows and keep terminal-looking ones
//   4. score a window by: ancestor proximity + session title + project name
//      + terminal keyword, then foreground the best match
//
// Windows Terminal gives every tab its own top-level CASCADIA_HOSTING_WINDOW_CLASS
// window whose title carries the tab name (for example "OC | project handoff"),
// which is enough to pick the right tab rather than just some terminal.
//
// The whole lookup runs off the UI thread: GetWindowText / EnumWindows touch
// other processes and a hung window must never stall the window procedure.

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	th32csSnapProcess = 0x00000002
	maxFocusDepth     = 6
	swRestore         = 9
)

var (
	focusUser32 = syscall.NewLazyDLL("user32.dll")

	createToolhelp32SnapshotProc = kernel32.NewProc("CreateToolhelp32Snapshot")
	process32FirstProc           = kernel32.NewProc("Process32FirstW")
	process32NextProc            = kernel32.NewProc("Process32NextW")
	enumWindowsProc              = focusUser32.NewProc("EnumWindows")
	isWindowVisibleProc          = focusUser32.NewProc("IsWindowVisible")
	getWindowThreadProcessIdProc = focusUser32.NewProc("GetWindowThreadProcessId")
	getClassNameProc             = focusUser32.NewProc("GetClassNameW")
	getWindowTextProc            = focusUser32.NewProc("GetWindowTextW")
	isIconicProc                 = focusUser32.NewProc("IsIconic")
	showWindowProc               = focusUser32.NewProc("ShowWindow")
	setForegroundWindowProc      = focusUser32.NewProc("SetForegroundWindow")
	getForegroundWindowProc      = focusUser32.NewProc("GetForegroundWindow")
)

// FocusTarget is the part of a session that identifies its terminal.
type FocusTarget struct {
	PID       int
	ParentPID int
	Title     string
	Project   string
	Source    string
	Terminal  string
}

type focusProcess struct {
	parent int
	name   string
}

type focusWindow struct {
	handle uintptr
	pid    int
	class  string
	title  string
	exe    string
}

type processEntry32W struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

// ResolveTarget locates the terminal window for a session WITHOUT touching
// the foreground or any window's z-order. It is safe from a background
// goroutine: all it does is enumerate processes and windows.
func ResolveTarget(target FocusTarget) (uintptr, string, bool) {
	best, ok := resolveFocusWindow(target)
	if !ok {
		return 0, "", false
	}
	return best.handle, best.title, true
}

// ActivateTarget brings exactly one window forward.
//
// It deliberately does NOT use AttachThreadInput or SetWindowPos fallbacks.
// Those old fallbacks reached into whichever window the user was working in:
// AttachThreadInput merged our background thread with the foreground thread's
// input queue, and the TOPMOST/NOTOPMOST ping reordered every window's
// z-order. Both could make the window the user was in the middle of using
// lose focus or close. The overlay has just been clicked, so it already owns
// the foreground and SetForegroundWindow succeeds on its own; if it does not,
// we report failure instead of forcing it.
//
// Call this on the UI thread. It must not LockOSThread/UnlockOSThread: the UI
// thread is pinned for the whole window's lifetime and unpinning it would
// recreate the ghost-window bug.
func ActivateTarget(handle uintptr) bool {
	if handle == 0 {
		return false
	}
	if minimized, _, _ := isIconicProc.Call(handle); minimized != 0 {
		showWindowProc.Call(handle, swRestore)
	}
	if result, _, _ := setForegroundWindowProc.Call(handle); result != 0 {
		return true
	}
	// Some systems report failure even though the switch landed.
	now, _, _ := getForegroundWindowProc.Call()
	return now == handle
}

// FocusTerminal is the convenience combination for non-UI callers (probes,
// tests). The overlay must use ResolveTarget + a posted wmFocusActivate so
// activation happens on the UI thread - see internal/ui.jumpTo.
func FocusTerminal(target FocusTarget) (string, bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	handle, title, ok := ResolveTarget(target)
	if !ok {
		return "", false
	}
	return title, ActivateTarget(handle)
}

// resolveFocusWindow performs the lookup without changing the foreground
// window. Keeping it separate lets the live-desktop test inspect what would
// be selected without stealing focus from whatever the user is doing.
func resolveFocusWindow(target FocusTarget) (focusWindow, bool) {
	processes := snapshotProcesses()
	windows := visibleWindows(processes)
	parentOf := func(pid int) (int, bool) {
		info, ok := processes[pid]
		return info.parent, ok
	}
	depths := ancestorDepths(target, parentOf, maxFocusDepth)
	return selectFocusWindow(windows, target, depths, os.Getpid())
}

func snapshotProcesses() map[int]focusProcess {
	result := make(map[int]focusProcess)
	handle, _, _ := createToolhelp32SnapshotProc.Call(th32csSnapProcess, 0)
	if handle == 0 || handle == ^uintptr(0) {
		return result
	}
	defer closeHandle.Call(handle)
	var entry processEntry32W
	entry.Size = uint32(unsafe.Sizeof(entry))
	ok, _, _ := process32FirstProc.Call(handle, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		if entry.ProcessID != 0 {
			result[int(entry.ProcessID)] = focusProcess{
				parent: int(entry.ParentProcessID),
				name:   syscall.UTF16ToString(entry.ExeFile[:]),
			}
		}
		ok, _, _ = process32NextProc.Call(handle, uintptr(unsafe.Pointer(&entry)))
	}
	return result
}

// EnumWindows needs a C callback, which cannot capture Go locals. A package
// level collector guarded by a mutex keeps the call synchronous and safe.
var (
	focusEnumMu     sync.Mutex
	focusEnumTarget *[]focusWindow
	focusEnumProcs  map[int]focusProcess
)

var focusEnumCallback = syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
	if focusEnumTarget == nil {
		return 1
	}
	if visible, _, _ := isWindowVisibleProc.Call(hwnd); visible == 0 {
		return 1
	}
	var pid uint32
	getWindowThreadProcessIdProc.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	classBuf := make([]uint16, 256)
	getClassNameProc.Call(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), uintptr(len(classBuf)))
	titleBuf := make([]uint16, 512)
	getWindowTextProc.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), uintptr(len(titleBuf)))
	*focusEnumTarget = append(*focusEnumTarget, focusWindow{
		handle: hwnd,
		pid:    int(pid),
		class:  syscall.UTF16ToString(classBuf),
		title:  syscall.UTF16ToString(titleBuf),
		exe:    focusEnumProcs[int(pid)].name,
	})
	return 1
})

func visibleWindows(processes map[int]focusProcess) []focusWindow {
	focusEnumMu.Lock()
	defer focusEnumMu.Unlock()
	collected := make([]focusWindow, 0, 32)
	focusEnumTarget = &collected
	focusEnumProcs = processes
	enumWindowsProc.Call(uintptr(focusEnumCallback), 0)
	focusEnumTarget = nil
	focusEnumProcs = nil
	return collected
}

// ancestorDepths maps every process in the session's ancestry to its distance.
// parentPid is the reliable root for hook reporters (their own pid is often a
// synthetic hash), so it starts at depth 0 and pid starts at depth 1.
func ancestorDepths(target FocusTarget, parentOf func(int) (int, bool), maxDepth int) map[int]int {
	depths := make(map[int]int)
	seen := make(map[int]bool)
	var walk func(pid, depth int)
	walk = func(pid, depth int) {
		if pid <= 0 || depth > maxDepth || seen[pid] {
			return
		}
		seen[pid] = true
		if previous, ok := depths[pid]; !ok || depth < previous {
			depths[pid] = depth
		}
		parent, ok := parentOf(pid)
		if ok && parent > 0 && parent != pid {
			walk(parent, depth+1)
		}
	}
	walk(target.ParentPID, 0)
	walk(target.PID, 1)
	return depths
}

// terminalWindowScore rejects windows that cannot plausibly host a CLI.
func terminalWindowScore(class, exe string) int {
	lowerClass := strings.ToLower(class)
	lowerExe := strings.ToLower(exe)
	switch {
	case strings.Contains(lowerClass, "cascadia_hosting_window_class"):
		return 30
	case lowerClass == "consolewindowclass" || strings.Contains(lowerClass, "mintty"):
		return 30
	case (lowerClass == "chrome_widgetwin_1" || lowerClass == "chrome_widgetwin_0") && isEditorProcess(lowerExe):
		return 24
	case isCLIProcess(lowerExe):
		return 12
	}
	return -1
}

func isEditorProcess(exe string) bool {
	for _, name := range []string{
		"code.exe", "code - insiders.exe", "cursor.exe", "windsurf.exe",
		"workbuddyai.exe", "trae.exe",
		// DeepSeek Harness is an Electron desktop app (Chrome_WidgetWin_1)
		// hosting its own terminal sessions. Its window is the only jump
		// target a Harness session can have.
		"deepseek harness.exe",
	} {
		if strings.Contains(exe, name) {
			return true
		}
	}
	return false
}

func isCLIProcess(exe string) bool {
	for _, name := range []string{
		"windowsterminal.exe", "openconsole.exe", "conhost.exe", "pwsh.exe",
		"powershell.exe", "cmd.exe", "wsl.exe", "bash.exe", "node.exe",
		"opencode.exe", "claude.exe", "antigravity.exe", "agy.exe", "codex.exe",
		"deepseek harness.exe",
	} {
		if strings.Contains(exe, name) {
			return true
		}
	}
	return false
}

// terminalKeyword is a lowercase token commonly found in the tab title of
// that CLI. OpenCode titles look like "OC | <session title>", so "oc |" is
// deliberately more specific than "oc" (which appears in unrelated words).
func terminalKeyword(source, terminal string) string {
	combined := strings.ToLower(source + " " + terminal)
	switch {
	case strings.Contains(combined, "opencode"):
		return "oc |"
	case strings.Contains(combined, "antigravity") || strings.Contains(combined, "agy"):
		return "agy"
	case strings.Contains(combined, "claude"):
		return "claude"
	case strings.Contains(combined, "codex"):
		return "codex"
	case strings.Contains(combined, "powershell") || strings.Contains(combined, "pwsh"):
		return "powershell"
	case strings.Contains(combined, "deepseek") || strings.Contains(combined, "harness") || strings.Contains(combined, "dsh-"):
		return "deepseek"
	}
	return ""
}

func projectBase(project string) string {
	trimmed := strings.TrimSpace(project)
	if trimmed == "" {
		return ""
	}
	parts := strings.FieldsFunc(trimmed, func(r rune) bool { return r == '\\' || r == '/' })
	if len(parts) == 0 {
		return ""
	}
	base := strings.TrimSpace(parts[len(parts)-1])
	if len([]rune(base)) < 3 {
		return ""
	}
	return base
}

// selectFocusWindow is pure: the OS-dependent parts are passed in as data so
// the ranking can be unit tested without touching the desktop.
func selectFocusWindow(windows []focusWindow, target FocusTarget, depths map[int]int, selfPID int) (focusWindow, bool) {
	best := focusWindow{}
	bestScore := 0
	title := strings.ToLower(strings.TrimSpace(target.Title))
	base := strings.ToLower(projectBase(target.Project))
	keyword := terminalKeyword(target.Source, target.Terminal)

	for _, candidate := range windows {
		if candidate.handle == 0 || candidate.pid == selfPID ||
			strings.EqualFold(candidate.class, "AITerminalPetWindow") ||
			strings.TrimSpace(candidate.title) == "" {
			continue
		}
		classScore := terminalWindowScore(candidate.class, candidate.exe)
		if classScore < 0 {
			continue
		}
		score := classScore
		if depth, ok := depths[candidate.pid]; ok {
			score += 80 - minInt(depth, 5)*8
		} else {
			score -= 40
		}
		windowTitle := strings.ToLower(candidate.title)
		if len([]rune(title)) >= 3 && strings.Contains(windowTitle, title) {
			score += 100
		}
		if base != "" && strings.Contains(windowTitle, base) {
			score += 55
		}
		if keyword != "" && strings.Contains(windowTitle, keyword) {
			score += 35
		}
		if score > bestScore {
			bestScore, best = score, candidate
		}
	}
	// A weak match would yank an unrelated window to the front, which is
	// worse than telling the user we could not find the terminal.
	if bestScore < 40 {
		return focusWindow{}, false
	}
	return best, true
}

// The unsafe fallbacks that used to live here (AttachThreadInput on the
// foreground thread, BringWindowToTop, and SetWindowPos TOPMOST/NOTOPMOST)
// were removed. They reached into the window the user was working in and
// could make it lose focus or close. See ActivateTarget above.

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
