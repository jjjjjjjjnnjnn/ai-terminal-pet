//go:build linux || darwin

package ui

// A failed jump on POSIX. The windows-tagged tests cannot see this file's
// subject, and the behaviour it replaces (silence) is exactly what nobody
// notices in a screenshot.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
)

// focusTestApp builds a real App without a window: New wires the approval
// tracker, the receipt and the sound settings, and nothing in these paths
// touches the display, so present() returns early and beep() finds no player on
// a test machine.
func focusTestApp() *App {
	return New(store.New(), Options{Settings: DefaultSettings()})
}

func TestFocusFailedArmsTheJumpPill(t *testing.T) {
	app := focusTestApp()
	target := session.Session{Source: "claude-hook", SessionID: "ses_1", State: session.StateIdle}
	now := time.Now()
	request := focusRequest{handle: 0, row: 0, identity: session.Identity(target), at: now}
	app.focusFailed(request, now)

	if !app.toast.active(now) {
		t.Fatal("a failed jump must say something")
	}
	frame := app.toast.frame(now, []session.Session{target})
	if !frame.visible() {
		t.Fatal("the pill should be on screen over the row it belongs to")
	}
	if frame.kind != toastJumpFailed {
		t.Fatalf("kind = %d, want the jump failure", frame.kind)
	}
}

// A resolution that took longer than the person's patience is not news.
func TestFocusFailedIgnoresAStaleResolution(t *testing.T) {
	app := focusTestApp()
	target := session.Session{Source: "claude-hook", SessionID: "ses_1", State: session.StateIdle}
	now := time.Now()
	request := focusRequest{row: 0, identity: session.Identity(target), at: now.Add(-jumpStaleAfter - time.Second)}
	app.focusFailed(request, now)
	if app.toast.active(now) {
		t.Fatal("a stale failure painted a pill about a tap the user has forgotten")
	}
}

// The pill follows the conversation, not the row index: a jump that failed on a
// session which has since left the list must not be drawn over somebody else.
func TestFocusFailedDiesWithItsSubject(t *testing.T) {
	app := focusTestApp()
	target := session.Session{Source: "claude-hook", SessionID: "ses_1", State: session.StateIdle}
	now := time.Now()
	app.focusFailed(focusRequest{row: 0, identity: session.Identity(target), at: now}, now)
	other := []session.Session{{Source: "opencode", SessionID: "ses_9", State: session.StateIdle}}
	if frame := app.toast.frame(now, other); frame.visible() {
		t.Fatal("the pill outlived the session it was about")
	}
	// ... and it is visible while the subject is still there, so the check above
	// is about the subject leaving rather than about the pill never being drawn.
	if frame := app.toast.frame(now, []session.Session{target}); !frame.visible() {
		t.Fatal("the pill should be visible over its own session")
	}
}
