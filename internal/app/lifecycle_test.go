package app

import (
	"testing"

	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
)

// Removing a snapshot is how reporters say "this conversation is over".
// Before this, the in-memory pipe entry survived forever because the long-
// lived parent pid kept passing the liveness check: the ghost-pile bug.
func TestSnapshotAwareAliveDropsRemovedSnapshots(t *testing.T) {
	value := session.Session{
		SchemaVersion: "ai-terminal-pet.v1", Source: "opencode",
		SessionID: "ses_ghost", PID: 5228, State: session.StateWorking,
	}
	identity := store.Key(value)
	alwaysAlive := func(session.Session) bool { return true }

	seen := map[string]bool{identity: true}
	present := map[string]bool{identity: true}
	if !snapshotAwareAlive(present, seen, alwaysAlive)(value) {
		t.Fatal("a session with a live snapshot must stay alive")
	}

	delete(present, identity)
	if snapshotAwareAlive(present, seen, alwaysAlive)(value) {
		t.Fatal("a removed snapshot must end the session even when the OS host lives")
	}

	// A reporter that never wrote a snapshot is judged by OS liveness alone.
	if !snapshotAwareAlive(present, map[string]bool{}, alwaysAlive)(value) {
		t.Fatal("pipe-only sessions must not be penalised for having no file")
	}
}

// A durable snapshot must outrank weak OS evidence. Otherwise the overlay
// evicts a session that --json (a fresh store reading the same file) still
// shows, HasAvailable() goes false, and the window self-destructs ten minutes
// later even though the terminal is still there.
func TestSnapshotAwareAliveTrustsTheDisk(t *testing.T) {
	value := session.Session{
		SchemaVersion: "ai-terminal-pet.v1", Source: "opencode",
		SessionID: "ses_disk", PID: 5228, State: session.StateWorking,
	}
	identity := store.Key(value)
	present := map[string]bool{identity: true}
	seen := map[string]bool{identity: true}
	noOSEvidence := func(session.Session) bool { return false }

	if !snapshotAwareAlive(present, seen, noOSEvidence)(value) {
		t.Fatal("a session with a durable snapshot was evicted on weak OS evidence")
	}

	closed := value
	closed.State = session.StateClosed
	if snapshotAwareAlive(present, seen, func(session.Session) bool { return true })(closed) {
		t.Fatal("a closed session must never be kept alive")
	}

	delete(present, identity)
	if snapshotAwareAlive(present, seen, func(session.Session) bool { return true })(value) {
		t.Fatal("removing the snapshot must still end the session")
	}
}
