package ui

import (
	"encoding/binary"
	"hash/fnv"
	"path/filepath"

	"ai-terminal-pet/internal/session"
)

const (
	dtVCenter     = 0x0004
	dtWordBreak   = 0x0010
	dtSingleLine  = 0x0020
	dtNoPrefix    = 0x0800
	dtEndEllipsis = 0x8000
)

var (
	colorPrimary   = rgba{237, 242, 255, 255}
	colorSecondary = rgba{169, 181, 207, 255}
	colorMuted     = rgba{116, 129, 157, 255}
	colorWaiting   = rgba{255, 184, 77, 255}
	colorError     = rgba{255, 107, 107, 255}
	colorWorking   = rgba{88, 214, 141, 255}
	colorTool      = rgba{93, 166, 255, 255}
	colorThinking  = rgba{181, 136, 255, 255}
	colorIdle      = rgba{151, 163, 190, 255}
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

func stateLabel(value session.State) string {
	switch value {
	case session.StateWaitingPermission:
		return "等待确认"
	case session.StateError:
		return "错误"
	case session.StateWorking:
		return "执行中"
	case session.StateToolUse:
		return "工具调用"
	case session.StateThinking:
		return "思考中"
	default:
		return "空闲"
	}
}

func projectLabel(current session.Session) string {
	value := current.Project
	if value == "" && current.Title != "" {
		value = current.Title
	}
	if value == "" {
		return "未标注项目"
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
func reasonLabel(current session.Session) string {
	if current.Reason != "" {
		return current.Reason
	}
	if current.Tool != "" {
		return current.Tool
	}
	return stateLabel(current.State)
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
