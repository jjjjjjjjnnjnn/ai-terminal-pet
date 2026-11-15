package ui

// What a tap on the panel means, in one pure function.
//
// This used to be an if-chain inside endDrag, which made the one rule that
// users actually feel - "what happens when I click here" - impossible to test
// without a window. Every branch below is an interaction decision the project
// has argued about at least once (the arrow versus the veto, the row versus
// the collapse), so they are pinned here instead of in a handler.

// tapAction is the outcome of one tap.
type tapAction int

const (
	// tapCollapse is the default: the panel toggles compact/expanded.
	tapCollapse tapAction = iota
	// tapJump switches to the terminal owning the session; row -1 is the
	// compact strip.
	tapJump
	// tapVeto denies the pending auto-approve for the armed session.
	tapVeto
	// tapCopyRow copies one expanded row's project path.
	tapCopyRow
	// tapCopyStrip copies the compact strip's session project path.
	tapCopyStrip
)

// tapInputs is everything the decision reads, resolved by the caller from the
// same geometry and state the frame was drawn from.
type tapInputs struct {
	// expanded is which panel the tap landed on.
	expanded bool
	// jumpHit is whether the tap hit a drawn jump control (arrow or strip's
	// jump half), and jumpRow which row (-1 for the strip).
	jumpHit   bool
	jumpRow   int
	jumpValid bool
	// copyStripHit is whether the tap hit the compact strip's copy half.
	copyStripHit bool
	// copyStripValid says the strip currently has something to copy.
	copyStripValid bool
	// vetoArmed says an auto-approve countdown is actually running; copyRowHit
	// and copyRow together describe an expanded row that has a path.
	vetoArmed  bool
	copyRowHit bool
	copyRow    int
}

// decideTap applies the interaction rules in their fixed order:
//
//  1. a jump control, because it is an explicit arrow the user aimed at;
//  2. the compact strip's copy half, which is a distinct half of the same pill;
//  3. the auto-approve veto, but only while a countdown is really running -
//     "don't approve" is the most expensive thing a tap can mean and it must
//     not be answered with a clipboard write;
//  4. an expanded session row, which copies its project path;
//  5. otherwise the panel collapses, exactly as it did before any of this.
func decideTap(in tapInputs) (tapAction, int) {
	if in.jumpHit && in.jumpValid {
		return tapJump, in.jumpRow
	}
	if in.copyStripHit && in.copyStripValid {
		return tapCopyStrip, 0
	}
	if in.expanded && in.vetoArmed {
		return tapVeto, 0
	}
	if in.expanded && in.copyRowHit {
		return tapCopyRow, in.copyRow
	}
	return tapCollapse, 0
}
