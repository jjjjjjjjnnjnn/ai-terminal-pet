package platform

// The round-61 -> round-63 regression, pinned as a pure function so it runs on
// every platform.
//
// 1.0.11 made pet-statusline send the console title in the payload's `title`
// field, because that string IS the Windows Terminal tab title and the jump
// resolver can therefore find the right tab. presence.go also reads `title`, as
// timeless identity evidence. Together those two facts kept finished Antigravity
// conversations on the panel: one tab named "命令提示符 - agy" is shared by every
// agy conversation that ever ran in it, so a conversation whose host died hours
// ago still had "a window that names it".
//
// Measured before the fix on the live runtime directory: four antigravity rows
// aged 2h01m / 2h08m / 2h56m, all with a dead parentPid, all answered
// SessionAlive=true - and `--json` listed them because SessionAlive is what the
// scanner's KeepAlive consults.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestWindowIdentifiesSessionIgnoresATabTitleAsIdentity(t *testing.T) {
	tab := []focusWindow{{handle: 1, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "命令提示符 - agy", exe: "WindowsTerminal.exe"}}
	cases := []struct {
		name  string
		title string
		want  bool
	}{
		// Exactly what the Antigravity statusline sends since 1.0.11: the
		// console title, which is the tab title, which is shared by every
		// conversation in that tab. Not evidence about THIS conversation.
		{"a title equal to the window title", "命令提示符 - agy", false},
	}
	// What a conversation title looks like - a PART of the tab title - must
	// still count, so the fix cannot quietly become "no evidence at all".
	if !windowIdentifiesSession(
		[]focusWindow{{handle: 2, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "OC | 项目交接", exe: "WindowsTerminal.exe"}},
		FocusTarget{Title: "项目交接"}) {
		t.Fatal("a title the window contains must still identify the conversation")
	}
	for _, test := range cases {
		if got := windowIdentifiesSession(tab, FocusTarget{Title: test.title}); got != test.want {
			t.Errorf("%s: windowIdentifiesSession(%q) = %v, want %v", test.name, test.title, got, test.want)
		}
	}
	if !windowIdentifiesSession(tab, FocusTarget{Project: `C:\Users\example\Desktop\命令提示符 - agy`}) {
		t.Fatal("a project name that the window contains must still be evidence")
	}
}

// TestPresenceDoesNotKeepAFinishedConversationAliveByItsTabTitle is the whole
// regression in one call: dead host, four hours old, title equal to the live tab's.
func TestPresenceDoesNotKeepAFinishedConversationAliveByItsTabTitle(t *testing.T) {
	windows := []focusWindow{{handle: 1, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "命令提示符 - agy", exe: "WindowsTerminal.exe"}}
	processes := map[int]focusProcess{11492: {parent: 1, name: "WindowsTerminal.exe"}}
	presence := &Presence{
		ttl:         DefaultPresenceTTL,
		orphanAfter: DefaultOrphanAfter,
		now:         func() time.Time { return time.Unix(4000, 0) },
		load:        func() (map[int]focusProcess, []focusWindow) { return processes, windows },
	}
	value := session.Session{
		Source: "antigravity-statusline", SessionID: "finished-hours-ago",
		PID: 2044089446, ParentPID: 72172,
		Title: "命令提示符 - agy", Project: `C:\Users\example\Desktop\probe`,
		State:     session.StateIdle,
		StartedAt: time.Unix(4000, 0).Add(-5 * time.Hour),
		UpdatedAt: time.Unix(4000, 0).Add(-4 * time.Hour),
	}
	if presence.SessionAlive(value) {
		t.Fatal("a conversation whose host died hours ago was kept alive by the tab title it was reported from")
	}
}

// And the other direction, so the fix cannot quietly turn into "no evidence at
// all": a live host keeps its conversation whatever the title says.
func TestPresenceKeepsALiveHostWhateverTheTitleSays(t *testing.T) {
	windows := []focusWindow{{handle: 1, pid: 11492, class: "CASCADIA_HOSTING_WINDOW_CLASS", title: "命令提示符 - agy", exe: "WindowsTerminal.exe"}}
	processes := map[int]focusProcess{
		11492: {parent: 1, name: "WindowsTerminal.exe"},
		17564: {parent: 11492, name: "agy.exe"},
	}
	presence := &Presence{
		ttl:         DefaultPresenceTTL,
		orphanAfter: DefaultOrphanAfter,
		now:         func() time.Time { return time.Unix(4000, 0) },
		load:        func() (map[int]focusProcess, []focusWindow) { return processes, windows },
	}
	value := session.Session{
		Source: "antigravity-statusline", SessionID: "live-now",
		PID: 909284829, ParentPID: 17564, Title: "命令提示符 - agy",
		State:     session.StateIdle,
		StartedAt: time.Unix(4000, 0).Add(-time.Minute),
		UpdatedAt: time.Unix(4000, 0).Add(-time.Minute),
	}
	if !presence.SessionAlive(value) {
		t.Fatal("a conversation whose host is alive was dropped")
	}
}
