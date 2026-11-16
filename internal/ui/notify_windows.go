//go:build windows

package ui

import "github.com/lxn/win"

// notify delivers an alert through the tray balloon. The tray can be briefly
// absent at startup and after an Explorer restart, which balloonText already
// treats as a no-op.
func (a *App) notify(value notification) {
	a.balloonText(value.Title, value.Body, windowsBalloonFlags(value.Level))
}

func windowsBalloonFlags(level notifyLevel) uint32 {
	switch level {
	case notifyInfo:
		return niifInfo
	case notifyError:
		return uint32(win.NIIF_ERROR)
	default:
		return uint32(win.NIIF_WARNING)
	}
}

// NotificationSurface names what `notifications` reaches on this machine, for
// --doctor. Shell_NotifyIconW ships with the shell, so the surface exists on
// every Windows install.
func NotificationSurface() string {
	return "tray balloon (Shell_NotifyIconW, built in)"
}

// notificationSurfaceAvailable answers for the alert receipt. Windows always
// has the surface; whether the tray icon is currently registered is a startup
// detail, not a property of the machine.
func notificationSurfaceAvailable() bool { return true }
