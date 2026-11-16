//go:build darwin

package ui

// Delivery on macOS is `osascript -e 'display notification ...'`, for the same
// reason the Linux path shells out: the native route (UNUserNotificationCenter)
// requires an app bundle and a signed identifier - this project installs a bare
// binary - and NSUserNotification is deprecated and cannot be verified without
// Apple hardware.
//
// That makes this the one alert surface in the project that is implemented but
// unverified live: the argument list and the escaping are pinned by tests, and
// --doctor says so out loud. Everything that can be checked without a Mac is.

import (
	"os/exec"
	"sync"
)

var (
	osascriptMu       sync.Mutex
	osascriptResolved bool
	osascriptPath     string
)

func osascriptBinary() string {
	osascriptMu.Lock()
	defer osascriptMu.Unlock()
	if !osascriptResolved {
		if path, err := exec.LookPath("osascript"); err == nil {
			osascriptPath = path
		}
		osascriptResolved = true
	}
	return osascriptPath
}

// notify delivers an alert through the notification centre.
func (a *App) notify(value notification) {
	path := osascriptBinary()
	if path == "" {
		return
	}
	command := exec.Command(path, "-e", osascriptCommand(value))
	if err := command.Start(); err != nil {
		return
	}
	go func() { _ = command.Wait() }()
}

// osascriptCommand is the whole script, split out so a test can pin it.
func osascriptCommand(value notification) string {
	return "display notification " + appleScriptLiteral(value.Body) +
		" with title " + appleScriptLiteral(value.Title)
}

// NotificationSurface names what `notifications` reaches on this machine, for
// --doctor. The wording keeps the unverified half visible: osascript exists on
// every macOS install, but whether the banner appears was never observed on
// real Apple hardware from this project.
func NotificationSurface() string {
	path := osascriptBinary()
	if path == "" {
		return "none (osascript not found: alerts keep the unread mark and the tone only)"
	}
	return "osascript display notification (" + path + "; best-effort, live display unverified)"
}

// notificationSurfaceAvailable answers for the alert receipt.
func notificationSurfaceAvailable() bool { return osascriptBinary() != "" }
