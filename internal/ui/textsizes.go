package ui

// The three text-size names, outside the window file.
//
// They are the persisted names config.json stores ("small" / "standard" / "large"), so they are
// part of the config contract, not part of the window: paint_cost_test.go names
// them to document what its measurement table covers, and it has to compile without
// a platform window layer (handover §55.6 step 2).

const (
	textSizeSmall    = "small"
	textSizeStandard = "standard"
	textSizeLarge    = "large"
)
