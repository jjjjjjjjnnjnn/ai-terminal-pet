package ui

// ControlAction is the set of commands one process can send to the live overlay,
// and the type is platform-neutral on purpose: the CLI, the tray menu and an
// installer script all name the same five actions, and only the delivery differs.
//
// The Windows implementation finds the window by class name and posts a message
// (postWindowMessage in control_windows.go). A platform without a window layer has
// nothing to post to, so its Control answers false - which every caller already
// treats as "no running instance" and uses to decide whether to start one
// (handover §55.6 step 7).
type ControlAction int

const (
	// ControlActivate reveals the overlay (same path as a tray click).
	ControlActivate ControlAction = iota
	// ControlToggle flips compact/expanded.
	ControlToggle
	// ControlQuit asks the overlay to close and exit.
	ControlQuit
	// ControlReset drops the remembered geometry and re-anchors.
	ControlReset
	// ControlSettings opens the settings window. It is a separate action so a
	// hotkey or a shortcut can reach it without going through the tray.
	ControlSettings
)

// Control sends action to an already-running instance and reports whether one was
// found. A missing window is a plain false, never an error: callers use it to decide
// whether to fall back to starting a fresh instance.
func Control(action ControlAction) bool {
	if _, known := controlActionIsKnown(action); !known {
		return false
	}
	return postControlMessage(action)
}

// controlActionIsKnown keeps the switch in one place: an unknown action must never
// be posted as a windows message with a zero id (which would be WM_NULL).
func controlActionIsKnown(action ControlAction) (ControlAction, bool) {
	switch action {
	case ControlActivate, ControlToggle, ControlQuit, ControlReset, ControlSettings:
		return action, true
	}
	return action, false
}
