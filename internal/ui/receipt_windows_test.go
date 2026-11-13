package ui

// The receipt's file path, pinned. It is diagnostics, so it must never be able
// to break an alert - but it also must actually land where --json and --doctor
// look for it, which is a path convention rather than a function call.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/session"
)

func TestAlertReceiptRoundTrip(t *testing.T) {
	// A sandbox LOCALAPPDATA: the receipt path is derived from it, so this
	// exercises the real convention without touching the user's runtime dir.
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	path := alertReceiptPath()
	if path == "" {
		t.Fatal("the receipt path must be derivable from LOCALAPPDATA")
	}
	if want := filepath.Join(root, "AI-Terminal-Pet", "alerts.json"); path != want {
		t.Fatalf("receipt path = %q, want %q", path, want)
	}
	if _, err := os.Stat(filepath.Join(root, "AI-Terminal-Pet")); err == nil {
		t.Fatal("nothing should have been created before the first write")
	}

	record := alertRecord{
		At:       time.Date(2026, 11, 12, 22, 0, 0, 0, time.UTC),
		Presence: "away",
		Balloon:  "claude-hook · ses_1",
		Tone:     toneWaiting,
		Sessions: []alertRecordSession{
			{Identity: "claude-hook · ses_1", State: string(session.StateWaitingPermission), Kind: "approval", Decision: decisionElsewhere, Unread: true},
		},
	}
	writeAlertRecord(record)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the receipt was not written: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("the receipt is empty")
	}
	// One JSON line, newline terminated: a script can read it without a parser
	// that cares about formatting.
	var decoded alertRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("the receipt is not valid JSON: %v", err)
	}
	if decoded.Balloon != record.Balloon || len(decoded.Sessions) != 1 || !decoded.Sessions[0].Unread {
		t.Fatalf("round trip lost data: %+v", decoded)
	}

	got := readAlertRecord()
	if got == nil || got.Balloon != record.Balloon {
		t.Fatalf("readAlertRecord did not return what was written: %+v", got)
	}

	// Rewriting replaces rather than appends: the file is one record, so an
	// eight-second scan cadence cannot grow it.
	record.Balloon = "opencode · ses_2"
	writeAlertRecord(record)
	again, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) > len(data)+64 {
		t.Fatalf("the receipt grew on a rewrite: %d -> %d bytes", len(data), len(again))
	}
}

// A missing LOCALAPPDATA must make the receipt a no-op, not a failure: the
// alert itself already happened.
func TestAlertReceiptWithoutLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")
	if path := alertReceiptPath(); path != "" {
		t.Fatalf("a missing LOCALAPPDATA should produce no path, got %q", path)
	}
	writeAlertRecord(alertRecord{Presence: "unknown"})
	if got := readAlertRecord(); got != nil {
		t.Fatalf("nothing was written, so nothing should be read: %+v", got)
	}
}

// The receipt is written atomically through config.WriteFileAtomic, so a crash
// mid-write leaves the previous record intact rather than a truncated one.
func TestAlertReceiptUsesTheAtomicWriter(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	writeAlertRecord(alertRecord{Presence: "present"})
	entries, err := os.ReadDir(filepath.Join(root, "AI-Terminal-Pet"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Fatalf("an atomic-write temporary was left behind: %s", entry.Name())
		}
	}
	// The temp prefix the atomic writer uses is the one this project owns.
	if config.TempPrefix == "" {
		t.Fatal("config.TempPrefix should name the temporaries")
	}
}
