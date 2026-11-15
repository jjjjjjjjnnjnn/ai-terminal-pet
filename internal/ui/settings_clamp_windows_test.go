package ui

// The third candidate from HANDOVER §51.3, as a table: "the window ended up
// outside the work area, so it looked like it had gone away".
//
// Work areas are not all positive. A monitor to the left of the primary has
// negative coordinates, a 4K laptop panel and a 1080p external display produce
// two different work areas on the same desktop, and the settings window can be
// taller than the short one. Each case below is one of those, checked as a
// pure function so no window has to exist to prove it.

import (
	"testing"

	win "github.com/lxn/win"
)

func TestClampOriginKeepsTheWindowOnTheWorkArea(t *testing.T) {
	margin := int32(8)
	cases := []struct {
		name   string
		work   win.RECT
		width  int32
		height int32
		x, y   int32
		wantX  int32
		wantY  int32
	}{
		{
			name:  "already inside stays put",
			work:  win.RECT{Left: 0, Top: 0, Right: 2560, Bottom: 1528},
			width: 938, height: 1198, x: 811, y: 165,
			wantX: 811, wantY: 165,
		},
		{
			name:  "origin off the left edge is pinned to the margin",
			work:  win.RECT{Left: 0, Top: 0, Right: 2560, Bottom: 1528},
			width: 938, height: 1198, x: -400, y: 165,
			wantX: 8, wantY: 165,
		},
		{
			name:  "origin past the bottom is pulled back",
			work:  win.RECT{Left: 0, Top: 0, Right: 2560, Bottom: 1528},
			width: 938, height: 1198, x: 811, y: 900,
			// 1528 - 1198 - 8
			wantX: 811, wantY: 322,
		},
		{
			name:  "a monitor with negative coordinates",
			work:  win.RECT{Left: -1920, Top: 0, Right: 0, Bottom: 1080},
			width: 750, height: 826, x: -3000, y: -200,
			wantX: -1912, wantY: 8,
		},
		{
			name:  "a short work area (taskbar on the left, 1366x768)",
			work:  win.RECT{Left: 0, Top: 0, Right: 1366, Bottom: 728},
			width: 938, height: 1198, x: 214, y: -40,
			// taller than the work area: the top margin wins, not a negative range
			wantX: 214, wantY: 8,
		},
		{
			name:  "wider than the work area",
			work:  win.RECT{Left: 0, Top: 0, Right: 800, Bottom: 600},
			width: 938, height: 1198, x: 400, y: 300,
			wantX: 8, wantY: 8,
		},
		{
			name:  "exactly at the far corner is not past it",
			work:  win.RECT{Left: 0, Top: 0, Right: 2560, Bottom: 1528},
			width: 500, height: 500, x: 2560 - 500 - 8, y: 1528 - 500 - 8,
			wantX: 2560 - 500 - 8, wantY: 1528 - 500 - 8,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotX, gotY := clampOriginToWorkArea(tc.work, margin, tc.width, tc.height, tc.x, tc.y)
			if gotX != tc.wantX || gotY != tc.wantY {
				t.Fatalf("clamp = (%d,%d), want (%d,%d)", gotX, gotY, tc.wantX, tc.wantY)
			}
			// The point of the clamp: the frame's top-left and its bottom-right
			// are both inside the work area, whatever the inputs were.
			if gotX < tc.work.Left || gotY < tc.work.Top {
				t.Fatalf("origin (%d,%d) is above/left of the work area", gotX, gotY)
			}
			if gotX+tc.width > tc.work.Right && tc.width <= tc.work.Right-tc.work.Left-margin*2 {
				t.Fatalf("origin x %d puts a %d wide window past %d", gotX, tc.width, tc.work.Right)
			}
			if gotY+tc.height > tc.work.Bottom && tc.height <= tc.work.Bottom-tc.work.Top-margin*2 {
				t.Fatalf("origin y %d puts a %d tall window past %d", gotY, tc.height, tc.work.Bottom)
			}
		})
	}
}
