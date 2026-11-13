package ui

// The row geometry, pinned against the rasterizer's own numbers. These tests
// exist because focus_windows_test.go already forced this pattern for the jump
// arrow: a hit test that re-derives its own arithmetic drifts from the drawing
// one rounding step at a time, and the symptom is "I clicked the row and nothing
// happened" on exactly one zoom level.

import (
	"testing"

	"ai-terminal-pet/internal/session"
)

func TestExpandedRowRectMatchesTheRasterizer(t *testing.T) {
	// The rasterizer draws rows at value(14), width-value(28),
	// top+index*value(expandedRowHeight), value(expandedRowHeight)-value(4).
	// expandedRowRect must produce the same numbers for the widths this panel
	// can actually take (65%..160% of 200 logical units, plus DPI).
	for _, width := range []int{168, 200, 300, 320, 424} {
		unit := float64(width) / expandedWidth
		value := func(base int) int {
			if scaled := int(float64(base)*unit + 0.5); scaled > 1 {
				return scaled
			}
			return 1
		}
		rowTop := value(expandedHeaderHeight + expandedSummaryHeight)
		for index := 0; index < expandedMaxRows; index++ {
			left, top, w, h := expandedRowRect(width, index)
			if want := value(14); left != want {
				t.Fatalf("width %d row %d: left = %d, want %d", width, index, left, want)
			}
			if want := width - value(28); w != want {
				t.Fatalf("width %d row %d: width = %d, want %d", width, index, w, want)
			}
			if want := rowTop + index*value(expandedRowHeight); top != want {
				t.Fatalf("width %d row %d: top = %d, want %d", width, index, top, want)
			}
			if want := value(expandedRowHeight) - value(4); h != want {
				t.Fatalf("width %d row %d: height = %d, want %d", width, index, h, want)
			}
		}
	}
}

func TestExpandedRowHitFindsEveryDrawnRow(t *testing.T) {
	width := expandedWidth
	height := expandedPanelHeight(expandedMaxRows)
	for index := 0; index < expandedMaxRows; index++ {
		left, top, w, h := expandedRowRect(width, index)
		if _, ok := expandedRowHit(left+w/2, top+h/2, width, height, expandedMaxRows); !ok {
			t.Fatalf("the centre of row %d is not a hit", index)
		}
		if got, ok := expandedRowHit(left+1, top+1, width, height, expandedMaxRows); !ok || got != index {
			t.Fatalf("the top-left corner of row %d hit %d/%v", index, got, ok)
		}
	}
}

func TestExpandedRowHitRejectsEverythingElse(t *testing.T) {
	width := expandedWidth
	height := expandedPanelHeight(3)
	rows := 3
	// Header, summary line, the gap between rows, the footer, and one pixel
	// above the first row: all of these keep their existing meaning (they are
	// the click-to-collapse surface), so none of them may be a row hit.
	cases := []struct{ x, y int }{
		{width / 2, 5},                        // header
		{width / 2, expandedHeaderHeight + 1}, // summary
		{width / 2, height - 4},               // footer
	}
	for _, tc := range cases {
		if index, ok := expandedRowHit(tc.x, tc.y, width, height, rows); ok {
			t.Fatalf("(%d,%d) hit row %d but is not a row", tc.x, tc.y, index)
		}
	}
	// One pixel above the first row.
	_, firstTop, _, _ := expandedRowRect(width, 0)
	if _, ok := expandedRowHit(width/2, firstTop-1, width, height, rows); ok {
		t.Fatal("one pixel above the first row must not be a hit")
	}
	// The gap between two rows belongs to neither: rows are drawn
	// expandedRowHeight tall minus scale(4), so the 4-unit gap is a gap.
	_, top, _, h := expandedRowRect(width, 0)
	if _, ok := expandedRowHit(width/2, top+h+1, width, height, rows); ok {
		t.Fatal("the gap between rows must not belong to a row")
	}
	// Past the last drawn row.
	if _, ok := expandedRowHit(width/2, height, width, height, rows); ok {
		t.Fatal("a point below the last row must not be a row hit")
	}
}

func TestExpandedRowHitCapsAtTheDrawnRows(t *testing.T) {
	width := expandedWidth
	height := expandedPanelHeight(expandedMaxRows + 3)
	// The panel says there are more sessions than get drawn; a point where the
	// overflow line lives must not resolve to an undrawn row.
	_, top, _, h := expandedRowRect(width, expandedMaxRows)
	if _, ok := expandedRowHit(width/2, top+h/2, width, height, expandedMaxRows+3); ok {
		t.Fatal("an undrawn row must not be clickable")
	}
}

func TestExpandedRowHitHandlesDegenerateInput(t *testing.T) {
	if _, ok := expandedRowHit(0, 0, 0, 0, 3); ok {
		t.Fatal("a zero-sized panel cannot have rows")
	}
	if _, ok := expandedRowHit(10, 10, expandedWidth, 400, 0); ok {
		t.Fatal("no sessions means no rows")
	}
}

// A row is only a copy target when it actually knows where its session lives:
// reporters may send no working directory at all, and a receipt over an empty
// clipboard is worse than a tap that does nothing.
func TestRowCopyTextNeedsAProjectPath(t *testing.T) {
	cases := []struct {
		name  string
		items []session.Session
		index int
		want  string
		ok    bool
	}{
		{"full path", []session.Session{{Project: `C:\Users\me\work\app`}}, 0, `C:\Users\me\work\app`, true},
		{"empty project", []session.Session{{Project: ""}}, 0, "", false},
		{"whitespace only", []session.Session{{Project: "   "}}, 0, "", false},
		{"no sessions", nil, 0, "", false},
		{"negative index", []session.Session{{Project: `C:\x`}}, -1, "", false},
		{"past the end", []session.Session{{Project: `C:\x`}}, 1, "", false},
		// Past expandedMaxRows nothing is drawn, so there is no row to tap.
		{"not drawn", make([]session.Session, expandedMaxRows+1), expandedMaxRows, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := rowCopyText(tc.items, tc.index)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("rowCopyText = (%q,%v), want (%q,%v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

// The clipboard gets the full path, not the label the row prints: filepath.Base
// is a 200-unit-wide compromise, and the path is what gets pasted.
func TestRowCopyTextUsesTheWholePathNotTheLabel(t *testing.T) {
	value := session.Session{Project: `C:\Users\me\work\ai-terminal-pet`}
	got, ok := rowCopyText([]session.Session{value}, 0)
	if !ok || got != value.Project {
		t.Fatalf("rowCopyText = (%q,%v), want the whole path", got, ok)
	}
}
