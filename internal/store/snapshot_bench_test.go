package store

// ROADMAP §3 L5 asks one question before anybody optimises anything: how long
// does the snapshot scan actually take, and does it grow with the file count?
// The answer decides whether "skip the directory when its mtime has not changed"
// or "only scan the newest N" is worth its complexity. This file is the
// measurement, and its numbers are quoted in docs/ROADMAP.md §3.
//
// It is a *_test.go file on purpose: test sources are excluded from the overlay's
// vendored source digest, so recording a number here costs no version bump and
// cannot make packages/dsh-ai-terminal-pet stale.
//
//	go test ./internal/store -run XXX -bench SnapshotScan -benchtime 50x

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkSnapshotScan1(b *testing.B)   { benchmarkSnapshotScan(b, 1) }
func BenchmarkSnapshotScan50(b *testing.B)  { benchmarkSnapshotScan(b, 50) }
func BenchmarkSnapshotScan200(b *testing.B) { benchmarkSnapshotScan(b, 200) }

func benchmarkSnapshotScan(b *testing.B, count int) {
	b.Helper()
	directory := b.TempDir()
	for index := 0; index < count; index++ {
		name := filepath.Join(directory, fmt.Sprintf("bench-54972-bench-%d.json", index))
		if err := os.WriteFile(name, snapshotBenchPayload(index), 0o600); err != nil {
			b.Fatalf("write snapshot %d: %v", index, err)
		}
	}
	scanner := SnapshotScanner{Directory: directory, Now: time.Now}
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if items := scanner.Scan(); len(items) != count {
			b.Fatalf("scan returned %d of %d snapshots", len(items), count)
		}
	}
}

// snapshotBenchPayload is one realistic snapshot: what a reporter writes, at
// roughly the size a real one has on this machine (about 470 bytes live).
func snapshotBenchPayload(index int) []byte {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return []byte(fmt.Sprintf(`{"schemaVersion":"ai-terminal-pet.v1","source":"bench","sessionId":"bench-%d",`+
		`"parentSessionId":"","pid":%d,"parentPid":94756,"project":"C:\\Users\\example\\Desktop\\project",`+
		`"title":"benchmark session %d","model":"provider/model","state":"working","reason":"step_streamed",`+
		`"startedAt":"%s","updatedAt":"%s","tool":"shell","terminal":"bench"}`,
		index, 5000+index, index, now, now))
}
