package ui

import "time"

// The overlay's own lifetime, as one question with one answer.
//
// It used to be an `if` inside the timer handler: "the store is empty for ten
// minutes, so quit". That is the right rule for the case it was written for -
// nobody has an AI terminal open, so a floating panel is dead weight - but it
// conflated "no sessions" with "nobody is using this process", and the settings
// window is the case where those disagree. Quitting underneath an open settings
// window does not just close a panel: the settings window lives on the same UI
// thread, so it disappears with the process, which is exactly the "the settings
// window flashes and goes away" report in docs/HANDOVER.md §51.3.
//
// The rule is here, as a pure function, for the same reason attention.go is:
// the policy is the part that gets argued about, so it should be testable
// without a window, a store or a clock.
//
// Why the exit paths are traced, and why the shipping build is silent:
//
// Guessing which of the three exits produced that report cost a whole round,
// because none of them wrote anything down - no WER report, no CrashDumps
// entry, no error.log, and zero Application-log events. So they get a trace,
// but only in a build that exists to diagnose, because "no network, nothing
// written unless there is a reason" is a product promise:
//
//	go build -tags petdebug ./cmd/ai-terminal-pet
//
// In a default build traceExit and traceSettings are no-ops
// (exit_trace_release.go). The petdebug build appends one line per event to
// exit.log beside config.json (exit_trace_debug.go), and honours
// PET_IDLE_GRACE_SEC so the idle window can be shortened to seconds without
// editing the constant under test. Both are called from the UI thread only and
// never change the decision being reported: an observer that can alter what it
// observes is worse than none.

// idleGrace is how long the overlay may sit with no sessions before it destroys
// itself and takes the settings window with it. It is a variable so the debug
// build can shorten it for a repro; the release build never assigns it, so the
// shipped ten minutes is the only value a user can get. lifetime_test.go pins
// that default.
var idleGrace = 10 * time.Minute

// idleDecision answers two things at once, because the caller needs both and
// deriving them separately is how the two used to drift:
//
//   - quit reports whether the overlay should destroy itself now.
//   - anchor is where the "last available" grace anchor moves to. It moves
//     forward whenever there is a reason to stay alive, so a quiet overlay
//     cannot accumulate a debt it pays the moment the reason goes away.
//
// hasSessions means the store still holds a live conversation. settingsVisible
// means the settings window is on screen right now - not merely created, and
// not hidden behind a close click.
func idleDecision(now, anchor time.Time, hasSessions, settingsVisible bool, grace time.Duration) (quit bool, nextAnchor time.Time) {
	if grace <= 0 {
		grace = time.Minute
	}
	if hasSessions || settingsVisible {
		return false, now
	}
	if anchor.IsZero() {
		return false, now
	}
	if now.Sub(anchor) >= grace {
		return true, anchor
	}
	return false, anchor
}
