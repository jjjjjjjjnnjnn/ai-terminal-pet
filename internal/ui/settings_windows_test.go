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
	for _, tab := range []int{0, 1, 2} {
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
	for _, tab := range []int{0, 1, 2} {
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
		defaults.Opacity != expected.Opacity {
		t.Fatalf("ui.DefaultSettings() drifted from config.Default():\n ui: %+v\ncfg: %+v", defaults, expected)
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
	return settingsRowsIn(tab, i18n.LangZhCN)
}

func settingsRowsIn(tab int, lang i18n.Lang) []settingsRow {
	switch tab {
	case 1:
		return soundRows(lang)
	case 2:
		return sessionRows(lang)
	default:
		return generalRows(lang)
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
	}
	for _, tc := range cases {
		rows := settingsRowsIn(tc.tab, tc.lang)
		geometry := settingsGeometryFor(96, tc.tab, rows)
		target, err := newCanvas(geometry.width, geometry.height)
		if err != nil {
			t.Fatalf("%s: canvas: %v", tc.name, err)
		}
		drawSettings(target, geometry, view, tc.hover, tc.hoverAt, tc.press, tc.lang)
		img := compositeOver(target, desktop)
		target.close()
		writePNG(t, filepath.Join(dir, tc.name+".png"), upscale(img, 2))
	}
}
