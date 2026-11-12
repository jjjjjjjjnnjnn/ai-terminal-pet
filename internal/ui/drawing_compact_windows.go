package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

const dtCenter = 0x0001

// dtRight right-aligns a label inside its box. It is only used by the settings
// window, where values sit on the far side of a row from their title.
const dtRight = 0x0002

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

// renderedKey is the set of inputs that actually change what the panel looks
// like, and where it sits.
//
// Two frames with the same key are pixel-identical, so present() can skip the
// whole raster. Reporters push on every statusline render and most pushes
// repeat the same state, which used to mean a full GDI repaint several times a
// second for no visible change.
type renderedKey struct {
	width    int
	height   int
	digest   uint64
	state    session.State
	count    int
	expanded bool
	pressed  bool
	hovered  bool
	// tuck is scaled by 1000 so the slide, which is continuous, still produces
	// a new key at every visible step.
	tuck int32
	// lang is the display language: switching it repaints every label.
	lang i18n.Lang
	// approveSec is the hottest auto-approve countdown in whole seconds, -1
	// when nothing is pending. It makes the ticking numbers repaint.
	approveSec int64
	// clock is what makes a time-driven frame differ from the last one.
	clock int64
	// theme is the palette name the frame was rasterised with. It belongs in
	// the key for the same reason lang does: assigning colors changes no
	// other field, so without this a theme switch would be skipped by the
	// "same key, same pixels" guard and the panel would keep its old colours.
	theme string
	// perSource colors the expanded rows' source names by terminal, so toggling
	// it has to repaint too.
	perSource bool
	// waitMinute is the minute bucket of the compact strip's duration, and 0
	// when the strip is not showing one. It has to be separate from clock:
	// purePulseFrame deliberately ignores clock (the breath is expressed as
	// window alpha, not as new pixels), so a field that only exists to make the
	// duration text advance would otherwise be thrown away and "12m" would
	// freeze until the next reporter push.
	waitMinute int64
}

func frameKey(width, height int, digest uint64, state session.State, count int, expanded, pressed, hovered bool, tuck float64, animating bool, now time.Time, lang i18n.Lang, approveSec int64, theme string, perSource bool) renderedKey {
	key := renderedKey{
		width:      width,
		height:     height,
		digest:     digest,
		state:      state,
		count:      count,
		expanded:   expanded,
		pressed:    pressed,
		hovered:    hovered,
		tuck:       int32(math.Round(tuck * 1000)),
		lang:       lang,
		approveSec: approveSec,
		theme:      normalizeTheme(theme),
		perSource:  perSource,
	}
	// The compact strip prints a duration only while somebody is waiting, and it
	// is minute-grained, so one repaint per minute is enough. This is the same
	// reasoning as the expanded panel's bucket below - except that here the
	// duration is drawn even while the pulse reuses the bitmap, so it cannot
	// ride on clock (see waitMinute).
	if !expanded && (state == session.StateWaitingPermission || state == session.StateError) {
		key.waitMinute = now.Unix() / 60
	}
	switch {
	case animating:
		// A pulse or a slide is driven by the clock, so every frame really is
		// different.
		key.clock = now.UnixNano()
	case expanded:
		// Relative timestamps are minute-grained; refreshing them every two
		// seconds repaints the largest panel for no visible change.
		key.clock = now.Unix() / 15
	}
	return key
}

// panelState is everything the rasterizer needs about the frame it is about to
// draw. It is a struct rather than a parameter list because that list had grown
// to six booleans and numbers whose order was easy to get wrong at the call
// site - and getting it wrong draws a plausible-looking wrong panel.
type panelState struct {
	Items    []session.Session
	Now      time.Time
	Expanded bool
	Pressed  bool
	Hovered  bool
	// Lang is the display language every label is rendered in.
	Lang i18n.Lang
	// Countdown carries the live auto-approve numbers by session key
	// (see approvalSessionKey). Nil unless auto-approve is on and a
	// permission turn is actually pending.
	Countdown map[string]approvalView
	// Tuck is 0 when the panel is fully on screen and 1 when it has slid all
	// the way into its docked edge.
	Tuck float64
	// Sliver is the width in canvas pixels that stays visible at Tuck 1.
	Sliver int
	// Dock is the edge being slid into, which decides which end of the frame
	// stays on screen.
	Dock dockSide
}

func drawVerticalPanel(target *canvas, state panelState) {
	width := target.width
	height := target.height
	items := state.Items
	aggregate := session.Aggregate(items)
	accent := stateColor(aggregate)
	inset := 0
	if state.Pressed {
		inset = 1
	}
	panel := colors.Panel
	raised := colors.Raised
	// Hover is the only cue that the panel is a control and not just a
	// picture. Lifting the fill as well as the outline matters because in the
	// alert states the outline is tinted, so an outline-only cue was invisible
	// exactly when the panel most wanted to be clicked.
	if state.Hovered {
		panel = colors.PanelHover
	}
	radius := minInt(width/2, 24)
	target.fillRounded(inset, inset, width-inset*2, height-inset*2, radius, panel)
	// The outline is a hairline, not a coloured frame: a fully saturated
	// amber border made the whole compact strip look like a warning light.
	border := colors.Border
	if aggregate == session.StateWaitingPermission || aggregate == session.StateError {
		border = accent
		border.A = 120
	}
	if state.Hovered {
		border.A = uint8(minInt(int(border.A)+55, 255))
	}
	if state.Pressed {
		border.A = uint8(minInt(int(border.A)+30, 255))
	}
	target.outlineRounded(inset, inset, width-inset*2, height-inset*2, radius, 1, border)

	// The silhouette deliberately stays opaque while the panel slides: it is
	// the shape that becomes the docked tab, so fading it would make the strip
	// blink out of existence on the way in. Only the contents fade.
	if state.Tuck < 1 {
		target.setAlphaScale(1 - state.Tuck)
		if state.Expanded {
			drawVerticalExpanded(target, items, state.Now, aggregate, accent, raised, state.Lang, state.Countdown)
		} else {
			drawVerticalCompact(target, items, state.Now, aggregate, accent, state.Pressed, state.Lang)
		}
		target.setAlphaScale(1)
	}
	if state.Tuck > 0 {
		drawDockedTab(target, aggregate, state.Dock, state.Sliver, state.Tuck)
	}
}

// drawDockedTab is what is left of the panel once it has slid into the screen
// edge: a slim spine of state colour on the sliver that stays visible.
//
// This is the whole point of the tuck. The strip still says "something is
// running" from across the room - green for busy, grey for idle - while
// occupying a dozen pixels of the desktop. The states that actually need the
// user never get here at all; wantsTuck vetoes them.
func drawDockedTab(target *canvas, state session.State, dock dockSide, sliver int, tuck float64) {
	if sliver <= 0 || dock == dockNone {
		return
	}
	width := target.width
	height := target.height
	left := 0
	if dock == dockLeft {
		left = width - sliver
	}
	spineWidth := maxInt(sliver*2/5, 3)
	spineLeft := left + (sliver-spineWidth)/2
	top := height / 4
	spineHeight := height / 2
	radius := spineWidth / 2

	target.setAlphaScale(tuck)
	accent := stateColor(state)
	target.fillRounded(spineLeft, top, spineWidth, spineHeight, radius, accent)
	core := colors.Primary
	core.A = 110
	target.fillRounded(spineLeft+spineWidth/4, top, maxInt(spineWidth/2, 1), spineHeight, radius, core)
	target.setAlphaScale(1)
}

// drawVerticalCompact answers three questions in the order a glance asks
// them: what is the state, how many sessions are behind it, and which one
// wants me. Everything else the strip used to carry - a 24-segment ring, two
// arrow-labelled counters, a divider and a chevron - encoded those same two
// facts six different ways, which made the panel dense without making it
// informative.
func drawVerticalCompact(target *canvas, items []session.Session, now time.Time, state session.State, accent rgba, pressed bool, lang i18n.Lang) {
	width := target.width
	unit := float64(width) / compactWidth
	value := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}
	center := width / 2

	drawStatusOrb(target, center, value(14), value(28), state, now, pressed)

	// The big slot answers whichever question the panel is actually asking.
	// While somebody is blocked on the user it shows HOW LONG ("12m" / "12分",
	// 16-bold: the widest case measured is "99小时+" at 44 px, and fitText is
	// still applied as a guard); the count moves down to the caption, so the
	// strip reads 12m / 3 sessions / Waiting / <project>. Everything busy keeps
	// the original count-then-noun pair, unchanged.
	if waited, ok := compactWait(items, now, lang); ok {
		target.text(value(4), value(50), width-value(8), value(26), target.fitText(waited, value(16), 700, width-value(8)), value(16), 700, colors.Primary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
		noun := fmt.Sprintf(i18n.T(lang, i18n.KeySummarySessions), len(items))
		target.text(value(4), value(76), width-value(8), value(12), target.fitText(noun, value(9), 400, width-value(8)), value(9), 400, colors.Secondary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	} else {
		target.text(value(4), value(50), width-value(8), value(26), compactCount(len(items)), value(22), 700, colors.Primary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
		target.text(value(4), value(76), width-value(8), value(12), i18n.T(lang, i18n.KeyPanelSessions), value(9), 400, colors.Secondary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	}

	target.fillRect(value(12), value(96), width-value(24), 1, colors.Divider)

	target.text(value(2), value(102), width-value(4), value(16), stateLabel(state, lang), value(11), 700, accent, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)

	// When a session needs a decision - or has finished and is waiting to be
	// read - the subject line becomes a one-tap jump bar. It still names the
	// project, so "who wants me" is not lost. A panel with everything busy
	// keeps the plain subject line and its original silhouette.
	if jump, ok := mostUrgentJump(items); ok {
		// Keep "who wants me" visible: lift the subject line a few units to
		// make room for the one-tap bar, then draw the bar below it.
		subject, subjectColor := compactSubject(items, state, lang)
		subject = target.fitText(subject, value(8), 400, width-value(8))
		target.text(value(4), value(119), width-value(8), value(13), subject, value(8), 400, subjectColor, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
		drawCompactJump(target, jump, lang)
		return
	}
	subject, subjectColor := compactSubject(items, state, lang)
	subject = target.fitText(subject, value(9), 400, width-value(8))
	target.text(value(4), value(124), width-value(8), value(14), subject, value(9), 400, subjectColor, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}

// compactWait renders the longest wait among the sessions that want the user.
//
// Only waiting_permission and error qualify. Those are the states where "how
// long" answers a question the user is asking; working / tool_use / thinking are
// busy (nobody is waiting on anybody), and idle / closed carry a timestamp that
// means "when it finished", not "how long it held me up".
//
// The OLDEST such session wins, not the highest-priority one: when a permission
// prompt and a failed build are both on screen, the prompt is the one that has
// been blocking longest, and that is the fact a glance should return.
func compactWait(items []session.Session, now time.Time, lang i18n.Lang) (string, bool) {
	var oldest time.Time
	for _, current := range items {
		if current.State != session.StateWaitingPermission && current.State != session.StateError {
			continue
		}
		stamp := current.UpdatedAt
		if stamp.IsZero() {
			// A snapshot without a timestamp still deserves the slot; it just
			// cannot contribute a duration, so it counts as "right now" rather
			// than dropping the whole answer.
			stamp = now
		}
		if oldest.IsZero() || stamp.Before(oldest) {
			oldest = stamp
		}
	}
	if oldest.IsZero() {
		return "", false
	}
	return shortElapsed(now.Sub(oldest), lang), true
}

// shortElapsed is the strip's compact duration. Its bucket boundaries are
// relativeTime's on purpose: two formatters with their own thresholds is how a
// panel ends up saying "刚刚" in one place and "1 分钟前" in another for the same
// instant.
//
// Anything under a minute collapses to the smallest label, because "0分" reads
// like a bug while "1分" is merely cautious.
func shortElapsed(delta time.Duration, lang i18n.Lang) string {
	if delta < 0 {
		delta = 0
	}
	switch {
	case delta < time.Minute:
		return i18n.T(lang, i18n.KeyTimeShortMin1)
	case delta < time.Hour:
		// 59 minutes is the longest text in this bucket: "59分" is 36 px at
		// 16-bold, comfortably inside the 48 the strip has between its insets.
		return fmt.Sprintf(i18n.T(lang, i18n.KeyTimeShortMins), int(delta.Minutes()))
	case delta < 2*time.Hour:
		return i18n.T(lang, i18n.KeyTimeShortHour1)
	default:
		hours := int(delta.Hours())
		if hours > 99 {
			// A week-old snapshot is not informative at this size, and an
			// unbounded integer would eventually overrun the slot. 99+ is the
			// ceiling; the long form in the expanded list still says the exact
			// number.
			hours = 99
		}
		return fmt.Sprintf(i18n.T(lang, i18n.KeyTimeShortHours), hours)
	}
}

// compactSubject picks the one line of text worth showing under the count:
// the session that wants the user when one does, otherwise the most recently
// active one, otherwise a hint at what to do about the empty panel.
//
// items arrive priority-sorted from the store, so the first match is already
// the most urgent. The label is deliberately not tinted with the state colour:
// the word directly above it already carries that, and colouring both made the
// strip read as one amber smear instead of two distinct facts.
func compactSubject(items []session.Session, state session.State, lang i18n.Lang) (string, rgba) {
	if len(items) == 0 {
		return i18n.T(lang, i18n.KeyPanelWaiting), colors.Muted
	}
	if state == session.StateWaitingPermission || state == session.StateError {
		for _, current := range items {
			if current.State == session.StateWaitingPermission || current.State == session.StateError {
				return projectLabel(current, lang), colors.Primary
			}
		}
	}
	return projectLabel(items[0], lang), colors.Secondary
}

func drawVerticalExpanded(target *canvas, items []session.Session, now time.Time, state session.State, accent, raised rgba, lang i18n.Lang, countdown map[string]approvalView) {
	width := target.width
	height := target.height
	unit := float64(width) / expandedWidth
	value := func(base int) int {
		return maxInt(int(float64(base)*unit+0.5), 1)
	}

	// Header: title, live state, close.
	target.text(value(14), value(13), width-value(120), value(22), i18n.T(lang, i18n.KeyPanelTitle), value(13), 700, colors.Primary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	pillWidth := value(74)
	pillX := width - pillWidth - value(40)
	pillY := value(13)
	pill := accent
	pill.A = 46
	target.fillRounded(pillX, pillY, pillWidth, value(20), 10, pill)
	target.text(pillX, pillY, pillWidth, value(20), stateLabel(state, lang), value(10), 700, accent, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)

	closeX, closeY, closeSize := expandedCloseButtonRect(width)
	target.fillEllipse(closeX, closeY, closeSize, closeSize, raised)
	target.line(closeX+value(6), closeY+value(6), closeX+closeSize-value(6), closeY+closeSize-value(6), 1, colors.Secondary)
	target.line(closeX+closeSize-value(6), closeY+value(6), closeX+value(6), closeY+closeSize-value(6), 1, colors.Secondary)

	target.fillRect(value(14), value(expandedHeaderHeight-6), width-value(28), 1, colors.Divider)
	target.text(value(14), value(expandedHeaderHeight), width-value(28), value(expandedSummaryHeight), summaryLine(items, lang), value(9), 400, colors.Secondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)

	top := value(expandedHeaderHeight + expandedSummaryHeight)
	if len(items) == 0 {
		drawExpandedEmpty(target, top, width, value, lang)
	} else {
		rows := minInt(len(items), expandedMaxRows)
		for index := 0; index < rows; index++ {
			view, ok := lookupCountdown(countdown, items[index])
			drawVerticalSession(target, items[index], value(14), top+index*value(expandedRowHeight), width-value(28), value(expandedRowHeight)-value(4), now, value, lang, view, ok)
		}
		if hidden := len(items) - rows; hidden > 0 {
			y := top + rows*value(expandedRowHeight)
			target.text(value(14), y, width-value(28), value(expandedOverflowHeight), fmt.Sprintf(i18n.T(lang, i18n.KeyPanelMore), hidden), value(8), 400, colors.Muted, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
		}
	}

	// The footer teaches the controls once, so the panel does not need a
	// visible handle or a chevron to advertise that it is interactive.
	footerTop := height - value(expandedFooterHeight)
	softDivider := colors.Divider
	softDivider.A = 58
	target.fillRect(value(14), footerTop-value(6), width-value(28), 1, softDivider)
	hint := target.fitText(i18n.T(lang, i18n.KeyPanelHint), value(8), 400, width-value(24))
	target.text(value(12), footerTop, width-value(24), value(expandedFooterHeight), hint, value(8), 400, colors.Muted, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}

func drawExpandedEmpty(target *canvas, top, width int, scale func(int) int, lang i18n.Lang) {
	// The summary line already says there are no sessions, so this block
	// carries the actionable half instead of repeating it.
	target.text(scale(16), top+scale(6), width-scale(32), scale(34), i18n.T(lang, i18n.KeyPanelEmpty), scale(9), 400, colors.Muted, dtVCenter|dtCenter|dtNoPrefix|dtWordBreak)
}

func drawStatusOrb(target *canvas, centerX, top, size int, state session.State, now time.Time, pressed bool) {
	accent := stateColor(state)
	if state == session.StateWaitingPermission || state == session.StateError {
		pulse := alertPulse(state, now)
		outer := size + 4 + int(float64(pulse)*5+0.5)
		outerColor := accent
		outerColor.A = uint8(18 + float64(pulse)*22)
		target.fillEllipse(centerX-outer/2, top-(outer-size)/2, outer, outer, outerColor)
	}
	// A thin ring of state colour around a dark core, instead of a solid
	// disc. The old 3.5px saturated ring plus the pulse halo filled the top
	// third of the strip with one colour; a hairline reads as an indicator.
	ring := accent
	ring.A = 225
	if pressed {
		ring.A = 200
	}
	target.fillEllipse(centerX-size/2, top, size, size, ring)
	inner := size - 4
	target.fillEllipse(centerX-inner/2, top+2, inner, inner, colors.OrbInner)
	target.text(centerX-inner/2, top+2, inner, inner, statusGlyph(state), maxInt(size/2, 11), 700, accent, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}

// lookupCountdown finds the live auto-approve numbers for one session, if
// auto-approve is on and this exact permission turn is still pending.
func lookupCountdown(countdown map[string]approvalView, current session.Session) (approvalView, bool) {
	if len(countdown) == 0 {
		return approvalView{}, false
	}
	view, ok := countdown[approvalSessionKey(current.Source, current.SessionID, current.PID)]
	return view, ok
}

// drawVerticalSession is one row of the expanded list: project and state on
// the first line, where it came from and how long ago it spoke on the second.
// The relative time is the part that makes the list actionable - "waiting for
// approval" reads very differently at 10 seconds and at 40 minutes.
//
// While an auto-approve countdown runs for the session, the second line's
// right half becomes the ticking Allow (Ns) plus a draining bar, so the
// user sees exactly how long they have to veto.
func drawVerticalSession(target *canvas, current session.Session, left, top, width, height int, now time.Time, scale func(int) int, lang i18n.Lang, view approvalView, hasCountdown bool) {
	target.fillRounded(left, top, width, height, 6, colors.RaisedRow)
	accent := stateColor(current.State)
	// The rail is the one place a row uses a large patch of colour. Keeping it
	// slightly translucent stops it from overpowering the neutral panel.
	rail := accent
	rail.A = 190
	target.fillRounded(left+scale(1), top+scale(1), scale(3), height-scale(2), 2, rail)

	textLeft := left + scale(9)
	stateWidth := scale(48)
	// Every row carries its own jump arrow. The geometry comes from the
	// shared helper so the painted button and the clickable button coincide.
	focusX, focusY, focusSize := expandedRowFocusRect(target.width, top, height)
	stateX := left + width - stateWidth - focusSize - scale(8)

	target.text(textLeft, top+scale(3), stateX-textLeft-scale(6), scale(12), projectLabel(current, lang), scale(10), 700, colors.Primary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	target.text(stateX, top+scale(3), stateWidth, scale(12), stateLabel(current.State, lang), scale(8), 700, accent, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)

	detail := fmt.Sprintf("%s · %s", current.Source, reasonLabel(current, lang))
	detailWidth := stateX - textLeft - scale(6)
	if perSourceColors && current.Source != "" {
		// Two draws instead of one: the source name takes its provenance
		// colour, the reason stays neutral. The split width comes from the
		// same measure() the ellipsis uses, so a long source name still
		// truncates instead of pushing the reason off the right edge.
		source := target.fitText(current.Source, scale(8), 400, detailWidth)
		sourceWidth := target.measure(source, scale(8), 400)
		target.text(textLeft, top+scale(17), sourceWidth+scale(4), scale(10), source, scale(8), 700, sourceColor(current.Source), dtSingleLine|dtVCenter|dtNoPrefix)
		rest := fmt.Sprintf(" · %s", reasonLabel(current, lang))
		if remain := detailWidth - sourceWidth; remain > scale(10) {
			target.text(textLeft+sourceWidth, top+scale(17), remain, scale(10), target.fitText(rest, scale(8), 400, remain), scale(8), 400, colors.Secondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
		}
	} else {
		target.text(textLeft, top+scale(17), detailWidth, scale(10), detail, scale(8), 400, colors.Secondary, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	}
	stamp := relativeTime(current.UpdatedAt, now, lang)
	if hasCountdown {
		stamp = fmt.Sprintf(i18n.T(lang, i18n.KeyApproveCountdown), view.RemainingSec)
	}
	target.text(stateX, top+scale(17), stateWidth, scale(10), stamp, scale(8), 400, colors.Muted, dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
	if hasCountdown && view.Fraction > 0 {
		barWidth := int(float64(width-scale(18)) * view.Fraction)
		if barWidth > 0 {
			bar := accent
			bar.A = 160
			target.fillRounded(textLeft, top+height-scale(3), barWidth, maxInt(scale(2), 1), 1, bar)
		}
	}
	drawExpandedFocusButton(target, focusX, focusY, focusSize, accent)
}

// summaryLine is the one-line census under the header. "运行" deliberately
// excludes sessions that are waiting on the user: from the user's side those
// are not running, they are blocking, and counting them as both made the old
// metric rows double-count.
func summaryLine(items []session.Session, lang i18n.Lang) string {
	if len(items) == 0 {
		return i18n.T(lang, i18n.KeySummaryEmpty)
	}
	parts := []string{fmt.Sprintf(i18n.T(lang, i18n.KeySummarySessions), len(items))}
	if busy := busySessions(items); busy > 0 {
		parts = append(parts, fmt.Sprintf(i18n.T(lang, i18n.KeySummaryRunning), busy))
	}
	if waiting := waitingSessions(items); waiting > 0 {
		parts = append(parts, fmt.Sprintf(i18n.T(lang, i18n.KeySummaryWaiting), waiting))
	}
	return strings.Join(parts, " · ")
}

// relativeTime renders coarse buckets rather than a clock: sessions live for
// minutes, and "3 分钟前" is read faster than "14:07".
func relativeTime(value, now time.Time, lang i18n.Lang) string {
	if value.IsZero() {
		return ""
	}
	delta := now.Sub(value)
	if delta < 0 {
		delta = 0
	}
	switch {
	case delta < 45*time.Second:
		return i18n.T(lang, i18n.KeyTimeNow)
	case delta < 90*time.Second:
		return i18n.T(lang, i18n.KeyTimeMin1)
	case delta < time.Hour:
		return fmt.Sprintf(i18n.T(lang, i18n.KeyTimeMins), int(delta.Minutes()))
	case delta < 2*time.Hour:
		return i18n.T(lang, i18n.KeyTimeHour1)
	default:
		return fmt.Sprintf(i18n.T(lang, i18n.KeyTimeHours), int(delta.Hours()))
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
