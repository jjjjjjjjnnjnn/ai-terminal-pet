package app

import (
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/ui"
)

// TestThemeListsAgree pins the one duplication this feature needs: config owns
// the persisted name and ui owns the palette, and both normalise an unknown value
// to slate. They are separate on purpose - ui does not import config (see
// ui.DefaultSettings) - so a test has to hold them together.
func TestThemeListsAgree(t *testing.T) {
	if len(config.Themes) != len(ui.ThemeNames) {
		t.Fatalf("config offers %d themes, ui offers %d", len(config.Themes), len(ui.ThemeNames))
	}
	for index, name := range config.Themes {
		if ui.ThemeNames[index] != name {
			t.Fatalf("theme %d: config says %q, ui says %q", index, name, ui.ThemeNames[index])
		}
	}
}

// TestUISettingsRoundTripKeepsAppearance is what makes the settings window
// actually able to persist a palette: applyUISettings is the only bridge from the
// window's struct to config.json, so a field it forgets is a field that silently
// resets on the next launch.
func TestUISettingsRoundTripKeepsAppearance(t *testing.T) {
	base := config.Default()
	uiSettings := ui.Settings{
		Notifications:         base.Notifications,
		Sound:                 base.Sound,
		SoundVolume:           base.SoundVolume,
		SoundScheme:           base.SoundScheme,
		SoundOnWaiting:        base.SoundOnWaiting,
		SoundOnError:          base.SoundOnError,
		NotifyOnFinish:        base.NotifyOnFinish,
		AlwaysOnTop:           base.AlwaysOnTop,
		EdgeRetract:           base.EdgeRetract,
		Language:              base.Language,
		AutoApprove:           base.AutoApprove,
		AutoApproveTimeoutSec: base.AutoApproveTimeoutSec,
		Opacity:               base.Opacity,
		Theme:                 config.ThemeContrast,
		PerSourceColor:        true,
		HiddenAtLaunch:        true,
	}
	merged := base
	applyUISettings(&merged, uiSettings)
	if merged.Theme != config.ThemeContrast {
		t.Fatalf("theme did not survive the merge: %q", merged.Theme)
	}
	if !merged.PerSourceColor || !merged.HiddenAtLaunch {
		t.Fatalf("appearance switches did not survive the merge: %+v", merged)
	}

	// And an unknown theme coming back from the window is normalised rather than
	// written to disk.
	uiSettings.Theme = "nonsense"
	merged = base
	applyUISettings(&merged, uiSettings)
	if merged.Theme != config.DefaultTheme {
		t.Fatalf("an unknown theme was persisted as %q", merged.Theme)
	}
}
