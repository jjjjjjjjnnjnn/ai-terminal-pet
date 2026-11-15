package platform

import (
	"os"
	"runtime"
	"strings"
)

const maxFocusDepth = 6

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

// FocusTerminal is the convenience combination for non-UI callers (probes,
// tests). The overlay must use ResolveTarget + a UI-thread activation post
// so activation happens on the UI thread.
func FocusTerminal(target FocusTarget) (string, bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	handle, title, ok := ResolveTarget(target)
	if !ok {
		return "", false
	}
	return title, ActivateTarget(handle)
}

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
	case isPOSIXTerminalClass(lowerClass):
		return 30
	case (lowerClass == "chrome_widgetwin_1" || lowerClass == "chrome_widgetwin_0") && isEditorProcess(lowerExe):
		return 24
	case isEditorProcess(lowerClass) || isEditorProcess(lowerExe):
		return 24
	case isCLIProcess(lowerExe):
		return 12
	}
	return -1
}

func isPOSIXTerminalClass(lowerClass string) bool {
	for _, token := range []string{
		"gnome-terminal", "kitty", "alacritty", "wezterm", "konsole",
		"foot", "xterm", "terminator", "tilix", "xfce4-terminal",
		"lxterminal", "mate-terminal", "urxvt", "ghostty", "tabby",
		"warp", "hyper", "st-256color", "sakura", "qterminal", "weston-terminal",
	} {
		if strings.Contains(lowerClass, token) {
			return true
		}
	}
	return false
}

func isEditorProcess(exe string) bool {
	for _, name := range []string{
		"code.exe", "code - insiders.exe", "cursor.exe", "windsurf.exe",
		"workbuddyai.exe", "trae.exe",
		// DeepSeek Harness is an Electron desktop app (Chrome_WidgetWin_1)
		// hosting its own terminal sessions.
		"deepseek harness.exe", "deepseek harness", "deepseek-harness",
		"vscode", "vscodium", "code-oss", "code - insiders", "cursor", "windsurf", "trae",
	} {
		if strings.Contains(exe, name) {
			return true
		}
	}
	// Exact short name check for Linux X11 wm_class/comm "code".
	for _, field := range strings.Fields(strings.NewReplacer(",", " ", ";", " ", "/", " ", "\\", " ").Replace(exe)) {
		if field == "code" {
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
		"gnome-terminal", "kitty", "alacritty", "wezterm", "konsole",
		"foot", "xterm", "ghostty", "tmux", "zsh", "fish", "bash", "pwsh",
		"node", "opencode", "claude", "antigravity", "agy", "codex",
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

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
