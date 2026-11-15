package ui

import (
	"encoding/binary"
	"hash/fnv"
	"path/filepath"
	"strings"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

const (
	dtVCenter     = 0x0004
	dtWordBreak   = 0x0010
	dtSingleLine  = 0x0020
	dtNoPrefix    = 0x0800
	dtEndEllipsis = 0x8000
)

// Slate palette.
//
// The old look was a blue-black slab with six fully saturated state colours.
// In an alert state the outline, orb, label and jump bar were all the same
// loud amber, so the whole strip read as one smear; in the expanded list four
// neon hues competed with the text. Everything structural is now neutral
// graphite, and state colour is reserved for a thin ring, a state word and an
// arrow. Keep every structural colour here: scattering rgba literals through
// the draw code is how the palette drifted apart in the first place.
// The palette: every structural colour in one place.
//
// The original comment here said "keep every structural colour in this file:
// scattering rgba literals through the draw code is how the palette drifted
// apart in the first place". That is still the rule - what changed is that there
// are now three of these blocks instead of one, and the draw code reads
// `colors.<Field>` so a theme switch is a single assignment rather than a
// hundred literals to hunt down.
//
// `colors` is process-wide and written only on the UI thread (see the
// single-threaded note above the font cache in canvas_windows.go), which is the
// same assumption every constant here already relied on.
type palette struct {
	Panel      rgba
	PanelHover rgba
	Raised     rgba
	RaisedRow  rgba
	Border     rgba
	Divider    rgba
	OrbInner   rgba

	Primary   rgba
	Secondary rgba
	Muted     rgba
	Hint      rgba

	Accent    rgba
	ControlBg rgba
	TrackOff  rgba

	Waiting  rgba
	Error    rgba
	Working  rgba
	Tool     rgba
	Thinking rgba
	Idle     rgba

	// Provenance hues, used only for a row's source name when perSourceColor is
	// on. Kept separate from the state family on purpose.
	SourceClaude      rgba
	SourceOpenCode    rgba
	SourceAntigravity rgba
	SourceHarness     rgba
}

// slate is the palette shipped since 0.9.0 and used by every screenshot in the
// repository. It must stay byte-identical, because a reader comparing their
// panel to the README has to see the same thing.
var slatePalette = palette{
	Panel:      rgba{19, 22, 28, 238},
	PanelHover: rgba{28, 33, 41, 246},
	Raised:     rgba{33, 38, 47, 225},
	RaisedRow:  rgba{33, 38, 47, 200},
	Border:     rgba{98, 108, 124, 132},
	Divider:    rgba{98, 108, 124, 76},
	OrbInner:   rgba{16, 19, 24, 235},

	Primary:   rgba{240, 244, 250, 255},
	Secondary: rgba{170, 180, 196, 255},
	Muted:     rgba{130, 140, 156, 255},
	// Hint is the settings-window caption colour. Muted is tuned for the
	// overlay's small print on a darker fill; on the settings panel it was too
	// dim to read at 10 px.
	Hint: rgba{150, 160, 176, 255},

	// Settings chrome. Accent marks "on" and "selected" - a control state, not a
	// session state - so it stays out of the state palette and is never used on
	// the overlay itself.
	Accent:    rgba{132, 174, 222, 255},
	ControlBg: rgba{13, 16, 21, 200},
	TrackOff:  rgba{96, 106, 122, 255},

	// Muted, near-equal-luminance state hues: distinct at a glance without any
	// one of them shouting over the text.
	Waiting:  rgba{231, 184, 112, 255},
	Error:    rgba{226, 124, 130, 255},
	Working:  rgba{126, 201, 167, 255},
	Tool:     rgba{132, 174, 222, 255},
	Thinking: rgba{174, 155, 214, 255},
	Idle:     rgba{146, 157, 174, 255},

	SourceClaude:      rgba{198, 148, 108, 255},
	SourceOpenCode:    rgba{110, 184, 196, 255},
	SourceAntigravity: rgba{150, 150, 206, 255},
	SourceHarness:     rgba{186, 168, 122, 255},
}

// paper is the light theme. The one rule that is easy to get wrong: "muted" has
// to move *towards black* on a light fill, or the small print disappears
// exactly as it did when the overlay was a single dark palette. The state hues
// are the same families darkened until they clear ~4.5:1 on the panel.
var paperPalette = palette{
	Panel:      rgba{246, 247, 250, 238},
	PanelHover: rgba{255, 255, 255, 248},
	Raised:     rgba{232, 235, 241, 235},
	RaisedRow:  rgba{225, 229, 236, 210},
	Border:     rgba{120, 128, 144, 150},
	Divider:    rgba{120, 128, 144, 90},
	OrbInner:   rgba{252, 252, 254, 240},

	Primary:   rgba{26, 30, 38, 255},
	Secondary: rgba{64, 72, 86, 255},
	Muted:     rgba{104, 112, 126, 255},
	Hint:      rgba{92, 100, 114, 255},

	Accent:    rgba{38, 100, 176, 255},
	ControlBg: rgba{214, 219, 227, 220},
	TrackOff:  rgba{168, 174, 186, 255},

	Waiting:  rgba{150, 100, 16, 255},
	Error:    rgba{180, 44, 52, 255},
	Working:  rgba{18, 114, 84, 255},
	Tool:     rgba{36, 88, 148, 255},
	Thinking: rgba{104, 78, 160, 255},
	Idle:     rgba{96, 104, 118, 255},

	SourceClaude:      rgba{132, 82, 40, 255},
	SourceOpenCode:    rgba{16, 96, 108, 255},
	SourceAntigravity: rgba{74, 74, 150, 255},
	SourceHarness:     rgba{122, 100, 44, 255},
}

// contrast is for a room where the strip is read from a distance: pure black
// fill, near-white text, and state colours pushed to full saturation. The
// state-to-state separation matters more here than the subdued look.
var contrastPalette = palette{
	Panel:      rgba{0, 0, 0, 246},
	PanelHover: rgba{16, 16, 16, 250},
	Raised:     rgba{28, 28, 28, 235},
	RaisedRow:  rgba{20, 20, 20, 225},
	Border:     rgba{210, 210, 210, 160},
	Divider:    rgba{210, 210, 210, 90},
	OrbInner:   rgba{0, 0, 0, 255},

	Primary:   rgba{255, 255, 255, 255},
	Secondary: rgba{214, 214, 214, 255},
	Muted:     rgba{178, 178, 178, 255},
	Hint:      rgba{190, 190, 190, 255},

	Accent:    rgba{120, 180, 255, 255},
	ControlBg: rgba{8, 8, 8, 230},
	TrackOff:  rgba{140, 140, 140, 255},

	Waiting:  rgba{255, 196, 64, 255},
	Error:    rgba{255, 96, 96, 255},
	Working:  rgba{64, 226, 154, 255},
	Tool:     rgba{120, 180, 255, 255},
	Thinking: rgba{196, 150, 255, 255},
	Idle:     rgba{196, 196, 196, 255},

	SourceClaude:      rgba{255, 196, 140, 255},
	SourceOpenCode:    rgba{120, 226, 236, 255},
	SourceAntigravity: rgba{186, 178, 255, 255},
	SourceHarness:     rgba{226, 206, 150, 255},
}

// perSourceColors is the live switch for the expanded rows' provenance colour.
// It is a package-level flag for the same reason the palette is: everything here
// rasterises on the UI thread, and threading another bool through six draw
// signatures for a cosmetic toggle is noise.
var perSourceColors bool

// colors is the palette actually in use. Slate until someone chooses otherwise.
var colors = slatePalette

// Theme names, mirrored from internal/config so the ui package keeps its rule
// of not importing the config package (the app layer owns file paths and file
// formats; see DefaultSettings).
const (
	themeSlate    = "slate"
	themePaper    = "paper"
	themeContrast = "contrast"
)

// ThemeNames is the complete set, in the order the settings window offers them.
var ThemeNames = []string{themeSlate, themePaper, themeContrast}

// normalizeTheme is deliberately duplicated from config.ValidTheme instead of
// imported: the ui package treating an unknown name as slate is what keeps a
// hand-edited config.json from rendering transparent black, and a test in the
// app package pins that the two agree.
func normalizeTheme(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case themePaper:
		return themePaper
	case themeContrast:
		return themeContrast
	default:
		return themeSlate
	}
}

// applyTheme installs a named palette. Called on the UI thread only, before any
// raster and again on every settings reload; assigning the same palette again
// costs one struct copy and changes nothing.
//
// The frame key carries the theme name (see renderedKey), so the caller's refresh
// repaints exactly once - the palette alone would be invisible to the
// "same key, same pixels" guard.
func applyTheme(name string) {
	colors = themePalette(name)
}

// applyAppearance applies the whole visual preference group. Both fields feed
// the frame key, so a change repaints exactly once.
func applyAppearance(theme string, perSource bool) {
	applyTheme(theme)
	perSourceColors = perSource
}

func themePalette(name string) palette {
	switch normalizeTheme(name) {
	case themePaper:
		return paperPalette
	case themeContrast:
		return contrastPalette
	default:
		return slatePalette
	}
}

func stateColor(value session.State) rgba {
	switch value {
	case session.StateWaitingPermission:
		return colors.Waiting
	case session.StateError:
		return colors.Error
	case session.StateWorking:
		return colors.Working
	case session.StateToolUse:
		return colors.Tool
	case session.StateThinking:
		return colors.Thinking
	default:
		return colors.Idle
	}
}

// sourceColor is the provenance palette: one hue per reporter family, chosen to
// stay distinguishable from each other and from the state palette above them in
// the row (the state word keeps its own colour, so these must not collide with
// the alert hues - hence the teal/indigo/plum/bronze spread rather than greens
// and ambers).
//
// Unknown sources fall back to Secondary rather than guessing: a reporter this
// build has never heard of should look neutral, not pretend to be a terminal it
// is not.
func sourceColor(source string) rgba {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "claude-hook", "claude-code", "claude":
		return colors.SourceClaude
	case "opencode":
		return colors.SourceOpenCode
	case "antigravity", "antigravity-statusline", "antigravity-hook":
		return colors.SourceAntigravity
	case "deepseek-harness", "deepseek":
		return colors.SourceHarness
	}
	return colors.Secondary
}

func stateLabel(value session.State, lang i18n.Lang) string {
	switch value {
	case session.StateWaitingPermission:
		return i18n.T(lang, i18n.KeyStateWaiting)
	case session.StateError:
		return i18n.T(lang, i18n.KeyStateError)
	case session.StateWorking:
		return i18n.T(lang, i18n.KeyStateWorking)
	case session.StateToolUse:
		return i18n.T(lang, i18n.KeyStateTool)
	case session.StateThinking:
		return i18n.T(lang, i18n.KeyStateThinking)
	default:
		return i18n.T(lang, i18n.KeyStateIdle)
	}
}

func projectLabel(current session.Session, lang i18n.Lang) string {
	value := current.Project
	if value == "" && current.Title != "" {
		value = current.Title
	}
	if value == "" {
		return i18n.T(lang, i18n.KeyProjectUnknown)
	}
	base := filepath.Base(value)
	if base == "." || base == string(filepath.Separator) {
		return value
	}
	return base
}

// reasonLabel is the "what is it doing" half of a session row. The fallback
// used to be a flat "等待用户响应", which was simply wrong for a session that
// was busy or idle; naming the state is always accurate.
func reasonLabel(current session.Session, lang i18n.Lang) string {
	if current.Reason != "" {
		return current.Reason
	}
	if current.Tool != "" {
		return current.Tool
	}
	return stateLabel(current.State, lang)
}

// sessionDigest fingerprints the parts of the session list the panel actually
// draws.
//
// The store bumps its revision on every push because updatedAt moves, so the
// revision alone would make every refresh look like a change and defeat the
// redraw skip entirely. This is a cheap O(n) walk over the fields that reach
// the screen, with updatedAt bucketed to 15 seconds so a ticking clock does
// not keep the panel repainting either.
func sessionDigest(items []session.Session) uint64 {
	hash := fnv.New64a()
	var stamp [8]byte
	for _, current := range items {
		hash.Write([]byte(current.State))
		hash.Write([]byte{0})
		hash.Write([]byte(current.Source))
		hash.Write([]byte{0})
		hash.Write([]byte(current.Project))
		hash.Write([]byte{0})
		hash.Write([]byte(current.Reason))
		hash.Write([]byte{0})
		binary.LittleEndian.PutUint64(stamp[:], uint64(current.UpdatedAt.Unix()/15))
		hash.Write(stamp[:])
	}
	return hash.Sum64()
}
