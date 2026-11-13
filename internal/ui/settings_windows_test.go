package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
)

// TestSettingsLayoutFitsTheWindow locks the two things that are easy to break
// when a row is added: nothing may hang below the panel, and the sidebar may not
// run into the content column.
func TestSettingsLayoutFitsTheWindow(t *testing.T) {
	for _, tab := range []int{0, 1, 2, 3} {
		rows := settingsRowsFor(t, tab)
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometryFor(dpi, tab, rows)
			if geometry.width <= 0 || geometry.height <= 0 {
				t.Fatalf("tab %d dpi %d: empty geometry", tab, dpi)
			}
			last := geometry.frames[len(geometry.frames)-1]
			if bottom := last.bottom(); bottom > geometry.height-geometry.scale(settingsMargin)/2 {
				t.Fatalf("tab %d dpi %d: last row ends at %d, panel is %d tall", tab, dpi, bottom, geometry.height)
			}
			for index, frame := range geometry.frames {
				if frame.Left < geometry.scale(settingsContentLeft)-1 {
					t.Fatalf("tab %d dpi %d: row %d starts left of the content column", tab, dpi, index)
				}
				if frame.right() > geometry.width-geometry.scale(settingsMargin)+1 {
					t.Fatalf("tab %d dpi %d: row %d overflows the right margin", tab, dpi, index)
				}
				if index > 0 && frame.Top < geometry.frames[index-1].bottom() {
					t.Fatalf("tab %d dpi %d: row %d overlaps row %d", tab, dpi, index, index-1)
				}
			}
			for index := range geometry.tabs {
				tab := geometry.tabs[index]
				if tab.right() >= geometry.scale(settingsContentLeft) {
					t.Fatalf("sidebar tab %d runs into the content column", index)
				}
			}
			if !geometry.close.contains(geometry.width-geometry.scale(settingsMargin)-1, geometry.close.Top+1) {
				t.Fatal("close button is not inside the right margin")
			}
		}
	}
}

// TestSettingsHitTestFindsEveryControl is the guard against a control that is
// drawn but unclickable: for every interactive row the centre of the drawn
// sub-rect must resolve back to that row's control.
func TestSettingsHitTestFindsEveryControl(t *testing.T) {
	for _, tab := range []int{0, 1, 2, 3} {
		rows := settingsRowsFor(t, tab)
		geometry := settingsGeometryFor(96, tab, rows)
		for index, row := range rows {
			if row.kind == rowNote {
				continue
			}
			rect := geometry.controls[index]
			x, y := rect.Left+rect.Width/2, rect.Top+rect.Height/2
			control, _, _, ok := geometry.hit(x, y)
			if !ok {
				t.Fatalf("tab %d row %d (%d): nothing hit at the control centre", tab, index, row.control)
			}
			// A stepper resolves to either half depending on x, and a segment
			// resolves to the option under the cursor; both are correct as long
			// as they belong to the same row.
			if control != row.control && control != row.alt {
				t.Fatalf("tab %d row %d: hit %d, want %d (or %d)", tab, index, control, row.control, row.alt)
			}
		}
		for index := range geometry.tabs {
			rect := geometry.tabs[index]
			control, _, _, ok := geometry.hit(rect.Left+rect.Width/2, rect.Top+rect.Height/2)
			if !ok || control != settingsTabControl(index) {
				t.Fatalf("sidebar tab %d did not hit its control", index)
			}
		}
	}
}

// TestSettingsSegmentsTileTheControl: the selected pill and the hit test derive
// their geometry from the same helper, so the options must cover the control
// exactly - no gap, no overlap, no rounding drift.
func TestSettingsSegmentsTileTheControl(t *testing.T) {
	control := settingsRect{Left: 100, Top: 40, Width: 272, Height: 28}
	for _, count := range []int{2, 3, 4} {
		previous := control.Left
		for index := 0; index < count; index++ {
			cell := settingsSegmentRect(control, count, index)
			if cell.Left != previous {
				t.Fatalf("count %d index %d: cell starts at %d, previous ended at %d", count, index, cell.Left, previous)
			}
			if cell.Height != control.Height {
				t.Fatalf("count %d index %d: height %d, want %d", count, index, cell.Height, control.Height)
			}
			previous = cell.right()
		}
		if previous != control.right() {
			t.Fatalf("count %d: cells end at %d, control ends at %d", count, previous, control.right())
		}
	}
}

func TestSettingsSliderMapping(t *testing.T) {
	track := settingsRect{Left: 10, Top: 0, Width: 200, Height: 4}
	row := settingsRow{kind: rowSlider, minimum: 40, maximum: 255}
	if got := settingsSliderValue(track, row, 10); got != 40 {
		t.Fatalf("left edge = %d, want the minimum", got)
	}
	if got := settingsSliderValue(track, row, 210); got != 255 {
		t.Fatalf("right edge = %d, want the maximum", got)
	}
	if got := settingsSliderValue(track, row, -500); got != 40 {
		t.Fatalf("far left = %d, want the minimum", got)
	}
	if got := settingsSliderValue(track, row, 9999); got != 255 {
		t.Fatalf("far right = %d, want the maximum", got)
	}
	if got := settingsSliderValue(track, row, 110); got != 148 {
		t.Fatalf("midpoint = %d, want 148 (round(40+107.5))", got)
	}
}

// TestDefaultSettingsMirrorConfig is the pin between the two packages that must
// not drift: ui cannot import config (the app layer owns file paths), so the
// defaults and the scheme names are mirrored, and this test is the contract.
func TestDefaultSettingsMirrorConfig(t *testing.T) {
	defaults := DefaultSettings()
	expected := config.Default()
	if defaults.Notifications != expected.Notifications ||
		defaults.Sound != expected.Sound ||
		defaults.SoundVolume != expected.SoundVolume ||
		defaults.SoundScheme != expected.SoundScheme ||
		defaults.SoundOnWaiting != expected.SoundOnWaiting ||
		defaults.SoundOnError != expected.SoundOnError ||
		defaults.NotifyOnFinish != expected.NotifyOnFinish ||
		defaults.AlwaysOnTop != expected.AlwaysOnTop ||
		defaults.EdgeRetract != expected.EdgeRetract ||
		defaults.Language != expected.Language ||
		defaults.AutoApprove != expected.AutoApprove ||
		defaults.AutoApproveTimeoutSec != expected.AutoApproveTimeoutSec ||
		defaults.Opacity != expected.Opacity ||
		defaults.ForegroundAware != expected.ForegroundAware ||
		defaults.IdleSeconds != expected.IdleMinutes*60 ||
		defaults.DonePolicy != expected.DonePolicy ||
		defaults.QuietCurrent != expected.QuietCurrent ||
		defaults.QuietSubagents != expected.QuietSubagents ||
		defaults.SettingsTextSize != expected.SettingsTextSize ||
		defaults.SettingsKeyboard != expected.SettingsKeyboard ||
		defaults.SettingsTooltips != expected.SettingsTooltips {
		t.Fatalf("ui.DefaultSettings() drifted from config.Default():\n ui: %+v\ncfg: %+v", defaults, expected)
	}
	// Every text size config.json accepts must have a ramp, and every ramp must
	// be a size config.json can store. A name on one side only means a config
	// file that silently falls back to standard, which is the kind of "my
	// setting does nothing" report this round started with.
	if len(settingsTypeRamps) != len(config.TextSizes) {
		t.Fatalf("ui has %d text ramps, config has %d text sizes", len(settingsTypeRamps), len(config.TextSizes))
	}
	for _, name := range config.TextSizes {
		if _, ok := settingsTypeRamps[name]; !ok {
			t.Fatalf("config offers text size %q but ui has no ramp for it", name)
		}
		if got := config.ValidTextSize(name); got != name {
			t.Fatalf("config rejects its own text size %q (got %q)", name, got)
		}
		if got := normalizeTextSize(name); got != name {
			t.Fatalf("ui normalises %q to %q", name, got)
		}
	}
	for name := range settingsTypeRamps {
		if got := config.ValidTextSize(name); got != name {
			t.Fatalf("ui has a ramp for %q but config normalises it to %q", name, got)
		}
	}
	if len(SoundSchemes) != len(config.SoundSchemes) {
		t.Fatalf("ui has %d sound schemes, config has %d", len(SoundSchemes), len(config.SoundSchemes))
	}
	for index, name := range config.SoundSchemes {
		if SoundSchemes[index] != name {
			t.Fatalf("scheme %d is %q in ui and %q in config", index, SoundSchemes[index], name)
		}
		if got := config.ValidSoundScheme(name); got != name {
			t.Fatalf("config rejects its own scheme %q (got %q)", name, got)
		}
		if got := validSoundScheme(name); got != name {
			t.Fatalf("ui rejects scheme %q (got %q)", name, got)
		}
	}
}

func settingsRowsFor(t *testing.T, tab int) []settingsRow {
	t.Helper()
	return settingsRowsIn(t, tab, i18n.LangZhCN)
}

func settingsRowsIn(t *testing.T, tab int, lang i18n.Lang) []settingsRow {
	return settingsRowsMeasuredFor(t, tab, lang, textSizeStandard)
}

// settingsRowsMeasuredFor is the production path: build the rows, then measure
// how many lines each hint needs at this text size and column width.
func settingsRowsMeasuredFor(t *testing.T, tab int, lang i18n.Lang, textSize string) []settingsRow {
	t.Helper()
	var rows []settingsRow
	switch tab {
	case 1:
		rows = soundRows(lang)
	case 2:
		rows = sessionRows(lang, 0)
	case 3:
		rows = appearanceRows(lang)
	default:
		rows = generalRows(lang)
	}
	target, err := newCanvas(1, 1)
	if err != nil {
		t.Fatalf("measure canvas: %v", err)
	}
	defer target.close()
	probe := settingsGeometrySized(96, textSize, tab, nil)
	return measureSettingsRows(target, rows, settingsTypeRamps[textSize], probe.scale(settingsContentWidth))
}

// TestAppearanceTabIsLaidOutLikeEveryOtherTab pins the fourth tab into the same
// geometry contract as the other three: every row fits the window, every control
// is hit-testable at its drawn centre, and the sidebar entry maps back to the tab
// it names.
func TestAppearanceTabIsLaidOutLikeEveryOtherTab(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangZhCN, i18n.LangEn} {
		rows := appearanceRows(lang)
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometryFor(dpi, 3, rows)
			last := geometry.frames[len(geometry.frames)-1]
			if bottom := last.bottom(); bottom > geometry.height-geometry.scale(settingsMargin)/2 {
				t.Fatalf("%v dpi %d: appearance tab overflows the panel (%d > %d)", lang, dpi, bottom, geometry.height)
			}
			for index, row := range rows {
				// A rowNote is a caption, not a control: it owns no hit target
				// by design, which is why the loop skips it rather than
				// asserting something the geometry never promised.
				if row.control == settingsNone {
					continue
				}
				rect := geometry.controls[index]
				control, _, _, ok := geometry.hit(rect.Left+rect.Width/2, rect.Top+rect.Height/2)
				if !ok || control != row.control {
					t.Fatalf("%v dpi %d row %d: hit %d, want %d", lang, dpi, index, control, row.control)
				}
			}
			if tab := settingsTabControl(3); tab != settingsTabAppearance {
				t.Fatalf("sidebar entry 3 maps to %d, want settingsTabAppearance", tab)
			}
			if index := settingsTabIndex(settingsTabAppearance); index != 3 {
				t.Fatalf("settingsTabAppearance maps to index %d, want 3", index)
			}
		}
	}
	if settingsTabCount != 4 {
		t.Fatalf("settingsTabCount = %d, want 4", settingsTabCount)
	}
}

// TestSessionsTabIsLaidOutLikeEveryOtherTab is the same contract for the tab
// that grew: every row fits, every control is hit-testable at its drawn centre,
// and the attentiveness rows name the controls their handlers switch on. The
// height assertion is the one that catches a tab that grew past the window,
// which is what a four-row addition is most likely to do.
func TestSessionsTabIsLaidOutLikeEveryOtherTab(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangZhCN, i18n.LangEn} {
		rows := sessionRows(lang, 2)
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometryFor(dpi, 2, rows)
			last := geometry.frames[len(geometry.frames)-1]
			if bottom := last.bottom(); bottom > geometry.height-geometry.scale(settingsMargin)/2 {
				t.Fatalf("%v dpi %d: sessions tab overflows the panel (%d > %d)", lang, dpi, bottom, geometry.height)
			}
			for index, row := range rows {
				if row.control == settingsNone {
					continue
				}
				rect := geometry.controls[index]
				control, _, _, ok := geometry.hit(rect.Left+rect.Width/2, rect.Top+rect.Height/2)
				if !ok || control != row.control {
					t.Fatalf("%v dpi %d row %d: hit %d, want %d", lang, dpi, index, control, row.control)
				}
			}
		}
	}
}

// The Done segment and the persisted policy have to agree in both directions,
// or clicking "Off" would select "Background" on the next paint.
func TestDonePolicySegmentRoundTrip(t *testing.T) {
	row := settingsRow{}
	for _, candidate := range sessionRows(i18n.LangEn, 0) {
		if candidate.control == settingsDonePolicy {
			row = candidate
		}
	}
	if len(row.options) != len(DonePolicyNames()) {
		t.Fatalf("done segment offers %d options, DonePolicyNames has %d", len(row.options), len(DonePolicyNames()))
	}
	for index, policy := range DonePolicyNames() {
		if got := settingsDonePolicyIndex(string(policy)); got != index {
			t.Fatalf("settingsDonePolicyIndex(%q) = %d, want %d", policy, got, index)
		}
		back, ok := settingsDonePolicyAt(index)
		if !ok || back != policy {
			t.Fatalf("settingsDonePolicyAt(%d) = %q/%v, want %q", index, back, ok, policy)
		}
	}
	if got := settingsDonePolicyIndex("nonsense"); got != 0 {
		t.Fatalf("unknown policy selected segment %d, want 0", got)
	}
	// Out-of-range indices keep the current value rather than inventing one.
	if _, ok := settingsDonePolicyAt(7); ok {
		t.Fatal("index 7 is not an option and must not resolve")
	}
	view := settingsView{DonePolicy: string(DonePolicyOff)}
	if got := settingsSegmentIndex(row, view); got != 1 {
		t.Fatalf("off should select segment 1, got %d", got)
	}
}

// The mutable-view settings (view) are a derived snapshot; the live ones are
// what the handlers read. A field missing from settingsView cannot be drawn.
func TestSettingsViewCarriesAttentiveness(t *testing.T) {
	a := &App{
		options: Options{Settings: Settings{
			ForegroundAware: true,
			IdleSeconds:     10 * 60,
			DonePolicy:      string(DonePolicyAlways),
			QuietCurrent:    false,
			QuietSubagents:  false,
			MutedSessions:   []string{"src\x00id"},
		}},
		mutes: newMuteSet([]string{"src\x00id"}),
	}
	view := a.settingsView()
	if !view.ForegroundAware {
		t.Fatal("foreground awareness did not reach the view")
	}
	if view.IdleMinutes != 10 {
		t.Fatalf("idle minutes in view = %d, want 10", view.IdleMinutes)
	}
	if view.DonePolicy != string(DonePolicyAlways) {
		t.Fatalf("done policy in view = %q", view.DonePolicy)
	}
	if view.QuietCurrent || view.QuietSubagents {
		t.Fatal("quiet switches were inverted in the view")
	}
	if view.MuteCount != 1 {
		t.Fatalf("mute count in view = %d, want 1", view.MuteCount)
	}
}

// TestThemeSegmentMatchesTheThemeList keeps the segment order and the palette
// list from drifting: the index the user clicks has to be the palette it names.
func TestThemeSegmentMatchesTheThemeList(t *testing.T) {
	rows := appearanceRows(i18n.LangEn)
	var theme settingsRow
	for _, row := range rows {
		if row.control == settingsTheme {
			theme = row
		}
	}
	if len(theme.options) != len(ThemeNames) {
		t.Fatalf("theme segment offers %d options, ThemeNames has %d", len(theme.options), len(ThemeNames))
	}
	for index, name := range ThemeNames {
		if got := settingsThemeIndex(name); got != index {
			t.Fatalf("settingsThemeIndex(%q) = %d, want %d", name, got, index)
		}
	}
	if got := settingsThemeIndex("nonsense"); got != 0 {
		t.Fatalf("unknown theme selected segment %d, want 0", got)
	}
}

// TestRenderSettingsPreview rasterises the settings window so the layout can be
// reviewed without a live window (a layered window cannot be screenshotted).
//
//	PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderSettingsPreview
func TestRenderSettingsPreview(t *testing.T) {
	dir := os.Getenv("PET_PREVIEW_DIR")
	if dir == "" {
		t.Skip("set PET_PREVIEW_DIR to render settings previews")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	desktop := color.RGBA{R: 0x20, G: 0x22, B: 0x2b, A: 0xff}
	view := settingsView{
		Notifications: true, AlwaysOnTop: true, EdgeRetract: true,
		Sound: true, SoundOnWaiting: true, SoundOnError: true, NotifyOnFinish: true,
		AutoApprove: false, AutoStart: false,
		Opacity: 235, Volume: 85, ApproveTimeout: 5,
		Scheme: config.SoundSchemeChime, Language: "zh-CN",
		// The previews must show the shipped default, so the theme here is
		// slate and the provenance switch is off. The appearance tab is rendered
		// separately below, once per palette.
		Theme: config.ThemeSlate,
	}
	cases := []struct {
		name    string
		tab     int
		lang    i18n.Lang
		hover   settingsControl
		hoverAt int
		press   settingsControl
	}{
		{"settings-general", 0, i18n.LangZhCN, settingsEdgeRetract, -1, settingsNone},
		{"settings-general-en", 0, i18n.LangEn, settingsLanguage, 2, settingsNone},
		{"settings-sound", 1, i18n.LangZhCN, settingsScheme, 2, settingsNone},
		{"settings-sound-en", 1, i18n.LangEn, settingsVolume, -1, settingsNone},
		{"settings-sessions", 2, i18n.LangZhCN, settingsApproveUp, -1, settingsNone},
		{"settings-appearance", 3, i18n.LangZhCN, settingsPerSource, -1, settingsNone},
		{"settings-appearance-en", 3, i18n.LangEn, settingsTheme, 1, settingsNone},
	}
	for _, tc := range cases {
		rows := settingsRowsIn(t, tc.tab, tc.lang)
		geometry := settingsGeometryFor(96, tc.tab, rows)
		target, err := newCanvas(geometry.width, geometry.height)
		if err != nil {
			t.Fatalf("%s: canvas: %v", tc.name, err)
		}
		drawSettings(target, settingsPaint{
			geometry: geometry,
			view:     view,
			hover:    tc.hover,
			hoverAt:  tc.hoverAt,
			pressed:  tc.press,
			focus:    -1,
			lang:     tc.lang,
		})
		img := compositeOver(target, desktop)
		target.close()
		writePNG(t, filepath.Join(dir, tc.name+".png"), upscale(img, 2))
	}
}
