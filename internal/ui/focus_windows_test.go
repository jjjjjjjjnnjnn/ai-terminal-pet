package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestCompactFocusRectStaysInsidePanel(t *testing.T) {
	for _, width := range []int{compactWidth, 84, 120, 300} {
		left, top, w, h := compactFocusRect(width)
		if left < 0 || top < 0 || w <= 0 || h <= 0 {
			t.Fatalf("width %d: invalid rect (%d,%d,%d,%d)", width, left, top, w, h)
		}
		if left+w > width {
			t.Fatalf("width %d: jump bar overflows the canvas (right=%d)", width, left+w)
		}
		if top+h > focusScaled(compactHeight, width, compactWidth) {
			t.Fatalf("width %d: jump bar overflows the panel height (bottom=%d)", width, top+h)
		}
	}
}

func TestFocusHitCompactJumpBar(t *testing.T) {
	width := compactWidth
	height := scaledDimension(compactHeight, 96, 1)
	left, top, w, h := compactFocusRect(width)
	x, y := left+w/2, top+h/2

	if row, ok := focusHit(x, y, width, height, false, true, false, 3); !ok || row != -1 {
		t.Fatalf("centre of jump bar = (%d,%v), want (-1,true)", row, ok)
	}
	if _, ok := focusHit(x, y, width, height, false, false, false, 3); ok {
		t.Fatal("a panel with no jumpable session must not have a live jump bar")
	}
	if _, ok := focusHit(0, 0, width, height, false, true, false, 3); ok {
		t.Fatal("a tap in the top-left corner must not trigger a jump")
	}
	if _, ok := focusHit(-1, -1, 0, 0, false, true, false, 3); ok {
		t.Fatal("a zero-sized frame must not trigger a jump")
	}
}

func TestFocusHitExpandedRowsAreDistinct(t *testing.T) {
	width := expandedWidth
	height := expandedPanelHeight(3)
	rows := 3
	for index := 0; index < rows; index++ {
		left, top, size := expandedRowFocusRect(width, expandedFocusRowTop(width, index), expandedFocusRowHeight(width))
		if row, ok := focusHit(left+size/2, top+size/2, width, height, true, true, false, rows); !ok || row != index {
			t.Fatalf("row %d centre = (%d,%v), want (%d,true)", index, row, ok, index)
		}
	}
	if _, ok := focusHit(width/2, height/2, width, height, true, true, false, rows); ok {
		t.Fatal("the middle of a row is not a jump button (it must still toggle/drag)")
	}
	// Only the first expandedMaxRows are painted; a tap below them must not
	// address an invisible row.
	_, belowTop, belowSize := expandedRowFocusRect(width, expandedFocusRowTop(width, expandedMaxRows), expandedFocusRowHeight(width))
	if _, ok := focusHit(expandedWidth-20, belowTop+belowSize/2, width, height+400, true, true, false, 20); ok {
		t.Fatal("a click past the visible row limit must not jump to a hidden row")
	}
}

func TestJumpPriorityExcludesBusyStates(t *testing.T) {
	cases := []struct {
		state session.State
		want  int
	}{
		{session.StateWaitingPermission, 4},
		{session.StateError, 3},
		{session.StateIdle, 2},
		{session.StateClosed, 1},
		{session.StateWorking, -1},
		{session.StateToolUse, -1},
		{session.StateThinking, -1},
	}
	for _, test := range cases {
		if got := jumpPriority(test.state); got != test.want {
			t.Errorf("jumpPriority(%q) = %d, want %d", test.state, got, test.want)
		}
	}
}

func TestMostUrgentJumpPrefersDecisionThenRecency(t *testing.T) {
	now := time.Unix(1000, 0)
	items := []session.Session{
		{SessionID: "busy", State: session.StateWorking, UpdatedAt: now},
		{SessionID: "idle-old", State: session.StateIdle, UpdatedAt: now.Add(-30 * time.Second)},
		{SessionID: "idle-new", State: session.StateIdle, UpdatedAt: now},
		{SessionID: "error", State: session.StateError, UpdatedAt: now.Add(-time.Minute)},
		{SessionID: "waiting", State: session.StateWaitingPermission, UpdatedAt: now.Add(-2 * time.Minute)},
	}
	got, ok := mostUrgentJump(items)
	if !ok || got.SessionID != "waiting" {
		t.Fatalf("most urgent = %q (%v), want waiting", got.SessionID, ok)
	}

	withoutDecision := items[:4]
	got, ok = mostUrgentJump(withoutDecision)
	if !ok || got.SessionID != "error" {
		t.Fatalf("after removing the permission turn = %q, want error", got.SessionID)
	}

	tie := []session.Session{
		{SessionID: "idle-old", State: session.StateIdle, UpdatedAt: now.Add(-time.Minute)},
		{SessionID: "idle-new", State: session.StateIdle, UpdatedAt: now},
	}
	got, ok = mostUrgentJump(tie)
	if !ok || got.SessionID != "idle-new" {
		t.Fatalf("tie should go to the newer session, got %q", got.SessionID)
	}

	if _, ok := mostUrgentJump([]session.Session{{State: session.StateWorking}}); ok {
		t.Fatal("a busy-only panel has no jump target")
	}
}

func TestExpandedFocusButtonSitsInsideItsRow(t *testing.T) {
	width := expandedWidth
	rowTop := expandedFocusRowTop(width, 0)
	rowHeight := expandedFocusRowHeight(width)
	left, top, size := expandedRowFocusRect(width, rowTop, rowHeight)
	if left < 0 || left+size > width {
		t.Fatalf("button escapes the canvas: left=%d size=%d width=%d", left, size, width)
	}
	if top < rowTop || top+size > rowTop+rowHeight {
		t.Fatalf("button escapes its row: top=%d size=%d row=(%d,%d)", top, size, rowTop, rowTop+rowHeight)
	}
}
