package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/home"
)

// TestUsageErrorPrintsTheUsageVerbatim pins a bug that survived for years: the
// usage text contains "at 100% zoom", and usageError spliced it into a format
// string, so every usage error printed "100%!z(MISSING)oom".
func TestUsageErrorPrintsTheUsageVerbatim(t *testing.T) {
	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = writer
	defer func() { os.Stderr = original }()

	returned := usageError("unknown argument %q", "--bogus")
	writer.Close()
	captured, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !errors.Is(returned, errUsage) {
		t.Fatalf("usageError returned %v, want errUsage", returned)
	}
	text := string(captured)
	if strings.Contains(text, "MISSING") {
		t.Fatalf("the usage text was used as a format string again:\n%s", text)
	}
	if !strings.Contains(text, "100% zoom") {
		t.Fatalf("the usage text is missing its percent sign:\n%s", text)
	}
	if !strings.Contains(text, `unknown argument "--bogus"`) {
		t.Fatalf("the message lost its own formatting:\n%s", text)
	}
}

func TestControlSetWritesEveryPairAtOnce(t *testing.T) {
	setTestDataRoot(t)
	if err := controlSet("--set", []string{"theme=paper", "sound=off", "opacity=200"}); err != nil {
		t.Fatalf("controlSet: %v", err)
	}
	stored := config.Load(home.Join("config.json"))
	if stored.Theme != config.ThemePaper {
		t.Fatalf("theme = %q, want paper", stored.Theme)
	}
	if stored.Sound {
		t.Fatal("sound is still on")
	}
	if stored.Opacity != 200 {
		t.Fatalf("opacity = %d, want 200", stored.Opacity)
	}
}

// TestControlSetLeavesTheFileAloneWhenAPairIsBad is the promise that makes the
// multi-pair form safe: validation happens for every pair before the single
// write, so a typo in the last argument cannot leave the first ones applied.
func TestControlSetLeavesTheFileAloneWhenAPairIsBad(t *testing.T) {
	setTestDataRoot(t)
	if err := config.Save(home.Join("config.json"), config.Default()); err != nil {
		t.Fatal(err)
	}
	err := controlSet("--set", []string{"theme=paper", "opacity=999"})
	if !errors.Is(err, errUsage) {
		t.Fatalf("controlSet returned %v, want errUsage", err)
	}
	stored := config.Load(home.Join("config.json"))
	if stored.Theme != config.Default().Theme {
		t.Fatalf("the first pair was written despite the second failing: theme = %q", stored.Theme)
	}
}

func TestControlSetClearsTheMuteListOnlyWithClear(t *testing.T) {
	setTestDataRoot(t)
	muted := config.Default()
	muted.MutedSessions = []string{"claude-hook|a"}
	if err := config.Save(home.Join("config.json"), muted); err != nil {
		t.Fatal(err)
	}
	if err := controlSet("--set", []string{"mutedSessions=clear"}); err != nil {
		t.Fatalf("controlSet mutedSessions=clear: %v", err)
	}
	if left := config.Load(home.Join("config.json")).MutedSessions; len(left) != 0 {
		t.Fatalf("the mute list still has %v", left)
	}
	if err := controlSet("--set", []string{"mutedSessions=claude-hook|a"}); !errors.Is(err, errUsage) {
		t.Fatalf("setting the list by hand returned %v, want errUsage", err)
	}
}

func TestControlSetRejectsALoneWord(t *testing.T) {
	setTestDataRoot(t)
	if err := controlSet("--set", []string{"theme"}); !errors.Is(err, errUsage) {
		t.Fatalf("returned %v, want errUsage", err)
	}
}

func TestSettingsTerminalRejectsArguments(t *testing.T) {
	if err := controlSettingsTerminal([]string{"--json"}); !errors.Is(err, errUsage) {
		t.Fatalf("returned %v, want errUsage", err)
	}
}
