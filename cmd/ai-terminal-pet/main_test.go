package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/home"
)

func setTestDataRoot(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	local := filepath.Join(root, "local")
	xdg := filepath.Join(root, "xdg")
	homeDir := filepath.Join(root, "home")
	for _, dir := range []string{local, xdg, homeDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("XDG_DATA_HOME", xdg)
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
}

func TestClearStateAndWriteErrorUseHomeDirectory(t *testing.T) {
	setTestDataRoot(t)
	statePath := home.Join(config.StateFileName)
	if statePath == "" {
		t.Fatal("home.Join(StateFileName) returned empty")
	}
	if err := os.MkdirAll(filepath.Dir(statePath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte(`{"placed":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := clearState(); err != nil {
		t.Fatalf("clearState() = %v", err)
	}
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("state.json should be removed, stat err = %v", err)
	}

	writeError(errors.New("boom test"))
	logPath := home.Join("error.log")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) = %v", logPath, err)
	}
	if !strings.Contains(string(raw), "boom test") {
		t.Fatalf("error.log = %q, want boom test", string(raw))
	}
}
