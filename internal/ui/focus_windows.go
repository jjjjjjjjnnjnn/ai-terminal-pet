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
	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

const (
	// Logical units at 96 DPI, in the same space as compactWidth/expandedWidth.
	compactFocusLeft   = 4
	compactFocusTop    = 136
	compactFocusHeight = 18
	expandedFocusSize  = 20
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
// jumpable gates the compact bar so a tap on empty space still toggles.
func focusHit(x, y, width, height int, expanded, jumpable bool, rows int) (int, bool) {
	if width <= 0 || height <= 0 {
		return 0, false
	}
	if !expanded {
		if !jumpable {
			return 0, false
		}
		left, top, w, h := compactFocusRect(width)
		if hitRect(x, y, left, top, w, h, 2) {
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

// drawCompactJump renders the one-tap bar: state-coloured pill, arrow, label
// and the project name of the session it will switch to.
func drawCompactJump(target *canvas, current session.Session, lang i18n.Lang) {
	width := target.width
	left, top, w, h := compactFocusRect(width)
	accent := stateColor(current.State)

	// A neutral pill with a state-coloured arrow and label. The old solid
	// amber bar dominated the whole compact strip; this is a control, not an
	// alert, so the colour belongs to the icon and the verb only.
	target.fillRounded(left, top, w, h, maxInt(h/3, 1), colors.Raised)
	outline := accent
	outline.A = 135
	target.outlineRounded(left, top, w, h, maxInt(h/3, 1), 1, outline)

	arrowSize := maxInt(h*11/20, 4)
	arrowLeft := left + focusScaled(4, width, compactWidth)
	drawFocusArrow(target, arrowLeft, top+(h-arrowSize)/2, arrowSize, accent)

	textLeft := arrowLeft + arrowSize + focusScaled(4, width, compactWidth)
	textWidth := w - (textLeft - left) - focusScaled(3, width, compactWidth)
	if textWidth <= 0 {
		return
	}
	// The button carries only its verb. The project name stays on the line
	// just above it (drawVerticalCompact), where there is room for ellipsis;
	// squeezing both into 48 logical units produced "Jump · ..." and told the
	// user neither which session nor what the control did.
	label := target.fitText(i18n.T(lang, i18n.KeyPanelJump), focusScaled(8, width, compactWidth), 700, textWidth)
	target.text(textLeft, top, textWidth, h, label, focusScaled(8, width, compactWidth), 700,
		accent, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
}

// drawExpandedFocusButton renders the per-row arrow button.
func drawExpandedFocusButton(target *canvas, left, top, size int, accent rgba) {
	target.fillRounded(left, top, size, size, maxInt(size/4, 1), colors.Raised)
	outline := accent
	outline.A = 120
	target.outlineRounded(left, top, size, size, maxInt(size/4, 1), 1, outline)
	drawFocusArrow(target, left+size/6, top+size/6, size*2/3, accent)
}
