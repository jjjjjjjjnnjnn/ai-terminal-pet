package platform

import (
	"os"
	"testing"
)

func TestAncestorDepthsWalksFromParentPID(t *testing.T) {
	parents := map[int]int{
		2832:  2896,
		2896:  11492,
		11492: 14576,
		14576: 4068,
		4068:  9356,
		9356:  4,
		4:     0,
	}
	parentOf := func(pid int) (int, bool) {
		value, ok := parents[pid]
		return value, ok
	}
	target := FocusTarget{PID: 5228, ParentPID: 2832}
	depths := ancestorDepths(target, parentOf, maxFocusDepth)

	// The parent is the reliable root for hook reporters; the synthetic pid
	// is present but ranked one step further away.
	if depths[2832] != 0 {
		t.Fatalf("parentPid depth = %d, want 0", depths[2832])
	}
	if depths[11492] != 2 {
		t.Fatalf("terminal pid depth = %d, want 2", depths[11492])
	}
	if depths[5228] != 1 {
		t.Fatalf("synthetic pid depth = %d, want 1", depths[5228])
	}
	if _, ok := depths[0]; ok {
		t.Fatal("pid 0 must never enter the ancestry")
	}
}

func TestAncestorDepthsStopsAtDepthAndCycles(t *testing.T) {
	parentOf := func(pid int) (int, bool) { return pid + 1, true }
	depths := ancestorDepths(FocusTarget{ParentPID: 1}, parentOf, 3)
	for pid := 1; pid <= 4; pid++ {
		if _, ok := depths[pid]; !ok {
			t.Fatalf("pid %d missing from bounded chain", pid)
		}
	}
	if _, ok := depths[5]; ok {
		t.Fatal("chain exceeded maxFocusDepth")
	}

	// A process table can contain a parent cycle (reused PIDs). It must not
	// hang or overflow the recursion.
	cyclic := func(pid int) (int, bool) {
		if pid == 7 {
			return 8, true
		}
		return 7, true
	}
	if got := ancestorDepths(FocusTarget{ParentPID: 7}, cyclic, maxFocusDepth); len(got) != 2 {
		t.Fatalf("cyclic ancestry produced %d entries, want 2", len(got))
	}
}

func TestTerminalWindowScoreFiltersNonTerminals(t *testing.T) {
	cases := []struct {
		class string
		exe   string
		want  int
	}{
		{"CASCADIA_HOSTING_WINDOW_CLASS", "WindowsTerminal.exe", 30},
		{"ConsoleWindowClass", "conhost.exe", 30},
		{"Chrome_WidgetWin_1", "Code.exe", 24},
		{"Chrome_WidgetWin_1", "firefox.exe", -1},
		{"CabinetWClass", "explorer.exe", -1},
		{"Progman", "explorer.exe", -1},
		{"", "unknown.exe", -1},
	}
	for _, test := range cases {
		if got := terminalWindowScore(test.class, test.exe); got != test.want {
			t.Errorf("terminalWindowScore(%q, %q) = %d, want %d", test.class, test.exe, got, test.want)
		}
	}
}

func TestTerminalKeywordAndProjectBase(t *testing.T) {
	cases := []struct {
		source   string
		terminal string
		want     string
	}{
		{"opencode", "opencode", "oc |"},
		{"antigravity-statusline", "antigravity", "agy"},
		{"claude-ipc", "claude", "claude"},
		{"powershell-ipc", "powershell", "powershell"},
		{"", "", ""},
	}
	for _, test := range cases {
		if got := terminalKeyword(test.source, test.terminal); got != test.want {
			t.Errorf("terminalKeyword(%q, %q) = %q, want %q", test.source, test.terminal, got, test.want)
		}
	}
	baseCases := []struct{ in, want string }{
		{`C:\Users\rongj\Desktop\项目开发\DSH-Addition`, "DSH-Addition"},
		{"C:/Users/rongj/Desktop/ai-terminal-pet", "ai-terminal-pet"},
		{`C:\`, ""},
		{"  ", ""},
	}
	for _, test := range baseCases {
		if got := projectBase(test.in); got != test.want {
			t.Errorf("projectBase(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}

func TestSelectFocusWindowPrefersMatchingTab(t *testing.T) {
	windows := []focusWindow{
		{handle: 11, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "OC | 项目交接与后续计划", exe: "WindowsTerminal.exe"},
		{handle: 12, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "命令提示符 - agy", exe: "WindowsTerminal.exe"},
		{handle: 13, pid: 15032, class: "MozillaWindowClass", title: "OC | notes", exe: "firefox.exe"},
		{handle: 14, pid: 40788, class: "AITerminalPetWindow", title: "AI Terminal Pet", exe: "AI-Terminal-Pet.exe"},
	}
	depths := map[int]int{2832: 0, 2896: 1, 11492: 2, 5228: 1}
	target := FocusTarget{
		PID: 5228, ParentPID: 2832,
		Title: "项目交接与后续计划", Project: `C:\Users\rongj\Desktop\项目开发\DSH-Addition`,
		Source: "opencode", Terminal: "opencode",
	}
	got, ok := selectFocusWindow(windows, target, depths, 40788)
	if !ok {
		t.Fatal("expected a matching terminal window")
	}
	if got.handle != 11 {
		t.Fatalf("selected handle %d, want 11 (the tab whose title matches the session)", got.handle)
	}
}

func TestSelectFocusWindowRejectsWeakOrSelfMatches(t *testing.T) {
	self := []focusWindow{
		{handle: 14, pid: 40788, class: "AITerminalPetWindow", title: "AI Terminal Pet", exe: "AI-Terminal-Pet.exe"},
	}
	if _, ok := selectFocusWindow(self, FocusTarget{}, map[int]int{40788: 0}, 40788); ok {
		t.Fatal("the overlay window must never be its own jump target")
	}

	unrelated := []focusWindow{
		{handle: 21, pid: 99999, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "Unrelated tab", exe: "WindowsTerminal.exe"},
	}
	target := FocusTarget{Title: "Some Session", Source: "opencode", Terminal: "opencode"}
	if _, ok := selectFocusWindow(unrelated, target, map[int]int{}, 40788); ok {
		t.Fatal("an off-chain, title-mismatched terminal must not be jumped to")
	}

	// The same window becomes valid once it is on the session's process chain.
	linked := []focusWindow{
		{handle: 21, pid: 99999, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "Unrelated tab", exe: "WindowsTerminal.exe"},
	}
	if _, ok := selectFocusWindow(linked, target, map[int]int{99999: 1}, 40788); !ok {
		t.Fatal("a terminal on the session's ancestor chain should be a valid target")
	}
}

// TestResolveFocusWindowLive is an opt-in probe against the real desktop.
// It only resolves a target - it never moves the foreground window - so it is
// safe to run while other work is open.
//
//	PET_FOCUS_PROBE=1 go test ./internal/platform -run TestResolveFocusWindowLive -v
func TestResolveFocusWindowLive(t *testing.T) {
	if os.Getenv("PET_FOCUS_PROBE") == "" {
		t.Skip("set PET_FOCUS_PROBE=1 to probe the live desktop")
	}
	got, ok := resolveFocusWindow(FocusTarget{
		Title:    "项目交接与后续计划",
		Project:  `C:/Users/rongj/Desktop/项目开发/DSH-Addition`,
		Source:   "opencode",
		Terminal: "opencode",
	})
	if !ok {
		t.Fatal("no terminal window matched - is the OpenCode tab still open?")
	}
	t.Logf("matched window: pid=%d class=%q title=%q", got.pid, got.class, got.title)
}

func TestActivateTargetRejectsZeroHandle(t *testing.T) {
	if ActivateTarget(0) {
		t.Fatal("a zero HWND must never reach SetForegroundWindow")
	}
}

// Opt-in desktop probes: the jump used to close the window the user was
// working in. These pin the two halves of the fix.
//
//	PET_FOCUS_PROBE=1 go test ./internal/platform -run 'Test(ResolveTargetDoesNotStealForeground|ActivateTargetKeepsEveryWindow)' -v
func TestResolveTargetDoesNotStealForeground(t *testing.T) {
	if os.Getenv("PET_FOCUS_PROBE") == "" {
		t.Skip("set PET_FOCUS_PROBE=1 for desktop probes")
	}
	before, _, _ := getForegroundWindowProc.Call()
	ResolveTarget(FocusTarget{
		Title: "项目交接与后续计划", Source: "opencode", Terminal: "opencode",
		Project: `C:/Users/rongj/Desktop/项目开发/DSH-Addition`,
	})
	after, _, _ := getForegroundWindowProc.Call()
	if before != after {
		t.Fatalf("window lookup changed the foreground: %d -> %d", before, after)
	}
}

func TestActivateTargetKeepsEveryWindow(t *testing.T) {
	if os.Getenv("PET_FOCUS_PROBE") == "" {
		t.Skip("set PET_FOCUS_PROBE=1 for desktop probes")
	}
	visible := func() map[uintptr]bool {
		handles := make(map[uintptr]bool)
		for _, window := range visibleWindows(nil) {
			handles[window.handle] = true
		}
		return handles
	}
	before := visible()
	foreground, _, _ := getForegroundWindowProc.Call()
	if foreground == 0 {
		t.Skip("no foreground window")
	}
	if !ActivateTarget(foreground) {
		t.Fatal("activating the already-foreground window should succeed")
	}
	after := visible()
	for handle := range before {
		if !after[handle] {
			t.Fatalf("activating a window closed another one (HWND %d disappeared)", handle)
		}
	}
}

// DeepSeek Harness is an Electron desktop app, not a conhost/cascadia tab.
// Without recognising its window class and title token, a Harness session
// could be displayed but never jumped to.
func TestDeepSeekHarnessWindowIsAJumpTarget(t *testing.T) {
	if score := terminalWindowScore("Chrome_WidgetWin_1", "deepseek harness.exe"); score < 24 {
		t.Fatalf("Harness window scored %d, want >= 24", score)
	}
	if got := terminalKeyword("deepseek-harness", "deepseek-harness"); got != "deepseek" {
		t.Fatalf("terminalKeyword = %q, want %q", got, "deepseek")
	}
	window := terminalWindow(9001, 9001, "DeepSeek Harness")
	window.class = "Chrome_WidgetWin_1"
	window.exe = "DeepSeek Harness.exe"
	selected, ok := selectFocusWindow([]focusWindow{window}, FocusTarget{
		Source: "deepseek-harness", Terminal: "deepseek-harness",
	}, map[int]int{9001: 0}, 12345)
	if !ok || selected.handle != window.handle {
		t.Fatalf("Harness session did not resolve to its own window (ok=%v)", ok)
	}
}

// TestResolveTargetFindsDeepSeekHarness is an opt-in, read-only desktop probe:
// it resolves the Harness window without ever activating it.
//
//	PET_DSH_PROBE=1 go test ./internal/platform -run TestResolveTargetFindsDeepSeekHarness -v
func TestResolveTargetFindsDeepSeekHarness(t *testing.T) {
	if os.Getenv("PET_DSH_PROBE") == "" {
		t.Skip("set PET_DSH_PROBE=1 to probe the live Harness window")
	}
	before, _, _ := getForegroundWindowProc.Call()
	handle, title, ok := ResolveTarget(FocusTarget{
		Title:    "DeepSeek Harness",
		Source:   "deepseek-harness",
		Terminal: "deepseek-harness",
	})
	after, _, _ := getForegroundWindowProc.Call()
	if before != after {
		t.Fatalf("lookup changed the foreground window: %d -> %d", before, after)
	}
	if !ok || handle == 0 {
		t.Fatal("no Harness window matched - is DeepSeek Harness open?")
	}
	t.Logf("matched: hwnd=%d title=%q", handle, title)
}
