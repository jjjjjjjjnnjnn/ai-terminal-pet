//go:build linux || darwin

package ui

// Jump feedback, POSIX half.
//
// Windows answers a failed jump with a tray balloon and a beep (focusFailure in
// app_windows.go). The POSIX ports had no surface at all, so a tap on a session
// whose terminal was already gone did nothing - indistinguishable from a tap
// the window never received, and it silently dropped the unread mark as well.
//
// The pill the copy receipt already draws is the right answer here: it appears
// over the row that was tapped, needs no daemon and no permission, and blows up
// in the error colour. The tone makes it audible from behind another window,
// which is where the user usually is when a terminal disappears.

import "time"

// jumpStaleAfter bounds how late a failed resolution may still speak. Resolving
// a target walks the window list, and a failure that arrives seconds later is
// about a tap the user has already forgotten.
const jumpStaleAfter = 2 * time.Second

// focusRequest is everything the UI thread needs to finish a jump: the handle
// that was resolved, and the receipt's subject - the row the tap was drawn on
// and the conversation it was about, which is what keeps a pill from naming
// whoever the sort moved into that row meanwhile.
type focusRequest struct {
	handle   uintptr
	row      int
	identity string
	at       time.Time
}

// focusFailed says the jump did not happen.
func (a *App) focusFailed(request focusRequest, now time.Time) {
	if now.Sub(request.at) > jumpStaleAfter {
		return
	}
	a.toast.show(request.row, request.identity, toastJumpFailed, now)
	beep(beepExclamation)
	a.syncAnimation()
	_ = a.present()
}
