package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --json has to say what the alert layer decided, or a script cannot tell "nothing
// needed me" from "the overlay was not running". These tests cover the file
// contract only; the decision itself is pinned in internal/ui.

func TestAlertReportIsReadFromTheRuntimeDirectory(t *testing.T) {
	directory := t.TempDir()
	record := AlertReport{
		At:       time.Date(2026, 11, 12, 20, 0, 0, 0, time.UTC),
		Presence: "away",
		Present:  false,
		Balloon:  "claude-hook · ses_1",
		Tone:     "waiting",
		Sessions: []AlertReportSession{
			{Identity: "claude-hook · ses_1", State: "waiting_permission", Kind: "approval", Decision: "elsewhere", Unread: true},
			{Identity: "opencode · ses_2", State: "working", Kind: "", Decision: "not-asking"},
		},
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "alerts.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	got := readAlertReport(directory)
	if got == nil {
		t.Fatal("the receipt was not read")
	}
	if got.Balloon != record.Balloon || got.Presence != "away" || len(got.Sessions) != 2 {
		t.Fatalf("receipt round trip lost data: %+v", got)
	}
	if !got.Sessions[0].Unread {
		t.Fatal("the unread flag did not survive")
	}
}

// A missing receipt is null, not an error: an overlay that has not run since the
// upgrade has written nothing yet, and --json must still answer.
func TestAlertReportMissingIsNil(t *testing.T) {
	if got := readAlertReport(t.TempDir()); got != nil {
		t.Fatalf("a missing receipt should read as nil, got %+v", got)
	}
}

func TestAlertReportCorruptIsNil(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "alerts.json"), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readAlertReport(directory); got != nil {
		t.Fatalf("a corrupt receipt should read as nil, got %+v", got)
	}
}

// The report carries the receipt under "alerts" and says nothing else new: the
// existing keys scripts already parse must not move.
func TestReportJSONShape(t *testing.T) {
	report := Report{
		SchemaVersion: "ai-terminal-pet.v1",
		State:         "idle",
		Count:         0,
		GeneratedAt:   time.Date(2026, 11, 12, 20, 0, 0, 0, time.UTC),
		Sessions:      nil,
		Alerts:        &AlertReport{Presence: "unknown"},
	}
	encoded, err := EncodeReport(report, false)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schemaVersion", "state", "count", "generatedAt", "sessions", "alerts"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("--json lost the %q key", key)
		}
	}
	if len(decoded) != 6 {
		t.Fatalf("--json grew an undocumented key: %v", decoded)
	}
	// With no receipt the key is still present, and explicitly null.
	encoded, err = EncodeReport(Report{SchemaVersion: "ai-terminal-pet.v1"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if string(decoded["alerts"]) != "null" {
		t.Fatalf("an absent receipt should be null, got %s", decoded["alerts"])
	}
}
