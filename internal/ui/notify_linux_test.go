//go:build linux

package ui

// The delivery is a subprocess, so these tests install a fake notify-send on
// PATH instead of a notification daemon. They pin our side: the command line,
// the urgency mapping, and that an absent tool is silence rather than a panic.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func resetNotifySendProbe() {
	notifySendMu.Lock()
	notifySendResolved = false
	notifySendPath = ""
	notifySendMu.Unlock()
}

func TestNotificationArgsCarryLevelAndCopy(t *testing.T) {
	args := notificationArgs(notification{Title: "Waiting", Body: "ai-terminal-pet · approval", Level: notifyWarning})
	want := []string{
		"--app-name=" + notificationAppName,
		"--urgency=normal",
		"--icon=dialog-warning",
		"Waiting",
		"ai-terminal-pet · approval",
	}
	if len(args) != len(want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	for index := range want {
		if args[index] != want[index] {
			t.Fatalf("args[%d] = %q, want %q (whole: %v)", index, args[index], want[index], args)
		}
	}
}

func TestUrgencyNamesReserveCriticalForErrors(t *testing.T) {
	cases := map[notifyLevel]string{
		notifyInfo:    "low",
		notifyWarning: "normal",
		notifyError:   "critical",
	}
	for level, want := range cases {
		if got := urgencyName(level); got != want {
			t.Fatalf("urgencyName(%v) = %q, want %q", level, got, want)
		}
		if icon := iconName(level); !strings.HasPrefix(icon, "dialog-") {
			t.Fatalf("iconName(%v) = %q, want a themed dialog icon", level, icon)
		}
	}
}

func TestNotifyStartsNotifySendWithTheAlertCopy(t *testing.T) {
	dir := t.TempDir()
	argv := filepath.Join(dir, "argv.txt")
	player := filepath.Join(dir, "notify-send")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$PET_FAKE_ARGV\"\n"
	if err := os.WriteFile(player, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake notify-send: %v", err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("PET_FAKE_ARGV", argv)
	resetNotifySendProbe()
	t.Cleanup(resetNotifySendProbe)

	if surface := NotificationSurface(); !strings.HasPrefix(surface, "notify-send ") {
		t.Fatalf("NotificationSurface() = %q, want the found binary", surface)
	}

	app := &App{}
	app.notify(notification{Title: "Session finished", Body: "ai-terminal-pet", Level: notifyInfo})

	deadline := time.Now().Add(3 * time.Second)
	for {
		if data, err := os.ReadFile(argv); err == nil {
			text := string(data)
			for _, want := range []string{"--urgency=low", "Session finished", "ai-terminal-pet"} {
				if !strings.Contains(text, want) {
					t.Fatalf("notify-send received %q, missing %q", text, want)
				}
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("notify-send was never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestNotifyIsSilentWithoutNotifySend(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	resetNotifySendProbe()
	t.Cleanup(resetNotifySendProbe)

	if alertSurfaceAvailable() {
		t.Fatal("an empty PATH must report no notification surface")
	}
	if surface := NotificationSurface(); !strings.HasPrefix(surface, "none") {
		t.Fatalf("NotificationSurface() = %q, want none", surface)
	}
	// Must not panic and must not block: the unread mark already happened.
	(&App{}).notify(notification{Title: "t", Body: "b"})
}
