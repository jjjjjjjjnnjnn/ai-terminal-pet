//go:build !windows

package ui

// Control on a platform whose window layer does not exist yet.
//
// Answering false is the deliberate choice, and it is the honest one: every caller
// treats "no running instance" as a reason to fall back, and the overlay command
// then reports ErrNoWindowLayer instead of silently doing nothing. Answering true
// would make `--toggle` look like it worked on a process that has no panel to toggle
// (handover §55.6 step 7).

func postControlMessage(action ControlAction) bool {
	return false
}
