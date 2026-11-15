//go:build !windows

package autostart

// The tests that can run on a platform whose autostart mechanism is a file.
//
// Windows' is a registry key, so its tests live in autostart_windows_test.go and
// use the in-memory backend - a unit test must never turn on autostart for whoever
// ran `go test`. Here the mechanism is a file, so the test can drive the real
// backend against a temp directory and assert what a session would actually read:
// the plist or the desktop entry.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutostartRoundTripsThroughTheRealBackend(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	manager := New()
	if status, err := manager.Status(); err != nil || status.Enabled {
		t.Fatalf("a fresh tree must report off, got %+v / %v", status, err)
	}

	executable := filepath.Join(root, "AI-Terminal-Pet.exe")
	if err := manager.Enable(executable); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	status, err := manager.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Enabled || status.Target != executable {
		t.Fatalf("after Enable: %+v, want enabled -> %s", status, executable)
	}

	// The file has to name the executable in a way the session will run it.
	if contents := readAutostartFile(t, root); !strings.Contains(contents, executable) {
		t.Fatalf("the entry does not name the executable:\n%s", contents)
	}

	if changed, err := manager.Sync(executable); err != nil || changed {
		t.Fatalf("Sync on an already-correct entry changed something: %v / %v", changed, err)
	}
	moved := filepath.Join(root, "moved", "AI-Terminal-Pet.exe")
	if changed, err := manager.Sync(moved); err != nil || !changed {
		t.Fatalf("Sync must repair a moved install: changed=%v err=%v", changed, err)
	}
	if status, _ := manager.Status(); status.Target != moved {
		t.Fatalf("Sync left the entry pointing at %q", status.Target)
	}

	if err := manager.Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if status, _ := manager.Status(); status.Enabled {
		t.Fatalf("Disable left the entry enabled: %+v", status)
	}
	// Disabling twice is not an error: the intent is "make sure it is off".
	if err := manager.Disable(); err != nil {
		t.Fatalf("second Disable: %v", err)
	}
}

func TestSyncNeverTurnsAutostartOn(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	manager := New()
	if changed, err := manager.Sync(filepath.Join(root, "AI-Terminal-Pet.exe")); err != nil || changed {
		t.Fatalf("Sync on a disabled entry did something: %v / %v", changed, err)
	}
	if status, _ := manager.Status(); status.Enabled {
		t.Fatal("Sync turned autostart on behind the user's back")
	}
}

// readAutostartFile returns the entry file's contents, whichever mechanism this
// platform uses. It is the assertion's window into "what the session will read".
func readAutostartFile(t *testing.T, root string) string {
	t.Helper()
	candidates := []string{
		filepath.Join(root, "Library", "LaunchAgents", "org.ai-terminal-pet.overlay.plist"),
		filepath.Join(root, "config", "autostart", "ai-terminal-pet.desktop"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}
	t.Fatalf("no autostart entry was written under %s", root)
	return ""
}
