package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestSnapshotScanAndCleanup(t *testing.T) {
	now := time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "fresh.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking))
	writeSnapshot(t, directory, "stale.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Minute), session.StateIdle))
	writeSnapshot(t, directory, "dead.json", snapshotJSON(now, 999999, now.Add(-time.Second), session.StateIdle))
	writeSnapshot(t, directory, "closed.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateClosed))
	if err := os.WriteFile(filepath.Join(directory, "invalid.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	scanner := SnapshotScanner{Directory: directory, TTL: 15 * time.Second, Now: func() time.Time { return now }, Alive: func(pid int) bool { return pid == os.Getpid() }}
	items := scanner.Scan()
	if len(items) != 2 {
		t.Fatalf("unexpected snapshots: %#v", items)
	}
	states := map[session.State]bool{}
	for _, item := range items {
		states[item.State] = true
	}
	if !states[session.StateWorking] || !states[session.StateClosed] {
		t.Fatalf("closed marker was not preserved for store cleanup: %#v", items)
	}
	for _, name := range []string{"stale.json", "dead.json", "closed.json", "invalid.json"} {
		if _, err := os.Stat(filepath.Join(directory, name)); !errorsIsNotExist(err) {
			t.Fatalf("%s was not removed: %v", name, err)
		}
	}
}

func TestSnapshotScannerRequiresDirectory(t *testing.T) {
	if err := (SnapshotScanner{}).Err(); err == nil {
		t.Fatal("expected directory error")
	}
}

func TestSnapshotKeepsHookSourceWithDeadPID(t *testing.T) {
	now := time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	payload := fmt.Sprintf(`{"schemaVersion":"ai-terminal-pet.v1","source":"claude-hook","sessionId":"hook","pid":999999,"state":"working","startedAt":%q,"updatedAt":%q}`,
		now.Add(-time.Hour).Format(time.RFC3339Nano), now.Add(-time.Second).Format(time.RFC3339Nano))
	writeSnapshot(t, directory, "hook.json", payload)
	scanner := SnapshotScanner{Directory: directory, TTL: 15 * time.Second, Now: func() time.Time { return now }, Alive: func(int) bool { return false }}
	items := scanner.Scan()
	if len(items) != 1 || items[0].Source != "claude-hook" {
		t.Fatalf("hook snapshot was dropped on dead PID: %#v", items)
	}
	if _, err := os.Stat(filepath.Join(directory, "hook.json")); err != nil {
		t.Fatalf("hook snapshot was removed: %v", err)
	}
}

func writeSnapshot(t *testing.T, directory, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func snapshotJSON(now time.Time, pid int, updated time.Time, state session.State) string {
	// "opencode" on purpose: it is a process-backed source, which is what the
	// liveness assertions around this helper are about.
	return fmt.Sprintf(`{"schemaVersion":"ai-terminal-pet.v1","source":"opencode","sessionId":"session","pid":%d,"state":%q,"startedAt":%q,"updatedAt":%q}`, pid, state, now.Add(-time.Hour).Format(time.RFC3339Nano), updated.Format(time.RFC3339Nano))
}

// TestSnapshotKeepsSyntheticPIDSource is the regression guard for a bug that
// made generic.js useless: a reporter that keys its session on a stable hash
// rather than a real process id was treated as process-backed, so the liveness
// probe always failed and the snapshot was deleted the moment it was written.
// The reporter reported "ok" and nothing ever appeared in the overlay.
func TestSnapshotKeepsSyntheticPIDSource(t *testing.T) {
	now := time.Date(2026, 9, 26, 7, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	payload := fmt.Sprintf(`{"schemaVersion":"ai-terminal-pet.v1","source":"generic","sessionId":"generic","pid":741006029,"state":"working","startedAt":%q,"updatedAt":%q}`,
		now.Add(-time.Hour).Format(time.RFC3339Nano), now.Add(-time.Second).Format(time.RFC3339Nano))
	writeSnapshot(t, directory, "generic.json", payload)

	// Nothing is alive as far as this scanner is concerned.
	scanner := SnapshotScanner{Directory: directory, TTL: 15 * time.Second, Now: func() time.Time { return now }, Alive: func(int) bool { return false }}
	items := scanner.Scan()
	if len(items) != 1 || items[0].Source != "generic" {
		t.Fatalf("synthetic-PID snapshot was dropped: %#v", items)
	}
	if _, err := os.Stat(filepath.Join(directory, "generic.json")); err != nil {
		t.Fatalf("synthetic-PID snapshot was removed: %v", err)
	}
}

func TestSnapshotScannerRejectsOversizedFileBeforeReading(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "oversized.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(int64(session.MaxPayloadBytes) + 1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	items := (SnapshotScanner{Directory: directory, Now: time.Now}).Scan()
	if len(items) != 0 {
		t.Fatalf("oversized snapshot was accepted: %#v", items)
	}
	if _, err := os.Stat(path); !errorsIsNotExist(err) {
		t.Fatalf("oversized snapshot was not removed: %v", err)
	}
}

func TestSnapshotScannerCapsFiles(t *testing.T) {
	now := time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	for index := 0; index < 5; index++ {
		writeSnapshot(t, directory, fmt.Sprintf("snapshot-%d.json", index), snapshotJSON(now, os.Getpid(), now, session.StateWorking))
	}
	items := (SnapshotScanner{
		Directory: directory,
		Now:       func() time.Time { return now },
		Alive:     func(int) bool { return true },
		MaxFiles:  2,
	}).Scan()
	if len(items) != 2 {
		t.Fatalf("snapshot file cap was not enforced: %#v", items)
	}
	remaining, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) > 2 {
		t.Fatalf("snapshot files were not evicted: %d", len(remaining))
	}
}

func TestSnapshotRemovalKeepsReplacement(t *testing.T) {
	now := time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	path := filepath.Join(directory, "replacement.json")
	writeSnapshot(t, directory, "replacement.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Minute), session.StateIdle))
	oldInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	temporary := filepath.Join(directory, "replacement.tmp")
	writeSnapshot(t, directory, "replacement.tmp", snapshotJSON(now, os.Getpid(), now, session.StateWorking))
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temporary, path); err != nil {
		t.Fatal(err)
	}
	removeSnapshotIfUnchanged(path, oldInfo)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("replacement was removed: %v", err)
	}
}

func TestSnapshotPrunesStaleTmpFiles(t *testing.T) {
	now := time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "stale.tmp", snapshotJSON(now, os.Getpid(), now.Add(-time.Minute), session.StateIdle))
	writeSnapshot(t, directory, "fresh.tmp", snapshotJSON(now, os.Getpid(), now, session.StateIdle))
	if err := os.Chtimes(filepath.Join(directory, "stale.tmp"), now.Add(-time.Minute), now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(directory, "fresh.tmp"), now, now); err != nil {
		t.Fatal(err)
	}
	scanner := SnapshotScanner{Directory: directory, TTL: 15 * time.Second, Now: func() time.Time { return now }, Alive: func(int) bool { return true }}
	scanner.Scan()
	if _, err := os.Stat(filepath.Join(directory, "stale.tmp")); !errorsIsNotExist(err) {
		t.Fatalf("stale tmp file was not pruned: %v", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "fresh.tmp")); err != nil {
		t.Fatalf("fresh tmp file was pruned: %v", err)
	}
}

func errorsIsNotExist(err error) bool {
	return err != nil && os.IsNotExist(err)
}
