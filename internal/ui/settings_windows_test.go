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

// TestAppearanceTabIsLaidOutLikeEveryOtherTab pins the fourth tab into the same
// geometry contract as the other three: every row fits the window, every control
// is hit-testable at its drawn centre, and the sidebar entry maps back to the tab
// it names.

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
