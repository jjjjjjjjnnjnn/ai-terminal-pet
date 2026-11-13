package store

// runtime/ is a directory of reporter snapshots, and the scanner treats every
// *.json in it as one: anything it cannot parse as a session payload is
// unparseable evidence from a dead reporter, and it deletes it.
//
// That is the right rule for the directory, and it is why nothing else may
// write a JSON file there. The alert receipt learned this the hard way: it lived
// in runtime/ for one release and was deleted within a single scan cycle, so
// --json reported "no receipt" and --doctor said the overlay had never scanned.
// This test pins the rule from the directory's side, so the next feature that
// wants a JSON file next to the snapshots finds out here.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScannerDeletesAnyJsonThatIsNotASnapshot(t *testing.T) {
	directory := t.TempDir()
	stranger := filepath.Join(directory, "alerts.json")
	if err := os.WriteFile(stranger, []byte(`{"at":"2026-11-13T00:00:00Z","presence":"unknown"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	scanner := SnapshotScanner{
		Directory: directory,
		TTL:       DefaultSnapshotTTL,
		Now:       func() time.Time { return time.Unix(1000, 0) },
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	for range scanner.Scan() {
		// The scan only has to run; what matters is the file afterwards.
	}

	if _, err := os.Stat(stranger); !os.IsNotExist(err) {
		t.Fatalf("a non-snapshot .json survived a scan in runtime/: %v", err)
	}
}
