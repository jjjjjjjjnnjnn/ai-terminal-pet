package app

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/ui"
)

// The alert policy lives in internal/ui (it draws nothing, but it speaks the
// panel's vocabulary) and the persisted names live in internal/config. That is
// one deliberate duplication, exactly like the theme names: ui does not import
// config because the app layer owns the file paths. These tests hold the two
// copies together, so a rename on one side cannot silently reset a user's
// setting on the other.

func TestDonePolicyListsAgree(t *testing.T) {
	if len(config.DonePolicies) != len(ui.DonePolicyNames()) {
		t.Fatalf("config offers %d policies, ui offers %d", len(config.DonePolicies), len(ui.DonePolicyNames()))
	}
	for index, name := range config.DonePolicies {
		if string(ui.DonePolicyNames()[index]) != name {
			t.Fatalf("policy %d: config says %q, ui says %q", index, name, ui.DonePolicyNames()[index])
		}
	}
}

func TestIdleDefaultsAgree(t *testing.T) {
	if ui.DefaultIdleSeconds() != config.DefaultIdleMinutes*60 {
		t.Fatalf("config default is %d minutes, ui default is %d seconds",
			config.DefaultIdleMinutes, ui.DefaultIdleSeconds())
	}
}

// The mute key is a cross-package contract: config stores it, ui looks sessions
// up by it. If the two encodings drifted apart, every persisted mute would stop
// matching the session it was meant for - and the symptom would be "muting does
// nothing", with no error anywhere.
func TestMuteKeyEncodingAgrees(t *testing.T) {
	if got := config.SessionMuteKey("claude-hook", "ses_1"); got != ui.SessionIdentityFor("claude-hook", "ses_1") {
		t.Fatalf("config key %q, ui key %q", got, ui.SessionIdentityFor("claude-hook", "ses_1"))
	}
}

// Every new preference has to survive applyUISettings, which is the only bridge
// from the settings window to config.json: a field it forgets is a field that
// silently resets on the next launch.
func TestUISettingsRoundTripKeepsAttention(t *testing.T) {
	base := config.Default()
	base.MutedSessions = []string{config.SessionMuteKey("claude-hook", "ses_keep")}

	settings := ui.Settings{
		Notifications:         base.Notifications,
		Sound:                 base.Sound,
		SoundVolume:           base.SoundVolume,
		SoundScheme:           base.SoundScheme,
		SoundOnWaiting:        base.SoundOnWaiting,
		SoundOnError:          base.SoundOnError,
		NotifyOnFinish:        base.NotifyOnFinish,
		DonePolicy:            config.DonePolicyOff,
		ForegroundAware:       false,
		IdleSeconds:           10 * 60,
		QuietCurrent:          false,
		MutedSessions:         []string{config.SessionMuteKey("claude-hook", "ses_keep")},
		QuietSubagents:        false,
		AlwaysOnTop:           base.AlwaysOnTop,
		EdgeRetract:           base.EdgeRetract,
		Language:              base.Language,
		AutoApprove:           base.AutoApprove,
		AutoApproveTimeoutSec: base.AutoApproveTimeoutSec,
		Opacity:               base.Opacity,
		Theme:                 base.Theme,
		PerSourceColor:        base.PerSourceColor,
		HiddenAtLaunch:        base.HiddenAtLaunch,
	}

	merged := base
	applyUISettings(&merged, settings)
	if merged.DonePolicy != config.DonePolicyOff {
		t.Fatalf("done policy did not survive: %q", merged.DonePolicy)
	}
	if merged.ForegroundAware {
		t.Fatal("foreground awareness did not survive")
	}
	if merged.IdleMinutes != 10 {
		t.Fatalf("idle minutes did not survive: %d", merged.IdleMinutes)
	}
	if merged.QuietCurrent {
		t.Fatal("quiet-current did not survive")
	}
	if merged.QuietSubagents {
		t.Fatal("quiet-subagents did not survive")
	}
	if len(merged.MutedSessions) != 1 || merged.MutedSessions[0] != config.SessionMuteKey("claude-hook", "ses_keep") {
		t.Fatalf("mute list did not survive: %v", merged.MutedSessions)
	}

	// An unknown policy coming back from the window is normalised rather than
	// written to disk.
	settings.DonePolicy = "nonsense"
	merged = base
	applyUISettings(&merged, settings)
	if merged.DonePolicy != config.DefaultDonePolicy {
		t.Fatalf("an unknown policy was persisted as %q", merged.DonePolicy)
	}
}

// IdleSeconds maps onto whole minutes, and the "leave it alone" case has to be
// distinguishable from "zero minutes".
func TestIdleSecondsMapping(t *testing.T) {
	if got := ui.SecondsToIdleMinutes(90); got != 1 {
		t.Fatalf("90 seconds is 1 whole minute, got %d", got)
	}
	if got := ui.SecondsToIdleMinutes(0); got != 0 {
		t.Fatalf("zero is the window-focus-only sentinel, got %d", got)
	}
	if ui.DefaultIdleSeconds() != int((3 * time.Minute).Seconds()) {
		t.Fatalf("the ui default changed: %d", ui.DefaultIdleSeconds())
	}
}
