package reporter

// The portable half of the delivery tests: what a reporter does when the overlay
// answers, and what the snapshot/acknowledgement ordering has to guarantee.
//
// Two pieces moved to files of their own:
//
//   - the retry classification, because the errors it classifies are the OS's
//     (a named pipe's ERROR_PIPE_BUSY on Windows, EACCES on a socket owned by
//     another user here): send_windows_test.go and send_unix_test.go.
//   - the denied-pipe probe, which needs winio to build a pipe with a DACL that
//     refuses this caller - a shape unix does not have.

import (
	"os"
	"testing"
	"time"
)

// Send must file the snapshot before it dials, so a report that arrives while the
// overlay is still starting is read from disk by the next scan instead of being
// lost. This is the ordering protocol.js documents and the one reason the
// snapshot write comes first.
func TestSendWritesTheSnapshotEvenWhenDeliveryFails(t *testing.T) {
	dataDirectory(t)
	now := time.Now().UTC()
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "snapshot-only"
	payload.PID = 22
	payload = payload.Finish(now)

	path := SnapshotPath(payload)
	if path == "" {
		t.Fatal("no snapshot path for this payload")
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("the snapshot exists before Send ran")
	}

	// A pipe path nobody is listening on: delivery fails, the snapshot stays.
	ok := Send(payload, SendOptions{
		Retry:          30 * time.Millisecond,
		ConnectTimeout: 10 * time.Millisecond,
		PipePath:       "definitely-not-listening",
	})
	if ok {
		t.Fatal("Send claimed success with no listener")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the snapshot was not filed: %v", err)
	}
}

// A payload the server would reject must not leave a snapshot behind: the scanner
// would keep re-reading a file it cannot parse and deleting it, once per scan.
//
// The payload is built past Finish's reach on purpose. Finish repairs everything
// out of range (an unknown state becomes idle, a bad pid is replaced), so the only
// way to reach the validate branch is a field Finish does not correct - the schema
// version, which is the one thing a reporter must never guess.
func TestSendSkipsTheSnapshotForAnInvalidPayload(t *testing.T) {
	dataDirectory(t)
	now := time.Now().UTC()
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "invalid"
	payload.PID = 23
	payload = payload.Finish(now)
	// Finish repairs everything out of range (an unknown state becomes idle, a
	// bad pid is replaced) and rewrites the schema version, so the only way to
	// reach Send's validate branch is to change a field *after* Finish - which is
	// exactly the shape of a future caller that forgets to go through it.
	payload.SchemaVersion = "ai-terminal-pet.v999"

	path := SnapshotPath(payload)
	ok := Send(payload, SendOptions{
		Retry:          30 * time.Millisecond,
		ConnectTimeout: 10 * time.Millisecond,
		PipePath:       "definitely-not-listening",
	})
	if ok {
		t.Fatal("Send claimed success for an invalid payload")
	}
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("an invalid payload left a snapshot the scanner will loop on")
	}
}
