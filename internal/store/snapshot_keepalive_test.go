package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestSnapshotKeepAliveRetainsQuietSession(t *testing.T) {
	now := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	// Ten minutes old: far beyond the 15s payload TTL, but the terminal is
	// still open and the reporter only writes on state changes.
	writeSnapshot(t, directory, "quiet.json", snapshotJSON(now, os.Getpid(), now.Add(-10*time.Minute), session.StateWorking))
	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return now },
		Alive:     func(int) bool { return false },
		KeepAlive: func(session.Session) bool { return true },
	}
	items := scanner.Scan()
	if len(items) != 1 || items[0].State != session.StateWorking {
		t.Fatalf("a quiet but live session was dropped: %#v", items)
	}
	if _, err := os.Stat(filepath.Join(directory, "quiet.json")); err != nil {
		t.Fatalf("the durable snapshot of a live session was deleted: %v", err)
	}
}

func TestSnapshotRemovesStaleSessionWhenHostIsGone(t *testing.T) {
	now := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "gone.json", snapshotJSON(now, os.Getpid(), now.Add(-10*time.Minute), session.StateWorking))
	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return now },
		Alive:     func(int) bool { return false },
		KeepAlive: func(session.Session) bool { return false },
	}
	if items := scanner.Scan(); len(items) != 0 {
		t.Fatalf("a dead session was kept: %#v", items)
	}
	if _, err := os.Stat(filepath.Join(directory, "gone.json")); !errorsIsNotExist(err) {
		t.Fatalf("dead session snapshot was not removed: %v", err)
	}
}

// A process-backed source whose PID probe fails is still kept when the
// presence layer can see its terminal window: one failed observation is not
// grounds for deleting a session.
func TestSnapshotKeepAliveRescuesFailedPIDProbe(t *testing.T) {
	now := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "opencode.json", snapshotJSON(now, 999999, now.Add(-time.Second), session.StateWorking))
	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return now },
		Alive:     func(int) bool { return false },
		KeepAlive: func(session.Session) bool { return true },
	}
	if items := scanner.Scan(); len(items) != 1 {
		t.Fatalf("a single failed PID probe deleted a live session: %#v", items)
	}
}

// ScanWithKeys is the disk-side half of "which conversations still exist".
// A closed snapshot is consumed in the same pass, so it must not be reported
// as an identity that is still present on disk.
func TestScanWithKeysReportsDurableIdentities(t *testing.T) {
	now := time.Date(2026, 11, 5, 16, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "live.json", fmt.Sprintf(
		`{"schemaVersion":"ai-terminal-pet.v1","source":"opencode","sessionId":"live","pid":%d,"state":"working","startedAt":%q,"updatedAt":%q}`,
		os.Getpid(), now.Add(-time.Hour).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)))
	writeSnapshot(t, directory, "done.json", fmt.Sprintf(
		`{"schemaVersion":"ai-terminal-pet.v1","source":"opencode","sessionId":"done","pid":%d,"state":"closed","startedAt":%q,"updatedAt":%q}`,
		os.Getpid(), now.Add(-time.Hour).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)))
	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return now },
		KeepAlive: func(session.Session) bool { return true },
	}
	items, keys := scanner.ScanWithKeys()
	if len(keys) != 1 {
		t.Fatalf("want exactly one durable identity, got %v", keys)
	}
	for _, item := range items {
		if item.State == session.StateClosed && keys[Key(item)] {
			t.Fatal("a consumed closed snapshot was reported as still present")
		}
	}
	if !keys[Key(session.Session{Source: "opencode", SessionID: "live", PID: os.Getpid()})] {
		t.Fatal("the live snapshot key is missing")
	}
	// Once the reporter removes its file, the identity disappears from the
	// observed set - that is the signal the overlay uses to drop the session.
	_ = os.Remove(filepath.Join(directory, "live.json"))
	_, keys = scanner.ScanWithKeys()
	if len(keys) != 0 {
		t.Fatalf("a removed snapshot is still reported as present: %v", keys)
	}
}
