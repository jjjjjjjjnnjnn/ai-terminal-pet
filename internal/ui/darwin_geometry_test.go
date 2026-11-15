package ui

import "testing"

func TestCocoaScreenAndOriginRoundTrip(t *testing.T) {
	primary := nsRect{Origin: nsPoint{0, 60}, Size: nsSize{1440, 815}}
	primaryTop := 900.0
	rect := cocoaScreenToDisplayRect(primary, primaryTop)
	want := displayRect{Left: 0, Top: 25, Right: 1440, Bottom: 840}
	if rect != want {
		t.Fatalf("cocoaScreenToDisplayRect = %+v, want %+v", rect, want)
	}

	origin := topLeftToCocoaOrigin(120, 25, 100, primaryTop)
	if origin.X != 120 || origin.Y != 775 {
		t.Fatalf("topLeftToCocoaOrigin = %+v, want {120 775}", origin)
	}

	// Secondary monitor stacked above primary (Cocoa Y=[900..2340]).
	above := nsRect{Origin: nsPoint{0, 900}, Size: nsSize{2560, 1440}}
	aboveRect := cocoaScreenToDisplayRect(above, primaryTop)
	if aboveRect.Top != -1440 || aboveRect.Bottom != 0 {
		t.Fatalf("above screen rect = %+v, want Top=-1440 Bottom=0", aboveRect)
	}
}

func TestCocoaWindowAndScreenPointConversion(t *testing.T) {
	tx, ty := cocoaPointToTopLeft(500.4, 875.0, 900.0)
	if tx != 500 || ty != 25 {
		t.Fatalf("cocoaPointToTopLeft = (%d,%d), want (500,25)", tx, ty)
	}

	cx, cy := cocoaWindowPointToCanvas(42.8, 99.6, 240, 100)
	if cx != 42 || cy != 0 {
		t.Fatalf("top edge canvas point = (%d,%d), want (42,0)", cx, cy)
	}
	cx, cy = cocoaWindowPointToCanvas(42.2, 0.3, 240, 100)
	if cx != 42 || cy != 99 {
		t.Fatalf("bottom edge canvas point = (%d,%d), want (42,99)", cx, cy)
	}
}

func TestBGRAStraightToPremulRGBA(t *testing.T) {
	src := []byte{
		10, 20, 30, 0, // transparent
		10, 20, 30, 255, // opaque BGRA=(10,20,30)
		0, 200, 100, 128, // half alpha: R=100, G=200, B=0, A=128
	}
	got := bgraStraightToPremulRGBA(src)
	if len(got) != len(src) {
		t.Fatalf("len = %d, want %d", len(got), len(src))
	}
	if got[0] != 0 || got[1] != 0 || got[2] != 0 || got[3] != 0 {
		t.Fatalf("transparent pixel = %v", got[0:4])
	}
	if got[4] != 30 || got[5] != 20 || got[6] != 10 || got[7] != 255 {
		t.Fatalf("opaque pixel = %v, want [30 20 10 255]", got[4:8])
	}
	if got[8] != 50 || got[9] != 100 || got[10] != 0 || got[11] != 128 {
		t.Fatalf("half-alpha pixel = %v, want [50 100 0 128]", got[8:12])
	}
}

func TestDarwinKeyCodeAndZoomMapping(t *testing.T) {
	if got := darwinKeyCodeToPanelKey(darwinKeyCodeUpArrow); got != panelKeyUp {
		t.Fatalf("up arrow = %v, want %v", got, panelKeyUp)
	}
	if got := darwinKeyCodeToPanelKey(darwinKeyCodeReturn); got != panelKeyEnter {
		t.Fatalf("return = %v, want %v", got, panelKeyEnter)
	}
	if got := darwinKeyCodeToPanelKey(darwinKeyCodeEscape); got != panelKeyEscape {
		t.Fatalf("escape = %v, want %v", got, panelKeyEscape)
	}
	if got := darwinKeyCodeToPanelKey(0x99); got != panelKeyNone {
		t.Fatalf("unknown key = %v, want panelKeyNone", got)
	}
	if delta, reset, ok := darwinZoomShortcut(darwinKeyCodeEqual, true); !ok || delta <= 0 || reset {
		t.Fatalf("Cmd+= = (%v,%v,%v)", delta, reset, ok)
	}
	if _, _, ok := darwinZoomShortcut(darwinKeyCodeEqual, false); ok {
		t.Fatal("= without Command should not trigger zoom")
	}
}
