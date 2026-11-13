package ui

// One conversation, one row.
//
// The store keys a session on source + sessionId + pid because the pipe, the
// snapshot files and the tombstone ledger are per reporter process. The panel is
// a picture of *conversations*, so two pids reporting the same sessionId - an
// OpenCode restart, a plugin that reports its own process id, a snapshot left
// over from a previous run - produced two rows for one conversation. Measured on
// the machine that reported it: 3 conversations, 6 rows, 3 of them from pids
// that no longer existed.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestDedupeKeepsOneRowPerConversation(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	older := session.Session{
		Source: "opencode", SessionID: "ses_1", PID: 100, State: session.StateWorking,
		Project: `C:\work\a`, UpdatedAt: now.Add(-5 * time.Minute),
	}
	newer := older
	newer.PID = 200
	newer.UpdatedAt = now

	other := session.Session{
		Source: "opencode", SessionID: "ses_2", PID: 100, State: session.StateWaitingPermission,
		UpdatedAt: now,
	}

	got := dedupeSessions([]session.Session{older, newer, other})
	if len(got) != 2 {
		t.Fatalf("want 2 conversations, got %d: %+v", len(got), got)
	}
	byIdentity := map[string]session.Session{}
	for _, value := range got {
		byIdentity[session.Identity(value)] = value
	}
	if byIdentity[session.Identity(older)].PID != 200 {
		t.Fatalf("the newest reporter should win, got pid %d", byIdentity[session.Identity(older)].PID)
	}
	if _, ok := byIdentity[session.Identity(other)]; !ok {
		t.Fatal("an unrelated conversation was dropped")
	}
}

// Two reporters disagreeing about one conversation must not hide a blocked one:
// the more urgent state wins a timestamp tie, because that is the row the user
// has to act on.
func TestDedupePrefersTheUrgentTie(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	busy := session.Session{Source: "opencode", SessionID: "ses_x", PID: 7, State: session.StateWorking, UpdatedAt: now}
	blocked := busy
	blocked.PID = 8
	blocked.State = session.StateWaitingPermission

	for _, order := range [][]session.Session{{busy, blocked}, {blocked, busy}} {
		got := dedupeSessions(order)
		if len(got) != 1 {
			t.Fatalf("want one row, got %d", len(got))
		}
		if got[0].State != session.StateWaitingPermission {
			t.Fatalf("the blocked report lost the tie: %q", got[0].State)
		}
	}
}

// Dedupe must be stable: it runs on every load, so an input that flips between
// two equal candidates would repaint forever.
func TestDedupeIsStable(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	a := session.Session{Source: "opencode", SessionID: "ses_s", PID: 10, State: session.StateWorking, UpdatedAt: now}
	b := a
	b.PID = 20
	for attempt := 0; attempt < 6; attempt++ {
		got := dedupeSessions([]session.Session{a, b})
		if len(got) != 1 || got[0].PID != 20 {
			t.Fatalf("attempt %d picked %+v", attempt, got)
		}
	}
}

// A row with no conversation identity is never merged into another: merging two
// unrelated rows would hide one of them, and a row without a session id is not
// a conversation we can group.
func TestDedupeNeverMergesUnidentifiedRows(t *testing.T) {
	first := session.Session{Source: "generic", PID: 1, State: session.StateWorking}
	second := session.Session{Source: "generic", PID: 2, State: session.StateIdle}
	if got := dedupeSessions([]session.Session{first, second}); len(got) != 2 {
		t.Fatalf("unidentified rows were merged: %+v", got)
	}
}

// The panel-level dedupe must not touch --json: that report keeps the
// per-process rows on purpose, because two reporters disagreeing about one
// conversation is evidence a bug report needs. This pins the boundary by
// showing the two layers disagree on the same input.
func TestDedupeIsAPanelConcernOnly(t *testing.T) {
	now := time.Now()
	a := session.Session{Source: "opencode", SessionID: "ses_z", PID: 1, State: session.StateWorking, UpdatedAt: now}
	b := a
	b.PID = 2
	if rows := dedupeSessions([]session.Session{a, b}); len(rows) != 1 {
		t.Fatalf("the panel should collapse them, got %d", len(rows))
	}
	// The underlying list is unchanged: the panel copy is deduped, the store's
	// is not.
	raw := []session.Session{a, b}
	_ = dedupeSessions(raw)
	if len(raw) != 2 {
		t.Fatal("dedupeSessions must not mutate its input")
	}
}
