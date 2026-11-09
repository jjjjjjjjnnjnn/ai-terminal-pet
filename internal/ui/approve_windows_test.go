package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func waitingSession(source, sessionID string, pid int, updatedAt time.Time) session.Session {
	return session.Session{
		Source:    source,
		SessionID: sessionID,
		PID:       pid,
		State:     session.StateWaitingPermission,
		UpdatedAt: updatedAt,
	}
}

func TestApprovalTrackerStartsAndExpires(t *testing.T) {
	tracker := newApprovalTracker()
	now := time.Now()
	items := []session.Session{waitingSession("claude-hook", "s1", 11, now)}
	tracker.update(items, now, true, 5*time.Second)
	if got := tracker.maxRemainingSeconds(now); got != 5 {
		t.Fatalf("fresh countdown must read 5s, got %d", got)
	}
	// Six seconds later the request is still waiting (the hook is about to
	// approve): it must read as expired, but stay tracked until the
	// session actually leaves waiting.
	later := now.Add(6 * time.Second)
	if got := tracker.maxRemainingSeconds(later); got != 0 {
		t.Fatalf("expired countdown must read 0s, got %d", got)
	}
	// The session leaves waiting -> the countdown must be gone, so no
	// timer can fire for a request that no longer exists.
	tracker.update([]session.Session{{Source: "claude-hook", SessionID: "s1", PID: 11, State: session.StateWorking}}, later, true, 5*time.Second)
	if got := tracker.maxRemainingSeconds(later); got != -1 {
		t.Fatalf("resolved session must drop its countdown, got %d", got)
	}
}

func TestApprovalTrackerRestartsOnNewTurn(t *testing.T) {
	tracker := newApprovalTracker()
	start := time.Now()
	first := waitingSession("claude-hook", "s1", 11, start)
	tracker.update([]session.Session{first}, start, true, 5*time.Second)
	mid := start.Add(4 * time.Second)
	// Same UpdatedAt: same permission turn, countdown keeps running down.
	tracker.update([]session.Session{first}, mid, true, 5*time.Second)
	if got := tracker.maxRemainingSeconds(mid); got != 1 {
		t.Fatalf("same turn must keep counting down, got %d", got)
	}
	// Newer UpdatedAt: a new permission turn, the clock restarts.
	second := waitingSession("claude-hook", "s1", 11, mid)
	tracker.update([]session.Session{second}, mid, true, 5*time.Second)
	if got := tracker.maxRemainingSeconds(mid); got != 5 {
		t.Fatalf("new turn must restart the countdown, got %d", got)
	}
}

func TestApprovalTrackerIsolatesSessions(t *testing.T) {
	tracker := newApprovalTracker()
	now := time.Now()
	items := []session.Session{
		waitingSession("claude-hook", "shared", 11, now),
		waitingSession("opencode", "shared", 22, now.Add(-4*time.Second)),
	}
	tracker.update(items, now, true, 5*time.Second)
	views := tracker.views(now)
	// Same session id, different source/pid: two independent countdowns.
	if len(views) != 2 {
		t.Fatalf("two sessions must have two countdowns, got %d", len(views))
	}
	keyA := approvalSessionKey("claude-hook", "shared", 11)
	keyB := approvalSessionKey("opencode", "shared", 22)
	if views[keyA].RemainingSec != 5 || views[keyB].RemainingSec != 5 {
		t.Fatalf("both fresh countdowns must read 5s, got %+v", views)
	}
	// One resolves, the other keeps counting.
	tracker.update(items[:1], now, true, 5*time.Second)
	if got := tracker.maxRemainingSeconds(now); got != 5 {
		t.Fatalf("surviving countdown must be untouched, got %d", got)
	}
	if _, ok := tracker.views(now)[keyB]; ok {
		t.Fatal("resolved countdown must be gone")
	}
}

func TestApprovalTrackerDisabledClears(t *testing.T) {
	tracker := newApprovalTracker()
	now := time.Now()
	tracker.update([]session.Session{waitingSession("claude-hook", "s1", 11, now)}, now, true, 5*time.Second)
	tracker.update([]session.Session{waitingSession("claude-hook", "s1", 11, now)}, now, false, 5*time.Second)
	if got := tracker.maxRemainingSeconds(now); got != -1 {
		t.Fatalf("disabling must clear all countdowns, got %d", got)
	}
}

func TestApprovalUrgentKey(t *testing.T) {
	tracker := newApprovalTracker()
	now := time.Now()
	if got := tracker.urgentKey(nil); got != "" {
		t.Fatalf("no sessions must yield no key, got %q", got)
	}
	items := []session.Session{
		{Source: "opencode", SessionID: "busy", PID: 7, State: session.StateWorking},
		waitingSession("claude-hook", "s1", 11, now),
	}
	tracker.update(items, now, true, 5*time.Second)
	want := approvalSessionKey("claude-hook", "s1", 11)
	if got := tracker.urgentKey(items); got != want {
		t.Fatalf("urgent key = %q, want %q", got, want)
	}
}

func TestSanitizeApprovalStem(t *testing.T) {
	if got := sanitizeApprovalStem("abc-123_X"); got != "abc-123_X" {
		t.Fatalf("clean stem must survive, got %q", got)
	}
	if got := sanitizeApprovalStem(`C:\work\proj:1`); got != "C__work_proj_1" {
		t.Fatalf("path chars must be flattened, got %q", got)
	}
	if got := sanitizeApprovalStem(""); got != "session" {
		t.Fatalf("empty stem must fall back, got %q", got)
	}
	long := string(make([]byte, 200))
	for i := range long {
		long = long[:i] + "a" + long[i+1:]
	}
	if got := sanitizeApprovalStem(long); len(got) != 64 {
		t.Fatalf("long stem must truncate to 64, got %d", len(got))
	}
}

// TestWriteApprovalDenyMatchesByIdentity pins the veto routing: a deny for
// one session must veto that session's hook files and leave a concurrent
// session's files alone.
func TestWriteApprovalDenyMatchesByIdentity(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)
	if err := os.MkdirAll(filepath.Join(local, "AI-Terminal-Pet", "approvals"), 0700); err != nil {
		t.Fatal(err)
	}
	writePending := func(name, source, sessionID string, pid int) {
		body := `{"source":` + quote(source) + `,"sessionId":` + quote(sessionID) + `,"pid":` + itoa(pid) + `,"tool":"Bash","createdAt":"t"}`
		if err := os.WriteFile(filepath.Join(local, "AI-Terminal-Pet", "approvals", name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writePending("pending-a.json", "claude-hook", "s1", 11)
	writePending("pending-b.json", "claude-hook", "s2", 22)

	target := waitingSession("claude-hook", "s1", 11, time.Now())
	if !writeApprovalDeny(target) {
		t.Fatal("deny must write a veto file")
	}
	denyA := filepath.Join(local, "AI-Terminal-Pet", "approvals", "pending-a.json.deny")
	denyB := filepath.Join(local, "AI-Terminal-Pet", "approvals", "pending-b.json.deny")
	if _, err := os.Stat(denyA); err != nil {
		t.Fatalf("matching request must be vetoed: %v", err)
	}
	if _, err := os.Stat(denyB); !os.IsNotExist(err) {
		t.Fatal("unrelated request must not be vetoed")
	}
}

// TestSweepApprovalsOnlyTouchesHandshakeFiles pins the reaper: stale
// handshake files go, everything else (including user files that happen
// to sit in the directory) stays.
func TestSweepApprovalsOnlyTouchesHandshakeFiles(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)
	directory := filepath.Join(local, "AI-Terminal-Pet", "approvals")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(directory, "pending-old.json")
	fresh := filepath.Join(directory, "pending-new.json")
	innocent := filepath.Join(directory, "notes.txt")
	for _, path := range []string{stale, fresh, innocent} {
		if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}
	if removed := sweepApprovals(time.Hour); removed != 1 {
		t.Fatalf("exactly the stale handshake must go, removed %d", removed)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("stale pending must be gone")
	}
	for _, path := range []string{fresh, innocent} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s must survive: %v", path, err)
		}
	}
}

func quote(value string) string {
	return "\"" + value + "\""
}
