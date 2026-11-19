package ui

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-terminal-pet/internal/home"
)

func timeForLog(index int) time.Time {
	base := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(index-(alertLogMax+6)) * time.Second)
}

func TestAlertLogKeepsOnlyRecentRings(t *testing.T) {
	root := t.TempDir()
	// home.Directory resolves differently per OS; pin every input so the test
	// writes into the sandbox on all three platforms.
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("XDG_DATA_HOME", root)
	t.Setenv("HOME", root)
	path := home.Join(alertLogName)
	if path == "" {
		t.Fatal("the alert log needs a data directory")
	}
	for index := 0; index < alertLogMax+7; index++ {
		appendAlertLog(alertRecord{
			Balloon:  "opencode · ses_1",
			At:       timeForLog(index),
			Presence: "present",
		})
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	count := 0
	last := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			t.Fatal("the log must not contain blank lines")
		}
		var record alertRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %d is not JSON: %v", count+1, err)
		}
		count++
		last = line
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != alertLogMax {
		t.Fatalf("log has %d lines, want %d", count, alertLogMax)
	}
	if !strings.Contains(last, `"at":"2026-11-19T09:00:00Z"`) {
		t.Fatalf("the newest ring is missing: %s", filepath.Base(path))
	}
}
