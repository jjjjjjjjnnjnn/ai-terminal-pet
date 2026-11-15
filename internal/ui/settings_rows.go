package ui

// Row kinds and their height arithmetic, outside the window file.
//
// settingsRowKind classifies what a settings line contains, and the height of a
// row follows from that: a toggle shares its line with the label, a slider and a
// segment need a line below it, and a note wraps with no control beside it at all.
//
// Why this is its own file (handover §55.6 step 2): settings_text.go - the file
// that wraps and measures every hint, and the one §55.4 called out as the piece
// with the most measured numbers in it - only needs these constants and the kind
// of the row it is laying out. It never touches a window handle. Leaving them in
// settings_windows.go made the layout math look Windows-specific when it is not.

// Row kinds.
const (
	rowToggle = iota
	rowSlider
	rowSegment
	rowStepper
	rowButton
	rowNote
)

// Row height parts, in logical pixels at 96 DPI. A row's height is these parts
// added up rather than one number per kind, because a hint may need one or two
// lines and a fixed height either wastes space or clips the second line.
const (
	settingsTitleBand = 24
	settingsHintLine  = 17
	settingsRowPad    = 10
	settingsBandGap   = 4
)
