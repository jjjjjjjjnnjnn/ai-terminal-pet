package ui

// The settings window's platform-neutral half: the tab table, the row shape,
// the geometry, the hit test, the keyboard focus ring and the drawing.
//
// It has no build tag on purpose. These are the rules that decide what a click
// on a slider means and where the focus ring goes, and they lived in
// settings_windows.go only because the window did - which meant the tests
// pinning them ran on Windows alone. What genuinely needs a window (class
// registration, DPI, the message pump, the layered surface) stays in
// settings_windows.go, together with the six adapters that read App state.

import (
	"fmt"
	"math"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
)

const (
	settingsClassName   = "AITerminalPetSettings"
	settingsWindowTitle = "AI Terminal Pet — Settings"

	// Logical layout at 96 DPI, at the standard text size.
	settingsWidth = 500
	// settingsMinHeight keeps the sidebar (four tabs plus margins) from being
	// clipped on the shortest tab; the real height follows the content.
	settingsMinHeight    = 280
	settingsHeaderHeight = 56
	settingsMargin       = 20
	settingsSidebarWidth = 112
	settingsTabsTop      = 70
	settingsTabHeight    = 36
	settingsTabGap       = 6
	settingsContentLeft  = 176
	settingsRowsTop      = 70
	settingsPanelRadius  = 14

	// The footer: two links and the version, under a divider. It exists because
	// both are things a user cannot reach from inside the app - the repository is
	// where the issue tracker lives, and a version string is the first line of any
	// useful bug report. Nothing here talks to the network; a click opens the
	// user's browser and nothing else happens.
	settingsFooterGap    = 10
	settingsFooterHeight = 26
	settingsLinkWidth    = 150
	settingsIssueWidth   = 130
)

// settingsContentWidth is the width of the right-hand column.
const settingsContentWidth = settingsWidth - settingsMargin - settingsContentLeft

// Settings text sizes are in textsizes.go: config.json stores those names
// (internal/config owns the list), the ui package compares against them, and a
// test in internal/app pins the two together.

// settingsRowsTop and the text-size names live in textsizes.go / settings_rows.go: they
// are config names and layout numbers, not window state. Keeping them here made
// paint_cost_test.go unbuildable on a platform without a window layer.

// settingsTypeRamp is the type scale of the settings window: one requested cell
// height per role, in logical pixels at 96 DPI.
//
// These numbers are not arbitrary and not a single table multiplied by a factor.
// GDI rounds a requested cell height onto the face's own ladder, measured on
// this machine as: requests 9-10 render 7 px glyphs, 11-13 render 8 px, 14-15
// render 9 px, 16 renders 10 px, 17-18 render 12 px, 20 renders 14 px. So
//
//   - a 1 px ramp step is usually a no-op. The first version of this window used
//     15/12/12/10 and actually rendered only two sizes (9 px and 8 px), which is
//     why it read as flat as well as small;
//   - scaling one table by 1.2 collapses two roles onto one rung (a 12 px title
//     and a 9 px hint both land on 48), so the three ramps below are hand-picked
//     on distinct rungs instead.
//
// TestSettingsTypeRampStaysOrdered measures every ramp at four DPIs and fails if
// the roles stop being visually ordered.
type settingsTypeRamp struct {
	header   int
	subtitle int
	tab      int
	title    int
	hint     int
	note     int
	value    int
	segment  int
	button   int
}

// settingsSpacing is how much room each text size gets. It scales margins, row
// heights and the window width, never the glyphs themselves.
var settingsSpacing = map[string]float64{
	textSizeSmall:    0.9,
	textSizeStandard: 1.0,
	textSizeLarge:    1.2,
}

var settingsTypeRamps = map[string]settingsTypeRamp{
	// 8 px hint, 10 px title, 12 px header.
	textSizeSmall: {header: 18, subtitle: 11, tab: 13, title: 16, hint: 12, note: 12, value: 13, segment: 13, button: 13},
	// 9 px hint, 10 px title, 14 px header - the shipped default.
	textSizeStandard: {header: 20, subtitle: 12, tab: 14, title: 16, hint: 14, note: 14, value: 14, segment: 14, button: 14},
	// 12 px hint, 14 px title, 17 px header.
	textSizeLarge: {header: 24, subtitle: 15, tab: 18, title: 20, hint: 17, note: 17, value: 18, segment: 17, button: 17},
}

// normalizeTextSize maps a persisted or hand-edited name onto one of the three
// ramps. The control id settingsTextSize is a different thing, hence the verb.
func normalizeTextSize(value string) string {
	if _, ok := settingsTypeRamps[value]; ok {
		return value
	}
	return textSizeStandard
}

// Row kinds and their height arithmetic live in settings_rows.go: the layout is
// platform-neutral and settings_text.go needs it without a window.

func settingsRowHeight(kind, hintLines int) int {
	if hintLines < 1 {
		hintLines = 1
	}
	base := settingsTitleBand + hintLines*settingsHintLine + settingsRowPad
	switch kind {
	case rowSlider:
		// The track sits in its own band under the hint.
		return base + settingsBandGap + 4 + 8
	case rowSegment:
		return base + settingsBandGap + 30
	case rowNote:
		// A note has no title line and no control.
		return hintLines*settingsHintLine + settingsRowPad
	default:
		return base
	}
}

// settingsControl identifies every interactive element. Values are stable and
// are used as map keys for hover/press state, so they must never be reused for
// a different meaning.
type settingsControl int

const (
	settingsNone settingsControl = iota
	settingsTabGeneral
	settingsTabSound
	settingsTabSessions
	settingsTabAppearance
	settingsClose

	settingsNotifications
	settingsAlwaysOnTop
	settingsEdgeRetract
	settingsOpacity
	settingsLanguage
	settingsAutoStart

	settingsSound
	settingsScheme
	settingsVolume
	settingsSoundWaiting
	settingsSoundError
	settingsSoundFinish
	settingsPreview

	settingsTheme
	settingsPerSource
	settingsHiddenAtLaunch
	settingsAnimations

	// How the settings window itself behaves. They live on the two tabs the
	// window is about: text size on Look (it is a look), keyboard and tooltips
	// on General (they are interaction).
	settingsTextSize
	settingsKeyboard
	settingsTooltips

	settingsAutoApprove
	settingsApproveDown
	settingsApproveUp
	settingsReset

	// Attentiveness: the rows that decide whether an alert is useful. They
	// share the Sessions tab with auto-approve, because both answer "what does
	// this thing do while I am not watching".
	settingsForegroundAware
	settingsIdleMinutes
	settingsDonePolicy
	settingsQuietCurrent
	settingsQuietSubagents
	settingsMuteClear
	settingsTestWaiting
	settingsTestDone

	// The footer links. They are controls rather than drawing because they have
	// to be hoverable, focusable by Tab and clickable, and every one of those is
	// what a settingsControl already means.
	settingsStar
	settingsIssue
)

// settingsTabControl maps a sidebar index onto its control id.
func settingsTabControl(index int) settingsControl {
	switch index {
	case 1:
		return settingsTabSound
	case 2:
		return settingsTabSessions
	case 3:
		return settingsTabAppearance
	default:
		return settingsTabGeneral
	}
}

// settingsTabIndex maps a control id back onto the sidebar index.
func settingsTabIndex(control settingsControl) int {
	switch control {
	case settingsTabSound:
		return 1
	case settingsTabSessions:
		return 2
	case settingsTabAppearance:
		return 3
	default:
		return 0
	}
}

// settingsTabCount is how many sidebar entries exist. It is a constant rather
// than len(geometry.tabs) so the geometry, the labels and the tests cannot
// disagree about how many tabs a build has.
const settingsTabCount = 4

// settingsRow is one line of the right-hand column.
type settingsRow struct {
	control settingsControl
	// alt is the second control a stepper needs (its minus button). It is
	// settingsNone for every other kind.
	alt   settingsControl
	kind  int
	title string
	hint  string
	// hintLines is how many lines the hint needs at the current width. It is
	// filled in by measureSettingsRows, because it depends on the font and the
	// column width; the row builders leave it zero and settingsRowHeight reads
	// that as one line.
	hintLines int
	// action is the caption inside a rowButton's button. Buttons carry their
	// own verb so the row title can stay descriptive.
	action  string
	options []string
	// values holds the slider range for rowSlider rows.
	minimum int
	maximum int
}

// settingsView is a plain snapshot of what the window displays. It is
// comparable on purpose: the frame key compares two of them to decide whether a
// repaint is needed at all.
type settingsView struct {
	Notifications  bool
	AlwaysOnTop    bool
	EdgeRetract    bool
	Sound          bool
	SoundOnWaiting bool
	SoundOnError   bool
	NotifyOnFinish bool
	AutoApprove    bool
	AutoStart      bool
	Opacity        int
	Volume         int
	ApproveTimeout int
	Scheme         string
	Language       string
	Theme          string
	PerSource      bool
	HiddenAtLaunch bool
	Animations     bool
	TextSize       string
	Keyboard       bool
	Tooltips       bool
	// Attentiveness. DonePolicy is the persisted string so the segment and
	// the tick can both be derived from one field.
	ForegroundAware bool
	IdleMinutes     int
	DonePolicy      string
	QuietCurrent    bool
	QuietSubagents  bool
	MuteCount       int
}

type settingsRect struct {
	Left   int
	Top    int
	Width  int
	Height int
}

func (r settingsRect) right() int  { return r.Left + r.Width }
func (r settingsRect) bottom() int { return r.Top + r.Height }

func (r settingsRect) contains(x, y int) bool {
	return x >= r.Left && x < r.right() && y >= r.Top && y < r.bottom()
}

func (r settingsRect) inset(by int) settingsRect {
	return settingsRect{Left: r.Left + by, Top: r.Top + by, Width: r.Width - by*2, Height: r.Height - by*2}
}

// settingsGeometry is the resolved layout in device pixels. Drawing and hit
// testing both read it, which is what keeps the two in step.
type settingsGeometry struct {
	width  int
	height int
	dpi    uint32
	// textSize is the persisted name this layout was built for, and fonts is the
	// ramp it resolved to. Both are carried so a repaint can tell "the user
	// changed the text size" from "the window merely moved".
	textSize string
	fonts    settingsTypeRamp
	// lang is the language the row labels were built in. Switching language
	// changes every string but not the layout, so the geometry has to be
	// rebuilt to pick the new labels up.
	lang  i18n.Lang
	close settingsRect
	tabs  [settingsTabCount]settingsRect
	tab   int
	rows  []settingsRow
	// frames is the clickable band of each row (the whole line, so a label is
	// as clickable as its switch), controls the drawn sub-control inside it.
	frames   []settingsRect
	controls []settingsRect
	// contentHeight is where the last row ends, which is what "does this tab fit
	// the screen" is measured against.
	contentHeight int
	// footer is the band under the last row, star and issue its two links. The
	// build string is not here: the footer is one 304 px column and a third item
	// would have squeezed the links to nothing. It goes in the header subtitle
	// instead, where there is room and where it reads as "what am I looking at".
	footer settingsRect
	star   settingsRect
	issue  settingsRect
}

// scale converts a logical dimension to device pixels at the current DPI and
// text size.
func (g settingsGeometry) scale(base int) int {
	spacing := g.spacing()
	return maxInt(int(math.Round(float64(base)*float64(g.dpi)/96*spacing)), 1)
}

// font converts a logical type size to device pixels. It deliberately does not
// apply the spacing factor: the ramp table already says how big each role is at
// this text size, and multiplying it again would land on the wrong GDI rung.
func (g settingsGeometry) font(base int) int {
	return maxInt(int(math.Round(float64(base)*float64(g.dpi)/96)), 1)
}

func (g settingsGeometry) spacing() float64 {
	if factor, ok := settingsSpacing[g.textSize]; ok {
		return factor
	}
	return 1
}

// settingsGeometryFor resolves the layout at the standard text size. Kept as the
// default entry point so the layout tests read as "the shipped layout".
func settingsGeometryFor(dpi uint32, tab int, rows []settingsRow) settingsGeometry {
	return settingsGeometrySized(dpi, textSizeStandard, tab, rows)
}

// settingsGeometrySized resolves the layout for one text size. Pure: same
// inputs, same output, no DPI query, which is what makes it testable off a real
// window.
func settingsGeometrySized(dpi uint32, textSize string, tab int, rows []settingsRow) settingsGeometry {
	if dpi == 0 {
		dpi = 96
	}
	size := normalizeTextSize(textSize)
	geometry := settingsGeometry{
		dpi:      dpi,
		textSize: size,
		fonts:    settingsTypeRamps[size],
		tab:      tab,
		rows:     rows,
	}
	geometry.width = geometry.scale(settingsWidth)
	s := geometry.scale

	closeSize := s(24)
	geometry.close = settingsRect{
		Left:   geometry.width - s(settingsMargin) - closeSize,
		Top:    s(14),
		Width:  closeSize,
		Height: closeSize,
	}
	for index := 0; index < settingsTabCount; index++ {
		geometry.tabs[index] = settingsRect{
			Left:   s(settingsMargin),
			Top:    s(settingsTabsTop) + index*s(settingsTabHeight+settingsTabGap),
			Width:  s(settingsSidebarWidth),
			Height: s(settingsTabHeight),
		}
	}

	contentLeft := s(settingsContentLeft)
	contentWidth := s(settingsContentWidth)
	y := s(settingsRowsTop)
	geometry.frames = make([]settingsRect, 0, len(rows))
	geometry.controls = make([]settingsRect, 0, len(rows))
	for _, row := range rows {
		height := s(settingsRowHeight(row.kind, row.hintLines))
		frame := settingsRect{Left: contentLeft, Top: y, Width: contentWidth, Height: height}
		geometry.frames = append(geometry.frames, frame)
		geometry.controls = append(geometry.controls, settingsControlRect(frame, row, s))
		y += height
	}
	geometry.contentHeight = y
	footerTop := y + s(settingsFooterGap)
	geometry.footer = settingsRect{Left: contentLeft, Top: footerTop, Width: contentWidth, Height: s(settingsFooterHeight)}
	linkTop := footerTop
	geometry.star = settingsRect{Left: contentLeft, Top: linkTop, Width: s(settingsLinkWidth), Height: s(settingsFooterHeight)}
	linkLeft := contentLeft + s(settingsLinkWidth) + s(8)
	geometry.issue = settingsRect{
		Left:   linkLeft,
		Top:    linkTop,
		Width:  s(settingsIssueWidth),
		Height: s(settingsFooterHeight),
	}
	// The window is exactly as tall as its content, and the footer is part of
	// the content: it used to be cut off on the shortest tab (Sessions) when the
	// height followed the rows alone.
	geometry.height = maxInt(
		geometry.footer.bottom()+geometry.scale(settingsMargin),
		geometry.scale(settingsMinHeight))
	return geometry
}

// settingsFitSizes is the order the layout is tried in when the preferred text
// size makes the window taller than the screen: the choice is a preference, the
// work area is a constraint, and a control nobody can reach is worse than text
// one step smaller.
var settingsFitSizes = []string{textSizeLarge, textSizeStandard, textSizeSmall}

// settingsGeometryFitting picks the largest text size at or below the preferred
// one whose window still fits maxHeight, and falls back to the smallest size
// when even that does not fit (the window is then taller than the screen, but
// nothing else can be done without scrolling).
func settingsGeometryFitting(dpi uint32, textSize string, tab int, rows []settingsRow, maxHeight int) settingsGeometry {
	preferred := normalizeTextSize(textSize)
	start := 0
	for index, name := range settingsFitSizes {
		if name == preferred {
			start = index
			break
		}
	}
	var smallest settingsGeometry
	for index := start; index < len(settingsFitSizes); index++ {
		geometry := settingsGeometrySized(dpi, settingsFitSizes[index], tab, rows)
		smallest = geometry
		if maxHeight <= 0 || geometry.height <= maxHeight {
			return geometry
		}
	}
	return smallest
}

// settingsControlRect places the interactive part of a row. A toggle and a
// stepper sit on the title line at the right edge; a slider, a segment and a
// button get their own band underneath it.
func settingsControlRect(frame settingsRect, row settingsRow, s func(int) int) settingsRect {
	titleLine := s(settingsTitleBand)
	switch row.kind {
	case rowToggle:
		width, height := s(40), s(22)
		return settingsRect{Left: frame.right() - width, Top: frame.Top + s(1), Width: width, Height: height}
	case rowSlider:
		return settingsRect{Left: frame.Left, Top: frame.Top + titleLine + s(settingsHintLine*row.hintLines+settingsBandGap), Width: frame.Width, Height: s(4)}
	case rowSegment:
		return settingsRect{Left: frame.Left, Top: frame.Top + titleLine + s(settingsHintLine*row.hintLines+settingsBandGap), Width: frame.Width, Height: s(30)}
	case rowStepper:
		return settingsRect{Left: frame.right() - s(96), Top: frame.Top, Width: s(96), Height: s(26)}
	case rowButton:
		width := s(84)
		return settingsRect{Left: frame.right() - width, Top: frame.Top, Width: width, Height: s(26)}
	default:
		return frame
	}
}

// hit resolves a device-pixel point onto a control. The returned index is the
// segment option under the cursor, or -1.
func (g settingsGeometry) hit(x, y int) (settingsControl, int, settingsRect, bool) {
	if g.close.contains(x, y) {
		return settingsClose, -1, g.close, true
	}
	for index := range g.tabs {
		if g.tabs[index].contains(x, y) {
			return settingsTabControl(index), -1, g.tabs[index], true
		}
	}
	if g.star.contains(x, y) {
		return settingsStar, -1, g.star, true
	}
	if g.issue.contains(x, y) {
		return settingsIssue, -1, g.issue, true
	}
	for index := range g.rows {
		if !g.frames[index].contains(x, y) {
			continue
		}
		row := g.rows[index]
		switch row.kind {
		case rowToggle:
			return row.control, -1, g.controls[index], true
		case rowSlider:
			return row.control, -1, g.controls[index], true
		case rowSegment:
			control := g.controls[index]
			count := len(row.options)
			if count == 0 {
				return settingsNone, -1, control, false
			}
			option := (x - control.Left) * count / maxInt(control.Width, 1)
			if option < 0 {
				option = 0
			}
			if option >= count {
				option = count - 1
			}
			return row.control, option, control, true
		case rowStepper:
			control := g.controls[index]
			if x < control.Left+control.Width/2 {
				return row.alt, -1, control, true
			}
			return row.control, -1, control, true
		case rowButton:
			return row.control, -1, g.controls[index], true
		}
	}
	return settingsNone, -1, settingsRect{}, false
}

// settingsFocusOrder is the order Tab walks: the close button, the four sidebar
// tabs, then every interactive row in reading order. Notes are skipped because
// there is nothing to activate on them.
func settingsFocusOrder(geometry settingsGeometry) []settingsControl {
	order := make([]settingsControl, 0, settingsTabCount+1+len(geometry.rows))
	order = append(order, settingsClose)
	for index := 0; index < settingsTabCount; index++ {
		order = append(order, settingsTabControl(index))
	}
	for _, row := range geometry.rows {
		if row.control == settingsNone || row.kind == rowNote {
			continue
		}
		order = append(order, row.control)
	}
	// The footer links are reachable by Tab too: a keyboard user has the same
	// reason to want the bug tracker as a mouse user does.
	order = append(order, settingsStar, settingsIssue)
	return order
}

// settingsFocusControl is the control at one focus index, or settingsNone.
func settingsFocusControl(geometry settingsGeometry, index int) settingsControl {
	order := settingsFocusOrder(geometry)
	if index < 0 || index >= len(order) {
		return settingsNone
	}
	return order[index]
}

// settingsFocusRect is where the focus ring goes for one focus index.
func settingsFocusRect(geometry settingsGeometry, index int) (settingsRect, bool) {
	control := settingsFocusControl(geometry, index)
	if control == settingsNone {
		return settingsRect{}, false
	}
	if control == settingsClose {
		return geometry.close, true
	}
	if tab := settingsTabIndex(control); tab >= 0 && settingsTabControl(tab) == control {
		return geometry.tabs[tab], true
	}
	for rowIndex, row := range geometry.rows {
		if row.control == control || row.alt == control {
			return geometry.controls[rowIndex], true
		}
	}
	return settingsRect{}, false
}

// moveSettingsFocus walks the focus ring. delta is +1 for Tab and -1 for
// Shift+Tab; wrapping is intentional, so Tab never dead-ends.
func settingsIndexForControl(geometry settingsGeometry, control settingsControl, view settingsView) int {
	for _, row := range geometry.rows {
		if row.control != control {
			continue
		}
		if row.kind == rowSegment {
			if len(row.options) == 0 {
				return -1
			}
			return (settingsSegmentIndex(row, view) + 1) % len(row.options)
		}
		if row.kind == rowStepper {
			return 1
		}
	}
	return -1
}

// settingsSliderFraction turns an x coordinate into a 0..1 position on a track.
func settingsSliderFraction(track settingsRect, x int) float64 {
	if track.Width <= 0 {
		return 0
	}
	fraction := float64(x-track.Left) / float64(track.Width)
	if fraction < 0 {
		return 0
	}
	if fraction > 1 {
		return 1
	}
	return fraction
}

// settingsSliderValue maps an x coordinate onto the row's integer range.
func settingsSliderValue(track settingsRect, row settingsRow, x int) int {
	span := row.maximum - row.minimum
	if span <= 0 {
		return row.minimum
	}
	value := row.minimum + int(math.Round(settingsSliderFraction(track, x)*float64(span)))
	if value < row.minimum {
		value = row.minimum
	}
	if value > row.maximum {
		value = row.maximum
	}
	return value
}

// settingsSegmentRect is the sub-rect of one option inside a segment control,
// so the selected pill and the hover highlight line up with the hit test.
func settingsSegmentRect(control settingsRect, count, index int) settingsRect {
	if count <= 0 {
		return control
	}
	left := control.Left + control.Width*index/count
	right := control.Left + control.Width*(index+1)/count
	return settingsRect{Left: left, Top: control.Top, Width: right - left, Height: control.Height}
}

// ---------------------------------------------------------------- row lists

func appearanceRows(lang i18n.Lang) []settingsRow {
	return []settingsRow{
		{control: settingsTextSize, kind: rowSegment,
			title: i18n.T(lang, i18n.KeySettingsTextSize),
			hint:  i18n.T(lang, i18n.KeySettingsTextSizeHint),
			options: []string{
				i18n.T(lang, i18n.KeyTextSizeSmall),
				i18n.T(lang, i18n.KeyTextSizeStandard),
				i18n.T(lang, i18n.KeyTextSizeLarge),
			}},
		{control: settingsTheme, kind: rowSegment,
			title: i18n.T(lang, i18n.KeySettingsTheme),
			hint:  i18n.T(lang, i18n.KeySettingsThemeHint),
			options: []string{
				i18n.T(lang, i18n.KeyThemeSlate),
				i18n.T(lang, i18n.KeyThemePaper),
				i18n.T(lang, i18n.KeyThemeContrast),
			}},
		{control: settingsPerSource, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsPerSource),
			hint:  i18n.T(lang, i18n.KeySettingsPerSourceHint)},
		{control: settingsHiddenAtLaunch, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsHiddenAtLaunch),
			hint:  i18n.T(lang, i18n.KeySettingsHiddenAtLaunchHint)},
		{control: settingsAnimations, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsAnimations),
			hint:  i18n.T(lang, i18n.KeySettingsAnimationsHint)},
		{control: settingsNone, kind: rowNote,
			title: i18n.T(lang, i18n.KeySettingsAppearanceNote)},
	}
}

func generalRows(lang i18n.Lang) []settingsRow {
	return []settingsRow{
		{control: settingsNotifications, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsNotify), hint: i18n.T(lang, i18n.KeySettingsNotifyHint)},
		{control: settingsAlwaysOnTop, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsTopmost), hint: i18n.T(lang, i18n.KeySettingsTopmostHint)},
		{control: settingsEdgeRetract, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsRetract), hint: i18n.T(lang, i18n.KeySettingsRetractHint)},
		{control: settingsOpacity, kind: rowSlider, minimum: 40, maximum: 255,
			title: i18n.T(lang, i18n.KeySettingsOpacity), hint: i18n.T(lang, i18n.KeySettingsOpacityHint)},
		{control: settingsLanguage, kind: rowSegment,
			title:   i18n.T(lang, i18n.KeySettingsLanguage),
			hint:    i18n.T(lang, i18n.KeySettingsLanguageHint),
			options: []string{i18n.T(lang, i18n.KeyTrayLangAuto), "简体中文", "English"}},
		{control: settingsAutoStart, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsAutostart), hint: i18n.T(lang, i18n.KeySettingsAutostartHint)},
		// The two switches that make this window usable with a keyboard, and
		// that show what a clipped label said. Both are opt-in for the keyboard
		// half: a dialog that starts eating Tab is a surprise for a mouse user.
		{control: settingsKeyboard, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsKeyboard), hint: i18n.T(lang, i18n.KeySettingsKeyboardHint)},
		{control: settingsTooltips, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsTooltips), hint: i18n.T(lang, i18n.KeySettingsTooltipsHint)},
	}
}

func soundRows(lang i18n.Lang) []settingsRow {
	return []settingsRow{
		{control: settingsSound, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsSound), hint: i18n.T(lang, i18n.KeySettingsSoundHint)},
		{control: settingsScheme, kind: rowSegment,
			title: i18n.T(lang, i18n.KeySettingsScheme),
			hint:  i18n.T(lang, i18n.KeySettingsSchemeHint),
			options: []string{
				i18n.T(lang, i18n.KeySchemeChime),
				i18n.T(lang, i18n.KeySchemeBell),
				i18n.T(lang, i18n.KeySchemeSoft),
				i18n.T(lang, i18n.KeySchemePulse),
			}},
		{control: settingsVolume, kind: rowSlider, minimum: 0, maximum: 100,
			title: i18n.T(lang, i18n.KeySettingsVolume), hint: i18n.T(lang, i18n.KeySettingsVolumeHint)},
		{control: settingsSoundWaiting, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsSoundWaiting), hint: i18n.T(lang, i18n.KeySettingsSoundWaitingHint)},
		{control: settingsSoundError, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsSoundError), hint: i18n.T(lang, i18n.KeySettingsSoundErrorHint)},
		{control: settingsSoundFinish, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsSoundFinish), hint: i18n.T(lang, i18n.KeySettingsSoundFinishHint)},
		{control: settingsPreview, kind: rowButton,
			title:  i18n.T(lang, i18n.KeySettingsPreview),
			hint:   i18n.T(lang, i18n.KeySettingsPreviewHint),
			action: i18n.T(lang, i18n.KeySettingsPreviewPlay)},
		// Two test buttons, not one: "did the tone play" and "did the balloon
		// arrive" are different questions, and the second one is the only way
		// to check notification permission without waiting for a real blocker.
		{control: settingsTestWaiting, kind: rowButton,
			title:  i18n.T(lang, i18n.KeySettingsTestWaiting),
			hint:   i18n.T(lang, i18n.KeySettingsTestWaitingHint),
			action: i18n.T(lang, i18n.KeySettingsTestPlay)},
		{control: settingsTestDone, kind: rowButton,
			title:  i18n.T(lang, i18n.KeySettingsTestDone),
			hint:   i18n.T(lang, i18n.KeySettingsTestDoneHint),
			action: i18n.T(lang, i18n.KeySettingsTestPlay)},
	}
}

func sessionRows(lang i18n.Lang, muteCount int) []settingsRow {
	return []settingsRow{
		{control: settingsForegroundAware, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsForegroundAware), hint: i18n.T(lang, i18n.KeySettingsForegroundAwareHint)},
		{control: settingsIdleMinutes, kind: rowStepper, minimum: 0, maximum: 120,
			title: i18n.T(lang, i18n.KeySettingsIdleMinutes), hint: i18n.T(lang, i18n.KeySettingsIdleMinutesHint)},
		{control: settingsDonePolicy, kind: rowSegment,
			title: i18n.T(lang, i18n.KeySettingsDonePolicy),
			hint:  i18n.T(lang, i18n.KeySettingsDonePolicyHint),
			options: []string{
				i18n.T(lang, i18n.KeySettingsDoneBackground),
				i18n.T(lang, i18n.KeySettingsDoneOff),
				i18n.T(lang, i18n.KeySettingsDoneAlways),
			}},
		{control: settingsQuietCurrent, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsQuietCurrent), hint: i18n.T(lang, i18n.KeySettingsQuietCurrentHint)},
		{control: settingsQuietSubagents, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsQuietSubagents), hint: i18n.T(lang, i18n.KeySettingsQuietSubagentsHint)},
		{control: settingsMuteClear, kind: rowButton,
			title:  fmt.Sprintf("%s · %d", i18n.T(lang, i18n.KeySettingsMuteCount), muteCount),
			hint:   i18n.T(lang, i18n.KeySettingsMuteCountHint),
			action: i18n.T(lang, i18n.KeySettingsMuteClearAction)},
		{control: settingsAutoApprove, kind: rowToggle,
			title: i18n.T(lang, i18n.KeySettingsAutoApprove), hint: i18n.T(lang, i18n.KeySettingsAutoApproveHint)},
		{control: settingsApproveUp, alt: settingsApproveDown, kind: rowStepper, minimum: 1, maximum: 60,
			title: i18n.T(lang, i18n.KeySettingsApproveTimeout), hint: i18n.T(lang, i18n.KeySettingsApproveTimeoutHint)},
		{control: settingsReset, kind: rowButton,
			title:  i18n.T(lang, i18n.KeySettingsReset),
			hint:   i18n.T(lang, i18n.KeySettingsResetHint),
			action: i18n.T(lang, i18n.KeySettingsResetAction)},
		{control: settingsNone, kind: rowNote,
			title: i18n.T(lang, i18n.KeySettingsSessionsNote)},
		{control: settingsNone, kind: rowNote,
			title: i18n.T(lang, i18n.KeySettingsAttendNote)},
	}
}

// settingsSchemeIndex maps the persisted scheme name onto the segment index.
func settingsSchemeIndex(scheme string) int {
	scheme = validSoundScheme(scheme)
	for index, name := range SoundSchemes {
		if name == scheme {
			return index
		}
	}
	return 0
}

// settingsLanguageIndex maps the persisted language onto the segment index.
func settingsLanguageIndex(value string) int {
	switch i18n.NormalizeSetting(value) {
	case i18n.SettingZhCN:
		return 1
	case i18n.SettingEn:
		return 2
	default:
		return 0
	}
}

// settingsSegmentIndex is the selected option for a segment row, or -1.
func settingsSegmentIndex(row settingsRow, view settingsView) int {
	switch row.control {
	case settingsScheme:
		return settingsSchemeIndex(view.Scheme)
	case settingsLanguage:
		return settingsLanguageIndex(view.Language)
	case settingsTextSize:
		return settingsTextSizeIndex(view.TextSize)
	case settingsTheme:
		return settingsThemeIndex(view.Theme)
	case settingsDonePolicy:
		return settingsDonePolicyIndex(view.DonePolicy)
	}
	return -1
}

// settingsDonePolicyIndex maps the persisted policy onto the segment, whose
// option order is the one config.DonePolicies declares: background, off, always.
func settingsDonePolicyIndex(value string) int {
	switch ValidDonePolicy(value) {
	case DonePolicyOff:
		return 1
	case DonePolicyAlways:
		return 2
	default:
		return 0
	}
}

// settingsDonePolicyAt is the inverse, and it is exhaustive on purpose: a
// segment index outside 0..2 keeps the current value rather than inventing one.
func settingsDonePolicyAt(index int) (DonePolicy, bool) {
	switch index {
	case 0:
		return DonePolicyBackground, true
	case 1:
		return DonePolicyOff, true
	case 2:
		return DonePolicyAlways, true
	}
	return "", false
}

// settingsTextSizeIndex maps the persisted text size onto the segment index.
func settingsTextSizeIndex(value string) int {
	size := normalizeTextSize(value)
	for index, name := range config.TextSizes {
		if name == size {
			return index
		}
	}
	return 0
}

// settingsThemeIndex maps the persisted palette onto the segment index.
func settingsThemeIndex(name string) int {
	for index, candidate := range ThemeNames {
		if candidate == normalizeTheme(name) {
			return index
		}
	}
	return 0
}

// settingsSliderValueFor is the 0..1 position a slider row should show.
func settingsSliderValueFor(row settingsRow, view settingsView) float64 {
	value := 0
	switch row.control {
	case settingsOpacity:
		value = view.Opacity
	case settingsVolume:
		value = view.Volume
	}
	span := row.maximum - row.minimum
	if span <= 0 {
		return 0
	}
	fraction := float64(value-row.minimum) / float64(span)
	if fraction < 0 {
		return 0
	}
	if fraction > 1 {
		return 1
	}
	return fraction
}

// settingsToggleState is the on/off state of a toggle row.
func settingsToggleState(row settingsRow, view settingsView) bool {
	switch row.control {
	case settingsNotifications:
		return view.Notifications
	case settingsAlwaysOnTop:
		return view.AlwaysOnTop
	case settingsEdgeRetract:
		return view.EdgeRetract
	case settingsAutoStart:
		return view.AutoStart
	case settingsSound:
		return view.Sound
	case settingsSoundWaiting:
		return view.SoundOnWaiting
	case settingsSoundError:
		return view.SoundOnError
	case settingsSoundFinish:
		return view.NotifyOnFinish
	case settingsAutoApprove:
		return view.AutoApprove
	case settingsPerSource:
		return view.PerSource
	case settingsHiddenAtLaunch:
		return view.HiddenAtLaunch
	case settingsAnimations:
		return view.Animations
	case settingsForegroundAware:
		return view.ForegroundAware
	case settingsQuietCurrent:
		return view.QuietCurrent
	case settingsQuietSubagents:
		return view.QuietSubagents
	case settingsDonePolicy:
		// The segment draws its own selection; there is no tick to derive, so
		// this reports the value the "on" style would use only for completeness.
		return view.DonePolicy != string(DonePolicyOff)
	}
	return false
}

// settingsValueLabel is the right-aligned value shown on a slider row.
func settingsValueLabel(row settingsRow, view settingsView) string {
	switch row.control {
	case settingsOpacity:
		// Alpha is stored 40..255 but people think in percent.
		return itoaPercent(int(math.Round(float64(view.Opacity) * 100 / 255)))
	case settingsVolume:
		return itoaPercent(view.Volume)
	}
	return ""
}

// itoa avoids strconv in the draw path, where an allocation per frame is
// avoidable for a value that is at most three digits.
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := make([]byte, 0, 4)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	if negative {
		digits = append(digits, '-')
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}
	return string(digits)
}

func itoaPercent(value int) string {
	return itoa(value) + "%"
}

// ------------------------------------------------------------------ drawing

// settingsPaint is everything one frame of the settings window needs. It is a
// struct rather than eight parameters because the window now has three kinds of
// transient state (hover, press, keyboard focus) and a positional argument list
// that long is where the wrong argument gets passed silently.
type settingsPaint struct {
	geometry settingsGeometry
	view     settingsView
	hover    settingsControl
	hoverAt  int
	pressed  settingsControl
	focus    int
	focusOn  bool
	lang     i18n.Lang
	// version is the footer's right-aligned build string. Empty in tests that do
	// not care, which keeps the rasterised previews free of a moving value.
	version string
}

func drawSettings(target *canvas, paint settingsPaint) {
	geometry, view := paint.geometry, paint.view
	lang, fonts := paint.lang, geometry.fonts
	hover, hoverIndex, pressed := paint.hover, paint.hoverAt, paint.pressed
	s := geometry.scale
	f := geometry.font
	width, height := geometry.width, geometry.height

	target.clear()
	target.fillRounded(0, 0, width, height, s(settingsPanelRadius), colors.Panel)
	target.outlineRounded(0, 0, width, height, s(settingsPanelRadius), 1, colors.Border)

	// Header.
	target.text(s(settingsMargin), s(12), width-s(settingsMargin)*2-s(40), s(30),
		i18n.T(lang, i18n.KeySettingsTitle), f(fonts.header), 600, colors.Primary,
		dtSingleLine|dtVCenter|dtNoPrefix)
	subtitleWidth := s(220)
	target.text(width-s(settingsMargin)-s(30)-subtitleWidth, s(12), subtitleWidth, s(30),
		settingsSubtitle(lang, paint.version), f(fonts.subtitle), 400, colors.Muted,
		dtSingleLine|dtVCenter|dtRight|dtNoPrefix|dtEndEllipsis)
	target.fillRect(s(settingsMargin), s(settingsHeaderHeight), width-s(settingsMargin)*2, 1, colors.Divider)
	drawSettingsClose(target, geometry, hover == settingsClose || pressed == settingsClose)
	drawSettingsFocus(target, geometry, paint, settingsClose)

	// Sidebar.
	for index := range geometry.tabs {
		rect := geometry.tabs[index]
		control := settingsTabControl(index)
		selected := index == geometry.tab
		fill := rgba{}
		if selected {
			fill = colors.Raised
		} else if hover == control {
			fill = rgba{colors.Raised.R, colors.Raised.G, colors.Raised.B, 120}
		}
		if fill.A != 0 {
			target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(8), fill)
		}
		if selected {
			target.fillRounded(rect.Left+s(1), rect.Top+s(9), s(2), rect.Height-s(18), s(1), colors.Accent)
		}
		color := colors.Secondary
		if selected {
			color = colors.Primary
		}
		label := settingsTabLabel(index, lang)
		target.text(rect.Left+s(14), rect.Top, rect.Width-s(18), rect.Height, label, f(fonts.tab), 500, color,
			dtSingleLine|dtVCenter|dtNoPrefix)
		drawSettingsFocus(target, geometry, paint, control)
	}
	dividerX := s(settingsContentLeft) - s(18)
	target.fillRect(dividerX, s(settingsRowsTop)-s(6), 1, height-s(settingsRowsTop)-s(settingsMargin)+s(6), colors.Divider)

	// Rows.
	for index, row := range geometry.rows {
		frame := geometry.frames[index]
		control := geometry.controls[index]
		hot := hover == row.control || hover == row.alt
		if row.kind == rowNote {
			lines := wrapHint(target, row.title, f(fonts.note), 400, frame.Width, settingsNoteMaxLines)
			for line, value := range lines {
				target.text(frame.Left, frame.Top+line*s(settingsHintLine), frame.Width, s(settingsHintLine),
					value, f(fonts.note), 400, colors.Hint, dtSingleLine|dtVCenter|dtNoPrefix)
			}
			continue
		}
		titleWidth := settingsTitleWidth(geometry, row, frame)
		target.text(frame.Left, frame.Top, titleWidth, s(settingsTitleBand), row.title, f(fonts.title), 500, colors.Primary,
			dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
		if row.hint != "" {
			lines := wrapHint(target, row.hint, f(fonts.hint), 400, frame.Width, settingsMaxLines(row.kind))
			for line, value := range lines {
				top := frame.Top + s(settingsTitleBand) + line*s(settingsHintLine)
				target.text(frame.Left, top, frame.Width, s(settingsHintLine), value, f(fonts.hint), 400, colors.Hint,
					dtSingleLine|dtVCenter|dtNoPrefix)
			}
		}

		switch row.kind {
		case rowToggle:
			drawSettingsToggle(target, control, settingsToggleState(row, view), hot, pressed == row.control)
		case rowSlider:
			label := settingsValueLabel(row, view)
			target.text(frame.right()-s(56), frame.Top, s(56), s(settingsTitleBand), label, f(fonts.value), 500, colors.Secondary,
				dtSingleLine|dtVCenter|dtRight|dtNoPrefix)
			drawSettingsSlider(target, control, settingsSliderValueFor(row, view), hot, s(14))
		case rowSegment:
			drawSettingsSegment(target, geometry, index, row, view, hover, hoverIndex)
		case rowStepper:
			drawSettingsStepper(target, geometry, index, row, view, hover, pressed)
		case rowButton:
			fill := colors.Raised
			if hot {
				fill = colors.PanelHover
			}
			if pressed == row.control {
				fill = colors.ControlBg
			}
			target.fillRounded(control.Left, control.Top, control.Width, control.Height, s(8), fill)
			target.outlineRounded(control.Left, control.Top, control.Width, control.Height, s(8), 1, colors.Border)
			target.text(control.Left, control.Top, control.Width, control.Height, row.action, f(fonts.button), 500, colors.Primary,
				dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
		}
		drawSettingsFocus(target, geometry, paint, row.control)
	}
	drawSettingsFooter(target, paint)
	drawSettingsTooltip(target, paint)
}

// drawSettingsFooter draws the two links and the version.
//
// The links are drawn as text with an underline on hover rather than as buttons:
// they leave the app, and a button shape would promise something the window
// cannot do (nothing comes back). The version is on the same line because a bug
// report that has to be chased for its build number is a bug report that does not
// get filed.
func drawSettingsFooter(target *canvas, paint settingsPaint) {
	geometry, lang, fonts := paint.geometry, paint.lang, paint.geometry.fonts
	s, f := geometry.scale, geometry.font
	if geometry.footer.Width <= 0 {
		return
	}
	target.fillRect(geometry.footer.Left, geometry.footer.Top-s(6), geometry.footer.Width, 1, colors.Divider)

	links := []struct {
		control settingsControl
		rect    settingsRect
		label   string
	}{
		{settingsStar, geometry.star, i18n.T(lang, i18n.KeySettingsStar)},
		{settingsIssue, geometry.issue, i18n.T(lang, i18n.KeySettingsIssue)},
	}
	for _, link := range links {
		colour := colors.Secondary
		hot := paint.hover == link.control || paint.pressed == link.control
		if hot {
			colour = colors.Accent
		}
		target.text(link.rect.Left, link.rect.Top, link.rect.Width, link.rect.Height, link.label,
			f(fonts.hint), 500, colour, dtSingleLine|dtVCenter|dtNoPrefix)
		if hot {
			width := target.measure(link.label, f(fonts.hint), 500)
			target.fillRect(link.rect.Left, link.rect.Top+link.rect.Height-s(4), minInt(width, link.rect.Width), 1, colour)
		}
		drawSettingsFocus(target, geometry, paint, link.control)
	}

}

// settingsSubtitle is the header's right-hand text: what this window belongs to,
// and which build it is. A bug report that has to be chased for its version is a
// bug report that does not get filed, and the header is where somebody copying
// the window already looks.
func settingsSubtitle(lang i18n.Lang, version string) string {
	label := i18n.T(lang, i18n.KeySettingsSubtitle)
	if version == "" {
		return label
	}
	return label + " · " + version
}

// drawSettingsTooltip shows the full text of a label that had to be ellipsised.
//
// It is drawn inside the window rather than as a shell tooltip window on
// purpose: this window is a layered surface with no child controls, so a real
// tooltip would need its own top-level window and its own activation plumbing
// for a case that is now rare - after the type ramp and the wrapping fix, only
// an unusually long row title is ever cut off. A strip attached to the row does
// the job with no new window, and it cannot end up off-screen.
func drawSettingsTooltip(target *canvas, paint settingsPaint) {
	if !paint.view.Tooltips {
		return
	}
	geometry := paint.geometry
	if paint.hover == settingsNone || paint.hover == settingsClose {
		return
	}
	index := -1
	for rowIndex, row := range geometry.rows {
		if row.control == paint.hover || row.alt == paint.hover {
			index = rowIndex
			break
		}
	}
	if index < 0 {
		return
	}
	row := geometry.rows[index]
	if row.kind == rowNote || row.title == "" {
		return
	}
	frame := geometry.frames[index]
	titleWidth := settingsTitleWidth(geometry, row, frame)
	titleFont := geometry.font(geometry.fonts.title)
	if target.measure(row.title, titleFont, 500) <= titleWidth {
		return
	}

	s := geometry.scale
	padding := s(10)
	width := frame.Width
	lines := wrapHint(target, row.title, titleFont, 500, width-padding*2, 2)
	if len(lines) == 0 {
		return
	}
	height := s(settingsHintLine)*len(lines) + padding
	top := frame.bottom() + s(4)
	if top+height > geometry.height-s(settingsMargin) {
		top = frame.Top - height - s(4)
	}
	if top < s(settingsHeaderHeight) {
		top = s(settingsHeaderHeight) + s(4)
	}
	target.fillRounded(frame.Left, top, width, height, s(8), colors.ControlBg)
	target.outlineRounded(frame.Left, top, width, height, s(8), 1, colors.Border)
	for line, value := range lines {
		target.text(frame.Left+padding, top+padding/2+line*s(settingsHintLine), width-padding*2, s(settingsHintLine),
			value, titleFont, 500, colors.Primary, dtSingleLine|dtVCenter|dtNoPrefix)
	}
}

// settingsTitleWidth is the space a row title has before its control, shared by
// the row drawing and the tooltip decision so the two cannot disagree about
// whether a title fits.
func settingsTitleWidth(geometry settingsGeometry, row settingsRow, frame settingsRect) int {
	s := geometry.scale
	switch row.kind {
	case rowToggle:
		return frame.Width - s(48)
	case rowStepper:
		return frame.Width - s(104)
	case rowButton:
		return frame.Width - s(96)
	case rowSlider:
		return frame.Width - s(56)
	}
	return frame.Width
}

// drawSettingsFocus draws the keyboard focus ring for one control, and only when
// the user has actually used the keyboard: a ring that appears because the window
// opened is the visual equivalent of shouting at someone who reached for the
// mouse.
func drawSettingsFocus(target *canvas, geometry settingsGeometry, paint settingsPaint, control settingsControl) {
	if !paint.focusOn || control == settingsNone {
		return
	}
	if settingsFocusControl(geometry, paint.focus) != control {
		return
	}
	rect, ok := settingsFocusRect(geometry, paint.focus)
	if !ok {
		return
	}
	thickness := maxInt(geometry.scale(1), 1)
	target.outlineRounded(rect.Left-thickness, rect.Top-thickness, rect.Width+thickness*2,
		rect.Height+thickness*2, geometry.scale(8), thickness, colors.Accent)
}

func settingsTabLabel(index int, lang i18n.Lang) string {
	switch index {
	case 1:
		return i18n.T(lang, i18n.KeySettingsTabSound)
	case 2:
		return i18n.T(lang, i18n.KeySettingsTabSessions)
	case 3:
		return i18n.T(lang, i18n.KeySettingsTabAppearance)
	default:
		return i18n.T(lang, i18n.KeySettingsTabGeneral)
	}
}

// drawSettingsClose draws the × in the header. Two strokes rather than a glyph
// keeps the weight consistent with the rest of the chrome at any DPI.
func drawSettingsClose(target *canvas, geometry settingsGeometry, hot bool) {
	s := geometry.scale
	rect := geometry.close
	if hot {
		target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(7), colors.Raised)
	}
	color := colors.Secondary
	if hot {
		color = colors.Primary
	}
	inset := rect.Width / 3
	left, top := rect.Left+inset, rect.Top+inset
	right, bottom := rect.right()-inset, rect.bottom()-inset
	thickness := maxInt(s(1), 1)
	target.line(left, top, right, bottom, thickness, color)
	target.line(right, top, left, bottom, thickness, color)
}

func drawSettingsToggle(target *canvas, rect settingsRect, on, hot, pressed bool) {
	track := colors.TrackOff
	if on {
		track = colors.Accent
	}
	if hot && !on {
		track = rgba{track.R, track.G, track.B, 255}
		track = rgba{colors.Secondary.R, colors.Secondary.G, colors.Secondary.B, 210}
	}
	if pressed {
		track.A = 230
	}
	radius := rect.Height / 2
	target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, radius, track)
	gap := maxInt(rect.Height/8, 1)
	knob := rect.Height - gap*2
	knobLeft := rect.Left + gap
	if on {
		knobLeft = rect.right() - gap - knob
	}
	target.fillEllipse(knobLeft, rect.Top+gap, knob, knob, colors.Primary)
}

func drawSettingsSlider(target *canvas, track settingsRect, fraction float64, hot bool, knobSize int) {
	target.fillRounded(track.Left, track.Top, track.Width, track.Height, track.Height/2, colors.ControlBg)
	filled := int(math.Round(float64(track.Width) * fraction))
	if filled > 0 {
		target.fillRounded(track.Left, track.Top, filled, track.Height, track.Height/2, colors.Accent)
	}
	knobLeft := track.Left + filled - knobSize/2
	if knobLeft < track.Left-knobSize/2 {
		knobLeft = track.Left - knobSize/2
	}
	if knobLeft > track.right()-knobSize/2 {
		knobLeft = track.right() - knobSize/2
	}
	knobTop := track.Top + track.Height/2 - knobSize/2
	ring := colors.Border
	if hot {
		ring = colors.Accent
	}
	target.fillEllipse(knobLeft, knobTop, knobSize, knobSize, ring)
	target.fillEllipse(knobLeft+2, knobTop+2, knobSize-4, knobSize-4, colors.Primary)
}

func drawSettingsSegment(target *canvas, geometry settingsGeometry, index int, row settingsRow, view settingsView, hover settingsControl, hoverIndex int) {
	s := geometry.scale
	rect := geometry.controls[index]
	target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(8), colors.ControlBg)
	target.outlineRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(8), 1, colors.Divider)
	count := len(row.options)
	selected := settingsSegmentIndex(row, view)
	for option := 0; option < count; option++ {
		cell := settingsSegmentRect(rect, count, option).inset(s(2))
		if option == selected {
			target.fillRounded(cell.Left, cell.Top, cell.Width, cell.Height, s(6), colors.Raised)
			target.outlineRounded(cell.Left, cell.Top, cell.Width, cell.Height, s(6), 1, colors.Accent)
		} else if hover == row.control && hoverIndex == option {
			target.fillRounded(cell.Left, cell.Top, cell.Width, cell.Height, s(6), colors.RaisedRow)
		}
		color := colors.Secondary
		if option == selected {
			color = colors.Primary
		}
		weight := 400
		if option == selected {
			weight = 600
		}
		target.text(cell.Left, cell.Top, cell.Width, cell.Height, row.options[option], s(11), weight, color,
			dtSingleLine|dtVCenter|dtCenter|dtNoPrefix|dtEndEllipsis)
	}
}

func drawSettingsStepper(target *canvas, geometry settingsGeometry, index int, row settingsRow, view settingsView, hover settingsControl, pressed settingsControl) {
	s := geometry.scale
	rect := geometry.controls[index]
	button := rect.Height
	minus := settingsRect{Left: rect.Left, Top: rect.Top, Width: button, Height: button}
	plus := settingsRect{Left: rect.right() - button, Top: rect.Top, Width: button, Height: button}
	value := settingsRect{Left: minus.right(), Top: rect.Top, Width: plus.Left - minus.right(), Height: rect.Height}

	for _, item := range []struct {
		rect    settingsRect
		control settingsControl
		glyph   string
	}{
		{minus, row.alt, "−"},
		{plus, row.control, "+"},
	} {
		fill := colors.Raised
		if hover == item.control {
			fill = colors.PanelHover
		}
		if pressed == item.control {
			fill = colors.ControlBg
		}
		target.fillRounded(item.rect.Left, item.rect.Top, item.rect.Width, item.rect.Height, s(6), fill)
		target.text(item.rect.Left, item.rect.Top, item.rect.Width, item.rect.Height, item.glyph, s(12), 500, colors.Primary,
			dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	}
	target.text(value.Left, value.Top, value.Width, value.Height, itoa(view.ApproveTimeout)+"s",
		s(11), 600, colors.Primary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
}
