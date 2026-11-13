package platform

// The two "ghost row" rules, both found by measuring the machine that reported
// the bug rather than by reading the code:
//
//   - a dead reporter process ends its conversation, and no window title may
//     overrule that (three of six rows on the real machine came from pids that
//     no longer existed);
//   - a title that names the tool rather than the conversation is not identity
//     (the plugin writes "opencode" before it has learned the session title, and
//     the tab is often titled "OpenCode" too).

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// The exact shape that produced the ghost rows: a trusted pid that has exited,
// a title that matches a live tab, and a stale timestamp.
func TestPresenceDeadTrustedPIDIsNotRescuedByATab(t *testing.T) {
	clock := time.Unix(5000, 0)
	procs := map[int]focusProcess{
		20992: {parent: 1, name: "opencode.exe"},
	}
	windows := []focusWindow{
		{handle: 1, pid: 20992, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "OC | 项目交接与后续计划", exe: "WindowsTerminal.exe"},
	}
	presence := fakePresence(time.Minute, &clock, procs, windows)

	dead := session.Session{
		Source: "opencode", SessionID: "ses_dead", PID: 46108, ParentPID: 26600,
		Title: "项目交接与后续计划", State: session.StateWorking,
		UpdatedAt: clock.Add(-12 * time.Hour),
	}
	if presence.SessionAlive(dead) {
		t.Fatal("a dead reporter pid was resurrected by a tab belonging to another process")
	}

	// The same conversation, now reported by the live process: kept.
	live := dead
	live.PID = 20992
	if !presence.SessionAlive(live) {
		t.Fatal("the live reporter lost its own conversation")
	}
}

// Hook reporters carry a synthetic pid, so process death proves nothing and the
// title must stay usable as evidence. This is the case the title rule exists
// for, and the fix above must not take it away.
func TestPresenceSyntheticPIDStillUsesTheTabTitle(t *testing.T) {
	clock := time.Unix(6000, 0)
	procs := map[int]focusProcess{500: {parent: 400, name: "claude.exe"}}
	windows := []focusWindow{terminalWindow(1, 500, "claude · refactor the parser")}
	presence := fakePresence(time.Minute, &clock, procs, windows)

	value := session.Session{
		Source: "claude-hook", SessionID: "ses_hook", PID: 1701354615, ParentPID: 500,
		Title: "refactor the parser", State: session.StateIdle,
		UpdatedAt: clock.Add(-6 * time.Hour),
	}
	if !presence.SessionAlive(value) {
		t.Fatal("a hook session with a named tab was dropped for silence")
	}
}

// The plugin writes the tool name as the title until it has learned the real
// one. That is the literal string "opencode", and the tab is often "OpenCode".
func TestBareToolTitlesAreNotIdentity(t *testing.T) {
	for _, title := range []string{
		"opencode", "OpenCode", "OC", "claude", "Claude Code", "agy",
		"antigravity", "codex", "deepseek harness", "Terminal", "powershell",
	} {
		if titleIdentifiesConversation(title) {
			t.Errorf("%q is a tool name, not a conversation title", title)
		}
	}
	for _, title := range []string{
		"项目交接与后续计划", "refactor the parser", "OC | fix the parser",
		"claude · migrate to sqlite",
	} {
		if !titleIdentifiesConversation(title) {
			t.Errorf("%q names a conversation and should be identity", title)
		}
	}
	if titleIdentifiesConversation("ab") {
		t.Error("a two-character title is too short to identify anything")
	}
}

// The measured bug in one test: a session titled "opencode" must not be kept
// alive by a window titled "OpenCode".
func TestSessionTitledWithTheToolNameIsNotKeptAlive(t *testing.T) {
	target := FocusTarget{Title: "opencode", Source: "opencode", Terminal: "opencode"}
	window := focusWindow{
		handle: 1, pid: 30, class: "CASCADIA_HOSTING_WINDOW_CLASS",
		title: "OpenCode", exe: "WindowsTerminal.exe",
	}
	if windowIdentifiesSession([]focusWindow{window}, target) {
		t.Fatal("a tool name matched a window titled with the same tool name")
	}
	// A real session title still matches its real tab.
	real := FocusTarget{Title: "接手项目并检查现状", Source: "opencode"}
	tab := window
	tab.title = "OC | 接手项目并检查现状"
	if !windowIdentifiesSession([]focusWindow{tab}, real) {
		t.Fatal("a real session title no longer matches its tab")
	}
}
