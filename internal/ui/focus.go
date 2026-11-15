package ui

// Jump-to-terminal affordances.
//
// A session that needs a decision (waiting_permission / error), or one that
// has finished and is waiting to be read (idle / closed), should be one click
// away from the terminal that owns it. Resolving *which* window is platform
// work (internal/platform); everything in this file is about where the click
// target is drawn and how a tap maps onto it.
//
// The geometry here is the single source of truth: the rasterizer and the hit
// test both call compactFocusRect / expandedRowFocusRect, so the arrow that is
// drawn is exactly the arrow that can be clicked.

import (
	"strings"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

const (
	// Logical units at 96 DPI, in the same space as compactWidth/expandedWidth.
	compactFocusLeft   = 4
	compactFocusTop    = 136
	compactFocusHeight = 18
	expandedFocusSize  = 20

	// The compact strip carries two controls: copy the most urgent session's
	// working directory, and jump to its terminal. The copy half is sized for
	// a glyph, the jump half keeps the arrow plus its verb, so a zoomed-out
	// strip gives up the copy target rather than drawing two unreadable
	// halves (see compactActionRects).
	compactCopyWidth = 17
	compactActionGap = 3
	// Minimums are device pixels, not logical units: a four-pixel target is a
	// trap regardless of what it claims to be. At 65% zoom the shipped strip
	// still splits (11 px copy, 17 px jump); below that there is only the
	// jump bar, exactly as 1.0.4 shipped it.
	compactCopyMinPixels = 10
	compactJumpMinPixels = 14
)

// focusScaled maps a logical unit onto canvas pixels. The rasterizer uses
// width/compactWidth (or width/expandedWidth) for exactly the same purpose.
func focusScaled(base, width, logicalWidth int) int {
	if logicalWidth <= 0 {
		return 1
	}
	return maxInt(int(float64(base)*float64(width)/float64(logicalWidth)+0.5), 1)
}

// compactFocusRect is the one-tap "jump to the most urgent session" bar. It
// replaces the compact subject line only when a jumpable session exists, so a
// panel with everything busy keeps its original silhouette and size.
func compactFocusRect(width int) (left, top, w, h int) {
	left = focusScaled(compactFocusLeft, width, compactWidth)
	top = focusScaled(compactFocusTop, width, compactWidth)
	w = width - focusScaled(compactFocusLeft*2, width, compactWidth)
	h = focusScaled(compactFocusHeight, width, compactWidth)
	return
}

// actionRect is one clickable zone inside a control.
type actionRect struct {
	left   int
	top    int
	width  int
	height int
}

func (r actionRect) contains(x, y int) bool {
	return x >= r.left && x < r.left+r.width && y >= r.top && y < r.top+r.height
}

// compactActionRects splits the one-tap strip into its two targets: copy this
// session's working directory, and jump to its terminal.
//
// The strip's outline is unchanged - compactFocusRect still describes the pill
// that gets painted - so the panel keeps the silhouette 1.0.4 shipped. Only the
// inside is divided, with a visible divider at the boundary, because "two
// controls in one pill" is a claim the picture has to make before the click
// can be believed.
//
// The split is gated on two things, and both are deliberate:
//
//   - copyable: a strip with nothing to copy must not advertise a copy target.
//     The session list changes constantly, and a permanent half that is dead
//     most of the time teaches the user to distrust the whole strip.
//   - width: at 56 logical units there is room for a glyph and an arrow, not
//     for a word as well. Below the minimums the split is dropped and the
//     strip is the plain jump bar again - two four-pixel targets are worse
//     than the single control 1.0.4 shipped. The minimums are in device
//     pixels rather than logical units because what makes a target usable is
//     how many pixels it occupies, not how many units it claims.
func compactActionRects(width int, copyable bool) (copyRect, jumpRect actionRect) {
	left, top, w, h := compactFocusRect(width)
	jumpRect = actionRect{left: left, top: top, width: w, height: h}
	if !copyable {
		return actionRect{}, jumpRect
	}
	copyWidth := focusScaled(compactCopyWidth, width, compactWidth)
	gap := focusScaled(compactActionGap, width, compactWidth)
	if copyWidth < compactCopyMinPixels || h < compactCopyMinPixels {
		return actionRect{}, jumpRect
	}
	remainder := w - copyWidth - gap
	if remainder < compactJumpMinPixels {
		return actionRect{}, jumpRect
	}
	copyRect = actionRect{left: left, top: top, width: copyWidth, height: h}
	jumpRect.left = left + copyWidth + gap
	jumpRect.width = remainder
	return copyRect, jumpRect
}

// compactCopyHit answers only for the copy half of the strip. There is
// deliberately no click tolerance here: the two targets are adjacent, and
// padding on both sides would make one tap hit both.
func compactCopyHit(x, y, width, height int, copyable bool) bool {
	if width <= 0 || height <= 0 || !copyable {
		return false
	}
	copyRect, _ := compactActionRects(width, copyable)
	if copyRect.width <= 0 {
		return false
	}
	return copyRect.contains(x, y)
}

// expandedFocusRowTop / expandedFocusRowHeight mirror drawVerticalExpanded's
// row layout. Keep them in sync with expandedRowHeight and the -scale(4) gap.
func expandedFocusRowTop(width, index int) int {
	return focusScaled(expandedHeaderHeight+expandedSummaryHeight, width, expandedWidth) +
		index*focusScaled(expandedRowHeight, width, expandedWidth)
}

func expandedFocusRowHeight(width int) int {
	return focusScaled(expandedRowHeight, width, expandedWidth) - focusScaled(4, width, expandedWidth)
}

// expandedRowFocusRect is the per-row arrow button, flush with the row's
// right edge (rows themselves run from scale(14) to width-scale(14)).
func expandedRowFocusRect(width, rowTop, rowHeight int) (left, top, size int) {
	size = focusScaled(expandedFocusSize, width, expandedWidth)
	left = width - focusScaled(14, width, expandedWidth) - focusScaled(4, width, expandedWidth) - size
	top = rowTop + (rowHeight-size)/2
	return
}

// hitRect adds a small tolerance: a 20-unit arrow is a small target at 100%
// zoom, and a pixel of slop costs nothing because the surrounding panel is
// still the ordinary click-to-toggle surface.
func hitRect(x, y, left, top, w, h, padding int) bool {
	return x >= left-padding && x <= left+w+padding &&
		y >= top-padding && y <= top+h+padding
}

// focusHit maps a client-space tap onto a jump target.
// row < 0 is the compact one-tap bar; row >= 0 is an expanded session row.
// jumpable gates the compact bar so a tap on empty space still toggles, and
// copyable says whether the bar is currently split into copy + jump.
func focusHit(x, y, width, height int, expanded, jumpable, copyable bool, rows int) (int, bool) {
	if width <= 0 || height <= 0 {
		return 0, false
	}
	if !expanded {
		if !jumpable {
			return 0, false
		}
		// The jump target is the strip's right half once the copy control has
		// taken the left one: a tap on the copy glyph must not also jump.
		_, jumpRect := compactActionRects(width, copyable)
		if jumpRect.contains(x, y) {
			return -1, true
		}
		return 0, false
	}
	limit := minInt(rows, expandedMaxRows)
	for index := 0; index < limit; index++ {
		left, top, size := expandedRowFocusRect(width, expandedFocusRowTop(width, index), expandedFocusRowHeight(width))
		if hitRect(x, y, left, top, size, size, 2) {
			return index, true
		}
	}
	return 0, false
}

// compactCopyText is what the compact strip's copy control puts on the
// clipboard: the working directory of the same session the jump bar would
// switch to. One selection rule for both halves - "the session that wants you"
// - is why the two controls sit in one pill; two different notions of "the
// current session" in one 48-unit bar would be indistinguishable.
func compactCopyText(items []session.Session) (string, bool) {
	current, ok := mostUrgentJump(items)
	if !ok {
		return "", false
	}
	value := strings.TrimSpace(current.Project)
	if value == "" {
		return "", false
	}
	return value, true
}

// jumpPriority ranks the states where the user is the one who must act next.
// Busy states are deliberately excluded: there is nothing to decide or read
// while the agent is still working, and jumping to it would just be a click
// that steals focus for no reason.
func jumpPriority(state session.State) int {
	switch state {
	case session.StateWaitingPermission:
		return 4
	case session.StateError:
		return 3
	case session.StateIdle:
		return 2
	case session.StateClosed:
		return 1
	default:
		return -1
	}
}

// mostUrgentJump picks the session the compact bar should jump to. Ties go to
// the most recently updated session, matching session.Sort's ordering.
func mostUrgentJump(items []session.Session) (session.Session, bool) {
	var best session.Session
	bestRank := -1
	for _, current := range items {
		rank := jumpPriority(current.State)
		if rank < 0 {
			continue
		}
		if rank > bestRank || (rank == bestRank && current.UpdatedAt.After(best.UpdatedAt)) {
			best, bestRank = current, rank
		}
	}
	return best, bestRank >= 0
}

// drawFocusArrow paints a north-east arrow with lines rather than a glyph:
// the panel font is not guaranteed to contain U+2197 on every system.
func drawFocusArrow(target *canvas, left, top, size int, color rgba) {
	if size < 4 {
		return
	}
	pad := size / 4
	x0, y0 := left+pad, top+size-pad
	x1, y1 := left+size-pad, top+pad
	thickness := maxInt(size/9, 1)
	target.line(x0, y0, x1, y1, thickness, color)
	target.line(x1-size/2, y1, x1, y1, thickness, color)
	target.line(x1, y1+size/2, x1, y1, thickness, color)
}

// drawCompactJump renders the one-tap strip: state-coloured pill, a copy
// control on its left half and the jump control on its right half.
func drawCompactJump(target *canvas, current session.Session, lang i18n.Lang, copyable bool) {
	width := target.width
	left, top, w, h := compactFocusRect(width)
	copyRect, jumpRect := compactActionRects(width, copyable)
	accent := stateColor(current.State)

	// A neutral pill with a state-coloured arrow and label. The old solid
	// amber bar dominated the whole compact strip; this is a control, not an
	// alert, so the colour belongs to the icon and the verb only. The pill
	// spans the full strip even when only the jump half is live, so the
	// silhouette is the same shape 1.0.4 shipped.
	target.fillRounded(left, top, w, h, maxInt(h/3, 1), colors.Raised)
	outline := accent
	outline.A = 135
	target.outlineRounded(left, top, w, h, maxInt(h/3, 1), 1, outline)

	if copyRect.width > 0 {
		// The divider is the picture's claim that the strip is two controls;
		// without it a tap on the left half looks like it should jump.
		divider := colors.Divider
		divider.A = 120
		boundary := copyRect.left + copyRect.width + (jumpRect.left-(copyRect.left+copyRect.width))/2
		target.fillRect(boundary, top+focusScaled(3, width, compactWidth), 1, h-focusScaled(6, width, compactWidth), divider)
		drawCopyGlyph(target, copyRect, accent)
	}

	arrowSize := maxInt(h*11/20, 4)
	inset := focusScaled(4, width, compactWidth)
	arrowLeft := jumpRect.left + inset
	drawFocusArrow(target, arrowLeft, top+(h-arrowSize)/2, arrowSize, accent)

	textLeft := arrowLeft + arrowSize + inset
	textWidth := jumpRect.left + jumpRect.width - textLeft - focusScaled(3, width, compactWidth)
	label := i18n.T(lang, i18n.KeyPanelJump)
	// An 18-unit bar has no room for an ellipsis: once the strip is split, the
	// arrow is the whole control and a lone "…" next to it would read as a
	// broken label rather than as a button. So the verb is drawn only when it
	// fits, and the project name on the line just above it carries the
	// meaning either way.
	if textWidth <= 0 || target.measure(label, focusScaled(8, width, compactWidth), 700) > textWidth {
		return
	}
	target.text(textLeft, top, textWidth, h, label, focusScaled(8, width, compactWidth), 700,
		accent, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
}

// drawCopyGlyph is the compact copy control's icon: two overlapping sheets.
// It is drawn rather than written because a word would not fit in the 16 units
// the strip can spare, and a glyph needs no font guarantee beyond the arrow
// this panel already draws with lines.
func drawCopyGlyph(target *canvas, rect actionRect, colour rgba) {
	size := maxInt(rect.height*9/20, 5)
	if rect.width < size+2 {
		size = maxInt(rect.width-2, 4)
	}
	sheet := maxInt(size*3/4, 3)
	offset := size - sheet
	x := rect.left + (rect.width-size)/2
	y := rect.top + (rect.height-size)/2
	radius := maxInt(size/5, 1)
	// Back sheet first (outline), then the front one, so the overlap reads as
	// "this is a copy of that".
	target.outlineRounded(x+offset, y, sheet, sheet, radius, 1, colour)
	target.outlineRounded(x, y+offset, sheet, sheet, radius, 1, colour)
}

// drawExpandedFocusButton renders the per-row arrow button.
func drawExpandedFocusButton(target *canvas, left, top, size int, accent rgba) {
	target.fillRounded(left, top, size, size, maxInt(size/4, 1), colors.Raised)
	outline := accent
	outline.A = 120
	target.outlineRounded(left, top, size, size, maxInt(size/4, 1), 1, outline)
	drawFocusArrow(target, left+size/6, top+size/6, size*2/3, accent)
}
