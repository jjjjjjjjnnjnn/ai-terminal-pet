//go:build linux

package ui

// Delivery on Linux is the freedesktop notification daemon, reached through the
// notify-send command that ships with libnotify. This project links no D-Bus
// client and no GTK: one subprocess per alert (rate-limited by PlanAlerts, never
// per frame) is the same trade the sound path makes with its player probe.
//
// No notify-send, or no daemon behind it, is silence - never an error. The panel
// still marks the row unread, and --doctor says which of the two is missing.

import (
	"os/exec"
	"sync"
)

const notificationAppName = "AI Terminal Pet"

var (
	notifySendMu       sync.Mutex
	notifySendResolved bool
	notifySendPath     string
)

// notifySendBinary resolves notify-send once per process: the answer cannot
// change while the overlay runs, and the alert path should not walk PATH.
func notifySendBinary() string {
	notifySendMu.Lock()
	defer notifySendMu.Unlock()
	if !notifySendResolved {
		if path, err := exec.LookPath("notify-send"); err == nil {
			notifySendPath = path
		}
		notifySendResolved = true
	}
	return notifySendPath
}

// notify delivers an alert through the notification daemon.
func (a *App) notify(value notification) {
	path := notifySendBinary()
	if path == "" {
		return
	}
	command := exec.Command(path, notificationArgs(value)...)
	if err := command.Start(); err != nil {
		return
	}
	// Wait reaps the child. Its exit status is deliberately not read: the daemon
	// may refuse (no session bus, D-Bus error) and there is nothing useful the
	// overlay could do about a refused notification except keep the unread mark.
	go func() { _ = command.Wait() }()
}

// notificationArgs is the whole command line, split out so a test can pin it
// without a daemon.
func notificationArgs(value notification) []string {
	args := []string{
		"--app-name=" + notificationAppName,
		"--urgency=" + urgencyName(value.Level),
		"--icon=" + iconName(value.Level),
	}
	return append(args, value.Title, value.Body)
}

// urgencyName maps our three levels onto the freedesktop ones. "critical" is
// reserved for errors because several daemons keep it on screen until it is
// dismissed and give it its own sound.
func urgencyName(level notifyLevel) string {
	switch level {
	case notifyInfo:
		return "low"
	case notifyError:
		return "critical"
	default:
		return "normal"
	}
}

// iconName is the themed icon each level asks for. A daemon that cannot find
// the theme name falls back to its own default, which is acceptable.
func iconName(level notifyLevel) string {
	switch level {
	case notifyInfo:
		return "dialog-information"
	case notifyError:
		return "dialog-error"
	default:
		return "dialog-warning"
	}
}

// NotificationSurface names what `notifications` reaches on this machine, for
// --doctor.
func NotificationSurface() string {
	path := notifySendBinary()
	if path == "" {
		return "none (notify-send not found: alerts keep the unread mark and the tone only)"
	}
	return "notify-send " + path
}

// notificationSurfaceAvailable answers for the alert receipt, which must not
// claim a balloon that nothing could have shown.
func notificationSurfaceAvailable() bool { return notifySendBinary() != "" }
