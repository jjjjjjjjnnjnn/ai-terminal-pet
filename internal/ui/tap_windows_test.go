package ui

// decideTap, pinned. These are the interaction rules a user feels, so every
// ordering claim below is asserted directly rather than inferred from a
// screenshot: a wrong order here is a tap that copies a path while a silent
// auto-approve is about to run the command.

import "testing"

func TestDecideTapPrefersTheJumpControl(t *testing.T) {
	// An aimed arrow wins over everything, including the veto: it is an
	// explicit control, not a tap on empty panel.
	for _, in := range []tapInputs{
		{expanded: true, jumpHit: true, jumpRow: 2, jumpValid: true, vetoArmed: true, copyRowHit: true, copyRow: 2},
		{expanded: false, jumpHit: true, jumpRow: -1, jumpValid: true, copyStripHit: true, copyStripValid: true},
	} {
		if action, row := decideTap(in); action != tapJump || row != in.jumpRow {
			t.Fatalf("decideTap(%+v) = (%v,%d), want a jump to row %d", in, action, row, in.jumpRow)
		}
	}
}

// A tap that hit nothing jumpable must not jump: the strip's dead half and a
// row's non-arrow area are both "somewhere else".
func TestDecideTapWithoutAValidJumpFallsThrough(t *testing.T) {
	if action, _ := decideTap(tapInputs{jumpHit: true, jumpRow: -1, jumpValid: false}); action != tapCollapse {
		t.Fatalf("an invalid jump target gave %v, want a collapse", action)
	}
}

// The strip's copy half is its own target and is checked before the veto,
// which only applies to the expanded panel anyway.
func TestDecideTapCopiesTheStripHalf(t *testing.T) {
	action, _ := decideTap(tapInputs{copyStripHit: true, copyStripValid: true})
	if action != tapCopyStrip {
		t.Fatalf("the strip's copy half gave %v", action)
	}
	if action, _ := decideTap(tapInputs{copyStripHit: true, copyStripValid: false}); action != tapCollapse {
		t.Fatalf("a dead copy half gave %v, want a collapse", action)
	}
}

// The veto outranks a row's copy target while a countdown is running: with
// auto-approve armed, "don't approve" is the most expensive thing a tap can
// mean, and answering it with a clipboard write is a silent approval.
func TestDecideTapPrefersTheVetoOverARowCopy(t *testing.T) {
	action, _ := decideTap(tapInputs{expanded: true, vetoArmed: true, copyRowHit: true, copyRow: 1})
	if action != tapVeto {
		t.Fatalf("an armed countdown gave %v, want a veto", action)
	}
	// With the countdown gone, the same tap copies again.
	action, row := decideTap(tapInputs{expanded: true, copyRowHit: true, copyRow: 1})
	if action != tapCopyRow || row != 1 {
		t.Fatalf("without a countdown the row gave (%v,%d), want a copy of row 1", action, row)
	}
}

// A row that is not a copy target keeps the old meaning: a tap anywhere else
// collapses. Otherwise a session that reported no working directory would
// become a spot where clicking does nothing at all.
func TestDecideTapFallsBackToCollapse(t *testing.T) {
	if action, _ := decideTap(tapInputs{expanded: true}); action != tapCollapse {
		t.Fatalf("an empty tap gave %v, want a collapse", action)
	}
	// The veto is an expanded-panel rule only: the compact strip still toggles
	// while a countdown runs, so the pending request stays one tap away.
	if action, _ := decideTap(tapInputs{vetoArmed: true}); action != tapCollapse {
		t.Fatalf("a compact tap during a countdown gave %v, want a collapse", action)
	}
}
