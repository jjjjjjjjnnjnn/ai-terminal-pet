package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// writeSnapshot drops a reporter snapshot into a throwaway runtime directory.
func writeSnapshot(t *testing.T, root, name, state string, updatedAt time.Time) {
	t.Helper()
	directory := filepath.Join(root, "AI-Terminal-Pet", "runtime")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatalf("mkdir runtime: %v", err)
	}
	payload := fmt.Sprintf(
		`{"schemaVersion":"ai-terminal-pet.v1","source":"unit","sessionId":%q,"pid":%d,"project":"C:\\unit","state":%q,"reason":"unit","startedAt":%q,"updatedAt":%q}`,
		name, os.Getpid(), state,
		updatedAt.Add(-time.Minute).Format(time.RFC3339), updatedAt.Format(time.RFC3339),
	)
	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(payload+"\n"), 0600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}

func TestSnapshotAggregatesOnDiskReporters(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	now := time.Now().UTC()
	writeSnapshot(t, root, "s-idle", string(session.StateIdle), now)
	writeSnapshot(t, root, "s-waiting", string(session.StateWaitingPermission), now)

	report, err := Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if report.SchemaVersion != session.SchemaVersion {
		t.Fatalf("schema must be reported, got %q", report.SchemaVersion)
	}
	if report.Count != 2 || len(report.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got count=%d len=%d", report.Count, len(report.Sessions))
	}
	// The aggregate must be the highest-priority state, not the last written.
	if report.State != session.StateWaitingPermission {
		t.Fatalf("aggregate must be waiting_permission, got %q", report.State)
	}
	if report.GeneratedAt.IsZero() {
		t.Fatal("generatedAt must be set")
	}
}

func TestSnapshotIgnoresStaleSnapshots(t *testing.T) {
	// The whole point of reading files instead of IPC is that the view stays
	// honest: a reporter that stopped writing must disappear, not linger.
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	writeSnapshot(t, root, "s-stale", string(session.StateWorking), time.Now().UTC().Add(-10*time.Minute))

	report, err := Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if report.Count != 0 {
		t.Fatalf("stale snapshot must not count, got %d", report.Count)
	}
	if report.State != session.StateIdle {
		t.Fatalf("empty view must be idle, got %q", report.State)
	}
	if report.Sessions == nil {
		t.Fatal("sessions must marshal as [] not null")
	}
}

func TestSnapshotWithoutRuntimeDirectory(t *testing.T) {
	// The overlay may never have run on this machine; that is an empty view,
	// not an error.
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	report, err := Snapshot()
	if err != nil {
		t.Fatalf("missing runtime directory must not error: %v", err)
	}
	if report.Count != 0 {
		t.Fatalf("expected an empty view, got %d", report.Count)
	}
}

func TestSnapshotRequiresLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")
	if _, err := Snapshot(); err == nil {
		t.Fatal("missing LOCALAPPDATA must be an error")
	}
}

func TestEncodeReport(t *testing.T) {
	report := Report{
		SchemaVersion: session.SchemaVersion,
		State:         session.StateIdle,
		Count:         0,
		GeneratedAt:   time.Unix(0, 0).UTC(),
		Sessions:      []session.Session{},
	}
	compact, err := EncodeReport(report, false)
	if err != nil {
		t.Fatalf("compact: %v", err)
	}
	// A single line is what a status line or a shell pipeline wants.
	if bytes := len(compact); bytes == 0 || compact[0] != '{' {
		t.Fatalf("compact output must be one JSON object, got %q", compact)
	}
	for _, b := range compact {
		if b == '\n' {
			t.Fatalf("compact output must not contain newlines: %q", compact)
		}
	}
	pretty, err := EncodeReport(report, true)
	if err != nil {
		t.Fatalf("pretty: %v", err)
	}
	if len(pretty) <= len(compact) {
		t.Fatal("pretty output must be longer than compact")
	}
	// Both must decode to the same thing.
	var fromCompact, fromPretty Report
	if err := json.Unmarshal(compact, &fromCompact); err != nil {
		t.Fatalf("compact must be valid JSON: %v", err)
	}
	if err := json.Unmarshal(pretty, &fromPretty); err != nil {
		t.Fatalf("pretty must be valid JSON: %v", err)
	}
	if fromCompact.Count != fromPretty.Count || fromCompact.State != fromPretty.State {
		t.Fatal("both encodings must describe the same report")
	}
}
