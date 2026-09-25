package ui

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"ai-terminal-pet/internal/session"
)

const dtCenter = 0x0001

// expandedCloseButtonRect is the single source of truth for the expanded
// panel's close button geometry. Both the rasterizer and the hit test go
// through it so a drawn button can never disagree with its click target -
// the two used to duplicate the same arithmetic in two files.
func expandedCloseButtonRect(width int) (left, top, size int) {
	unit := float64(width) / 200
	scale := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}
	size = scale(20)
	left = width - size - scale(12)
	top = scale(9)
	return left, top, size
}

// closeButtonHit reports whether a client-space point lands on the expanded
// panel's close button. Kept pure so the geometry is unit-testable.
func closeButtonHit(x, y, width int) bool {
	left, top, size := expandedCloseButtonRect(width)
	return x >= left && x < left+size && y >= top && y < top+size
}

func drawVerticalPanel(target *canvas, items []session.Session, now time.Time, expanded, pressed bool) {
	width := target.width
	height := target.height
	state := session.Aggregate(items)
	accent := stateColor(state)
	center := width / 2
	inset := 0
	if pressed {
		inset = 1
	}
	if state == session.StateWaitingPermission || state == session.StateError {
		pulse := alertPulse(state, now)
		glowSize := 48 + int(float64(pulse)*8+0.5)
		glow := accent
		glow.A = uint8(18 + float64(pulse)*24)
		target.fillEllipse(center-glowSize/2, 4, glowSize, glowSize, glow)
	}
	panel := rgba{10, 15, 29, 235}
	raised := rgba{24, 32, 52, 225}
	radius := minInt(width/2, 28)
	target.fillRounded(inset, inset, width-inset*2, height-inset*2, radius, panel)
	border := rgba{76, 91, 123, 150}
	if state == session.StateWaitingPermission || state == session.StateError {
		border = accent
		border.A = 210
	}
	if pressed {
		border.A = uint8(minInt(int(border.A)+30, 255))
	}
	target.outlineRounded(inset, inset, width-inset*2, height-inset*2, radius, 1, border)
	if expanded {
		drawVerticalExpanded(target, items, now, state, accent, raised)
		return
	}
	drawVerticalCompact(target, items, now, state, accent, raised, pressed)
}

func drawVerticalCompact(target *canvas, items []session.Session, now time.Time, state session.State, accent, raised rgba, pressed bool) {
	width := target.width
	unit := float64(width) / 40
	value := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}
	center := width / 2
	orbSize := value(24)
	drawStatusOrb(target, center, value(8), orbSize, state, now, pressed)
	ringSize := value(34)
	ringTop := value(38)
	drawSessionRing(target, center, ringTop, ringSize, len(items), state)
	target.text(center-ringSize/2, ringTop+value(9), ringSize, value(17), compactCount(len(items)), value(15), 700, colorPrimary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	target.text(center-ringSize/2, ringTop+value(37), ringSize, value(10), "会话", value(8), 400, colorSecondary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	metricTop := value(93)
	drawCompactMetric(target, value(6), metricTop, value(14), "↑", "运", activeSessions(items), colorWorking, width, value)
	drawCompactMetric(target, value(6), metricTop+value(20), value(14), "↓", "待", waitingSessions(items), colorWaiting, width, value)
	dividerY := value(140)
	target.fillRect(value(8), dividerY, width-value(16), 1, rgba{79, 94, 126, 120})
	footer := "详情"
	if state == session.StateWaitingPermission {
		footer = "待处理"
	}
	if state == session.StateError {
		footer = "异常"
	}
	target.text(value(6), value(151), width-value(18), value(14), footer, value(9), 400, footerColor(state), dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	target.text(width-value(13), value(150), value(9), value(15), "›", value(12), 400, colorMuted, dtSingleLine|dtVCenter|dtNoPrefix)
}

func drawVerticalExpanded(target *canvas, items []session.Session, now time.Time, state session.State, accent, raised rgba) {
	width := target.width
	unit := float64(width) / 200
	value := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}
	target.text(value(14), value(9), width-value(58), value(20), "终端", value(14), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	pillWidth := value(70)
	pillX := width - pillWidth - value(42)
	pillY := value(9)
	pill := accent
	pill.A = 48
	target.fillRounded(pillX, pillY, pillWidth, value(19), 9, pill)
	accent.A = 255
	target.text(pillX, pillY, pillWidth, value(19), stateLabel(state), value(10), 700, accent, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	closeX, closeY, closeSize := expandedCloseButtonRect(width)
	target.fillEllipse(closeX, closeY, closeSize, closeSize, raised)
	target.line(closeX+value(5), closeY+value(5), closeX+closeSize-value(5), closeY+closeSize-value(5), 1, colorSecondary)
	target.line(closeX+closeSize-value(5), closeY+value(5), closeX+value(5), closeY+closeSize-value(5), 1, colorSecondary)
	center := width / 2
	ringSize := value(78)
	ringTop := value(38)
	drawSessionRing(target, center, ringTop, ringSize, len(items), state)
	target.text(center-ringSize/2, ringTop+value(22), ringSize, value(25), compactCount(len(items)), value(22), 700, colorPrimary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	target.text(center-ringSize/2, ringTop+value(49), ringSize, value(13), "会话", value(10), 400, colorSecondary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	metricTop := value(126)
	drawWideMetric(target, value(16), metricTop, value(22), "↑", "运行", activeSessions(items), colorWorking, width, value)
	drawWideMetric(target, value(16), metricTop+value(27), value(22), "↓", "等待", waitingSessions(items), colorWaiting, width, value)
	target.fillRect(value(14), value(181), width-value(28), 1, rgba{79, 94, 126, 120})
	target.text(value(14), value(187), width-value(28), value(15), compactSummary(items), value(10), 400, colorSecondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	if len(items) == 0 {
		target.text(value(16), value(222), width-value(32), value(23), "等待终端会话", value(12), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix)
		return
	}
	visible := len(items)
	if visible > 2 {
		visible = 2
	}
	rowHeight := value(20)
	gap := value(2)
	top := value(206)
	for index := 0; index < visible; index++ {
		drawVerticalSession(target, items[index], value(14), top+index*(rowHeight+gap), width-value(28), rowHeight, gap, value)
	}

}

func drawStatusOrb(target *canvas, centerX, top, size int, state session.State, now time.Time, pressed bool) {
	if state == session.StateWaitingPermission || state == session.StateError {
		pulse := alertPulse(state, now)
		outer := size + 6 + int(float64(pulse)*6+0.5)
		outerColor := stateColor(state)
		outerColor.A = 60
		target.fillEllipse(centerX-outer/2, top-3, outer, outer, outerColor)
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

func drawSessionRing(target *canvas, centerX, top, size, count int, state session.State) {
	segments := 24
	activeSegments := count * 2
	if state == session.StateWaitingPermission || state == session.StateError {
		activeSegments = segments
	}
	if activeSegments > segments {
		activeSegments = segments
	}
	if activeSegments < 2 && count > 0 {
		activeSegments = 2
	}
	dot := maxInt(size/20, 2)
	radius := size/2 - dot
	centerY := top + size/2
	muted := rgba{64, 77, 106, 125}
	accent := stateColor(state)
	for index := 0; index < segments; index++ {
		angle := -math.Pi/2 + float64(index)*2*math.Pi/float64(segments)
		x := centerX + int(math.Cos(angle)*float64(radius)+0.5) - dot/2
		y := centerY + int(math.Sin(angle)*float64(radius)+0.5) - dot/2
		value := muted
		if index < activeSegments {
			value = accent
		}
		target.fillEllipse(x, y, dot, dot, value)
	}
	target.fillEllipse(centerX-size/2+5, top+5, size-10, size-10, rgba{7, 12, 25, 225})
}

func drawCompactMetric(target *canvas, left, top, height int, icon, label string, value int, accent rgba, width int, scale func(int) int) {
	target.text(left, top, height, height, icon, maxInt(height*3/4, 9), 700, accent, dtSingleLine|dtVCenter|dtNoPrefix)
	textLeft := left + height + scale(1)
	target.text(textLeft, top, scale(8), height, label, maxInt(height/2, 9), 400, colorSecondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	valueWidth := scale(9)
	target.text(width-valueWidth-scale(1), top, valueWidth, height, strconv.Itoa(value), maxInt(height*3/5, 9), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix)
}

func drawWideMetric(target *canvas, left, top, height int, icon, label string, value int, accent rgba, width int, scale func(int) int) {
	target.text(left, top, height, height, icon, maxInt(height*3/4, 10), 700, accent, dtSingleLine|dtVCenter|dtNoPrefix)
	textLeft := left + height + scale(6)
	target.text(textLeft, top, scale(70), height, label, maxInt(height/2, 10), 400, colorSecondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	valueWidth := scale(28)
	target.text(width-valueWidth-scale(14), top, valueWidth, height, strconv.Itoa(value), maxInt(height*3/5, 11), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix)
}

func drawVerticalSession(target *canvas, current session.Session, left, top, width, height, gap int, scale func(int) int) {
	target.fillRounded(left, top, width, height, 5, rgba{24, 32, 52, 225})
	accent := stateColor(current.State)
	target.fillRounded(left+gap, top+gap, 3, height-gap*2, 2, accent)
	textLeft := left + gap + 7
	stateWidth := scale(52)
	stateX := left + width - gap - stateWidth
	target.text(textLeft, top+2, stateX-textLeft-scale(4), scale(10), current.Source, scale(8), 700, colorPrimary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	target.text(stateX, top+2, stateWidth, scale(10), stateLabel(current.State), scale(8), 700, accent, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	line := fmt.Sprintf("%s · %s", projectLabel(current), reasonLabel(current))
	target.text(textLeft, top+scale(11), width-(textLeft-left)-gap-scale(4), scale(8), line, scale(8), 400, colorSecondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
}

func alertPulse(state session.State, now time.Time) float64 {
	if state != session.StateWaitingPermission && state != session.StateError {
		return 0
	}
	value := math.Sin(float64(now.UnixNano()) / float64(180*time.Millisecond))
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

func activeSessions(items []session.Session) int {
	count := 0
	for _, current := range items {
		if current.State != session.StateIdle && current.State != session.StateClosed {
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

func compactSummary(items []session.Session) string {
	if len(items) == 0 {
		return "等待会话"
	}
	return fmt.Sprintf("%d 个会话 · %s", len(items), stateLabel(session.Aggregate(items)))
}

func footerColor(state session.State) rgba {
	if state == session.StateWaitingPermission {
		return colorWaiting
	}
	if state == session.StateError {
		return colorError
	}
	return colorMuted
}
