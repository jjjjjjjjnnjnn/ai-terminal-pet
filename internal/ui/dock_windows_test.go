package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// TestDockRoundTripIsExact is the regression for the drift the panel used to
// show: expanding while docked to the right edge and collapsing again landed
// well to the left of where it started.
//
// Two separate mistakes produced it. The old code preserved the window centre,
// so the collapse measured from a position the clamp had already pushed inward.
// Anchoring on the frame's own edge was not enough either, because the clamp
// still moves a wide frame. The fix is that a docked frame is a pure function
// of the edge and a remembered gap - so however many times its width changes,
// and in whatever order, it comes back to the same place.
func TestDockRoundTripIsExact(t *testing.T) {
	const workRight = int32(2560)
	const compactW = 84

	// Every gap the panel can settle at, including flush (0) and the tightest
	// margin clampHome allows.
	for _, gap := range []int32{0, 12, 36, 60, 120} {
		start := dockedOrigin(dockRight, workRight, gap, compactW)
		// Expand, grow taller, expand again, collapse.
		for _, width := range []int{300, 200, 300, compactW} {
			got := dockedOrigin(dockRight, workRight, gap, width)
			if width == compactW && got != start {
				t.Fatalf("gap %d: round trip ended at %d, want %d", gap, got, start)
			}
		}
		// The mirror image on the left edge, where the panel grows the other
		// way and the gap is measured from the left.
		if got := dockedOrigin(dockLeft, 0, gap, compactW); got != gap {
			t.Fatalf("left dock gap %d = %d", gap, got)
		}
	}
}

// TestDockedOriginKeepsTheGap pins the rule itself: a docked frame is placed by
// how far it sits from the edge, never by where it last happened to be.
func TestDockedOriginKeepsTheGap(t *testing.T) {
	if got := dockedOrigin(dockRight, 2560, 24, 84); got != 2452 {
		t.Fatalf("right dock = %d, want 2452", got)
	}
	if got := dockedOrigin(dockLeft, 0, 24, 84); got != 24 {
		t.Fatalf("left dock = %d, want 24", got)
	}
	if got := dockedOrigin(dockNone, 2560, 24, 84); got != 0 {
		t.Fatalf("undocked frame must not be placed by the dock maths, got %d", got)
	}
}

// TestTuckedOriginLeavesExactlyTheSliver checks the arithmetic that decides how
// much of the panel is still on screen once it has slid into an edge.
func TestTuckedOriginLeavesExactlyTheSliver(t *testing.T) {
	const width, sliver = 84, 15

	left := tuckedOrigin(dockRight, 2560, width, sliver)
	if visible := 2560 - left; visible != sliver {
		t.Fatalf("right dock: %d units visible, want %d", visible, sliver)
	}
	left = tuckedOrigin(dockLeft, 0, width, sliver)
	if visible := left + width; visible != sliver {
		t.Fatalf("left dock: %d units visible, want %d", visible, sliver)
	}
	if got := tuckedOrigin(dockNone, 2560, width, sliver); got != 0 {
		t.Fatalf("undocked frame has nowhere to tuck to, got %d", got)
	}
}

func TestSlideBetween(t *testing.T) {
	if got := slideBetween(2440, 2545, 0); got != 2440 {
		t.Fatalf("tuck 0 = %d, want the untucked origin", got)
	}
	if got := slideBetween(2440, 2545, 1); got != 2545 {
		t.Fatalf("tuck 1 = %d, want the tucked origin", got)
	}
	if got := slideBetween(2440, 2544, 0.5); got != 2492 {
		t.Fatalf("tuck 0.5 = %d, want the midpoint", got)
	}
}

// TestSliverWidthSurvivesLowZoom pins the reason sliverWidth does not go
// through scaledDimension: that helper floors every dimension at 32 units,
// which would make the tucked strip wider than the panel it came from once the
// user zooms out.
func TestSliverWidthSurvivesLowZoom(t *testing.T) {
	app := &App{dpi: 96, zoom: 0.65}
	if got := app.sliverWidth(); got >= scaledDimension(compactWidth, 96, 0.65) {
		t.Fatalf("sliver %d must stay narrower than the panel", got)
	}
	if got := app.sliverWidth(); got < 6 {
		t.Fatalf("sliver %d is too thin to aim at", got)
	}
}

// TestSlideTuckIsTimeBased pins the reason the slide is driven by elapsed time
// rather than by a fraction of the remaining distance per tick: a busy desktop
// drops frames, and a fraction-based slide would then take proportionally
// longer - slowing down exactly when it should not.
func TestSlideTuckIsTimeBased(t *testing.T) {
	if got := slideTuck(1, 0, 0); got != 1 {
		t.Fatalf("at 0: %v, want the start value", got)
	}
	if got := slideTuck(1, 0, slideOutDuration); got != 0 {
		t.Fatalf("at the reveal duration: %v, want fully out", got)
	}
	if got := slideTuck(0, 1, slideInDuration); got != 1 {
		t.Fatalf("at the hide duration: %v, want fully tucked", got)
	}
	// Overshooting the duration must clamp, not overshoot.
	if got := slideTuck(0, 1, slideInDuration*10); got != 1 {
		t.Fatalf("past the duration: %v", got)
	}
	// Halfway must be more than halfway there (ease out), and monotonic.
	previous := 1.0
	for step := 0; step <= 10; step++ {
		got := slideTuck(1, 0, slideOutDuration*time.Duration(step)/10)
		if got > previous {
			t.Fatalf("slide is not monotonic at %d%%: %v after %v", step*10, got, previous)
		}
		previous = got
	}
	if mid := slideTuck(1, 0, slideOutDuration/2); mid > 0.5 {
		t.Fatalf("ease out should pass the midpoint early, got %v", mid)
	}
	// Revealing is quicker than hiding.
	if slideOutDuration >= slideInDuration {
		t.Fatal("a hover reveal must not be slower than settling away")
	}
}

// TestFrameKeySkipsUnchangedFrames is the redraw optimisation: a reporter that
// pushes the same state repeatedly must not cost a full GDI raster each time.
func TestFrameKeySkipsUnchangedFrames(t *testing.T) {
	base := frameKey(84, 234, 7, session.StateWorking, 2, false, false, false, 0, false, time.Unix(1000, 0))
	same := frameKey(84, 234, 7, session.StateWorking, 2, false, false, false, 0, false, time.Unix(1010, 0))
	if base != same {
		t.Fatal("an idle compact panel must produce the same key ten seconds later")
	}
	// Every input that is actually drawn must move the key.
	cases := map[string]renderedKey{
		"width":     frameKey(90, 234, 7, session.StateWorking, 2, false, false, false, 0, false, time.Unix(1000, 0)),
		"height":    frameKey(84, 300, 7, session.StateWorking, 2, false, false, false, 0, false, time.Unix(1000, 0)),
		"digest":    frameKey(84, 234, 8, session.StateWorking, 2, false, false, false, 0, false, time.Unix(1000, 0)),
		"state":     frameKey(84, 234, 7, session.StateError, 2, false, false, false, 0, false, time.Unix(1000, 0)),
		"count":     frameKey(84, 234, 7, session.StateWorking, 3, false, false, false, 0, false, time.Unix(1000, 0)),
		"expanded":  frameKey(84, 234, 7, session.StateWorking, 2, true, false, false, 0, false, time.Unix(1000, 0)),
		"pressed":   frameKey(84, 234, 7, session.StateWorking, 2, false, true, false, 0, false, time.Unix(1000, 0)),
		"hovered":   frameKey(84, 234, 7, session.StateWorking, 2, false, false, true, 0, false, time.Unix(1000, 0)),
		"tuck":      frameKey(84, 234, 7, session.StateWorking, 2, false, false, false, 0.5, false, time.Unix(1000, 0)),
		"animating": frameKey(84, 234, 7, session.StateWorking, 2, false, false, false, 0, true, time.Unix(1000, 0)),
	}
	for name, key := range cases {
		if key == base {
			t.Fatalf("%s must change the frame key", name)
		}
	}
	// While animating, every frame differs, because the pulse is clock-driven.
	first := frameKey(84, 234, 7, session.StateWaitingPermission, 2, false, false, false, 0, true, time.Unix(1000, 0))
	second := frameKey(84, 234, 7, session.StateWaitingPermission, 2, false, false, false, 0, true, time.Unix(1000, 1))
	if first == second {
		t.Fatal("an animating frame must not be skipped")
	}
}
