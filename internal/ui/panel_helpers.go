package ui

import (
	"math"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

const (
	windowClassName = "AITerminalPetWindow"
	windowTitle     = "AI Terminal Pet"

	pulseInterval = 160
	slideInterval = 33
)

func classifyHit(x, y, width, height int, hit []byte) bool {
	if width <= 0 || height <= 0 || len(hit) < width*height {
		return false
	}
	if x < 0 || y < 0 || x >= width || y >= height {
		return false
	}
	return hit[y*width+x] >= 20
}

type releaseAction int

const (
	releaseNothing releaseAction = iota
	releaseToggle
)

func decideRelease(active, moved bool) releaseAction {
	if !active || moved {
		return releaseNothing
	}
	return releaseToggle
}

func doneBalloonBody(value session.Session, lang i18n.Lang) string {
	body := i18n.T(lang, i18n.KeyBalloonDoneBody)
	if label := projectLabel(value, lang); label != "" {
		body = label + " · " + body
	}
	return body
}

func presenceLabel(sample attendanceState, known bool) string {
	if !known {
		return "unknown"
	}
	if sample.Present {
		return "present"
	}
	return "away"
}

func scaledDimension(base int, dpi uint32, zoom float64) int {
	value := int(math.Round(float64(base) * float64(dpi) / 96 * zoom))
	return maxInt(value, 32)
}

func abs32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
