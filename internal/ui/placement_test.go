package ui

import "testing"

func TestResolveInitialPlacementRestoresPerDisplayLayout(t *testing.T) {
	single := []displayRect{{Left: 0, Top: 0, Right: 1920, Bottom: 1040, Primary: true}}
	dual := []displayRect{
		{Left: 0, Top: 0, Right: 1920, Bottom: 1040, Primary: true},
		{Left: 1920, Top: 0, Right: 4480, Bottom: 1400, Primary: false},
	}

	singleSig := layoutSignature(single)
	dualSig := layoutSignature(dual)
	if singleSig == "" || dualSig == "" || singleSig == dualSig {
		t.Fatalf("signatures should be distinct and non-empty: single=%q dual=%q", singleSig, dualSig)
	}

	// Simulate user placing the overlay on Monitor 2 while dual-monitor is plugged,
	// then unplugging Monitor 2 (clamping onto Monitor 1), then plugging Monitor 2 back in.
	layouts := cloneLayoutsWithCurrent(nil, dualSig, 4390, 400)
	layouts = cloneLayoutsWithCurrent(layouts, singleSig, 1850, 500)

	initial := Placement{
		Placed:  true,
		X:       1850,
		Y:       500,
		Zoom:    1,
		Layouts: layouts,
	}

	// Unplugged: restores single-monitor coordinate.
	x1, y1, ok := resolveInitialPlacement(initial, single, 56, 156, 96)
	if !ok || x1 != 1850 || y1 != 500 {
		t.Fatalf("single display restore: got (%d, %d, %v), want (1850, 500, true)", x1, y1, ok)
	}

	// Replugged: restores dual-monitor coordinate on Monitor 2 even though
	// top-level X/Y was overwritten by the single-monitor session!
	x2, y2, ok := resolveInitialPlacement(initial, dual, 56, 156, 96)
	if !ok || x2 != 4390 || y2 != 400 {
		t.Fatalf("dual display restore: got (%d, %d, %v), want (4390, 400, true)", x2, y2, ok)
	}
}

func TestResolveInitialPlacementFallsBackToLegacyPlacedAndPrimaryAnchor(t *testing.T) {
	displays := []displayRect{
		{Left: 0, Top: 0, Right: 1920, Bottom: 1040, Primary: true},
	}

	// Legacy state.json without Layouts clamps global X/Y onto screen.
	x, y, ok := resolveInitialPlacement(Placement{Placed: true, X: 5000, Y: 200, Zoom: 1}, displays, 56, 156, 96)
	if !ok || x != 1920-56-8 || y != 200 {
		t.Fatalf("legacy clamp: got (%d, %d, %v), want (%d, 200, true)", x, y, ok, 1920-56-8)
	}

	// Unplaced falls back to right-edge anchor on primary display.
	ax, ay, ok := resolveInitialPlacement(Placement{Placed: false, Zoom: 1}, displays, 56, 156, 96)
	if !ok || ax != 1920-56-24 || ay != (1040-156)/2 {
		t.Fatalf("primary anchor: got (%d, %d, %v)", ax, ay, ok)
	}
}

func TestLayoutSignatureIsOrderIndependent(t *testing.T) {
	a := displayRect{Left: 0, Top: 0, Right: 1920, Bottom: 1040, Primary: true}
	b := displayRect{Left: 1920, Top: 0, Right: 3840, Bottom: 1080}
	if layoutSignature([]displayRect{a, b}) != layoutSignature([]displayRect{b, a}) {
		t.Fatal("layoutSignature must not depend on monitor enumeration order")
	}
}
