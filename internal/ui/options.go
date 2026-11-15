package ui

import "ai-terminal-pet/internal/platform"

// Options is everything the UI needs from outside. All callbacks are invoked
// on the UI thread and may be nil.
type Options struct {
	// Opacity is the layered window constant alpha, 1..255.
	Opacity uint8
	// Settings is the initial state of the tray switches.
	Settings Settings
	// AutoStart is the initial state of the run-at-logon switch.
	AutoStart bool
	// AutoApproveTimeoutSec is the countdown length for auto-approve.
	// Out-of-range values fall back to 5 seconds.
	AutoApproveTimeoutSec int
	// OpenSettings opens the settings window right after the first paint, so
	// `--settings` works even when it had to start the overlay itself.
	OpenSettings bool
	// Initial is the remembered geometry, or the zero value when the window
	// has never been placed.
	Initial Placement
	// OnPlacementChanged fires after the user moves, zooms or expands the
	// overlay so the app layer can persist it.
	OnPlacementChanged func(Placement)
	// OnSettingsChanged fires when a tray switch is flipped.
	OnSettingsChanged func(Settings)
	// OnAutoStartChanged fires when the run-at-logon switch is flipped.
	OnAutoStartChanged func(bool) bool
	// Presence is the process/window cache the attendance sampler reuses.
	Presence *platform.Presence
}
