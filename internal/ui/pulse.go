package ui

import (
	"time"

	"ai-terminal-pet/internal/session"
)

// A parked waiting_permission session used to cost ~8% of a CPU core forever:
// every animation tick rasterised the whole layered panel (GDI text, rounded
// rects, premultiplied alpha) just to change the breathing orb by a few
// shades. The pulse is now expressed as the layered window's constant alpha,
// so a pure pulse frame reuses the already-rendered bitmap and does no GDI
// drawing at all.

// purePulseFrame reports whether this frame differs from the previous one
// only by time. Sliding frames and any content change still need a full
// raster; a breathing alert does not.
func purePulseFrame(previous, current renderedKey, state session.State, tuck, tuckTarget float64) bool {
	if state != session.StateWaitingPermission && state != session.StateError {
		return false
	}
	if tuck != tuckTarget {
		return false
	}
	if previous.clock == current.clock {
		return false
	}
	previous.clock = 0
	current.clock = 0
	return previous == current
}

// pulseOpacity maps the breath onto a narrow band of constant alpha so the
// panel is never too dim to read. alertPulse returns 0..1.
func pulseOpacity(base uint8, state session.State, now time.Time) uint8 {
	if state != session.StateWaitingPermission && state != session.StateError {
		return base
	}
	const amplitude = 26
	minimum := int(base) - amplitude
	if minimum < 1 {
		minimum = 1
	}
	level := minimum + int(float64(amplitude)*alertPulse(state, now))
	if level > int(base) {
		level = int(base)
	}
	return uint8(level)
}
