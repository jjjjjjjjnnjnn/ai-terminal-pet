//go:build darwin

package ui

// The macOS tone path is the one piece of this project that has never been
// heard on real hardware, and the first time it was ever executed - a headless
// `go test` on macos-latest, 2026-11-16 - it faulted with SIGSEGV inside
// objc_msgSend, in the deferred NSAutoreleasePool drain, because there was no
// application session for NSSound to belong to.
//
// The gate is darwinAudioReady, set once by initCocoa. This test is the guard:
// remove the gate and the suite crashes instead of failing neatly.

import (
	"testing"
	"time"
)

func TestDarwinBeepIsInertWithoutAnAppKitSession(t *testing.T) {
	if darwinAudioReady.Load() {
		t.Skip("this process already started an AppKit session")
	}
	previousVolume := int(soundVolume.Load())
	t.Cleanup(func() { setSoundVolume(previousVolume) })
	setSoundVolume(80)

	beep(beepExclamation)
	beep(beepError)
	beep(beepFinished)

	if darwinAudioReady.Load() {
		t.Fatal("beep must not start an AppKit session")
	}
}

// The pill path is the same one the user sees after a failed jump; on macOS it
// must work with no window, no AppKit and no audio.
func TestDarwinFocusFailedNeedsNoAppKitSession(t *testing.T) {
	if darwinAudioReady.Load() {
		t.Skip("this process already started an AppKit session")
	}
	app := focusTestApp()
	now := time.Now()
	request := focusRequest{row: 0, identity: "claude-hook · ses_1", at: now}
	app.focusFailed(request, now)
	if !app.toast.active(now) {
		t.Fatal("a failed jump must arm the pill even without AppKit")
	}
}
