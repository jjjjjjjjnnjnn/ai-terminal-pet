package ui

// The copy receipt: a short-lived pill saying "copied", drawn over the row that
// was tapped (or over the compact subject line).
//
// Why it exists at all: a layered window has no cursor, no focus ring and no
// other way to say "that worked". The tap has to answer for itself, or the
// gesture reads as "nothing happened" - which is exactly the failure mode F8
// was written to avoid.
//
// It is NOT the alert receipt (receipt_windows.go). That file records what the
// alert policy decided, for a bug report; this one is a transient pixel. Two
// different questions, two different lifetimes, so they do not share a name.

import (
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

const (
	// copyToastDuration is long enough to read one short word and short enough
	// that a second tap on another row is not fighting the first pill.
	copyToastDuration = 1200 * time.Millisecond
	// jumpToastDuration is longer because "could not switch" is three words and
	// arrives after the jump already failed - there is no second signal behind
	// it to correct a missed read.
	jumpToastDuration = 1600 * time.Millisecond
	// toastCompactRow marks a toast that belongs to the compact strip, which
	// has no row index. A negative value cannot collide with a drawn row.
	toastCompactRow = -1
)

// toastKind is which sentence the pill states. Three, because a tap can fail in
// two different ways and telling somebody "could not copy" after a jump they
// know they attempted is worse than saying nothing.
type toastKind uint8

const (
	toastNone toastKind = iota
	toastCopied
	toastCopyFailed
	toastJumpFailed
)

// duration is how long this pill stays up. A zero kind has no lifetime, which
// is why show() clears the receipt rather than arming it with nothing to say.
func (k toastKind) duration() time.Duration {
	if k == toastJumpFailed {
		return jumpToastDuration
	}
	return copyToastDuration
}

// failed reports whether the gesture did not take effect. It picks the error
// colour, so "failed" in the picture and "failed" in the key cannot disagree.
func (k toastKind) failed() bool {
	return k == toastCopyFailed || k == toastJumpFailed
}

// copyToastKind maps a copy outcome onto its pill. One place, because three
// call sites passing !copied is how the two meanings drift apart.
func copyToastKind(failed bool) toastKind {
	if failed {
		return toastCopyFailed
	}
	return toastCopied
}

// copyToast is the live receipt. Its zero value is inert: only `until` decides
// whether anything is on screen, so an App that never copied anything behaves
// exactly as before F8 existed.
//
// identity is the conversation the receipt is about. The panel reorders itself
// whenever a reporter pushes, and a pill left over a row that now holds
// somebody else's session would claim "copied" about the wrong project; the
// receipt therefore dies when its subject moves (see frame).
type copyToast struct {
	row      int
	identity string
	kind     toastKind
	until    time.Time
}

// show arms the receipt for one target. kind says which sentence it will state;
// a failed jump and a failed copy are both failures but not the same claim.
func (t *copyToast) show(row int, identity string, kind toastKind, now time.Time) {
	if kind == toastNone {
		t.clear()
		return
	}
	t.row = row
	t.identity = identity
	t.kind = kind
	t.until = now.Add(kind.duration())
}

// clear removes the receipt early: a collapse, an expand toggle or a new
// session list must not leave a pill that belongs to the previous frame.
func (t *copyToast) clear() {
	*t = copyToast{}
}

func (t *copyToast) active(now time.Time) bool {
	return !t.until.IsZero() && now.Before(t.until)
}

// frame is what the rasterizer and the frame key both read. tick is whole
// milliseconds remaining, so a live toast changes the key on every frame that
// can see a difference - and an expired one contributes 0 with no row, which
// makes the frame that removes it a repaint rather than a reuse.
func (t *copyToast) frame(now time.Time, items []session.Session) panelToast {
	if !t.active(now) || !toastSubjectThere(items, t.row, t.identity) {
		return panelToast{}
	}
	return panelToast{row: int32(t.row), tick: tickOf(t.until, now), kind: t.kind}
}

func tickOf(until, now time.Time) int64 {
	remaining := until.Sub(now)
	tick := int64(remaining / time.Millisecond)
	if tick < 1 {
		tick = 1
	}
	return tick
}

// toastSubjectThere answers whether the conversation the receipt is about still
// owns the place it was drawn in: the same row (expanded) or still the most
// urgent session (compact). A receipt pinned to a row index alone would claim
// "copied" about whoever the sort moved there.
func toastSubjectThere(items []session.Session, row int, identity string) bool {
	if identity == "" {
		return false
	}
	if row == toastCompactRow {
		current, ok := mostUrgentJump(items)
		return ok && session.Identity(current) == identity
	}
	if row < 0 || row >= len(items) || row >= expandedMaxRows {
		return false
	}
	return session.Identity(items[row]) == identity
}

// panelToast is the toast's contribution to one frame. Zero means "no toast":
// tick == 0 is unreachable for a live receipt (see frame), which is what lets
// the key distinguish "pill gone" from "pill on row 0".
type panelToast struct {
	row  int32
	tick int64
	kind toastKind
}

func (p panelToast) visible() bool {
	return p.tick > 0
}

func (p panelToast) compact() bool {
	return p.visible() && int(p.row) == toastCompactRow
}

// toastColour is the pill's ink: primary for a confirmation, error for a
// gesture that did not take effect. Both the row pill and the compact subject
// line read it, so the two cannot disagree about what "failed" looks like.
func toastColour(p panelToast) rgba {
	if p.kind.failed() {
		return colors.Error
	}
	return colors.Primary
}

// copyToastLabel is the one phrase the pill says. It is deliberately a statement
// about what happened ("copied" / "copy failed" / "could not switch") rather
// than a noun, so the user does not have to infer the verb from the gesture.
func copyToastLabel(p panelToast, lang i18n.Lang) string {
	switch p.kind {
	case toastCopyFailed:
		return i18n.T(lang, i18n.KeyCopyFailed)
	case toastJumpFailed:
		return i18n.T(lang, i18n.KeyJumpFailed)
	default:
		return i18n.T(lang, i18n.KeyCopyDone)
	}
}

// drawCopyToastRow paints the pill over one expanded row's text area: the
// state column and the jump arrow stay visible, because the pill is an aside
// about the tap and not a replacement for the row.
func drawCopyToastRow(target *canvas, p panelToast, items []session.Session, lang i18n.Lang) {
	index := int(p.row)
	if index < 0 || index >= len(items) || index >= expandedMaxRows {
		return
	}
	scale := func(base int) int {
		return focusScaled(base, target.width, expandedWidth)
	}
	left, top, width, height := expandedRowRect(target.width, index)
	stateWidth := scale(48)
	focusSize := scale(expandedFocusSize)
	textLeft := left + scale(9)
	textRight := left + width - stateWidth - focusSize - scale(8)
	pillWidth := textRight - textLeft
	if pillWidth <= 0 {
		return
	}
	pillHeight := scale(18)
	if pillHeight > height {
		pillHeight = height
	}
	pillTop := top + (height-pillHeight)/2
	// Primary, not Accent: the palette comment pins Accent to the settings
	// chrome ("marks on/selected, never used on the overlay itself"), and a
	// receipt is a neutral confirmation rather than a control state.
	colour := toastColour(p)
	fill := colors.Raised
	fill.A = 236
	target.fillRounded(textLeft, pillTop, pillWidth, pillHeight, maxInt(scale(5), 2), fill)
	outline := colour
	outline.A = 170
	target.outlineRounded(textLeft, pillTop, pillWidth, pillHeight, maxInt(scale(5), 2), 1, outline)
	label := target.fitText(copyToastLabel(p, lang), scale(8), 700, pillWidth-scale(6))
	target.text(textLeft+scale(3), pillTop, pillWidth-scale(6), pillHeight, label, scale(8), 700, colour, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}
