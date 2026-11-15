package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

func TestKeyboardNavSelectionWrapAndEnter(t *testing.T) {
	now := time.Now()
	items := []session.Session{
		{Source: "claude-hook", SessionID: "s1", State: session.StateWorking, UpdatedAt: now},
		{Source: "opencode", SessionID: "s2", State: session.StateWaitingPermission, UpdatedAt: now},
		{Source: "antigravity", SessionID: "s3", State: session.StateIdle, UpdatedAt: now},
	}

	var nav keyboardNav
	if _, ok := nav.selected(); ok {
		t.Fatal("fresh keyboardNav should have no active selection")
	}

	// Down while expanded selects row 0, then advances and wraps.
	out := nav.handle(panelKeyDown, true, items)
	if !out.repainted {
		t.Fatal("first Down should repaint")
	}
	if row, ok := nav.selected(); !ok || row != 0 {
		t.Fatalf("after first Down: got (%d, %v), want (0, true)", row, ok)
	}

	nav.handle(panelKeyDown, true, items)
	if row, ok := nav.selected(); !ok || row != 1 {
		t.Fatalf("after second Down: got (%d, %v), want (1, true)", row, ok)
	}

	nav.handle(panelKeyDown, true, items)
	nav.handle(panelKeyDown, true, items)
	if row, ok := nav.selected(); !ok || row != 0 {
		t.Fatalf("Down should wrap to 0, got (%d, %v)", row, ok)
	}

	// Up from 0 wraps to last visible row.
	nav.handle(panelKeyUp, true, items)
	if row, ok := nav.selected(); !ok || row != 2 {
		t.Fatalf("Up from 0 should wrap to 2, got (%d, %v)", row, ok)
	}

	// Home / End move to bounds.
	nav.handle(panelKeyHome, true, items)
	if row, ok := nav.selected(); !ok || row != 0 {
		t.Fatalf("Home: got (%d, %v), want (0, true)", row, ok)
	}
	nav.handle(panelKeyEnd, true, items)
	if row, ok := nav.selected(); !ok || row != 2 {
		t.Fatalf("End: got (%d, %v), want (2, true)", row, ok)
	}

	// Enter jumps to the currently selected row.
	enter := nav.handle(panelKeyEnter, true, items)
	if !enter.jump || enter.jumpRow != 2 {
		t.Fatalf("Enter on row 2: got %+v, want jumpRow=2", enter)
	}

	// Escape collapses and clears selection.
	esc := nav.handle(panelKeyEscape, true, items)
	if !esc.toggleExpanded {
		t.Fatal("Escape while expanded should toggleExpanded")
	}
	if _, ok := nav.selected(); ok {
		t.Fatal("Escape should clear selection")
	}
}

func TestKeyboardNavTracksIdentityAcrossReorders(t *testing.T) {
	now := time.Now()
	first := []session.Session{
		{Source: "claude-hook", SessionID: "s1", State: session.StateWorking, UpdatedAt: now},
		{Source: "opencode", SessionID: "s2", State: session.StateWaitingPermission, UpdatedAt: now},
	}
	var nav keyboardNav
	nav.handle(panelKeyDown, true, first)
	nav.handle(panelKeyDown, true, first)
	if row, ok := nav.selected(); !ok || row != 1 {
		t.Fatalf("expected row 1, got (%d, %v)", row, ok)
	}

	// Reorder so s2 moves to index 0.
	reordered := []session.Session{first[1], first[0]}
	if !nav.sync(true, reordered) {
		t.Fatal("sync should report change when selected session moves index")
	}
	if row, ok := nav.selected(); !ok || row != 0 {
		t.Fatalf("selection should follow s2 to index 0, got (%d, %v)", row, ok)
	}

	// If s2 disappears, selection clears rather than jumping to s1 by mistake.
	if !nav.sync(true, []session.Session{first[0]}) {
		t.Fatal("sync should report change when selected session disappears")
	}
	if _, ok := nav.selected(); ok {
		t.Fatal("selection should clear when selected session leaves visible list")
	}
}

func TestKeyboardNavCompactActions(t *testing.T) {
	now := time.Now()
	items := []session.Session{
		{Source: "claude-hook", SessionID: "s1", State: session.StateWaitingPermission, UpdatedAt: now},
	}
	var nav keyboardNav
	// Enter when compact jumps to most urgent session (-1).
	out := nav.handle(panelKeyEnter, false, items)
	if !out.jump || out.jumpRow != -1 {
		t.Fatalf("compact Enter: got %+v, want jumpRow=-1", out)
	}
	// Escape when compact is a no-op.
	esc := nav.handle(panelKeyEscape, false, items)
	if esc.toggleExpanded || esc.jump || esc.repainted {
		t.Fatalf("compact Escape should be a no-op, got %+v", esc)
	}
	// Down when compact expands and selects row 0.
	down := nav.handle(panelKeyDown, false, items)
	if !down.toggleExpanded {
		t.Fatal("compact Down with sessions should expand")
	}
	if row, ok := nav.selected(); !ok || row != 0 {
		t.Fatalf("compact Down should select row 0, got (%d, %v)", row, ok)
	}
}

func TestRenderedKeyDistinguishesZeroRowFromNoSelection(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	none := frameKey(expandedWidth, 240, 1, session.StateIdle, 1, true, false, false, 0, false, now, i18n.LangEn, -1, themeSlate, false, 0, 0, -1)
	rowZero := none
	rowZero.withSelectedRow(0, true)
	rowOne := none
	rowOne.withSelectedRow(1, true)
	if none == rowZero {
		t.Fatal("no selection and row 0 must produce distinct frame keys")
	}
	if rowZero == rowOne {
		t.Fatal("row 0 and row 1 must produce distinct frame keys")
	}
}
