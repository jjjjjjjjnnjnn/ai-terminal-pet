package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"ai-terminal-pet/internal/session"
)

const dtCenter = 0x0001

// Panel geometry in logical units at 96 DPI. The window sizing code and the
// rasterizer both derive from these, so a layout change cannot leave the frame
// a different size from the picture drawn inside it.
const (
	compactWidth  = 56
	compactHeight = 156

	expandedWidth          = 200
	expandedHeaderHeight   = 46
	expandedSummaryHeight  = 30
	expandedRowHeight      = 34
	expandedEmptyHeight    = 56
	expandedOverflowHeight = 24
	expandedFooterHeight   = 34
	expandedMaxRows        = 6
)

// expandedPanelHeight sizes the expanded panel to what it actually has to
// show. The old panel was a fixed 250 units tall with two hard-coded rows, so
// a third session existed only as a number in the summary line - the panel
// could not display the very thing the user opened it for.
func expandedPanelHeight(count int) int {
	body := expandedEmptyHeight
	if count > 0 {
		body = minInt(count, expandedMaxRows) * expandedRowHeight
		if count > expandedMaxRows {
			body += expandedOverflowHeight
		}
	}
	return expandedHeaderHeight + expandedSummaryHeight + body + expandedFooterHeight
}

// expandedCloseButtonRect is the single source of truth for the expanded
// panel's close button geometry, used by the rasterizer. The button is a
// visual affordance only - a tap anywhere on the panel collapses it - but it
// still has to sit inside the header without colliding with the state pill.
func expandedCloseButtonRect(width int) (left, top, size int) {
	unit := float64(width) / expandedWidth
	scale := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}
	size = scale(20)
	left = width - size - scale(12)
	top = scale(13)
	return left, top, size
}

func drawVerticalPanel(target *canvas, items []session.Session, now time.Time, expanded, pressed, hovered bool) {
	width := target.width
	height := target.height
	state := session.Aggregate(items)
	accent := stateColor(state)
	inset := 0
	if pressed {
		inset = 1
	}
	panel := rgba{10, 15, 29, 235}
	raised := rgba{24, 32, 52, 225}
	// Hover is the only cue that the panel is a control and not just a
	// picture. Lifting the fill as well as the outline matters because in the
	// alert states the outline is already near-opaque state colour, so an
	// outline-only cue was invisible exactly when the panel most wanted to be
	// clicked.
	if hovered {
		panel = rgba{22, 30, 54, 242}
	}
	radius := minInt(width/2, 24)
	target.fillRounded(inset, inset, width-inset*2, height-inset*2, radius, panel)
	border := rgba{76, 91, 123, 150}
	if state == session.StateWaitingPermission || state == session.StateError {
		border = accent
		border.A = 210
	}
	if hovered {
		border.A = uint8(minInt(int(border.A)+55, 255))
	}
	if pressed {
		border.A = uint8(minInt(int(border.A)+30, 255))
	}
	target.outlineRounded(inset, inset, width-inset*2, height-inset*2, radius, 1, border)
	if expanded {
		drawVerticalExpanded(target, items, now, state, accent, raised)
		return
	}
	drawVerticalCompact(target, items, now, state, accent, pressed)
}

// drawVerticalCompact answers three questions in the order a glance asks
// them: what is the state, how many sessions are behind it, and which one
// wants me. Everything else the strip used to carry - a 24-segment ring, two
// arrow-labelled counters, a divider and a chevron - encoded those same two
// facts six different ways, which made the panel dense without making it
// informative.
func drawVerticalCompact(target *canvas, items []session.Session, now time.Time, state session.State, accent rgba, pressed bool) {
	width := target.width
	unit := float64(width) / compactWidth
	value := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}
	center := width / 2

	drawStatusOrb(target, center, value(14), value(28), state, now, pressed)

	target.text(value(4), value(50), width-value(8), value(26), compactCount(len(items)), value(22), 700, colorPrimary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	target.text(value(4), value(76), width-value(8), value(12), "会话", value(9), 400, colorSecondary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)

	target.fillRect(value(12), value(96), width-value(24), 1, rgba{79, 94, 126, 120})

	target.text(value(2), value(102), width-value(4), value(16), stateLabel(state), value(11), 700, accent, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)

	subject, subjectColor := compactSubject(items, state)
	subject = target.fitText(subject, value(9), 400, width-value(8))
	target.text(value(4), value(124), width-value(8), value(14), subject, value(9), 400, subjectColor, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}

// compactSubject picks the one line of text worth showing under the count:
// the session that wants the user when one does, otherwise the most recently
// active one, otherwise a hint at what to do about the empty panel.
//
// items arrive priority-sorted from the store, so the first match is already
// the most urgent. The label is deliberately not tinted with the state colour:
// the word directly above it already carries that, and colouring both made the
// strip read as one amber smear instead of two distinct facts.
func compactSubject(items []session.Session, state session.State) (string, rgba) {
	if len(items) == 0 {
		return "等待 AI CLI", colorMuted
	}
	if state == session.StateWaitingPermission || state == session.StateError {
		for _, current := range items {
			if current.State == session.StateWaitingPermission || current.State == session.StateError {
				return projectLabel(current), colorPrimary
			}
		}
	}
	return projectLabel(items[0]), colorSecondary
}

func drawVerticalExpanded(target *canvas, items []session.Session, now time.Time, state session.State, accent, raised rgba) {
	width := target.width
	height := target.height
	unit := float64(width) / expandedWidth
	value := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}

	// Header: title, live state, close.
	target.text(value(14), value(13), width-value(120), value(22), "终端会话", value(13), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	pillWidth := value(74)
	pillX := width - pillWidth - value(40)
	pillY := value(13)
	pill := accent
	pill.A = 46
	target.fillRounded(pillX, pillY, pillWidth, value(20), 10, pill)
	target.text(pillX, pillY, pillWidth, value(20), stateLabel(state), value(10), 700, accent, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)

	closeX, closeY, closeSize := expandedCloseButtonRect(width)
	target.fillEllipse(closeX, closeY, closeSize, closeSize, raised)
	target.line(closeX+value(6), closeY+value(6), closeX+closeSize-value(6), closeY+closeSize-value(6), 1, colorSecondary)
	target.line(closeX+closeSize-value(6), closeY+value(6), closeX+value(6), closeY+closeSize-value(6), 1, colorSecondary)

	target.fillRect(value(14), value(expandedHeaderHeight-6), width-value(28), 1, rgba{79, 94, 126, 120})
	target.text(value(14), value(expandedHeaderHeight), width-value(28), value(expandedSummaryHeight), summaryLine(items), value(9), 400, colorSecondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)

	top := value(expandedHeaderHeight + expandedSummaryHeight)
	if len(items) == 0 {
		drawExpandedEmpty(target, top, width, value)
	} else {
		rows := minInt(len(items), expandedMaxRows)
		for index := 0; index < rows; index++ {
			drawVerticalSession(target, items[index], value(14), top+index*value(expandedRowHeight), width-value(28), value(expandedRowHeight)-value(4), now, value)
		}
		if hidden := len(items) - rows; hidden > 0 {
			y := top + rows*value(expandedRowHeight)
			target.text(value(14), y, width-value(28), value(expandedOverflowHeight), fmt.Sprintf("还有 %d 个会话", hidden), value(8), 400, colorMuted, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
		}
	}

	// The footer teaches the controls once, so the panel does not need a
	// visible handle or a chevron to advertise that it is interactive.
	footerTop := height - value(expandedFooterHeight)
	target.fillRect(value(14), footerTop-value(6), width-value(28), 1, rgba{79, 94, 126, 90})
	hint := target.fitText("拖拽移动 · 滚轮缩放 · 点击收起", value(8), 400, width-value(24))
	target.text(value(12), footerTop, width-value(24), value(expandedFooterHeight), hint, value(8), 400, colorMuted, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}

func drawExpandedEmpty(target *canvas, top, width int, scale func(int) int) {
	// The summary line already says there are no sessions, so this block
	// carries the actionable half instead of repeating it.
	target.text(scale(16), top+scale(6), width-scale(32), scale(34), "在终端里启动 AI CLI，浮窗会自动出现", scale(9), 400, colorMuted, dtVCenter|dtCenter|dtNoPrefix|dtWordBreak)
}

func drawStatusOrb(target *canvas, centerX, top, size int, state session.State, now time.Time, pressed bool) {
	if state == session.StateWaitingPermission || state == session.StateError {
		pulse := alertPulse(state, now)
		outer := size + 6 + int(float64(pulse)*8+0.5)
		outerColor := stateColor(state)
		outerColor.A = uint8(48 + float64(pulse)*40)
		target.fillEllipse(centerX-outer/2, top-(outer-size)/2, outer, outer, outerColor)
	}
	orb := stateColor(state)
	if pressed {
		orb.A = 215
	}
	target.fillEllipse(centerX-size/2, top, size, size, orb)
	inner := size - 7
	target.fillEllipse(centerX-inner/2, top+4, inner, inner, rgba{7, 12, 24, 205})
	target.text(centerX-inner/2, top+4, inner, inner, statusGlyph(state), maxInt(size/2, 11), 700, colorPrimary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}

// drawVerticalSession is one row of the expanded list: project and state on
// the first line, where it came from and how long ago it spoke on the second.
// The relative time is the part that makes the list actionable - "waiting for
// approval" reads very differently at 10 seconds and at 40 minutes.
func drawVerticalSession(target *canvas, current session.Session, left, top, width, height int, now time.Time, scale func(int) int) {
	target.fillRounded(left, top, width, height, 6, rgba{24, 32, 52, 205})
	accent := stateColor(current.State)
	target.fillRounded(left+scale(1), top+scale(1), scale(3), height-scale(2), 2, accent)

	textLeft := left + scale(9)
	stateWidth := scale(56)
	stateX := left + width - stateWidth - scale(9)

	target.text(textLeft, top+scale(3), stateX-textLeft-scale(6), scale(12), projectLabel(current), scale(10), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	target.text(stateX, top+scale(3), stateWidth, scale(12), stateLabel(current.State), scale(8), 700, accent, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)

	detail := fmt.Sprintf("%s · %s", current.Source, reasonLabel(current))
	target.text(textLeft, top+scale(17), stateX-textLeft-scale(6), scale(10), detail, scale(8), 400, colorSecondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	target.text(stateX, top+scale(17), stateWidth, scale(10), relativeTime(current.UpdatedAt, now), scale(8), 400, colorMuted, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
}

// summaryLine is the one-line census under the header. "运行" deliberately
// excludes sessions that are waiting on the user: from the user's side those
// are not running, they are blocking, and counting them as both made the old
// metric rows double-count.
func summaryLine(items []session.Session) string {
	if len(items) == 0 {
		return "暂无会话"
	}
	parts := []string{fmt.Sprintf("%d 个会话", len(items))}
	if busy := busySessions(items); busy > 0 {
		parts = append(parts, fmt.Sprintf("%d 运行", busy))
	}
	if waiting := waitingSessions(items); waiting > 0 {
		parts = append(parts, fmt.Sprintf("%d 等待", waiting))
	}
	return strings.Join(parts, " · ")
}

// relativeTime renders coarse buckets rather than a clock: sessions live for
// minutes, and "3 分钟前" is read faster than "14:07".
func relativeTime(value, now time.Time) string {
	if value.IsZero() {
		return ""
	}
	delta := now.Sub(value)
	if delta < 0 {
		delta = 0
	}
	switch {
	case delta < 45*time.Second:
		return "刚刚"
	case delta < 90*time.Second:
		return "1 分钟前"
	case delta < time.Hour:
		return fmt.Sprintf("%d 分钟前", int(delta.Minutes()))
	case delta < 2*time.Hour:
		return "1 小时前"
	default:
		return fmt.Sprintf("%d 小时前", int(delta.Hours()))
	}
}

// alertPulse is a slow, calm breath (about 1.1 s per cycle). It used to run at
// 180 ms, which read as a glitch rather than a signal - a status light should
// be noticed once, not be hard to look at.
func alertPulse(state session.State, now time.Time) float64 {
	if state != session.StateWaitingPermission && state != session.StateError {
		return 0
	}
	value := math.Sin(float64(now.UnixNano()) / float64(175*time.Millisecond))
	return (value + 1) / 2
}

func statusGlyph(state session.State) string {
	switch state {
	case session.StateWaitingPermission:
		return "!"
	case session.StateError:
		return "×"
	case session.StateWorking:
		return "•"
	case session.StateToolUse, session.StateThinking:
		return "…"
	default:
		return "✓"
	}
}

// busySessions counts sessions doing work right now, excluding those blocked
// on the user.
func busySessions(items []session.Session) int {
	count := 0
	for _, current := range items {
		switch current.State {
		case session.StateWorking, session.StateToolUse, session.StateThinking:
			count++
		}
	}
	return count
}

func compactCount(value int) string {
	if value > 9 {
		return "9+"
	}
	return strconv.Itoa(value)
}

func waitingSessions(items []session.Session) int {
	count := 0
	for _, current := range items {
		if current.State == session.StateWaitingPermission {
			count++
		}
	}
	return count
}
