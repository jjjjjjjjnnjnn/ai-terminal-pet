package store

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-terminal-pet/internal/session"
)

const (
	DefaultSnapshotTTL      = 15 * time.Second
	DefaultMaxSnapshotFiles = 256
	maxSnapshotFilesLimit   = 4096
)

var (
	errSnapshotOversized  = errors.New("snapshot payload is too large")
	errSnapshotChanged    = errors.New("snapshot changed while reading")
	errSnapshotNotRegular = errors.New("snapshot is not a regular file")
)

type SnapshotScanner struct {
	Directory string
	TTL       time.Duration
	Now       func() time.Time
	Alive     func(int) bool
	// KeepAlive answers "is this session's host terminal still there?".
	// A snapshot whose payload has not been updated within TTL is retained
	// (and returned) while its host is alive. Reporters only write on state
	// changes, so a session that has been working for ten minutes has an old
	// timestamp but is very much alive. Deleting that file was what made the
	// panel - and --json - lose sessions that never actually went away.
	// Nil keeps the legacy TTL-only behaviour, which the offline tests use.
	KeepAlive func(session.Session) bool
	MaxFiles  int
}

type snapshotFile struct {
	path string
	info os.FileInfo
}

// Scan returns the sessions whose snapshot files are still on disk.
func (s SnapshotScanner) Scan() []session.Session {
	items, _ := s.ScanWithKeys()
	return items
}

// ScanWithKeys also reports which session identities were observed on disk in
// this pass. The running overlay uses that set as the authority for "does this
// conversation still exist?" - a reporter that ends a session removes (or
// closes) its snapshot, and the store can then drop the in-memory entry
// instead of keeping it forever on the strength of a long-lived parent pid.
//
// `closed` snapshots are intentionally absent from the key set: the scanner
// consumes and deletes them in the same pass, and the store already tombstones
// them from the payload itself.
func (s SnapshotScanner) ScanWithKeys() ([]session.Session, map[string]bool) {
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	now = normalizedTime(now)
	ttl := positiveDuration(s.TTL, DefaultSnapshotTTL)
	maxFiles := boundedSnapshotFiles(s.MaxFiles)

	directoryInfo, err := os.Lstat(s.Directory)
	if err != nil || !directoryInfo.IsDir() || directoryInfo.Mode()&os.ModeSymlink != 0 {
		return []session.Session{}, map[string]bool{}
	}
	directory, err := os.Open(s.Directory)
	if err != nil {
		return []session.Session{}, map[string]bool{}
	}
	defer directory.Close()

	batchSize := maxFiles
	if batchSize > 128 {
		batchSize = 128
	}
	candidates := make([]snapshotFile, 0, maxFiles)
	pruneCutoff := normalizedTime(now).Add(-positiveDuration(s.TTL, DefaultSnapshotTTL))
	prunedTmp := 0
	for {
		infos, readErr := directory.Readdir(batchSize)
		for _, info := range infos {
			if info == nil || info.IsDir() {
				continue
			}
			if strings.EqualFold(filepath.Ext(info.Name()), ".tmp") {
				if prunedTmp < 64 && !info.ModTime().After(pruneCutoff) {
					removeSnapshotIfUnchanged(filepath.Join(s.Directory, info.Name()), info)
					prunedTmp++
				}
				continue
			}
			if !strings.EqualFold(filepath.Ext(info.Name()), ".json") {
				continue
			}
			candidate := snapshotFile{path: filepath.Join(s.Directory, info.Name()), info: info}
			candidates = retainSnapshotCandidate(candidates, candidate, maxFiles)
		}
		if readErr != nil {
			break
		}
	}

	items := make([]session.Session, 0, len(candidates))
	keys := make(map[string]bool, len(candidates))
	cutoff := now.Add(-ttl)
	for _, candidate := range candidates {
		data, info, err := readSnapshot(candidate.path)
		if err != nil {
			if errors.Is(err, errSnapshotOversized) {
				removeSnapshotIfUnchanged(candidate.path, info)
			}
			continue
		}
		value, err := session.ParsePayload(data, now)
		if err != nil {
			removeSnapshotIfUnchanged(candidate.path, info)
			continue
		}
		if value.UpdatedAt.Before(cutoff) && (s.KeepAlive == nil || !s.KeepAlive(value)) {
			removeSnapshotIfUnchanged(candidate.path, info)
			continue
		}
		if value.State == session.StateClosed {
			if s.Alive != nil && processBacked(value.Source) && !s.Alive(value.PID) {
				removeSnapshotIfUnchanged(candidate.path, info)
				continue
			}
			removeSnapshotIfUnchanged(candidate.path, info)
			items = append(items, value)
			continue
		}
		if s.Alive != nil && processBacked(value.Source) && !s.Alive(value.PID) &&
			(s.KeepAlive == nil || !s.KeepAlive(value)) {
			removeSnapshotIfUnchanged(candidate.path, info)
			continue
		}
		keys[key(value)] = true
		items = append(items, value)
	}
	return items, keys
}

func (s SnapshotScanner) Err() error {
	if s.Directory == "" {
		return errors.New("runtime directory is empty")
	}
	return nil
}

func boundedSnapshotFiles(value int) int {
	if value <= 0 {
		return DefaultMaxSnapshotFiles
	}
	if value > maxSnapshotFilesLimit {
		return maxSnapshotFilesLimit
	}
	return value
}

func retainSnapshotCandidate(candidates []snapshotFile, candidate snapshotFile, maxFiles int) []snapshotFile {
	if len(candidates) < maxFiles {
		return append(candidates, candidate)
	}
	oldest := 0
	for index := 1; index < len(candidates); index++ {
		if snapshotFileBefore(candidates[index], candidates[oldest]) {
			oldest = index
		}
	}
	if snapshotFileBefore(candidates[oldest], candidate) {
		removeSnapshotIfUnchanged(candidates[oldest].path, candidates[oldest].info)
		candidates[oldest] = candidate
		return candidates
	}
	removeSnapshotIfUnchanged(candidate.path, candidate.info)
	return candidates
}

func snapshotFileBefore(left, right snapshotFile) bool {
	leftTime := left.info.ModTime()
	rightTime := right.info.ModTime()
	if !leftTime.Equal(rightTime) {
		return leftTime.Before(rightTime)
	}
	return left.info.Name() < right.info.Name()
}

func readSnapshot(path string) ([]byte, os.FileInfo, error) {
	pathInfo, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if !pathInfo.Mode().IsRegular() {
		return nil, pathInfo, errSnapshotNotRegular
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	before, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, before, errSnapshotNotRegular
	}
	if before.Size() > int64(session.MaxPayloadBytes) {
		return nil, before, errSnapshotOversized
	}

	data, err := io.ReadAll(io.LimitReader(file, int64(session.MaxPayloadBytes)+1))
	if err != nil {
		return nil, before, err
	}
	after, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	if !sameSnapshotFile(before, after) {
		return nil, after, errSnapshotChanged
	}
	if len(data) > session.MaxPayloadBytes {
		return nil, after, errSnapshotOversized
	}
	return data, after, nil
}

func removeSnapshotIfUnchanged(path string, expected os.FileInfo) {
	if expected == nil {
		return
	}
	current, err := os.Lstat(path)
	if err != nil || !sameSnapshotFile(expected, current) {
		return
	}
	_ = os.Remove(path)
}

func sameSnapshotFile(expected, current os.FileInfo) bool {
	if expected == nil || current == nil {
		return false
	}
	return os.SameFile(expected, current) &&
		expected.Size() == current.Size() &&
		expected.ModTime().Equal(current.ModTime()) &&
		expected.Mode() == current.Mode()
}
