package store

// The cache is a performance device, so the properties pinned here are the ones
// that make it safe to have at all: it must never answer for a file that changed,
// never survive a removed file, and never turn a cached session into a second
// copy of the same disk state. Everything else is the scanner's existing
// behaviour and stays covered by snapshot_test.go.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func cachedScanner(directory string, now time.Time, cache *SnapshotCache) SnapshotScanner {
	return SnapshotScanner{
		Directory: directory,
		TTL:       15 * time.Second,
		Now:       func() time.Time { return now },
		Alive:     func(pid int) bool { return pid == os.Getpid() },
		Cache:     cache,
	}
}

func TestSnapshotCacheSkipsUnchangedFiles(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "one.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking))
	writeSnapshot(t, directory, "two.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateThinking))
	cache := NewSnapshotCache()
	scanner := cachedScanner(directory, now, cache)

	for pass := 0; pass < 3; pass++ {
		items := scanner.Scan()
		if len(items) != 2 {
			t.Fatalf("pass %d returned %d sessions, want 2", pass, len(items))
		}
	}
	if cache.Len() != 2 {
		t.Fatalf("cache holds %d entries, want 2 (one per live snapshot)", cache.Len())
	}
	// Three passes over two unchanged files: the first reads both, the next two
	// are served entirely from the cache. This is the property the optimisation
	// claims, asserted rather than assumed.
	if cache.Misses() != 2 {
		t.Fatalf("cache missed %d times over three passes, want 2 (one per file)", cache.Misses())
	}
	if cache.Hits() != 4 {
		t.Fatalf("cache hit %d times, want 4 (two files x two later passes)", cache.Hits())
	}
}

// A rewritten snapshot must be re-read. This is the one way the cache could lie,
// so it is checked by content rather than by counting reads.
func TestSnapshotCacheNoticesARewrittenFile(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	path := filepath.Join(directory, "live.json")
	writeSnapshot(t, directory, "live.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking))
	cache := NewSnapshotCache()
	scanner := cachedScanner(directory, now, cache)
	if items := scanner.Scan(); len(items) != 1 || items[0].State != session.StateWorking {
		t.Fatalf("first pass did not read the working snapshot: %#v", items)
	}

	// Rewrite through a temp file + rename, exactly what reporter.WriteSnapshot
	// does, so the size can stay the same while the content changes - the
	// metadata-only comparison must not be fooled by that.
	rewritten := snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateError)
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, []byte(rewritten), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temporary, path); err != nil {
		t.Fatal(err)
	}
	items := scanner.Scan()
	if len(items) != 1 || items[0].State != session.StateError {
		t.Fatalf("rewritten snapshot was served from cache: %#v", items)
	}
}

// A file that disappears must not keep answering from cache, and the entry for
// it must be gone too - otherwise a session revived under the same name would
// start from a stale value.
func TestSnapshotCacheDropsRemovedAndInvalidFiles(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "gone.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking))
	cache := NewSnapshotCache()
	scanner := cachedScanner(directory, now, cache)
	if items := scanner.Scan(); len(items) != 1 {
		t.Fatalf("setup pass returned %d sessions, want 1", len(items))
	}
	if err := os.Remove(filepath.Join(directory, "gone.json")); err != nil {
		t.Fatal(err)
	}
	if items := scanner.Scan(); len(items) != 0 {
		t.Fatalf("a removed snapshot was still returned: %#v", items)
	}
	if cache.Len() != 0 {
		t.Fatalf("cache kept %d entries after the file disappeared", cache.Len())
	}
}

// Sessions that the scanner consumes (closed, expired) must leave the cache
// empty for their files: the file is deleted, so a cached value would be the
// only surviving copy of something that no longer exists on disk.
func TestSnapshotCacheForgetsConsumedFiles(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "closed.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateClosed))
	cache := NewSnapshotCache()
	scanner := cachedScanner(directory, now, cache)
	items := scanner.Scan()
	if len(items) != 1 || items[0].State != session.StateClosed {
		t.Fatalf("closed marker was not returned once: %#v", items)
	}
	if cache.Len() != 0 {
		t.Fatalf("cache kept %d entries for a consumed file", cache.Len())
	}
	if items := scanner.Scan(); len(items) != 0 {
		t.Fatalf("closed marker was returned twice: %#v", items)
	}
}

// The uncached scanner is what --json and every existing test use; the cache
// must be strictly opt-in, or a one-shot caller would keep state it never asked
// for.
func TestSnapshotCacheIsOptIn(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	writeSnapshot(t, directory, "one.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking))
	scanner := SnapshotScanner{Directory: directory, TTL: 15 * time.Second, Now: func() time.Time { return now }}
	for pass := 0; pass < 2; pass++ {
		if items := scanner.Scan(); len(items) != 1 {
			t.Fatalf("pass %d: uncached scanner returned %d sessions", pass, len(items))
		}
	}
}

// Two directories can hold a file with the same name, and the key is the full
// path, so a cache must never answer one from the other. It also only remembers
// the directory it last scanned: a cache belongs to one overlay, and the overlay
// only ever scans one directory.
func TestSnapshotCacheIsKeyedByPath(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	cache := NewSnapshotCache()
	first := t.TempDir()
	second := t.TempDir()
	writeSnapshot(t, first, "same-name.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking))
	writeSnapshot(t, second, "same-name.json", snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateThinking))

	if items := cachedScanner(first, now, cache).Scan(); len(items) != 1 || items[0].State != session.StateWorking {
		t.Fatalf("first directory read %d sessions: %#v", len(items), items)
	}
	if cache.Len() != 1 {
		t.Fatalf("cache holds %d entries after one scan, want 1", cache.Len())
	}
	// Same file name, different directory, different content: a cache keyed by
	// name would answer with the first directory's session here.
	if items := cachedScanner(second, now, cache).Scan(); len(items) != 1 || items[0].State != session.StateThinking {
		t.Fatalf("second directory was answered from the first one's path: %#v", items)
	}
	if cache.Len() != 1 {
		t.Fatalf("cache holds %d entries, want 1 (it keeps only the directory it last saw)", cache.Len())
	}
}

// A broken file must not poison the cache: it is removed, and a same-named file
// written afterwards must be read fresh.
func TestSnapshotCacheDoesNotCacheParseFailures(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	directory := t.TempDir()
	path := filepath.Join(directory, "broken.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	cache := NewSnapshotCache()
	scanner := cachedScanner(directory, now, cache)
	if items := scanner.Scan(); len(items) != 0 {
		t.Fatalf("an unparsable snapshot produced %d sessions", len(items))
	}
	if cache.Len() != 0 {
		t.Fatalf("cache kept %d entries for a file it could not parse", cache.Len())
	}
	if err := os.WriteFile(path, []byte(snapshotJSON(now, os.Getpid(), now.Add(-time.Second), session.StateWorking)), 0o600); err != nil {
		t.Fatal(err)
	}
	items := scanner.Scan()
	if len(items) != 1 || items[0].State != session.StateWorking {
		t.Fatalf("a repaired file was not read: %#v", items)
	}
}
