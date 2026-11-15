package store

// The one bug this cache could introduce, and the reason the fallback to the
// Readdir FileInfo exists.
//
// Before the fallback, a cache HIT carried info == nil, and every removal branch
// calls removeSnapshotIfUnchanged(path, info), which returns immediately when
// expected == nil. The ordinary path re-read the file and passed the post-read
// info, so nothing showed up in the existing tests. The cached path silently
// stopped deleting anything: a tombstone stayed on disk (and was therefore
// returned again on every tick) and an expired snapshot leaked one file per
// session, forever.
//
// These two tests scan TWICE through the cache, because one pass is exactly the
// case that hides the bug.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestSnapshotCacheConsumesAClosedTombstoneExactlyOnce(t *testing.T) {
	now := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "closed.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateClosed))
	cache := NewSnapshotCache()
	scanner := cachedScanner(directory, now, cache)

	first := scanner.Scan()
	if len(first) != 1 || first[0].State != session.StateClosed {
		t.Fatalf("first pass did not return the tombstone: %#v", first)
	}
	if _, err := os.Stat(filepath.Join(directory, "closed.json")); !errorsIsNotExist(err) {
		t.Fatalf("the tombstone file was not removed on the first pass: %v", err)
	}
	second := scanner.Scan()
	if len(second) != 0 {
		t.Fatalf("the tombstone was returned a second time: %#v", second)
	}
}

func TestSnapshotCacheDeletesSnapshotsThatBecomeExpired(t *testing.T) {
	// The sequence matters and it is the only one that reaches the bug: the
	// snapshot has to SURVIVE one pass first (so it gets stored in the cache),
	// and only then go stale - without being rewritten. A tombstone or an
	// already-expired file is deleted on the pass that reads it and never
	// enters the cache at all, which is why the obvious version of this test
	// passed even with the fallback missing.
	start := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	clock := start
	directory := t.TempDir()
	writeSnapshot(t, directory, "slow.json", snapshotJSON(start, os.Getpid(), start.Add(-time.Second), session.StateWorking))
	cache := NewSnapshotCache()
	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return clock },
		Alive:     func(pid int) bool { return pid == os.Getpid() },
		Cache:     cache,
	}

	if items := scanner.Scan(); len(items) != 1 {
		t.Fatalf("the fresh pass returned %d sessions, want 1", len(items))
	}
	if cache.Len() != 1 {
		t.Fatalf("the surviving snapshot was not cached")
	}

	// Ten minutes later the payload is far past the TTL and the host still
	// answers Alive, so KeepAlive is nil and only the TTL branch can fire.
	clock = start.Add(10 * time.Minute)
	_ = scanner.Scan()
	if _, err := os.Stat(filepath.Join(directory, "slow.json")); !errorsIsNotExist(err) {
		t.Fatalf("a snapshot that went stale while cached was never deleted: %v", err)
	}
}

func TestSnapshotCacheDeletesSnapshotsWhoseHostDied(t *testing.T) {
	// Same shape, the other removal branch: the file is fresh and stays fresh,
	// but the process behind it dies. Nothing rewrites the file, so the only
	// thing that can notice is this scan - and on a cache hit it is the only
	// branch that would have to act with Readdir's FileInfo.
	start := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	alive := true
	directory := t.TempDir()
	writeSnapshot(t, directory, "host.json", snapshotJSON(start, os.Getpid(), start.Add(-time.Second), session.StateWorking))
	cache := NewSnapshotCache()
	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return start },
		Alive:     func(pid int) bool { return alive && pid == os.Getpid() },
		Cache:     cache,
	}

	if items := scanner.Scan(); len(items) != 1 {
		t.Fatalf("the first pass returned %d sessions, want 1", len(items))
	}
	alive = false
	_ = scanner.Scan()
	if _, err := os.Stat(filepath.Join(directory, "host.json")); !errorsIsNotExist(err) {
		t.Fatalf("a snapshot whose host died was never deleted: %v", err)
	}
}
