package app

// The settings window's own preferences have to survive the same bridge as
// every other preference: applyUISettings is the only path from a click to
// config.json, and a field it forgets resets silently on the next launch. That
// is exactly the failure mode the dead Look tab produced for the theme, so the
// three fields that tab now owns are pinned here rather than eyeballed in a
// sandbox.

import (
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/ui"
)

func TestUISettingsRoundTripKeepsSettingsWindowPreferences(t *testing.T) {
	base := config.Default()
	settings := ui.Settings{
		Notifications:         base.Notifications,
		Sound:                 base.Sound,
		SoundVolume:           base.SoundVolume,
		SoundScheme:           base.SoundScheme,
		SoundOnWaiting:        base.SoundOnWaiting,
		SoundOnError:          base.SoundOnError,
		NotifyOnFinish:        base.NotifyOnFinish,
		DonePolicy:            base.DonePolicy,
		ForegroundAware:       base.ForegroundAware,
		IdleSeconds:           base.IdleMinutes * 60,
		QuietCurrent:          base.QuietCurrent,
		MutedSessions:         nil,
		QuietSubagents:        base.QuietSubagents,
		AlwaysOnTop:           base.AlwaysOnTop,
		EdgeRetract:           base.EdgeRetract,
		Language:              base.Language,
		AutoApprove:           base.AutoApprove,
		AutoApproveTimeoutSec: base.AutoApproveTimeoutSec,
		Opacity:               base.Opacity,
		Theme:                 base.Theme,
		PerSourceColor:        base.PerSourceColor,
		HiddenAtLaunch:        true,
		SettingsTextSize:      config.TextSizeLarge,
		SettingsKeyboard:      true,
		SettingsTooltips:      false,
	}

	merged := base
	applyUISettings(&merged, settings)
	merged = merged.Normalize()

	if merged.SettingsTextSize != config.TextSizeLarge {
		t.Fatalf("text size round trip: %q", merged.SettingsTextSize)
	}
	if !merged.SettingsKeyboard {
		t.Fatal("keyboard navigation was dropped")
	}
	if merged.SettingsTooltips {
		t.Fatal("tooltips were not turned off")
	}
	if !merged.HiddenAtLaunch {
		t.Fatal("hidden-at-launch was dropped")
	}

	// And through the real file format, because a name config.json cannot read
	// back is a preference that resets on every launch.
	path := t.TempDir() + "/config.json"
	if err := config.Save(path, merged); err != nil {
		t.Fatalf("save: %v", err)
	}
	reloaded := config.Load(path)
	if reloaded.SettingsTextSize != config.TextSizeLarge || !reloaded.SettingsKeyboard || reloaded.SettingsTooltips {
		t.Fatalf("reloaded %+v", reloaded)
	}
}
