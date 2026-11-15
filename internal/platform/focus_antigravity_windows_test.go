//go:build windows

package platform

// The Antigravity jump regression, pinning the two halves of §61's fix.
//
// The measured failure: an agy session's parentPid is a shell the host spawned
// for one statusline render, so by the time the panel reads the snapshot the pid
// is dead, the ancestry is empty, and every candidate lost its depth bonus. The
// console title the reporter now sends is the only surviving identity.

import "testing"

func TestAntigravityConsoleTitleResolvesWithoutLiveAncestry(t *testing.T) {
	windows := []focusWindow{
		{handle: 11, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "OC | 项目交接与下一步计划", exe: "WindowsTerminal.exe"},
		{handle: 12, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "命令提示符 - agy", exe: "WindowsTerminal.exe"},
		{handle: 13, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "Windows PowerShell", exe: "WindowsTerminal.exe"},
		// A PseudoConsoleWindow is what the shell inside that tab owns; it
		// carries the shell's pid but can never be activated.
		{handle: 14, pid: 71076, class: "PseudoConsoleWindow", title: "", exe: "cmd.exe"},
	}
	target := FocusTarget{
		PID:       273965046, // a synthetic reporter pid
		ParentPID: 32152,     // a transient shell, already gone
		Title:     "命令提示符 - agy",
		Project:   `C:\Users\rongj\Desktop\项目开发\ai-terminal-pet`,
		Source:    "antigravity-statusline",
		Terminal:  "antigravity",
	}
	// No ancestry at all: exactly what the dead parentPid produces.
	selected, ok := selectFocusWindow(windows, target, map[int]int{}, 99999)
	if !ok {
		t.Fatal("a session with a console title must still resolve with no live ancestry")
	}
	if selected.handle != 12 {
		t.Fatalf("selected handle %d (%q), want 12 (the tab whose title is the console title)", selected.handle, selected.title)
	}
	// Behaviour before the fix: no title, no ancestry, nothing to match on.
	blind := target
	blind.Title = ""
	if _, ok := selectFocusWindow(windows, blind, map[int]int{}, 99999); ok {
		t.Fatal("a session with neither ancestry nor title must stay unresolved rather than jump at random")
	}
}

// The same tab, reached the old way: a live agy host whose pid IS in the
// ancestry. The console title must not make the ranking worse.
func TestAntigravityConsoleTitleDoesNotBeatOwnTab(t *testing.T) {
	windows := []focusWindow{
		{handle: 12, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "命令提示符 - agy", exe: "WindowsTerminal.exe"},
		{handle: 13, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "Windows PowerShell", exe: "WindowsTerminal.exe"},
	}
	target := FocusTarget{
		PID: 1, ParentPID: 51688, Title: "Windows PowerShell",
		Project: `C:\Users\rongj\Desktop\项目开发\ai-terminal-pet`,
		Source:  "antigravity-statusline", Terminal: "antigravity",
	}
	selected, ok := selectFocusWindow(windows, target, map[int]int{51688: 0, 68340: 1, 20992: 2}, 99999)
	if !ok || selected.handle != 13 {
		t.Fatalf("selected handle %d (ok=%v), want 13 - the tab whose title is the console title", selected.handle, ok)
	}
}
