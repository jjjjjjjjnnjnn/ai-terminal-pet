package ui

// The settings *shape*: one plain struct and the defaults it starts from.
//
// These lived in app_windows.go, which made the whole settings vocabulary look
// Windows-only even though it is a value type the app layer maps config.json onto.
// internal/app's lifecycle could not be compiled for linux until this moved, and the
// reason it could not is the same reason handover §55.6 step 1 moved the alert policy:
// a declaration is not a platform dependency.
//
// Nothing here reads a window, a registry or a clock. The values are the persisted
// names config.json stores, and TestDefaultSettingsMirrorConfig pins them against
// config.Default() so the two cannot drift.

// Settings is the subset of user preferences the tray menu can flip.
type Settings struct {
	Notifications bool
	Sound         bool
	// SoundVolume is the tone amplitude, 0..100. It is mirrored into the
	// audio layer on startup, hot reload and every tray change.
	SoundVolume int
	// SoundScheme picks one of the built-in timbres; see ui.SoundSchemes.
	SoundScheme string
	// SoundOnWaiting / SoundOnError gate the tone per event, so someone who
	// only wants to hear genuine blockers can mute the rest.
	SoundOnWaiting bool
	SoundOnError   bool
	// NotifyOnFinish is kept as the legacy two-way view of DonePolicy: it is
	// true whenever a completion can be announced at all. New code should read
	// DonePolicy; this stays because the tray check mark and older config
	// files still speak it, and Normalize keeps the two consistent.
	NotifyOnFinish bool
	// DonePolicy is when a finished turn is announced: background / off /
	// always. It is the live preference; NotifyOnFinish is derived.
	DonePolicy string
	// ForegroundAware makes an alert depend on whether the user is already
	// looking at that session's terminal.
	ForegroundAware bool
	// IdleSeconds is how long without input counts as being away even while a
	// window is in front. Zero means window focus only. It is stored in
	// config.json as idleMinutes; seconds here because that is what the
	// comparison against GetLastInputInfo's result wants.
	IdleSeconds int
	// QuietCurrent suppresses the balloon for the session being looked at,
	// leaving its unread mark.
	QuietCurrent bool
	// MutedSessions is the persisted per-session silence list, as
	// "<source>\x00<sessionId>" keys.
	MutedSessions []string
	// QuietSubagents silences sessions that declare a parent session.
	QuietSubagents bool
	AlwaysOnTop    bool
	// EdgeRetract slides the panel into the screen edge it is hugging when
	// nothing needs attention.
	EdgeRetract bool
	// Language is the persisted language preference: "auto" follows the
	// host OS locale, "en" and "zh-CN" pin one. See internal/i18n.
	Language string
	// AutoApprove lets a permission prompt auto-approve after the timeout
	// when the user says nothing. Off by default.
	AutoApprove bool
	// AutoApproveTimeoutSec is the countdown length; the settings window can
	// change it at runtime (it used to be read only at startup).
	AutoApproveTimeoutSec int
	// Opacity is the layered window's constant alpha, 1..255.
	Opacity int
	// Theme picks one of the named palettes (see ThemeNames). It is a name,
	// not a colour value: the ui package owns the hues.
	Theme string
	// PerSourceColor colours each expanded row's source name by which terminal
	// it came from. Off by default, because the state colour is the one signal
	// the strip exists for and provenance is secondary.
	PerSourceColor bool
	// HiddenAtLaunch starts with the panel in the tray: the reporters start the
	// overlay, and someone who wants it out of the way until a session needs
	// attention should not have to hide it every logon.
	HiddenAtLaunch bool
	// SettingsTextSize scales the settings window: "small", "standard" or
	// "large". A name, like Theme, because the factor is a rendering decision
	// (settingsTextScales) and config.json should record the choice.
	SettingsTextSize string
	// SettingsKeyboard enables Tab focus and Space/Enter activation in the
	// settings window. Off by default; see config.Config.
	SettingsKeyboard bool
	// SettingsTooltips shows the untruncated text when a settings label had to
	// be cut off.
	SettingsTooltips bool
}

// DefaultSettings is what "restore defaults" puts back, and what the overlay
// starts with when no config file exists. It mirrors config.Default(): the ui
// package does not import internal/config (the app layer owns file paths), so
// a test in the app package pins the two together.
func DefaultSettings() Settings {
	return Settings{
		Notifications:         true,
		Sound:                 true,
		SoundVolume:           85,
		SoundScheme:           soundSchemeChime,
		SoundOnWaiting:        true,
		SoundOnError:          true,
		NotifyOnFinish:        true,
		AlwaysOnTop:           true,
		EdgeRetract:           true,
		Language:              "auto",
		AutoApprove:           false,
		AutoApproveTimeoutSec: 5,
		Opacity:               235,
		Theme:                 themeSlate,
		PerSourceColor:        false,
		HiddenAtLaunch:        false,
		ForegroundAware:       true,
		IdleSeconds:           defaultIdleSeconds,
		DonePolicy:            string(DonePolicyBackground),
		QuietCurrent:          true,
		QuietSubagents:        true,
		SettingsTextSize:      textSizeStandard,
		SettingsKeyboard:      false,
		SettingsTooltips:      true,
	}
}
