//go:build darwin

package ui

// macOS is the one platform whose alert surface cannot be exercised on the
// machine this project is built on, so the test pins the only part that is
// knowable without Apple hardware: the script we hand to osascript, and that
// the delivery starts it with the arguments we composed.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func resetOsascriptProbe() {
	osascriptMu.Lock()
	osascriptResolved = false
	osascriptPath = ""
	osascriptMu.Unlock()
}

func TestOsascriptCommandIsOneDisplayNotification(t *testing.T) {
	script := osascriptCommand(notification{Title: `say "hi"`, Body: `back\slash`})
	want := `display notification "back\\slash" with title "say \"hi\""`
	if script != want {
		t.Fatalf("script = %s, want %s", script, want)
	}
	if strings.Count(script, "display notification") != 1 {
		t.Fatalf("one alert must be one statement: %s", script)
	}
}

func TestNotifyStartsOsascriptWithTheScript(t *testing.T) {
	dir := t.TempDir()
	argv := filepath.Join(dir, "argv.txt")
	fake := filepath.Join(dir, "osascript")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$PET_FAKE_ARGV\"\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake osascript: %v", err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("PET_FAKE_ARGV", argv)
	resetOsascriptProbe()
	t.Cleanup(resetOsascriptProbe)

	if surface := NotificationSurface(); !strings.HasPrefix(surface, "osascript ") {
		t.Fatalf("NotificationSurface() = %q, want the found binary", surface)
	}

	(&App{}).notify(notification{Title: "Session finished", Body: "ai-terminal-pet"})

	deadline := time.Now().Add(3 * time.Second)
	for {
		if data, err := os.ReadFile(argv); err == nil {
			text := string(data)
			if !strings.Contains(text, "-e") || !strings.Contains(text, "display notification") {
				t.Fatalf("osascript received %q, want -e display notification", text)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("osascript was never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestNotifyIsSilentWithoutOsascript(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	resetOsascriptProbe()
	t.Cleanup(resetOsascriptProbe)

	if alertSurfaceAvailable() {
		t.Fatal("an empty PATH must report no notification surface")
	}
	(&App{}).notify(notification{Title: "t", Body: "b"})
}
