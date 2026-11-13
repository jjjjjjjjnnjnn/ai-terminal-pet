package ui

// The settings window.
//
// The overlay is deliberately a glanceable strip, not a control panel, so every
// preference used to live either in a tray menu of checkboxes or in
// config.json. That made the two things people actually want to change - which
// alert tone, and how loud - either invisible or a three-click cycle, and it
// left opacity, the auto-approve countdown and the language with no UI at all.
//
// This is a second top-level window on the SAME UI thread. It shares the Slate
// palette and the canvas rasteriser, so it reads as part of the same product
// without dragging in a widget toolkit (and without the DPI, theming and
// focus behaviour that a toolkit would impose on the overlay). Three rules from
// the overlay carry over verbatim:
//
//   - geometry is derived from the constants below, and the same layout is used
//     for drawing and for hit testing, so a control can never be drawn in one
//     place and clickable in another;
//   - the window is created, shown and destroyed on the thread that pumps its
//     messages (Run pins that thread);
//   - no file IO here. Every change is published to the app layer, which
//     coalesces it onto a background writer.

import (
	"fmt"
	"math"
	"syscall"
	"time"
	"unsafe"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
	win "github.com/lxn/win"
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
)

// settingsContentWidth is the width of the right-hand column.
const settingsContentWidth = settingsWidth - settingsMargin - settingsContentLeft

// Settings text sizes. config.json stores these names (internal/config owns the
// list); the ui package compares against them, and a test in internal/app pins
// the two together.
const (
	textSizeSmall    = "small"
	textSizeStandard = "standard"
	textSizeLarge    = "large"
)

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

// Row kinds. The height of a row follows from what it has to contain: a toggle
// shares its line with the label, a slider and a segment need a line below it.
const (
	rowToggle = iota
	rowSlider
	rowSegment
	rowStepper
	rowButton
	rowNote
)

// Row height parts, in logical pixels at 96 DPI. The height of a row is these
// parts added up rather than one number per kind, because a hint may need one or
// two lines and a fixed height either wastes space or clips the second line.
const (
	settingsTitleBand = 24
	settingsHintLine  = 17
	settingsRowPad    = 10
	settingsBandGap   = 4
)

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
	// The window is exactly as tall as its content. A fixed height left the
	// shortest tab with half a screen of emptiness, which reads as a broken
	// layout rather than a compact one.
	geometry.height = maxInt(y+geometry.scale(settingsMargin), geometry.scale(settingsMinHeight))
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
func (a *App) moveSettingsFocus(delta int) {
	order := settingsFocusOrder(a.settingsGeometry)
	if len(order) == 0 {
		return
	}
	index := a.settingsFocus + delta
	if index < 0 {
		index = len(order) - 1
	}
	if index >= len(order) {
		index = 0
	}
	a.settingsFocus = index
	a.settingsFocusVisible = true
	_ = a.paintSettings()
}

// activateSettingsFocus runs the focused control, the same path a click takes.
func (a *App) activateSettingsFocus() {
	control := settingsFocusControl(a.settingsGeometry, a.settingsFocus)
	if control == settingsNone {
		return
	}
	a.settingsApply(control, settingsIndexForControl(a.settingsGeometry, control, a.settingsView()))
}

// settingsIndexForControl answers the option index a control needs: the current
// one for a segment, the plus half for a stepper. Keyboard activation has no
// cursor to point at an option, so a segment re-applies the option it is
// already showing rather than guessing a new one.
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

func (a *App) settingsView() settingsView {
	settings := a.options.Settings
	return settingsView{
		Notifications:   settings.Notifications,
		AlwaysOnTop:     settings.AlwaysOnTop,
		EdgeRetract:     settings.EdgeRetract,
		Sound:           settings.Sound,
		SoundOnWaiting:  settings.SoundOnWaiting,
		SoundOnError:    settings.SoundOnError,
		NotifyOnFinish:  settings.NotifyOnFinish,
		AutoApprove:     settings.AutoApprove,
		AutoStart:       a.options.AutoStart,
		Opacity:         int(a.options.Opacity),
		Volume:          settings.SoundVolume,
		ApproveTimeout:  a.approveSeconds(),
		Scheme:          validSoundScheme(settings.SoundScheme),
		Language:        settings.Language,
		Theme:           normalizeTheme(settings.Theme),
		PerSource:       settings.PerSourceColor,
		HiddenAtLaunch:  settings.HiddenAtLaunch,
		TextSize:        normalizeTextSize(settings.SettingsTextSize),
		Keyboard:        settings.SettingsKeyboard,
		Tooltips:        settings.SettingsTooltips,
		ForegroundAware: settings.ForegroundAware,
		IdleMinutes:     settings.IdleSeconds / 60,
		DonePolicy:      string(ValidDonePolicy(settings.DonePolicy)),
		QuietCurrent:    settings.QuietCurrent,
		QuietSubagents:  settings.QuietSubagents,
		MuteCount:       a.mutes.size(),
	}
}

// approveSeconds is the countdown length the overlay is actually running with.
func (a *App) approveSeconds() int {
	seconds := int(a.approveTimeout / time.Second)
	if seconds < 1 {
		seconds = 5
	}
	return seconds
}

func (a *App) settingsRows(dpi uint32) []settingsRow {
	var rows []settingsRow
	switch a.settingsTab {
	case 1:
		rows = soundRows(a.lang)
	case 2:
		rows = sessionRows(a.lang, a.mutes.size())
	case 3:
		rows = appearanceRows(a.lang)
	default:
		rows = generalRows(a.lang)
	}
	// The number of lines each hint needs depends on the font and the column
	// width, so it is measured here rather than guessed by the builders. The
	// column width does not depend on the rows, so this cannot chase itself.
	size := normalizeTextSize(a.options.Settings.SettingsTextSize)
	probe := settingsGeometrySized(dpi, size, a.settingsTab, nil)
	fonts := settingsTypeRamps[size]
	return measureSettingsRows(a.settingsMeasureCanvas(), rows, fonts, probe.scale(settingsContentWidth))
}

// settingsMeasureCanvas is a one-pixel canvas used only to ask GDI how wide a
// string is. It is created lazily and lives as long as the window.
func (a *App) settingsMeasureCanvas() *canvas {
	if a.settingsMeasure != nil {
		return a.settingsMeasure
	}
	scratch, err := newCanvas(1, 1)
	if err != nil {
		return nil
	}
	a.settingsMeasure = scratch
	return scratch
}

// appearanceRows is the look of the product: palette, provenance colour, and
// whether the panel starts in the tray.
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
	subtitleWidth := s(160)
	target.text(width-s(settingsMargin)-s(30)-subtitleWidth, s(12), subtitleWidth, s(30),
		i18n.T(lang, i18n.KeySettingsSubtitle), f(fonts.subtitle), 400, colors.Muted,
		dtSingleLine|dtVCenter|dtRight|dtNoPrefix)
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
	drawSettingsTooltip(target, paint)
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

// ------------------------------------------------------------------- window

var settingsWindowProcedure = syscall.NewCallback(settingsProcedure)

func settingsProcedure(hwnd win.HWND, message uint32, wParam, lParam uintptr) uintptr {
	window := activeWindow.Load()
	if window == nil {
		return win.DefWindowProc(hwnd, message, wParam, lParam)
	}
	// The hang watchdog watches this timestamp, and the settings window runs on
	// the same thread as the overlay, so its messages have to refresh it too.
	window.lastDispatchNs.Store(time.Now().UnixNano())
	return window.handleSettingsMessage(hwnd, message, wParam, lParam)
}

// registerSettingsClass registers the settings window class once, from Run.
func registerSettingsClass(instance win.HINSTANCE) (*uint16, bool) {
	name, err := syscall.UTF16FromString(settingsClassName)
	if err != nil {
		return nil, false
	}
	class := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		Style:         win.CS_HREDRAW | win.CS_VREDRAW,
		LpfnWndProc:   settingsWindowProcedure,
		HInstance:     instance,
		HCursor:       win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW)),
		LpszClassName: &name[0],
	}
	if atom := win.RegisterClassEx(&class); atom == 0 {
		return nil, false
	}
	return &name[0], true
}

// openSettings shows the window, creating it the first time. It is deliberately
// created lazily: most runs never open it, and a window plus a DIB is real
// memory to hold for a panel nobody asked to configure.
func (a *App) openSettings() {
	if a.settingsHandle.Load() == 0 {
		a.createSettingsWindow()
	}
	handle := a.settingsHandle.Load()
	if handle == 0 {
		return
	}
	a.settingsFocus = 0
	a.settingsFocusVisible = false
	a.paintSettings()
	win.ShowWindow(win.HWND(handle), win.SW_SHOW)
	win.SetForegroundWindow(win.HWND(handle))
}

// settingsDpi is the DPI of the monitor the settings window is on, falling
// back to the overlay's. Using the overlay's DPI was wrong whenever the two were
// on different monitors: the dialog then laid itself out at the wrong scale.
func (a *App) settingsDpi() uint32 {
	if handle := a.settingsHandle.Load(); handle != 0 {
		if dpi := win.GetDpiForWindow(win.HWND(handle)); dpi != 0 {
			return dpi
		}
	}
	return a.windowDpi()
}

func (a *App) createSettingsWindow() {
	name, err := syscall.UTF16FromString(settingsClassName)
	if err != nil {
		return
	}
	title, err := syscall.UTF16FromString(settingsWindowTitle)
	if err != nil {
		return
	}
	dpi := a.windowDpi()
	geometry := a.settingsGeometryFitted(dpi)
	style := uint32(win.WS_EX_TOOLWINDOW | win.WS_EX_LAYERED | win.WS_EX_TOPMOST)
	x, y := a.settingsOrigin(geometry.width, geometry.height)
	created := win.CreateWindowEx(
		style, &name[0], &title[0], uint32(win.WS_POPUP),
		int32(x), int32(y), int32(geometry.width), int32(geometry.height),
		0, 0, win.GetModuleHandle(nil), nil,
	)
	if created == 0 {
		return
	}
	a.settingsHandle.Store(uintptr(created))
	a.settingsDPI = dpi
	a.settingsGeometry = geometry
	a.settingsX, a.settingsY = x, y
	a.settingsHoverIndex = -1
	a.settingsFrame = settingsFrameKey{}
}

// settingsOrigin centres the window on the monitor the overlay is on, nudged so
// it does not sit exactly on top of the panel.
func (a *App) settingsOrigin(width, height int) (int32, int32) {
	work, ok := a.workArea()
	if !ok {
		return 120, 120
	}
	x := work.Left + (work.Right-work.Left-int32(width))/2
	y := work.Top + (work.Bottom-work.Top-int32(height))/2
	return x, y
}

func (a *App) hideSettings() {
	handle := a.settingsHandle.Load()
	if handle == 0 {
		return
	}
	a.settingsDrag = settingsNone
	a.settingsPress = settingsNone
	a.settingsHover = settingsNone
	a.settingsMoving = false
	a.settingsHoverTracking = false
	a.settingsFocusVisible = false
	win.ShowWindow(win.HWND(handle), win.SW_HIDE)
}

func (a *App) destroySettings() {
	handle := a.settingsHandle.Load()
	if handle != 0 {
		a.settingsHandle.Store(0)
		win.DestroyWindow(win.HWND(handle))
	}
	if a.settingsSurface != nil {
		a.settingsSurface.close()
		a.settingsSurface = nil
	}
	if a.settingsMeasure != nil {
		a.settingsMeasure.close()
		a.settingsMeasure = nil
	}
}

// settingsFrameKey is everything that changes the picture. Two equal keys mean
// the raster can be skipped, which matters because hover tracking repaints on
// every mouse move.
type settingsFrameKey struct {
	width, height int
	tab           int
	hover         settingsControl
	hoverIndex    int
	pressed       settingsControl
	focus         int
	focusOn       bool
	lang          i18n.Lang
	view          settingsView
}

// reflowSettings rebuilds the layout for the current tab and DPI, keeping the
// window's centre fixed so switching tabs grows or shrinks it symmetrically
// instead of yanking one edge around.
//
// UpdateLayeredWindow takes the destination and the size together, so this only
// has to adjust the remembered origin - the next paint resizes the window.
func (a *App) reflowSettings(dpi uint32) {
	if dpi == 0 {
		dpi = a.settingsDPI
	}
	if dpi == 0 {
		dpi = 96
	}
	geometry := a.settingsGeometryFitted(dpi)
	geometry.lang = a.lang
	previous := a.settingsGeometry
	if previous.width > 0 && previous.height > 0 {
		a.settingsX += int32((previous.width - geometry.width) / 2)
		a.settingsY += int32((previous.height - geometry.height) / 2)
	}
	a.settingsDPI = dpi
	a.settingsGeometry = geometry
	a.clampSettingsToWorkArea()
}

// settingsGeometryFitted resolves the layout for the current tab and text
// size, then steps the text size down if the window would be taller than the
// work area. A control nobody can reach is worse than text one size smaller.
func (a *App) settingsGeometryFitted(dpi uint32) settingsGeometry {
	limit := 0
	if work, ok := a.workArea(); ok {
		margin := int(math.Round(16 * float64(dpi) / 96))
		limit = int(work.Bottom-work.Top) - margin
	}
	size := normalizeTextSize(a.options.Settings.SettingsTextSize)
	geometry := settingsGeometryFitting(dpi, size, a.settingsTab, a.settingsRows(dpi), limit)
	geometry.lang = a.lang
	return geometry
}

// clampSettingsToWorkArea keeps the window on screen. It matters because the
// window is centred on the overlay's monitor: a taller tab near the bottom edge
// would otherwise grow past it.
func (a *App) clampSettingsToWorkArea() {
	work, ok := a.workArea()
	if !ok {
		return
	}
	margin := int32(a.settingsGeometry.scale(8))
	width := int32(a.settingsGeometry.width)
	height := int32(a.settingsGeometry.height)
	minimumX := work.Left + margin
	maximumX := work.Right - width - margin
	if maximumX < minimumX {
		maximumX = minimumX
	}
	minimumY := work.Top + margin
	maximumY := work.Bottom - height - margin
	if maximumY < minimumY {
		maximumY = minimumY
	}
	if a.settingsX < minimumX {
		a.settingsX = minimumX
	}
	if a.settingsX > maximumX {
		a.settingsX = maximumX
	}
	if a.settingsY < minimumY {
		a.settingsY = minimumY
	}
	if a.settingsY > maximumY {
		a.settingsY = maximumY
	}
}

func (a *App) paintSettings() error {
	handle := a.settingsHandle.Load()
	if handle == 0 {
		return nil
	}
	dpi := a.settingsDpi()
	if dpi != a.settingsDPI || len(a.settingsGeometry.rows) == 0 ||
		a.settingsGeometry.tab != a.settingsTab || a.settingsGeometry.lang != a.lang ||
		a.settingsGeometry.textSize != normalizeTextSize(a.options.Settings.SettingsTextSize) {
		a.reflowSettings(dpi)
	}
	geometry := a.settingsGeometry
	view := a.settingsView()
	key := settingsFrameKey{
		width: geometry.width, height: geometry.height, tab: a.settingsTab,
		hover: a.settingsHover, hoverIndex: a.settingsHoverIndex, pressed: a.settingsPress,
		focus: a.settingsFocus, focusOn: a.settingsFocusVisible,
		lang: a.lang, view: view,
	}
	if key == a.settingsFrame && a.settingsSurface != nil {
		return nil
	}
	if a.settingsSurface != nil && (a.settingsSurface.width != geometry.width || a.settingsSurface.height != geometry.height) {
		a.settingsSurface.close()
		a.settingsSurface = nil
	}
	if a.settingsSurface == nil {
		surface, err := newCanvas(geometry.width, geometry.height)
		if err != nil {
			return err
		}
		a.settingsSurface = surface
	}
	drawSettings(a.settingsSurface, settingsPaint{
		geometry: geometry,
		view:     view,
		hover:    a.settingsHover,
		hoverAt:  a.settingsHoverIndex,
		pressed:  a.settingsPress,
		focus:    a.settingsFocus,
		focusOn:  a.settingsFocusVisible,
		lang:     a.lang,
	})
	a.settingsFrame = key

	screenDC := win.GetDC(0)
	if screenDC == 0 {
		return syscall.EINVAL
	}
	defer win.ReleaseDC(0, screenDC)
	destination := win.POINT{X: a.settingsX, Y: a.settingsY}
	dimensions := win.SIZE{CX: int32(geometry.width), CY: int32(geometry.height)}
	return updateLayeredWindow(win.HWND(handle), screenDC, &destination, dimensions, a.settingsSurface.dc, 255)
}

func (a *App) handleSettingsMessage(hwnd win.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmEraseBackground:
		return 1
	case wmMouseMove:
		a.settingsMouseMove(int(int16(uint16(lParam))), int(int16(uint16(lParam>>16))))
		return 0
	case wmMouseLeave:
		a.settingsHoverTracking = false
		a.settingsHover = settingsNone
		a.settingsHoverIndex = -1
		_ = a.paintSettings()
		return 0
	case wmCaptureChanged, wmCancelMode:
		// Losing capture mid-gesture (Alt-Tab, another window taking the
		// mouse) must not leave a slider stuck to the pointer.
		//
		// ReleaseCapture in our own mouse-up sends WM_CAPTURECHANGED straight
		// back into this procedure, so the press state is deliberately NOT
		// touched here: clearing it would erase the very click that is being
		// processed.
		if a.settingsDrag != settingsNone || a.settingsMoving {
			a.settingsDrag = settingsNone
			a.settingsMoving = false
			_ = a.paintSettings()
		}
		return 0
	case wmLButtonDown:
		a.settingsMouseDown(int(int16(uint16(lParam))), int(int16(uint16(lParam>>16))))
		return 0
	case wmLButtonUp:
		a.settingsMouseUp(int(int16(uint16(lParam))), int(int16(uint16(lParam>>16))))
		return 0
	case wmDPIChanged:
		// Dragging this window to a monitor with different scaling used to leave
		// it laid out for the old one, because the layout read the *overlay's*
		// DPI and nothing re-read it. The new DPI arrives in the low word of
		// wParam; the suggested rectangle Windows also passes is deliberately
		// ignored, because this window is placed by its own rules and clamped to
		// the work area straight after.
		a.settingsDPI = 0
		a.reflowSettings(uint32(wParam & 0xffff))
		a.clampSettingsToWorkArea()
		_ = a.paintSettings()
		return 0
	case wmKeyDown:
		if a.handleSettingsKey(wParam, lParam) {
			return 0
		}
	case wmClose:
		a.hideSettings()
		return 0
	case wmDestroy:
		a.settingsHandle.Store(0)
		return 0
	}
	return win.DefWindowProc(hwnd, message, wParam, lParam)
}

// handleSettingsKey runs the settings window's keyboard gestures and reports
// whether it consumed the key.
//
// Escape always works. Tab and Space/Enter only work when the user turned
// keyboard navigation on: this dialog is mouse-first, and a window that starts
// swallowing Tab because it happens to be open is a surprise. The flag exists so
// that is a choice rather than an opinion.
func (a *App) handleSettingsKey(key uintptr, lParam uintptr) bool {
	if key == wmKeyEscape {
		a.hideSettings()
		return true
	}
	if !a.options.Settings.SettingsKeyboard {
		return false
	}
	switch key {
	case wmKeyTab:
		delta := 1
		if win.GetKeyState(win.VK_SHIFT) < 0 {
			delta = -1
		}
		a.moveSettingsFocus(delta)
		return true
	case wmKeySpace, wmKeyReturn:
		a.activateSettingsFocus()
		return true
	}
	return false
}

func (a *App) settingsMouseMove(x, y int) {
	if a.settingsMoving {
		var cursor win.POINT
		win.GetCursorPos(&cursor)
		a.settingsX = cursor.X - a.settingsMoveDX
		a.settingsY = cursor.Y - a.settingsMoveDY
		win.SetWindowPos(win.HWND(a.settingsHandle.Load()), 0, a.settingsX, a.settingsY, 0, 0,
			win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
		return
	}
	if a.settingsDrag != settingsNone {
		a.settingsApplySlider(a.settingsDrag, x)
		return
	}
	a.trackSettingsHover()
	control, index, _, ok := a.settingsGeometry.hit(x, y)
	if !ok {
		control, index = settingsNone, -1
	}
	if control == a.settingsHover && index == a.settingsHoverIndex {
		return
	}
	a.settingsHover = control
	a.settingsHoverIndex = index
	_ = a.paintSettings()
}

// trackSettingsHover installs the leave notification once per entry. Without
// TrackMouseEvent Windows never sends WM_MOUSELEAVE, so a highlight would stay
// lit after the cursor left the window.
func (a *App) trackSettingsHover() {
	if a.settingsHoverTracking {
		return
	}
	a.settingsHoverTracking = true
	track := win.TRACKMOUSEEVENT{
		CbSize:    uint32(unsafe.Sizeof(win.TRACKMOUSEEVENT{})),
		DwFlags:   win.TME_LEAVE,
		HwndTrack: win.HWND(a.settingsHandle.Load()),
	}
	win.TrackMouseEvent(&track)
}

func (a *App) settingsMouseDown(x, y int) {
	// The header drags the window. This is the same gesture the overlay uses,
	// and it runs on the thread that owns the window, so SetWindowPos here
	// cannot block on another thread's queue.
	if y < a.settingsGeometry.scale(settingsHeaderHeight) && !a.settingsGeometry.close.contains(x, y) {
		var cursor win.POINT
		win.GetCursorPos(&cursor)
		a.settingsMoving = true
		a.settingsMoveDX = cursor.X - a.settingsX
		a.settingsMoveDY = cursor.Y - a.settingsY
		win.SetCapture(win.HWND(a.settingsHandle.Load()))
		return
	}
	control, _, _, ok := a.settingsGeometry.hit(x, y)
	if !ok || control == settingsNone {
		return
	}
	a.settingsPress = control
	// Capture every press, not just sliders: without it a release outside the
	// window is never delivered and the pressed style sticks.
	win.SetCapture(win.HWND(a.settingsHandle.Load()))
	if control == settingsOpacity || control == settingsVolume {
		a.settingsDrag = control
		a.settingsApplySlider(control, x)
		return
	}
	_ = a.paintSettings()
}

func (a *App) settingsMouseUp(x, y int) {
	// Snapshot first, then release capture. ReleaseCapture delivers
	// WM_CAPTURECHANGED synchronously, so reading the gesture state afterwards
	// would see whatever that handler just cleared.
	pressed := a.settingsPress
	dragging := a.settingsDrag
	moving := a.settingsMoving
	a.settingsPress = settingsNone
	a.settingsDrag = settingsNone
	a.settingsMoving = false
	win.ReleaseCapture()

	if moving {
		return
	}
	if dragging != settingsNone {
		// Persist once, on release: writing config.json per pixel of a drag
		// would be a file write per mouse move.
		a.publishSettings()
		if dragging == settingsVolume && a.options.Settings.Sound {
			beep(beepExclamation)
		}
		return
	}
	control, index, _, ok := a.settingsGeometry.hit(x, y)
	if !ok || control == settingsNone || control != pressed {
		_ = a.paintSettings()
		return
	}
	a.settingsApply(control, index)
}

// idleMinutes is the live idle threshold in whole minutes.
func (a *App) idleMinutes() int {
	seconds := a.options.Settings.IdleSeconds
	if seconds <= 0 {
		return 0
	}
	return seconds / 60
}

// setIdleMinutes clamps the idle threshold into the range config.json accepts
// and keeps the ui-side seconds in step. Zero is a real value there - "window
// focus only" - so it must survive the clamp.
func (a *App) setIdleMinutes(minutes int) {
	if minutes < 0 {
		minutes = 0
	}
	if minutes > 120 {
		minutes = 120
	}
	a.options.Settings.IdleSeconds = minutes * 60
}

// testAlert fires one balloon through the same path a real alert uses, so the
// three things that can go wrong (a balloon that never appears, a tone that
// never plays, a setting that is not the one being tested) are all visible at
// once. It deliberately counts against no session: a test alert must not mark a
// row unread or start a coalescing window on somebody's conversation.
func (a *App) testAlert(kind session.Kind) {
	switch kind {
	case session.KindDone:
		a.balloonText(i18n.T(a.lang, i18n.KeyBalloonDoneTitle),
			i18n.T(a.lang, i18n.KeyBalloonDoneBody), niifInfo)
		if a.options.Settings.Sound && ValidDonePolicy(a.options.Settings.DonePolicy) != DonePolicyOff {
			beep(beepFinished)
		}
	default:
		a.balloonText(i18n.T(a.lang, i18n.KeyBalloonWaitTitle),
			i18n.T(a.lang, i18n.KeySettingsTestWaitingHint), uint32(win.NIIF_WARNING))
		if soundWanted(session.KindApproval, a.attentionSettings()) {
			beep(beepExclamation)
		}
	}
}

// settingsApplySlider mirrors a drag position into the live settings without
// persisting. Release persists.
func (a *App) settingsApplySlider(control settingsControl, x int) {
	geometry := a.settingsGeometry
	for index := range geometry.rows {
		row := geometry.rows[index]
		if row.control != control {
			continue
		}
		value := settingsSliderValue(geometry.controls[index], row, x)
		switch control {
		case settingsOpacity:
			a.options.Opacity = uint8(value)
			a.postRefresh()
		case settingsVolume:
			a.options.Settings.SoundVolume = value
			setSoundVolume(value)
		}
		break
	}
	_ = a.paintSettings()
}

func (a *App) settingsApply(control settingsControl, index int) {
	settings := &a.options.Settings
	switch control {
	case settingsClose:
		a.hideSettings()
		return
	// Every tab in the sidebar must be listed here. settingsTabAppearance was
	// missing from this case for two releases: the tab hit-tested, hovered and
	// drew its selected state, and clicking it did nothing at all - which also
	// meant the three appearance settings had no reachable UI, because the tray
	// menu has no theme entry. TestSettingsEveryTabSwitches pins it.
	case settingsTabGeneral, settingsTabSound, settingsTabSessions, settingsTabAppearance:
		a.settingsTab = settingsTabIndex(control)
		a.reflowSettings(a.settingsDPI)
		a.settingsHover = settingsNone
		a.settingsHoverIndex = -1
		_ = a.paintSettings()
		return
	case settingsNotifications:
		settings.Notifications = !settings.Notifications
	case settingsAlwaysOnTop:
		settings.AlwaysOnTop = !settings.AlwaysOnTop
		a.applyTopmost()
	case settingsEdgeRetract:
		settings.EdgeRetract = !settings.EdgeRetract
		a.syncTuck()
	case settingsAutoStart:
		if a.options.OnAutoStartChanged != nil {
			a.options.AutoStart = a.options.OnAutoStartChanged(!a.options.AutoStart)
		}
	case settingsSound:
		settings.Sound = !settings.Sound
		if settings.Sound {
			beep(beepExclamation)
		}
	case settingsSoundWaiting:
		settings.SoundOnWaiting = !settings.SoundOnWaiting
	case settingsSoundError:
		settings.SoundOnError = !settings.SoundOnError
	case settingsSoundFinish:
		settings.NotifyOnFinish = !settings.NotifyOnFinish
		if settings.NotifyOnFinish {
			beep(beepFinished)
		}
	case settingsScheme:
		if index >= 0 && index < len(SoundSchemes) {
			settings.SoundScheme = SoundSchemes[index]
			setSoundScheme(settings.SoundScheme)
			// Audition immediately: choosing a tone by name alone is guesswork.
			beep(beepExclamation)
		}
	case settingsPreview:
		beep(beepExclamation)
	case settingsAutoApprove:
		settings.AutoApprove = !settings.AutoApprove
		if !settings.AutoApprove {
			a.approvals.update(nil, time.Now(), false, 0)
		}
		a.postRefresh()
	case settingsForegroundAware:
		settings.ForegroundAware = !settings.ForegroundAware
	// The stepper convention, unchanged: index 0 is the minus half, 1 the plus.
	case settingsIdleMinutes:
		if index == 0 {
			a.setIdleMinutes(a.idleMinutes() - 1)
		} else {
			a.setIdleMinutes(a.idleMinutes() + 1)
		}
	case settingsDonePolicy:
		if policy, ok := settingsDonePolicyAt(index); ok {
			settings.DonePolicy = string(policy)
			// Keep the legacy two-way view in step, so an older build reading
			// the same file still sees a coherent answer.
			settings.NotifyOnFinish = policy != DonePolicyOff
		}
	case settingsQuietCurrent:
		settings.QuietCurrent = !settings.QuietCurrent
	case settingsQuietSubagents:
		settings.QuietSubagents = !settings.QuietSubagents
	case settingsMuteClear:
		if a.mutes.clear() > 0 {
			settings.MutedSessions = a.mutes.list()
			a.postRefresh()
		}
	case settingsTestWaiting:
		a.testAlert(session.KindApproval)
	case settingsTestDone:
		a.testAlert(session.KindDone)
	case settingsApproveDown:
		a.setApproveSeconds(a.approveSeconds() - 1)
	case settingsApproveUp:
		a.setApproveSeconds(a.approveSeconds() + 1)
	case settingsLanguage:
		switch index {
		case 1:
			settings.Language = string(i18n.SettingZhCN)
		case 2:
			settings.Language = string(i18n.SettingEn)
		default:
			settings.Language = string(i18n.SettingAuto)
		}
		a.lang = i18n.Resolve(i18n.Setting(settings.Language), i18n.SystemLocale())
		a.updateTooltip()
		a.postRefresh()
	case settingsTextSize:
		if index >= 0 && index < len(config.TextSizes) {
			settings.SettingsTextSize = config.TextSizes[index]
			// The window is sized by its content, so changing the text size
			// changes the window. Reflow keeps the centre still.
			a.reflowSettings(a.settingsDPI)
		}
	case settingsKeyboard:
		settings.SettingsKeyboard = !settings.SettingsKeyboard
		if !settings.SettingsKeyboard {
			a.settingsFocusVisible = false
		}
	case settingsTooltips:
		settings.SettingsTooltips = !settings.SettingsTooltips
	case settingsTheme:
		if index >= 0 && index < len(ThemeNames) {
			settings.Theme = ThemeNames[index]
			applyAppearance(settings.Theme, settings.PerSourceColor)
			a.postRefresh()
			// The settings window shares the palette, so it repaints too.
			_ = a.paintSettings()
		}
	case settingsPerSource:
		settings.PerSourceColor = !settings.PerSourceColor
		applyAppearance(settings.Theme, settings.PerSourceColor)
		a.postRefresh()
	case settingsHiddenAtLaunch:
		settings.HiddenAtLaunch = !settings.HiddenAtLaunch
	case settingsReset:
		a.applyDefaultSettings()
		applyAppearance(a.options.Settings.Theme, a.options.Settings.PerSourceColor)
		a.postRefresh()
	}
	a.publishSettings()
	_ = a.paintSettings()
}

// setApproveSeconds clamps and applies the countdown length.
func (a *App) setApproveSeconds(seconds int) {
	if seconds < 1 {
		seconds = 1
	}
	if seconds > 60 {
		seconds = 60
	}
	a.approveTimeout = time.Duration(seconds) * time.Second
	a.options.Settings.AutoApproveTimeoutSec = seconds
	a.approvals.update(nil, time.Now(), false, 0)
	a.postRefresh()
}

// applyDefaultSettings restores every preference to its documented default.
// Window placement is deliberately untouched: "restore defaults" is about
// behaviour, and silently moving the panel would be a surprise.
//
// Two things are deliberately NOT restored: window placement (above) and the
// per-session mute list. A mute is a decision about one conversation - a
// background job the user silenced on purpose - and clearing it would start
// waking them again without being asked. The Sessions tab has an explicit
// "clear all" for that.
func (a *App) applyDefaultSettings() {
	defaults := DefaultSettings()
	current := a.options.Settings
	current.Notifications = defaults.Notifications
	current.Sound = defaults.Sound
	current.SoundVolume = defaults.SoundVolume
	current.SoundScheme = defaults.SoundScheme
	current.SoundOnWaiting = defaults.SoundOnWaiting
	current.SoundOnError = defaults.SoundOnError
	current.NotifyOnFinish = defaults.NotifyOnFinish
	current.AlwaysOnTop = defaults.AlwaysOnTop
	current.EdgeRetract = defaults.EdgeRetract
	current.Language = defaults.Language
	current.AutoApprove = defaults.AutoApprove
	current.AutoApproveTimeoutSec = defaults.AutoApproveTimeoutSec
	current.Theme = defaults.Theme
	current.PerSourceColor = defaults.PerSourceColor
	current.HiddenAtLaunch = defaults.HiddenAtLaunch
	current.ForegroundAware = defaults.ForegroundAware
	current.IdleSeconds = defaults.IdleSeconds
	current.DonePolicy = defaults.DonePolicy
	current.NotifyOnFinish = defaults.NotifyOnFinish
	current.QuietCurrent = defaults.QuietCurrent
	current.QuietSubagents = defaults.QuietSubagents
	// The mute list is a user decision about specific conversations, not a
	// preference: "restore defaults" restores behaviour, and silently unmuting
	// a background job the user silenced on purpose would start waking them
	// again. The Sessions tab has its own "clear all" button for that.
	current.MutedSessions = a.mutes.list()
	a.options.Settings = current
	a.options.Opacity = uint8(defaults.Opacity)
	setSoundVolume(current.SoundVolume)
	setSoundScheme(current.SoundScheme)
	a.approveTimeout = time.Duration(current.AutoApproveTimeoutSec) * time.Second
	a.lang = i18n.Resolve(i18n.Setting(current.Language), i18n.SystemLocale())
	a.applyTopmost()
	a.syncTuck()
	a.updateTooltip()
	a.postRefresh()
}
