package store

import (
	"sync"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestStoreCleanup(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := New()
	stale := session.Session{Source: "a", SessionID: "stale", PID: 1, State: session.StateIdle, StartedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)}
	live := session.Session{Source: "b", SessionID: "live", PID: 2, State: session.StateWorking, StartedAt: now.Add(-time.Hour), UpdatedAt: now}
	values.UpsertSnapshot(stale, now)
	values.UpsertPipe(live, now)
	if values.UpsertSnapshot(stale, now.Add(10*time.Second)) {
		t.Fatal("unchanged snapshot must not count as a change")
	}
	values.Cleanup(now.Add(20*time.Second), func(pid int) bool { return pid == 2 }, 15*time.Second)
	items := values.List()
	if len(items) != 1 || items[0].SessionID != "live" {
		t.Fatalf("unexpected sessions: %#v", items)
	}
	values.Cleanup(now.Add(21*time.Second), func(pid int) bool { return false }, 15*time.Second)
	if values.HasAvailable() {
		t.Fatal("dead sessions were not removed")
	}
}

func TestStoreClosedAndOrder(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := New()
	idle := session.Session{Source: "a", SessionID: "idle", PID: 10, State: session.StateIdle, StartedAt: now, UpdatedAt: now}
	wait := session.Session{Source: "b", SessionID: "wait", PID: 11, State: session.StateWaitingPermission, StartedAt: now, UpdatedAt: now}
	values.UpsertPipe(idle, now)
	values.UpsertPipe(wait, now.Add(time.Second))
	items := values.List()
	if len(items) != 2 || items[0].State != session.StateWaitingPermission {
		t.Fatalf("unexpected order: %#v", items)
	}
	wait.State = session.StateClosed
	wait.UpdatedAt = now.Add(2 * time.Second)
	values.UpsertPipe(wait, now.Add(2*time.Second))
	items = values.List()
	if len(items) != 1 || items[0].State != session.StateIdle {
		t.Fatalf("closed session remained: %#v", items)
	}
}

func TestStoreKeepsPipeSessionPastSnapshotTTL(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := New()
	value := session.Session{Source: "a", SessionID: "pipe", PID: 20, State: session.StateThinking, StartedAt: now, UpdatedAt: now}
	values.UpsertSnapshot(value, now)
	values.UpsertPipe(value, now.Add(10*time.Second))
	values.Cleanup(now.Add(30*time.Second), func(int) bool { return true }, 15*time.Second)
	if !values.HasAvailable() {
		t.Fatal("pipe event should keep the session available")
	}
}

func TestStoreIdentity(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := New()
	left := session.Session{Source: "a", SessionID: "same", State: session.StateIdle, StartedAt: now, UpdatedAt: now}
	right := left
	right.PID = 31
	values.UpsertPipe(left, now)
	values.UpsertPipe(right, now)
	if len(values.List()) != 2 {
		t.Fatal("PID is required in session identity")
	}
}

func TestStoreStaleCloseUsesTimestampTombstone(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := New()
	active := session.Session{Source: "a", SessionID: "ordered", PID: 40, State: session.StateWorking, StartedAt: now, UpdatedAt: now.Add(2 * time.Second)}
	values.UpsertPipe(active, now.Add(2*time.Second))

	staleClose := active
	staleClose.State = session.StateClosed
	staleClose.UpdatedAt = now.Add(time.Second)
	if values.UpsertPipe(staleClose, now.Add(3*time.Second)) {
		t.Fatal("stale close must not replace a newer open event")
	}
	items := values.List()
	if len(items) != 1 || items[0].State != session.StateWorking {
		t.Fatalf("newer session was removed: %#v", items)
	}

	closeValue := active
	closeValue.State = session.StateClosed
	closeValue.UpdatedAt = now.Add(4 * time.Second)
	if !values.UpsertPipe(closeValue, now.Add(4*time.Second)) {
		t.Fatal("newer close event was not applied")
	}
	if values.HasAvailable() {
		t.Fatal("closed session remained available")
	}

	staleOpen := active
	staleOpen.UpdatedAt = now.Add(3 * time.Second)
	if values.UpsertPipe(staleOpen, now.Add(5*time.Second)) {
		t.Fatal("stale open event resurrected a closed session")
	}
	newerOpen := active
	newerOpen.UpdatedAt = now.Add(5 * time.Second)
	if !values.UpsertPipe(newerOpen, now.Add(5*time.Second)) {
		t.Fatal("newer open event did not replace the tombstone")
	}
	if !values.HasAvailable() {
		t.Fatal("newer open event was not restored")
	}
}

func TestStorePipeTTLAndSessionEviction(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	pipeTTL := 10 * time.Second
	values := NewWithLimits(4, pipeTTL)
	value := session.Session{Source: "a", SessionID: "pipe", PID: 50, State: session.StateIdle, StartedAt: now, UpdatedAt: now}
	values.UpsertPipe(value, now)
	values.UpsertPipe(value, now.Add(4*time.Second))
	values.Cleanup(now.Add(11*time.Second), func(int) bool { return true }, time.Hour)
	if !values.HasAvailable() {
		t.Fatal("a recent pipe heartbeat was evicted")
	}
	values.Cleanup(now.Add(15*time.Second), func(int) bool { return true }, time.Hour)
	if values.HasAvailable() {
		t.Fatal("stale pipe session was not evicted")
	}
	newer := value
	newer.UpdatedAt = now.Add(20 * time.Second)
	if !values.UpsertPipe(newer, now.Add(20*time.Second)) {
		t.Fatal("newer pipe event did not replace the pipe tombstone")
	}

	capped := NewWithLimits(2, pipeTTL)
	first := session.Session{Source: "a", SessionID: "one", PID: 51, State: session.StateIdle, StartedAt: now, UpdatedAt: now}
	second := first
	second.SessionID = "two"
	second.PID = 52
	third := first
	third.SessionID = "three"
	third.PID = 53
	capped.UpsertPipe(first, now)
	capped.UpsertPipe(second, now.Add(time.Second))
	capped.UpsertPipe(third, now.Add(2*time.Second))
	if items := capped.List(); len(items) != 2 {
		t.Fatalf("session cap was not enforced: %#v", items)
	}
	if capped.UpsertPipe(first, now.Add(3*time.Second)) {
		t.Fatal("evicted session was not protected by a tombstone")
	}
}

func TestStoreConcurrentEventsAreRaceSafe(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := NewWithLimits(32, time.Hour)
	ids := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	var wait sync.WaitGroup
	for worker, id := range ids {
		wait.Add(1)
		go func(worker int, id string) {
			defer wait.Done()
			for iteration := 0; iteration < 100; iteration++ {
				value := session.Session{
					Source:    "concurrent",
					SessionID: id,
					PID:       worker + 1,
					State:     session.StateIdle,
					StartedAt: now,
					UpdatedAt: now.Add(time.Duration(iteration) * time.Millisecond),
				}
				values.UpsertPipe(value, now.Add(time.Duration(iteration)*time.Millisecond))
				values.UpsertSnapshot(value, now.Add(time.Duration(iteration)*time.Millisecond))
				values.Cleanup(now.Add(time.Second), func(int) bool { return true }, time.Hour)
				_ = values.List()
				_ = values.Revision()
			}
		}(worker, id)
	}
	wait.Wait()
	if items := values.List(); len(items) > 32 {
		t.Fatalf("session cap was exceeded under concurrency: %d", len(items))
	}
}

func TestStoreHookSourceIgnoresProcessLiveness(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	values := NewWithLimits(32, time.Hour)
	hook := session.Session{Source: "claude-hook", SessionID: "hook-1", PID: 999999, State: session.StateWorking, StartedAt: now, UpdatedAt: now}
	pipe := session.Session{Source: "opencode", SessionID: "pipe-1", PID: 999999, State: session.StateWorking, StartedAt: now, UpdatedAt: now}
	values.UpsertPipe(hook, now)
	values.UpsertPipe(pipe, now)
	values.Cleanup(now.Add(time.Second), func(int) bool { return false }, time.Hour)
	items := values.List()
	if len(items) != 1 || items[0].Source != "claude-hook" {
		t.Fatalf("hook session must survive dead-PID cleanup: %#v", items)
	}
}
