package ui

import (
	"strings"

	"ai-terminal-pet/internal/session"
)

// Row geometry for the expanded list, in one place.
//
// The rule this file exists to enforce is the same one focus_windows.go already
// follows for the jump arrow: the rasterizer and the hit test must call the
// same function, or the panel ends up drawing a row in one place and reacting
// to clicks in another. A right-click now means something per row (mute that
// conversation), which turns "close enough" from a cosmetic nit into a feature
// that silently does nothing.

// expandedRowRect is the drawn rectangle of one session row, in canvas pixels.
// index 0 is the first row under the summary line.
//
// The formulas are the rasterizer's own: left scale(14), width minus
// scale(28), each row starting at expandedFocusRowTop and standing
// expandedFocusRowHeight tall, with the 4-unit gap baked into that helper.
func expandedRowRect(width, index int) (left, top, w, h int) {
	left = focusScaled(14, width, expandedWidth)
	w = width - focusScaled(28, width, expandedWidth)
	top = expandedFocusRowTop(width, index)
	h = expandedFocusRowHeight(width)
	return
}

// expandedRowHit maps a client-space point onto a drawn session row. Rows that
// are not drawn (past expandedMaxRows, or when there are fewer sessions than
// pointer positions) are not hit: the overflow line and the footer below them
// keep their existing meaning.
func expandedRowHit(x, y, width, height, rows int) (int, bool) {
	if width <= 0 || height <= 0 || rows <= 0 {
		return 0, false
	}
	limit := minInt(rows, expandedMaxRows)
	for index := 0; index < limit; index++ {
		left, top, w, h := expandedRowRect(width, index)
		if x >= left && x < left+w && y >= top && y < top+h {
			return index, true
		}
	}
	return 0, false
}

// rowCopyText is what a tap on a drawn session row puts on the clipboard.
//
// A row is only a copy target when the session actually reported a working
// directory. Reporters are allowed to send no context at all, and a receipt
// that says "copied" over an empty clipboard would be worse than the click
// meaning nothing, so an empty project falls through to the ordinary
// collapse behaviour instead of copying "".
//
// The full path is used, not the displayed label: panel rows print
// filepath.Base(project) because a 200-unit-wide row cannot show
// C:\Users\me\work\ai-terminal-pet, and the path is the thing worth pasting
// into a terminal.
func rowCopyText(items []session.Session, index int) (string, bool) {
	if index < 0 || index >= len(items) || index >= expandedMaxRows {
		return "", false
	}
	value := strings.TrimSpace(items[index].Project)
	if value == "" {
		return "", false
	}
	return value, true
}
