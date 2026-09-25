package ui

import "testing"

func TestClassifyHit(t *testing.T) {
	// 4x3 canvas: opaque block at x in [1,3), y in [1,2).
	hit := make([]byte, 4*3)
	hit[1*4+1] = 255
	hit[1*4+2] = 19 // below threshold
	if !classifyHit(1, 1, 4, 3, hit) {
		t.Fatal("opaque pixel must hit")
	}
	if classifyHit(2, 1, 4, 3, hit) {
		t.Fatal("sub-threshold alpha must miss")
	}
	if classifyHit(0, 0, 4, 3, hit) {
		t.Fatal("transparent pixel must miss")
	}
	// Regression: X must be checked against width, not len(hit).
	// Window wider than canvas (zoom-out/DPI-down) used to escape the
	// slice (Y*width+X >= len) and panic the process with no error.log.
	wide, high := 28, 133
	wideHit := make([]byte, wide*high)
	if classifyHit(39, 132, wide, high, wideHit) {
		t.Fatal("out-of-canvas X must miss, not panic")
	}
	if classifyHit(39, 170, wide, high, wideHit) {
		t.Fatal("out-of-canvas corner must miss, not panic")
	}
	if classifyHit(-1, 0, wide, high, wideHit) || classifyHit(0, -1, wide, high, wideHit) {
		t.Fatal("negative coords must miss")
	}
	if classifyHit(0, 0, 0, 0, nil) || classifyHit(0, 0, 4, 3, nil) {
		t.Fatal("empty geometry must miss")
	}
}

// TestExpandedCloseButtonRect pins the shared geometry used by both the
// rasterizer and the hit test. The two used to compute the same numbers
// independently; if they ever drift, the drawn button stops matching its
// click target and collapsing the panel silently breaks.
func TestExpandedCloseButtonRect(t *testing.T) {
	for _, width := range []int{200, 220, 250, 300, 32, 40} {
		left, top, size := expandedCloseButtonRect(width)
		if size < 1 {
			t.Fatalf("width %d: size must be positive, got %d", width, size)
		}
		if left < 0 || left+size > width {
			t.Fatalf("width %d: button %d..%d escapes the panel", width, left, left+size)
		}
		if top < 0 {
			t.Fatalf("width %d: top must be non-negative, got %d", width, top)
		}
		// The centre must hit, and a point one pixel outside every edge
		// must not.
		cx, cy := left+size/2, top+size/2
		if !closeButtonHit(cx, cy, width) {
			t.Fatalf("width %d: centre (%d,%d) must be inside", width, cx, cy)
		}
		for _, probe := range [][2]int{
			{left - 1, cy},
			{left + size, cy},
			{cx, top - 1},
			{cx, top + size},
		} {
			if closeButtonHit(probe[0], probe[1], width) {
				t.Fatalf("width %d: (%d,%d) must be outside %d,%d,%d",
					width, probe[0], probe[1], left, top, size)
			}
		}
	}
}

// TestDecideRelease locks the rule that made dragging feel broken: a press
// that actually travelled must move the window only, never toggle it.
func TestDecideRelease(t *testing.T) {
	cases := []struct {
		name                    string
		active, moved, expanded bool
		x, y, width             int
		want                    releaseAction
	}{
		{name: "drag that moved does nothing", active: true, moved: true, expanded: false, x: 20, y: 95, width: 40, want: releaseNothing},
		{name: "drag that moved, expanded, does nothing", active: true, moved: true, expanded: true, x: 100, y: 125, width: 200, want: releaseNothing},
		{name: "tap compact expands", active: true, moved: false, expanded: false, x: 20, y: 95, width: 40, want: releaseToggle},
		{name: "tap expanded body does nothing", active: true, moved: false, expanded: true, x: 100, y: 125, width: 200, want: releaseNothing},
		{name: "inactive press does nothing", active: false, moved: false, expanded: false, x: 20, y: 95, width: 40, want: releaseNothing},
	}
	for _, tc := range cases {
		got := decideRelease(tc.active, tc.moved, tc.expanded, tc.x, tc.y, tc.width)
		if got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
	// The expanded close button must still collapse.
	left, top, size := expandedCloseButtonRect(200)
	if got := decideRelease(true, false, true, left+size/2, top+size/2, 200); got != releaseToggle {
		t.Fatalf("close button tap must toggle, got %v", got)
	}
}
