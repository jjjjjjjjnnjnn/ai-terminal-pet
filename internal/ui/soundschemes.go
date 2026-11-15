package ui

// The four sound scheme names, and the order the settings window offers them in.
//
// They are the values config.json stores ("chime" / "bell" / "soft" / "pulse"), so
// they belong to the settings vocabulary rather than to the Win32 synthesiser that
// happens to play them. Moving them here is what lets ui.Settings compile without a
// window layer: DefaultSettings names one of them, and until this moved the whole
// settings shape was stuck behind a Win32 build tag (handover §55.6 step 2).
//
// internal/config keeps its own mirror (config.SoundSchemes) because the app layer
// owns file paths and the ui package deliberately does not import it; a test pins the
// two lists together.

const (
	soundSchemeChime = "chime"
	soundSchemeBell  = "bell"
	soundSchemeSoft  = "soft"
	soundSchemePulse = "pulse"

	defaultScheme = soundSchemeChime
)

// SoundSchemes is the user-visible order shown by the settings window.
var SoundSchemes = []string{soundSchemeChime, soundSchemeBell, soundSchemeSoft, soundSchemePulse}

// validSoundScheme maps any stored or hand-edited value onto a scheme this build can
// play. An unknown name falls back to the default instead of failing: a config typo
// must not silence the overlay, and it must not make Settings invalid either.
func validSoundScheme(value string) string {
	for _, name := range SoundSchemes {
		if name == value {
			return value
		}
	}
	return defaultScheme
}
