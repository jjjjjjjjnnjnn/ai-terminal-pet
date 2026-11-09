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
	"math"
	"syscall"
	"time"
	"unsafe"

	"ai-terminal-pet/internal/i18n"
	win "github.com/lxn/win"
)

const (
	settingsClassName   = "AITerminalPetSettings"
	settingsWindowTitle = "AI Terminal Pet — Settings"

	// Logical layout at 96 DPI.
	settingsWidth = 452
	// settingsMinHeight keeps the sidebar (three tabs plus margins) from being
	// clipped on the shortest tab; the real height follows the content.
	settingsMinHeight    = 260
	settingsHeaderHeight = 52
	settingsMargin       = 20
	settingsSidebarWidth = 104
	settingsTabsTop      = 68
	settingsTabHeight    = 32
	settingsTabGap       = 4
	settingsContentLeft  = 160
	settingsRowsTop      = 66
	settingsPanelRadius  = 14
)

// settingsContentWidth is the width of the right-hand column.
const settingsContentWidth = settingsWidth - settingsMargin - settingsContentLeft

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

func settingsRowHeight(kind int) int {
	switch kind {
	case rowSlider:
		return 68
	case rowSegment:
		// A segment needs the title line, the hint line and then its own band;
		// 80 leaves a visible gap under the hint so the two never touch.
		return 80
	case rowNote:
		return 58
	default:
		// Toggles, steppers and buttons share the title line with their
		// control, so one line plus the hint is all the height they need.
		return 54
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

	settingsAutoApprove
	settingsApproveDown
	settingsApproveUp
	settingsReset
)

// settingsTabControl maps a sidebar index onto its control id.
func settingsTabControl(index int) settingsControl {
	switch index {
	case 1:
		return settingsTabSound
	case 2:
		return settingsTabSessions
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
	default:
		return 0
	}
}

// settingsRow is one line of the right-hand column.
type settingsRow struct {
	control settingsControl
	// alt is the second control a stepper needs (its minus button). It is
	// settingsNone for every other kind.
	alt   settingsControl
	kind  int
	title string
	hint  string
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
	// lang is the language the row labels were built in. Switching language
	// changes every string but not the layout, so the geometry has to be
	// rebuilt to pick the new labels up.
	lang  i18n.Lang
	close settingsRect
	tabs  [3]settingsRect
	tab   int
	rows  []settingsRow
	// frames is the clickable band of each row (the whole line, so a label is
	// as clickable as its switch), controls the drawn sub-control inside it.
	frames   []settingsRect
	controls []settingsRect
}

func (g settingsGeometry) scale(base int) int {
	return maxInt(int(math.Round(float64(base)*float64(g.dpi)/96)), 1)
}

// settingsGeometryFor resolves the layout. Pure: same inputs, same output, no
// DPI query, which is what makes it testable off a real window.
func settingsGeometryFor(dpi uint32, tab int, rows []settingsRow) settingsGeometry {
	if dpi == 0 {
		dpi = 96
	}
	geometry := settingsGeometry{
		width: maxInt(int(math.Round(float64(settingsWidth)*float64(dpi)/96)), 96),
		dpi:   dpi,
		tab:   tab,
		rows:  rows,
	}
	s := geometry.scale

	closeSize := s(24)
	geometry.close = settingsRect{
		Left:   geometry.width - s(settingsMargin) - closeSize,
		Top:    s(14),
		Width:  closeSize,
		Height: closeSize,
	}
	for index := 0; index < 3; index++ {
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
		height := s(settingsRowHeight(row.kind))
		frame := settingsRect{Left: contentLeft, Top: y, Width: contentWidth, Height: height}
		geometry.frames = append(geometry.frames, frame)
		geometry.controls = append(geometry.controls, settingsControlRect(frame, row, s))
		y += height
	}
	// The window is exactly as tall as its content. A fixed height left the
	// shortest tab with half a screen of emptiness, which reads as a broken
	// layout rather than a compact one.
	geometry.height = maxInt(y+geometry.scale(settingsMargin), geometry.scale(settingsMinHeight))
	return geometry
}

// settingsControlRect places the interactive part of a row. A toggle and a
// stepper sit on the title line at the right edge; a slider, a segment and a
// button get their own band underneath it.
func settingsControlRect(frame settingsRect, row settingsRow, s func(int) int) settingsRect {
	switch row.kind {
	case rowToggle:
		width, height := s(38), s(21)
		return settingsRect{Left: frame.right() - width, Top: frame.Top + s(3), Width: width, Height: height}
	case rowSlider:
		return settingsRect{Left: frame.Left, Top: frame.Top + s(38), Width: frame.Width, Height: s(4)}
	case rowSegment:
		return settingsRect{Left: frame.Left, Top: frame.Top + s(38), Width: frame.Width, Height: s(28)}
	case rowStepper:
		return settingsRect{Left: frame.right() - s(92), Top: frame.Top + s(2), Width: s(92), Height: s(24)}
	case rowButton:
		width := s(76)
		return settingsRect{Left: frame.right() - width, Top: frame.Top + s(1), Width: width, Height: s(24)}
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
		Notifications:  settings.Notifications,
		AlwaysOnTop:    settings.AlwaysOnTop,
		EdgeRetract:    settings.EdgeRetract,
		Sound:          settings.Sound,
		SoundOnWaiting: settings.SoundOnWaiting,
		SoundOnError:   settings.SoundOnError,
		NotifyOnFinish: settings.NotifyOnFinish,
		AutoApprove:    settings.AutoApprove,
		AutoStart:      a.options.AutoStart,
		Opacity:        int(a.options.Opacity),
		Volume:         settings.SoundVolume,
		ApproveTimeout: a.approveSeconds(),
		Scheme:         validSoundScheme(settings.SoundScheme),
		Language:       settings.Language,
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

func (a *App) settingsRows() []settingsRow {
	switch a.settingsTab {
	case 1:
		return soundRows(a.lang)
	case 2:
		return sessionRows(a.lang)
	default:
		return generalRows(a.lang)
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
	}
}

func sessionRows(lang i18n.Lang) []settingsRow {
	return []settingsRow{
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
	}
	return -1
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

func drawSettings(target *canvas, geometry settingsGeometry, view settingsView, hover settingsControl, hoverIndex int, pressed settingsControl, lang i18n.Lang) {
	s := geometry.scale
	width, height := geometry.width, geometry.height

	target.clear()
	target.fillRounded(0, 0, width, height, s(settingsPanelRadius), colorPanel)
	target.outlineRounded(0, 0, width, height, s(settingsPanelRadius), 1, colorBorder)

	// Header.
	target.text(s(settingsMargin), s(12), width-s(settingsMargin)*2-s(40), s(26),
		i18n.T(lang, i18n.KeySettingsTitle), s(15), 600, colorPrimary,
		dtSingleLine|dtVCenter|dtNoPrefix)
	subtitleWidth := s(150)
	target.text(width-s(settingsMargin)-s(30)-subtitleWidth, s(12), subtitleWidth, s(26),
		i18n.T(lang, i18n.KeySettingsSubtitle), s(10), 400, colorMuted,
		dtSingleLine|dtVCenter|dtRight|dtNoPrefix)
	target.fillRect(s(settingsMargin), s(settingsHeaderHeight), width-s(settingsMargin)*2, 1, colorDivider)
	drawSettingsClose(target, geometry, hover == settingsClose || pressed == settingsClose)

	// Sidebar.
	for index := range geometry.tabs {
		rect := geometry.tabs[index]
		control := settingsTabControl(index)
		selected := index == geometry.tab
		fill := rgba{}
		if selected {
			fill = colorRaised
		} else if hover == control {
			fill = rgba{colorRaised.R, colorRaised.G, colorRaised.B, 120}
		}
		if fill.A != 0 {
			target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(8), fill)
		}
		if selected {
			target.fillRounded(rect.Left+s(1), rect.Top+s(9), s(2), rect.Height-s(18), s(1), colorAccent)
		}
		color := colorSecondary
		if selected {
			color = colorPrimary
		}
		label := settingsTabLabel(index, lang)
		target.text(rect.Left+s(14), rect.Top, rect.Width-s(18), rect.Height, label, s(12), 500, color,
			dtSingleLine|dtVCenter|dtNoPrefix)
	}
	dividerX := s(settingsContentLeft) - s(18)
	target.fillRect(dividerX, s(settingsRowsTop)-s(6), 1, height-s(settingsRowsTop)-s(settingsMargin)+s(6), colorDivider)

	// Rows.
	for index, row := range geometry.rows {
		frame := geometry.frames[index]
		control := geometry.controls[index]
		hot := hover == row.control || hover == row.alt
		if row.kind == rowNote {
			target.text(frame.Left, frame.Top, frame.Width, frame.Height, row.title, s(10), 400, colorHint,
				dtWordBreak|dtNoPrefix)
			continue
		}
		titleWidth := frame.Width
		switch row.kind {
		case rowToggle:
			titleWidth = frame.Width - s(46)
		case rowStepper:
			titleWidth = frame.Width - s(100)
		case rowButton:
			titleWidth = frame.Width - s(88)
		case rowSlider:
			titleWidth = frame.Width - s(52)
		}
		target.text(frame.Left, frame.Top, titleWidth, s(20), row.title, s(12), 500, colorPrimary,
			dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
		if row.hint != "" {
			target.text(frame.Left, frame.Top+s(19), frame.Width, s(16), row.hint, s(10), 400, colorHint,
				dtSingleLine|dtVCenter|dtNoPrefix|dtEndEllipsis)
		}

		switch row.kind {
		case rowToggle:
			drawSettingsToggle(target, control, settingsToggleState(row, view), hot, pressed == row.control)
		case rowSlider:
			label := settingsValueLabel(row, view)
			target.text(frame.right()-s(52), frame.Top, s(52), s(20), label, s(11), 500, colorSecondary,
				dtSingleLine|dtVCenter|dtRight|dtNoPrefix)
			drawSettingsSlider(target, control, settingsSliderValueFor(row, view), hot, s(14))
		case rowSegment:
			drawSettingsSegment(target, geometry, index, row, view, hover, hoverIndex)
		case rowStepper:
			drawSettingsStepper(target, geometry, index, row, view, hover, pressed)
		case rowButton:
			fill := colorRaised
			if hot {
				fill = colorPanelHover
			}
			if pressed == row.control {
				fill = colorControlBg
			}
			target.fillRounded(control.Left, control.Top, control.Width, control.Height, s(8), fill)
			target.outlineRounded(control.Left, control.Top, control.Width, control.Height, s(8), 1, colorBorder)
			labelColor := colorPrimary
			if hot {
				labelColor = colorPrimary
			}
			target.text(control.Left, control.Top, control.Width, control.Height, row.action, s(11), 500, labelColor,
				dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
		}
	}
}

func settingsTabLabel(index int, lang i18n.Lang) string {
	switch index {
	case 1:
		return i18n.T(lang, i18n.KeySettingsTabSound)
	case 2:
		return i18n.T(lang, i18n.KeySettingsTabSessions)
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
		target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(7), colorRaised)
	}
	color := colorSecondary
	if hot {
		color = colorPrimary
	}
	inset := rect.Width / 3
	left, top := rect.Left+inset, rect.Top+inset
	right, bottom := rect.right()-inset, rect.bottom()-inset
	thickness := maxInt(s(1), 1)
	target.line(left, top, right, bottom, thickness, color)
	target.line(right, top, left, bottom, thickness, color)
}

func drawSettingsToggle(target *canvas, rect settingsRect, on, hot, pressed bool) {
	track := colorTrackOff
	if on {
		track = colorAccent
	}
	if hot && !on {
		track = rgba{track.R, track.G, track.B, 255}
		track = rgba{colorSecondary.R, colorSecondary.G, colorSecondary.B, 210}
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
	target.fillEllipse(knobLeft, rect.Top+gap, knob, knob, colorPrimary)
}

func drawSettingsSlider(target *canvas, track settingsRect, fraction float64, hot bool, knobSize int) {
	target.fillRounded(track.Left, track.Top, track.Width, track.Height, track.Height/2, colorControlBg)
	filled := int(math.Round(float64(track.Width) * fraction))
	if filled > 0 {
		target.fillRounded(track.Left, track.Top, filled, track.Height, track.Height/2, colorAccent)
	}
	knobLeft := track.Left + filled - knobSize/2
	if knobLeft < track.Left-knobSize/2 {
		knobLeft = track.Left - knobSize/2
	}
	if knobLeft > track.right()-knobSize/2 {
		knobLeft = track.right() - knobSize/2
	}
	knobTop := track.Top + track.Height/2 - knobSize/2
	ring := colorBorder
	if hot {
		ring = colorAccent
	}
	target.fillEllipse(knobLeft, knobTop, knobSize, knobSize, ring)
	target.fillEllipse(knobLeft+2, knobTop+2, knobSize-4, knobSize-4, colorPrimary)
}

func drawSettingsSegment(target *canvas, geometry settingsGeometry, index int, row settingsRow, view settingsView, hover settingsControl, hoverIndex int) {
	s := geometry.scale
	rect := geometry.controls[index]
	target.fillRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(8), colorControlBg)
	target.outlineRounded(rect.Left, rect.Top, rect.Width, rect.Height, s(8), 1, colorDivider)
	count := len(row.options)
	selected := settingsSegmentIndex(row, view)
	for option := 0; option < count; option++ {
		cell := settingsSegmentRect(rect, count, option).inset(s(2))
		if option == selected {
			target.fillRounded(cell.Left, cell.Top, cell.Width, cell.Height, s(6), colorRaised)
			target.outlineRounded(cell.Left, cell.Top, cell.Width, cell.Height, s(6), 1, colorAccent)
		} else if hover == row.control && hoverIndex == option {
			target.fillRounded(cell.Left, cell.Top, cell.Width, cell.Height, s(6), colorRaisedRow)
		}
		color := colorSecondary
		if option == selected {
			color = colorPrimary
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
		fill := colorRaised
		if hover == item.control {
			fill = colorPanelHover
		}
		if pressed == item.control {
			fill = colorControlBg
		}
		target.fillRounded(item.rect.Left, item.rect.Top, item.rect.Width, item.rect.Height, s(6), fill)
		target.text(item.rect.Left, item.rect.Top, item.rect.Width, item.rect.Height, item.glyph, s(12), 500, colorPrimary,
			dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
	}
	target.text(value.Left, value.Top, value.Width, value.Height, itoa(view.ApproveTimeout)+"s",
		s(11), 600, colorPrimary, dtSingleLine|dtVCenter|dtCenter|dtNoPrefix)
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
	a.paintSettings()
	win.ShowWindow(win.HWND(handle), win.SW_SHOW)
	win.SetForegroundWindow(win.HWND(handle))
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
	geometry := settingsGeometryFor(dpi, a.settingsTab, a.settingsRows())
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
	geometry := settingsGeometryFor(dpi, a.settingsTab, a.settingsRows())
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
	dpi := a.windowDpi()
	if dpi != a.settingsDPI || len(a.settingsGeometry.rows) == 0 ||
		a.settingsGeometry.tab != a.settingsTab || a.settingsGeometry.lang != a.lang {
		a.reflowSettings(dpi)
	}
	geometry := a.settingsGeometry
	view := a.settingsView()
	key := settingsFrameKey{
		width: geometry.width, height: geometry.height, tab: a.settingsTab,
		hover: a.settingsHover, hoverIndex: a.settingsHoverIndex, pressed: a.settingsPress,
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
	drawSettings(a.settingsSurface, geometry, view, a.settingsHover, a.settingsHoverIndex, a.settingsPress, a.lang)
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
	case wmKeyDown:
		if wParam == wmKeyEscape {
			a.hideSettings()
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
	case settingsTabGeneral, settingsTabSound, settingsTabSessions:
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
	case settingsReset:
		a.applyDefaultSettings()
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
