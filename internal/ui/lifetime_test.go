package ui

import (
	"testing"
	"time"
)

// The idle self-destruct, pinned as a policy rather than as a timer handler.
//
// Each case below is a user-visible behaviour that has either been wrong or
// would have been invisible in a test of the window procedure:
//
//   - sessions keep it alive (the rule it was written for);
//   - a visible settings window keeps it alive too, because the settings
//     window dies with the process (HANDOVER §51.3);
//   - closing the settings window does not pay an accumulated debt at once
//     (hideSettings restarts the anchor - pinned here as the policy half);
//   - an anchor in the future, or a zero anchor, never quits.
func TestIdleDecisionKeepsTheProcessForWork(t *testing.T) {
	start := time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)
	grace := 10 * time.Minute

	cases := []struct {
		name     string
		offset   time.Duration
		sessions bool
		settings bool
		wantQuit bool
	}{
		{name: "sessions, a moment in", offset: time.Minute, sessions: true},
		{name: "sessions, long past the grace", offset: time.Hour, sessions: true},
		{name: "empty, a moment in", offset: time.Minute},
		{name: "empty, just under the grace", offset: grace - time.Second},
		{name: "empty, at the grace", offset: grace, wantQuit: true},
		{name: "empty, well past the grace", offset: 2 * grace, wantQuit: true},
		{name: "settings open, a moment in", offset: time.Minute, settings: true},
		{name: "settings open, long past the grace", offset: time.Hour, settings: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			quit, anchor := idleDecision(start.Add(testCase.offset), start, testCase.sessions, testCase.settings, grace)
			if quit != testCase.wantQuit {
				t.Fatalf("quit = %v, want %v", quit, testCase.wantQuit)
			}
			// The anchor moves to "now" whenever the process is being kept
			// alive, and stays put when nothing is owed either way. A quit
			// leaves it alone; retrying that decision is the next caller's
			// problem, not a reason to rewrite history.
			if testCase.sessions || testCase.settings {
				if !anchor.Equal(start.Add(testCase.offset)) {
					t.Fatalf("anchor = %v, want now", anchor)
				}
			} else if !anchor.Equal(start) {
				t.Fatalf("anchor moved to %v, want %v", anchor, start)
			}
		})
	}
}

func TestIdleDecisionWithoutAnAnchorNeverQuits(t *testing.T) {
	// lastAvailable is seeded at New(), but a zero value must still be safe:
	// quitting on an uninitialised clock would be a random exit.
	now := time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)
	for _, settings := range []bool{false, true} {
		quit, anchor := idleDecision(now, time.Time{}, false, settings, time.Minute)
		if quit {
			t.Fatalf("a zero anchor quit (settings=%v)", settings)
		}
		if !anchor.Equal(now) {
			t.Fatalf("a zero anchor was not seeded: %v", anchor)
		}
	}
}

func TestIdleDecisionRejectsANonPositiveGrace(t *testing.T) {
	// A zero grace would mean "quit immediately", and the timer fires every
	// 30 s: that is the five-second version of this policy that made the panel
	// blink out while a terminal was still working.
	now := time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)
	quit, _ := idleDecision(now, now.Add(-time.Second), false, false, 0)
	if quit {
		t.Fatal("a non-positive grace quit on an empty overlay")
	}
}

func TestIdleDecisionTreatsAFutureAnchorAsFresh(t *testing.T) {
	now := time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)
	quit, anchor := idleDecision(now, now.Add(time.Minute), false, false, time.Minute)
	if quit {
		t.Fatal("an anchor in the future was treated as overdue")
	}
	if !anchor.Equal(now.Add(time.Minute)) {
		t.Fatalf("a future anchor was rewritten to %v", anchor)
	}
}

func TestShippedIdleGraceIsTenMinutes(t *testing.T) {
	// The number is documented (README's lifecycle line and --doctor's runtime
	// check). A shipping build must not honour an environment override, so this
	// pins the default: exit_trace_debug.go is the only place allowed to change
	// it, and that file is behind the petdebug build tag.
	if idleGrace != 10*time.Minute {
		t.Fatalf("shipped idle grace = %v, want 10m", idleGrace)
	}
}
