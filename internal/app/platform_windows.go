//go:build windows

package app

// The Windows half of the overlay's lifecycle: the single-instance mutex, the
// process/window presence cache, and --doctor's Windows-only pieces.
//
// RunWith lives in run.go and calls the four hooks below. Everything in this file
// is the part that genuinely needs Win32.

import (
	"ai-terminal-pet/internal/platform"
)

// newInstance takes the single-instance handle.
//
// The name is prefixed "Local\" so the handle is scoped to this logon session: two
// users may each run their own overlay, and a machine-wide object would let one
// user's instance block another's.
func newInstance(name string) (*platform.Instance, error) {
	return platform.AcquireInstance(`Local\` + name)
}

// errAlreadyRunning is what newInstance reports when somebody else holds it.
var errAlreadyRunning = platform.ErrAlreadyRunning

// newPresence builds the OS presence cache the scanner and the attendance sampler
// share.
func newPresence() *platform.Presence {
	return platform.NewPresence(platform.DefaultPresenceTTL)
}

// hideConsole detaches from any console the process inherited.
func hideConsole() {
	platform.HideConsole()
}

// uiAvailable reports whether this build has a window layer. It always does on
// Windows: internal/ui/app_windows.go is the window.
func uiAvailable() bool { return true }
