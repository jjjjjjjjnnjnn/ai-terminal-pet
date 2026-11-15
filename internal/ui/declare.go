package ui

// Pure declarations the drawing layer needs but that used to live in
// Windows-only files.
//
// Step 2 of handover §55.6 renamed the ten files that never call Win32 and let
// the compiler say what they still needed. The answer was two things, both of
// them values rather than code:
//
//	dockSide / dockNone / dockLeft / dockRight  - which screen edge the panel
//	    hugs. The arithmetic lives in dock_windows.go (it needs a work area);
//	    the type only has to exist for drawing_compact.go to switch on it.
//	rowNote                                     - one settings-row kind. The rest
//	    of the set is in settings_windows.go, which owns the window; rowNote is
//	    the only kind the platform-neutral text layout distinguishes, because a
//	    note wraps with no control beside it.
//
// Keeping them here instead of copying them per platform is the same rule as
// attention_keys.go: a shared type has one home.

// dockSide is the screen edge the panel is attached to.
type dockSide int

const (
	dockNone dockSide = iota
	dockLeft
	dockRight
)
