package ui

import (
	"encoding/binary"
	"hash/fnv"
	"path/filepath"

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
var (
	colorPanel      = rgba{19, 22, 28, 238}
	colorPanelHover = rgba{28, 33, 41, 246}
	colorRaised     = rgba{33, 38, 47, 225}
	colorRaisedRow  = rgba{33, 38, 47, 200}
	colorBorder     = rgba{98, 108, 124, 132}
	colorDivider    = rgba{98, 108, 124, 76}
	colorOrbInner   = rgba{16, 19, 24, 235}

	colorPrimary   = rgba{240, 244, 250, 255}
	colorSecondary = rgba{170, 180, 196, 255}
	colorMuted     = rgba{130, 140, 156, 255}
	// colorHint is the row caption colour in the settings window. colorMuted is
	// tuned for the overlay's small print, where it sits on a darker fill; on
	// the settings panel it was too dim to read at 10 px.
	colorHint = rgba{150, 160, 176, 255}

	// Settings-window chrome. Accent marks "on" and "selected" - a control
	// state, not a session state - so it deliberately stays out of the
	// state palette above and is never used on the overlay itself.
	colorAccent    = rgba{132, 174, 222, 255}
	colorControlBg = rgba{13, 16, 21, 200}
	colorTrackOff  = rgba{96, 106, 122, 255}

	// Muted, near-equal-luminance state hues: distinct at a glance without
	// any one of them shouting over the text or clashing with the others.
	colorWaiting  = rgba{231, 184, 112, 255}
	colorError    = rgba{226, 124, 130, 255}
	colorWorking  = rgba{126, 201, 167, 255}
	colorTool     = rgba{132, 174, 222, 255}
	colorThinking = rgba{174, 155, 214, 255}
	colorIdle     = rgba{146, 157, 174, 255}
)

func stateColor(value session.State) rgba {
	switch value {
	case session.StateWaitingPermission:
		return colorWaiting
	case session.StateError:
		return colorError
	case session.StateWorking:
		return colorWorking
	case session.StateToolUse:
		return colorTool
	case session.StateThinking:
		return colorThinking
	default:
		return colorIdle
	}
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
