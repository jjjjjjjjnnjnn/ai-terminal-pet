package store

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func reconcileSession() session.Session {
	now := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	return session.Session{
		Source: "claude-ipc", SessionID: "quiet", PID: 1234,
		State: session.StateWorking, StartedAt: now, UpdatedAt: now,
	}
}

// The core stability invariant: an event-driven reporter can stay quiet for
// as long as its terminal is open. Silence must neither evict the session nor
// dirty the frame (which would cause a repaint every scan).
func TestReconcileKeepsLiveQuietSessionWithoutRepaint(t *testing.T) {
	value := reconcileSession()
	values := New()
	values.UpsertPipe(value, value.UpdatedAt)
	revision := values.Revision()
	alive := func(session.Session) bool { return true }

	current := value.UpdatedAt
	for minute := 0; minute < 60; minute++ {
		current = current.Add(time.Minute)
		if values.Reconcile(current, alive, DefaultAbsenceGrace) {
			t.Fatalf("minute %d: an alive, unchanged session was reported as a change", minute)
		}
		if values.Revision() != revision {
			t.Fatalf("minute %d: revision changed without a visible change", minute)
		}
		if !values.HasAvailable() {
			t.Fatalf("minute %d: a live quiet session disappeared", minute)
		}
		items := values.List()
		if len(items) != 1 || !items[0].UpdatedAt.Equal(value.UpdatedAt) {
			t.Fatalf("minute %d: UpdatedAt was rewritten (would restart relative time and repaint)", minute)
		}
	}
}

// A single missed observation must never evict anything, and a brief
// reappearance must reset the absence streak rather than inherit it.
func TestReconcileRequiresContinuousAbsence(t *testing.T) {
	value := reconcileSession()
	values := New()
	values.UpsertPipe(value, value.UpdatedAt)

	alive := true
	check := func(session.Session) bool { return alive }
	now := value.UpdatedAt

	alive = false
	if values.Reconcile(now.Add(time.Second), check, DefaultAbsenceGrace) || !values.HasAvailable() {
		t.Fatal("first missed observation removed a session")
	}
	if values.Reconcile(now.Add(time.Minute), check, DefaultAbsenceGrace) || !values.HasAvailable() {
		t.Fatal("session removed before the grace elapsed")
	}

	// The host comes back for one scan, then disappears again. The old streak
	// must be forgotten.
	alive = true
	values.Reconcile(now.Add(90*time.Second), check, DefaultAbsenceGrace)
	alive = false
	if values.Reconcile(now.Add(3*time.Minute), check, DefaultAbsenceGrace) || !values.HasAvailable() {
		t.Fatal("a reappearance did not reset the absence streak")
	}
	if !values.Reconcile(now.Add(5*time.Minute), check, DefaultAbsenceGrace) {
		t.Fatal("a continuous absence past the grace was not removed")
	}
	if values.HasAvailable() {
		t.Fatal("absent session still available")
	}
}

// This is the self-heal regression. A silence-induced tombstone used to gate
// re-entry by UpdatedAt, so a quiet session reporting again with its unchanged
// timestamp could never come back - the panel flickered and stayed wrong.
func TestExpiredTombstoneAllowsUnchangedTimestampToReturn(t *testing.T) {
	value := reconcileSession()
	values := NewWithLimits(4, 10*time.Second)
	values.UpsertPipe(value, value.UpdatedAt)
	values.Cleanup(value.UpdatedAt.Add(15*time.Second), func(int) bool { return true }, time.Hour)
	if values.HasAvailable() {
		t.Fatal("setup: TTL eviction did not happen")
	}
	// Same payload, same timestamp: exactly what a de-duplicating reporter
	// sends when its state has not changed.
	if !values.UpsertPipe(value, value.UpdatedAt.Add(20*time.Second)) {
		t.Fatal("a silence tombstone blocked an unchanged timestamp from returning")
	}
	if !values.HasAvailable() {
		t.Fatal("session did not self-heal")
	}
}

func TestReconcileTombstoneAlsoAllowsReturn(t *testing.T) {
	value := reconcileSession()
	values := New()
	values.UpsertPipe(value, value.UpdatedAt)
	dead := func(session.Session) bool { return false }
	// Two observations: the first only starts the absence streak, the second
	// (past the grace) actually evicts. A single miss is never enough.
	values.Reconcile(value.UpdatedAt.Add(time.Second), dead, DefaultAbsenceGrace)
	if !values.Reconcile(value.UpdatedAt.Add(3*time.Minute), dead, DefaultAbsenceGrace) {
		t.Fatal("setup: absence eviction did not happen")
	}
	if values.HasAvailable() {
		t.Fatal("setup: absence eviction did not happen")
	}
	if !values.UpsertPipe(value, value.UpdatedAt.Add(4*time.Minute)) {
		t.Fatal("an absence tombstone blocked the returning session")
	}
}

// Explicit closes are different: a stale push from a session that really
// ended must NOT resurrect it.
func TestClosedTombstoneStillBlocksStalePush(t *testing.T) {
	value := reconcileSession()
	values := New()
	values.UpsertPipe(value, value.UpdatedAt)
	closed := value
	closed.State = session.StateClosed
	closed.UpdatedAt = value.UpdatedAt.Add(time.Minute)
	values.UpsertPipe(closed, closed.UpdatedAt)
	if values.UpsertPipe(value, closed.UpdatedAt.Add(time.Second)) {
		t.Fatal("an explicit close was undone by a stale open event")
	}
}
