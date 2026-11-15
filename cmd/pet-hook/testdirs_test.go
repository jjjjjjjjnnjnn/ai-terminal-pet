package main

// One helper for the hook's tests: point every environment variable that could
// decide where this project keeps its files at a temp tree, and return that tree.
//
// This exists because cmd/pet-hook's tests used to set LOCALAPPDATA alone. That was
// correct while the paths were spelled by hand and became wrong the moment
// internal/home learned three platforms - on linux the tests wrote pending
// approval files into the tester's real ~/.local/share and then looked for them
// in a temp directory that nothing had used.

import (
	"os"
	"path/filepath"
	"testing"

	"ai-terminal-pet/internal/home"
)

// dataDirectory isolates the project's data directory and returns the temp root.
func dataDirectory(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", filepath.Join(root, "local"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "xdg"))
	t.Setenv("HOME", filepath.Join(root, "profile"))
	t.Setenv("USERPROFILE", filepath.Join(root, "profile"))
	directory := home.Directory()
	if directory == "" {
		t.Fatal("the platform resolved no data directory from the test environment")
	}
	return root
}

// dataDirectoryFor returns the directory those variables now resolve to, for the
// tests that write a file into it by hand.
func dataDirectoryFor(t *testing.T) string {
	t.Helper()
	dataDirectory(t)
	directory := home.Directory()
	if directory == "" {
		t.Fatal("no data directory")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("mkdir data directory: %v", err)
	}
	return directory
}
