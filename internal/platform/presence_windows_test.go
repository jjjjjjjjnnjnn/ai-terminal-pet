package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func fakePresence(ttl time.Duration, clock *time.Time, procs map[int]focusProcess, windows []focusWindow) *Presence {
	presence := NewPresence(ttl)
	presence.now = func() time.Time { return *clock }
	presence.load = func() (map[int]focusProcess, []focusWindow) {
		return procs, windows
	}
	return presence
}

func terminalWindow(handle uintptr, pid int, title string) focusWindow {
	return focusWindow{
		handle: handle,
		pid:    pid,
		class:  "CASCADIA_HOSTING_WINDOW_CLASS",
		title:  title,
		exe:    "WindowsTerminal.exe",
	}
}

func TestPresenceCachesOneSnapshotPerWindow(t *testing.T) {
	clock := time.Unix(1000, 0)
	presence := fakePresence(10*time.Second, &clock,
		map[int]focusProcess{42: {parent: 7, name: "node.exe"}}, nil)

	presence.ProcessAlive(42)
	presence.ProcessAlive(42)
	presence.SessionAlive(session.Session{ParentPID: 42, State: session.StateIdle})
	if got := presence.Captures(); got != 1 {
		t.Fatalf("captured %d OS snapshots within one cache window, want 1", got)
	}
	clock = clock.Add(11 * time.Second)
	presence.ProcessAlive(42)
	if got := presence.Captures(); got != 2 {
		t.Fatalf("captured %d snapshots after expiry, want 2", got)
	}
}

func TestPresenceSessionAliveUsesStrongestEvidenceFirst(t *testing.T) {
	clock := time.Unix(2000, 0)
	procs := map[int]focusProcess{
		500: {parent: 400, name: "claude.exe"},
		600: {parent: 400, name: "opencode.exe"},
	}
	windows := []focusWindow{terminalWindow(11, 600, "OC | 项目交接与后续计划")}
	presence := fakePresence(time.Minute, &clock, procs, windows)

	cases := []struct {
		name  string
		value session.Session
		want  bool
	}{
		{"live parent pid", session.Session{Source: "claude-ipc", ParentPID: 500, State: session.StateWorking}, true},
		{"trusted real pid", session.Session{Source: "opencode", PID: 600, State: session.StateWorking}, true},
		{"matching tab title", session.Session{Source: "opencode", Title: "项目交接与后续计划", State: session.StateIdle}, true},
		{"cli keyword alone is not identity", session.Session{Source: "antigravity-statusline", Terminal: "antigravity", State: session.StateIdle}, false},
		{"closed is never alive", session.Session{Source: "opencode", ParentPID: 500, State: session.StateClosed}, false},
		{"no evidence", session.Session{Source: "generic", PID: 741006029, State: session.StateIdle}, false},
	}
	for _, test := range cases {
		if got := presence.SessionAlive(test.value); got != test.want {
			t.Errorf("%s: SessionAlive = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestPresenceDoesNotTrustSyntheticPID(t *testing.T) {
	clock := time.Unix(3000, 0)
	// A synthetic stableProcessId can accidentally collide with a real PID.
	// It must not be treated as evidence of a live host.
	procs := map[int]focusProcess{741006029: {parent: 1, name: "explorer.exe"}}
	presence := fakePresence(time.Minute, &clock, procs, nil)
	value := session.Session{
		Source: "generic", PID: 741006029, State: session.StateWorking,
		Title: "unrelated", Project: `C:\work\unrelated`,
	}
	if presence.SessionAlive(value) {
		t.Fatal("a synthetic PID collision kept an unobservable session alive")
	}
}

func TestWindowIdentifiesSessionNeedsAnIdentityMatch(t *testing.T) {
	target := FocusTarget{
		Title:    "项目交接与后续计划",
		Project:  `C:\work\项目\DSH-Addition`,
		Source:   "opencode",
		Terminal: "opencode",
	}
	if windowIdentifiesSession([]focusWindow{terminalWindow(1, 10, "Some other tab")}, target) {
		t.Fatal("a bare terminal window must not keep an unrelated session alive")
	}
	if !windowIdentifiesSession([]focusWindow{terminalWindow(1, 10, "OC | 项目交接与后续计划")}, target) {
		t.Fatal("a tab whose title carries the session title should match")
	}
	// Regression: a single "agy" tab must not keep every old conversation
	// alive. Those sessions have their own parentPid to stand on.
	if windowIdentifiesSession([]focusWindow{terminalWindow(1, 10, "命令提示符 - agy")},
		FocusTarget{Source: "antigravity-statusline", Terminal: "antigravity"}) {
		t.Fatal("a bare CLI keyword kept historical conversations alive")
	}
	// But a tab that names the session's project is real evidence.
	if !windowIdentifiesSession([]focusWindow{terminalWindow(1, 10, "agy · ai-terminal-pet")},
		FocusTarget{Project: `C:\work\项目\ai-terminal-pet`}) {
		t.Fatal("a project name in the tab title should identify the session")
	}
	// A non-terminal window with a matching title is not evidence.
	browser := focusWindow{handle: 2, pid: 20, class: "MozillaWindowClass", title: "OC | 项目交接与后续计划", exe: "firefox.exe"}
	if windowIdentifiesSession([]focusWindow{browser}, target) {
		t.Fatal("a browser window must not count as a terminal")
	}
	if windowIdentifiesSession([]focusWindow{terminalWindow(1, 10, "OC | anything")}, FocusTarget{}) {
		t.Fatal("no identity fields means no match")
	}
}

// TestPresenceLiveSessions is an opt-in probe against the real runtime dir.
// It never changes the desktop - it only reports which on-disk sessions the
// new existence model considers live.
//
//	PET_PRESENCE_PROBE=1 go test ./internal/platform -run TestPresenceLiveSessions -v
func TestPresenceLiveSessions(t *testing.T) {
	if os.Getenv("PET_PRESENCE_PROBE") == "" {
		t.Skip("set PET_PRESENCE_PROBE=1 to probe live runtime snapshots")
	}
	directory := filepath.Join(os.Getenv("LOCALAPPDATA"), "AI-Terminal-Pet", "runtime")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Skipf("runtime dir unavailable: %v", err)
	}
	presence := NewPresence(0)
	seen := 0
	for _, item := range entries {
		if item.IsDir() || !strings.HasSuffix(item.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, item.Name()))
		if err != nil {
			continue
		}
		value, err := session.ParsePayload(data, time.Now())
		if err != nil {
			continue
		}
		seen++
		t.Logf("%-58s source=%-24s parentPid=%-8d alive=%v",
			item.Name(), value.Source, value.ParentPID, presence.SessionAlive(value))
	}
	if got := presence.Captures(); got > 1 {
		t.Fatalf("live probe took %d OS snapshots, want 1 (cache failed)", got)
	}
	t.Logf("checked %d valid snapshots with %d OS capture(s)", seen, presence.Captures())
}

// An Antigravity statusline writes one snapshot per conversation and gives
// each a real parentPid. Historical conversations must disappear when their
// CLI process exits, even though another "agy" tab is still open.
func TestPresenceDropsHistoricalAntigravityConversation(t *testing.T) {
	clock := time.Unix(4000, 0)
	processes := map[int]focusProcess{
		11492: {parent: 1, name: "WindowsTerminal.exe"},
		17564: {parent: 11492, name: "antigravity.exe"},
	}
	windows := []focusWindow{terminalWindow(1, 11492, "命令提示符 - agy")}
	presence := fakePresence(time.Minute, &clock, processes, windows)

	historical := session.Session{
		Source: "antigravity-statusline", SessionID: "8a06fc0f-old",
		ParentPID: 999999, Title: "", Project: `C:/Users/pet-test`,
		PID: 760369550, State: session.StateIdle,
	}
	if presence.SessionAlive(historical) {
		t.Fatal("a historical conversation was kept alive by a bare 'agy' tab")
	}
	current := historical
	current.SessionID = "8a06fc0f-current"
	current.ParentPID = 17564
	if !presence.SessionAlive(current) {
		t.Fatal("the live conversation was dropped")
	}
}

// A host process can live for days. Its liveness alone must never keep the
// conversations it once hosted on the panel forever: that was the ghost-pile
// bug. Only a tab that actually names the session is timeless evidence.
func TestPresenceReclaimsOrphansUnderALongLivedHost(t *testing.T) {
	clock := time.Date(2026, 11, 5, 16, 0, 0, 0, time.UTC)
	procs := map[int]focusProcess{
		2832: {parent: 2896, name: "opencode.exe"},
		5228: {parent: 2832, name: "opencode.exe"},
	}
	base := session.Session{
		SchemaVersion: "ai-terminal-pet.v1",
		Source:        "claude-ipc", SessionID: "old-conversation",
		ParentPID: 2832, PID: 1701354615,
		Project: `C:/work/项目/oh_my_3dlab`,
		Title:   "opencode", State: session.StateWorking,
	}

	// No tab identifies this conversation; the host is still up.
	presence := fakePresence(time.Minute, &clock, procs, nil)
	stale := base
	stale.UpdatedAt = clock.Add(-2 * time.Hour)
	if presence.SessionAlive(stale) {
		t.Fatal("an old, unidentified conversation outlived its host's evidence")
	}
	fresh := base
	fresh.UpdatedAt = clock.Add(-5 * time.Minute)
	if !presence.SessionAlive(fresh) {
		t.Fatal("a recently reported conversation was dropped while its host lives")
	}

	// A terminal tab naming the project is real identity: silence is fine
	// forever, which preserves the v0.7.0 no-heartbeat guarantee.
	withTab := fakePresence(time.Minute, &clock, procs,
		[]focusWindow{terminalWindow(1, 11492, "OC | oh_my_3dlab")})
	ancient := base
	ancient.UpdatedAt = clock.Add(-72 * time.Hour)
	if !withTab.SessionAlive(ancient) {
		t.Fatal("a session whose tab is still open must never be evicted for silence")
	}
}

// For reporters with a real OS pid, that pid is the conversation's owner.
// Its parent terminal being open is not a reason to keep the session.
func TestPresenceTrustedPIDDeathEndsSession(t *testing.T) {
	clock := time.Date(2026, 11, 5, 16, 0, 0, 0, time.UTC)
	procs := map[int]focusProcess{2832: {parent: 2896, name: "opencode.exe"}}
	presence := fakePresence(time.Minute, &clock, procs, nil)
	value := session.Session{
		SchemaVersion: "ai-terminal-pet.v1", Source: "opencode",
		SessionID: "ses_dead", PID: 5228, ParentPID: 2832,
		Project: `C:/work/项目/DSH-Addition`,
		Title:   "opencode", State: session.StateWorking,
		UpdatedAt: clock,
	}
	if presence.SessionAlive(value) {
		t.Fatal("a dead reporter process kept its conversation alive via parentPid")
	}
	procs[5228] = focusProcess{parent: 2832, name: "opencode.exe"}
	presence = fakePresence(time.Minute, &clock, procs, nil)
	if !presence.SessionAlive(value) {
		t.Fatal("the live reporter process should keep its conversation")
	}
}
