package store

// Where the scan time actually goes, and what the cache below removes.
//
// Measured on this machine (NTFS + active AV filtering, 2026-11-15):
//
//	os.Lstat                       31.3 us
//	readSnapshot (4 stats + read)  89.6 us
//	session.ParsePayload            5.1 us
//	one candidate, end to end      98.1 us
//
// So parsing is 5% of the cost and metadata calls are 90%. The thing to remove
// is therefore not "parse less" but "look at the disk less": for a file whose
// identity, size and mtime have not moved, the payload cannot have changed, and
// its parse result can be reused as-is.
//
// Safety comes from the same three fields readSnapshot already compares before
// it hands bytes back (sameSnapshotFile). A reporter never writes in place: it
// writes a temp file and renames, so a rewrite lands as a NEW file and
// os.SameFile sees it (inode on Unix, file index on NTFS). Measured on this
// machine: after a temp+rename with the mtime put back by hand and an identical
// byte length, SameFile(snapshot from the cache, fresh stat) = false, so the
// rewrite cannot hide behind metadata.
//
// One measured caveat, recorded rather than papered over: NTFS did reuse the
// file index once in 40 temp+rename rounds on this machine. That is why the
// guard is "identity AND size AND mtime" rather than identity alone - a reused
// index still has to match both other fields, and the temp+rename path leaves
// the mtime moved. This cache also never deletes anything: removals go through
// removeSnapshotIfUnchanged, which re-stats before deleting, so the worst case
// of a coincidental match is a stale read, not a lost file.
//
// The cache is a performance device only - a missing entry simply costs what the
// uncached path cost, which is what every test and every one-shot caller (--json)
// still pay, because a nil cache means "read everything".

import (
	"os"
	"sync"

	"ai-terminal-pet/internal/session"
)

type cachedSnapshot struct {
	info  os.FileInfo
	value session.Session
}

// SnapshotCache remembers the parse result of the files a previous scan read.
//
// It is a pointer type on SnapshotScanner on purpose: the scanner is a value, is
// copied freely and is rebuilt per call in some places, while the cache must
// survive those copies. A nil cache means "read every file every time", which is
// what every existing test does and what a one-shot caller (--json) still pays.
type SnapshotCache struct {
	mutex   sync.Mutex
	entries map[string]cachedSnapshot
	// hits and misses are counters, not decisions: the cache is correct either
	// way, and the numbers exist so a test can assert the cache is actually
	// consulted (a second scan of unchanged files must miss zero times) and so
	// a future --doctor could report it. Nothing in the scan path reads them.
	hits   int
	misses int
}

// NewSnapshotCache returns an empty cache.
func NewSnapshotCache() *SnapshotCache {
	return &SnapshotCache{entries: make(map[string]cachedSnapshot, 64)}
}

// lookup answers whether this exact file was parsed on an earlier scan.
func (c *SnapshotCache) lookup(path string, info os.FileInfo) (session.Session, bool) {
	if c == nil || info == nil {
		return session.Session{}, false
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry, ok := c.entries[path]
	if !ok || !sameSnapshotFile(entry.info, info) {
		c.misses++
		return session.Session{}, false
	}
	c.hits++
	return entry.value, true
}

// Hits and Misses answer "was the cache actually used?". They are diagnostics,
// not policy: the scanner behaves identically whatever they say.
func (c *SnapshotCache) Hits() int {
	if c == nil {
		return 0
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.hits
}

func (c *SnapshotCache) Misses() int {
	if c == nil {
		return 0
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.misses
}

// store records one successfully parsed file, tagged with the info the caller
// read it through so the next scan can compare the same two fields.
func (c *SnapshotCache) store(path string, info os.FileInfo, value session.Session) {
	if c == nil || info == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[path] = cachedSnapshot{info: info, value: value}
}

// forget drops one path. It is called whenever the file could not be read or
// was consumed, so a broken or closed file never keeps answering from cache.
func (c *SnapshotCache) forget(path string) {
	if c == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.entries, path)
}

// retain keeps only the paths this scan saw. Without it a long-lived overlay
// would remember every conversation it has ever read.
func (c *SnapshotCache) retain(paths map[string]bool) {
	if c == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for path := range c.entries {
		if !paths[path] {
			delete(c.entries, path)
		}
	}
}

// Len is for tests and for the measurement in ROADMAP §3.
func (c *SnapshotCache) Len() int {
	if c == nil {
		return 0
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.entries)
}
