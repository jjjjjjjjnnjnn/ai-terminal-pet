package ui

import (
	"math"
	"strings"
	"time"

	"ai-terminal-pet/internal/session"
)

type dragState struct {
	active  bool
	moved   bool
	startX  int32
	startY  int32
	windowX int32
	windowY int32
}

func attentionSettingsFrom(settings Settings) AttentionSettings {
	return AttentionSettings{
		Notifications:   settings.Notifications,
		Sound:           settings.Sound,
		SoundOnWaiting:  settings.SoundOnWaiting,
		SoundOnError:    settings.SoundOnError,
		DonePolicy:      ValidDonePolicy(settings.DonePolicy),
		QuietCurrent:    settings.QuietCurrent,
		ForegroundAware: settings.ForegroundAware,
		IdleAfter:       time.Duration(settings.IdleSeconds) * time.Second,
		QuietSubagents:  settings.QuietSubagents,
	}
}

func buildAlertInputs(sessions []session.Session, mutes *muteSet, previous, current map[string]alertFacts, sample attendanceState, sampleKnown bool) []AlertInput {
	inputs := make([]AlertInput, 0, len(sessions))
	for _, value := range sessions {
		key := sessionAlertKey(value)
		kind := session.KindOfSession(value)
		old, existed := previous[key]
		if existed && old.State == value.State && old.Kind == kind {
			continue
		}
		in := AlertInput{
			Session:       value,
			Kind:          kind,
			Previous:      old.State,
			Existed:       existed,
			ChangedKind:   existed && old.Kind != kind,
			Muted:         mutes.has(value),
			Subagent:      strings.TrimSpace(value.ParentSessionID) != "",
			PresenceKnown: sampleKnown,
		}
		if sampleKnown {
			in.Present = sample.Present
			in.Attending = sample.Attending[session.Identity(value)]
		}
		inputs = append(inputs, in)
	}
	return inputs
}

func buildVanishedInputs(previous, current map[string]alertFacts, sample attendanceState) []AlertInput {
	var inputs []AlertInput
	for key, old := range previous {
		if _, stillHere := current[key]; stillHere {
			continue
		}
		if !busyState(old.State) {
			continue
		}
		source, sessionID := splitAlertKey(key)
		inputs = append(inputs, AlertInput{
			Session:  session.Session{Source: source, SessionID: sessionID, State: session.StateClosed},
			Kind:     session.KindNone,
			Previous: old.State,
			Existed:  true,
			Present:  sample.Present,
			Vanished: true,
		})
	}
	return inputs
}

func recordRememberedAlerts(lastAlertAt map[string]time.Time, inputs []AlertInput, plan AlertPlan, now time.Time) map[string]time.Time {
	if lastAlertAt == nil {
		lastAlertAt = make(map[string]time.Time, len(inputs))
	}
	for index := range inputs {
		key := sessionAlertKey(inputs[index].Session)
		if plan.Reasons[key] == decisionDoneAway || plan.Reasons[key] == decisionDoneSeen {
			lastAlertAt[key] = now
		}
	}
	return lastAlertAt
}

func alertRecordsEqual(a, b alertRecord) bool {
	if len(a.Sessions) != len(b.Sessions) {
		return false
	}
	if a.Balloon != b.Balloon ||
		a.Tone != b.Tone ||
		a.Presence != b.Presence ||
		a.Present != b.Present ||
		a.Attending != b.Attending ||
		a.Skipped != b.Skipped {
		return false
	}
	for index := range b.Sessions {
		if a.Sessions[index] != b.Sessions[index] {
			return false
		}
	}
	return true
}

func computeDock(x int32, width int, dpi uint32, work displayRect, hasWork bool) (dockSide, int32, int32) {
	if !hasWork {
		return dockNone, 0, 0
	}
	tolerance := int32(dockTolerance * dpi / 96)
	w := int32(width)
	switch {
	case x <= work.Left+tolerance:
		return dockLeft, work.Left, x - work.Left
	case x+w >= work.Right-tolerance:
		return dockRight, work.Right, work.Right - (x + w)
	}
	return dockNone, 0, 0
}

func panelSliverWidth(dpi uint32, zoom float64) int {
	value := int(math.Round(float64(dockedSliver) * float64(dpi) / 96 * zoom))
	return maxInt(value, 6)
}

func clampFrameLeftToWork(x int32, dpi uint32, work displayRect, hasWork bool) int32 {
	if !hasWork {
		return x
	}
	if minimum := work.Left + int32(8*dpi/96); x < minimum {
		return minimum
	}
	return x
}

func panelFrameX(dragActive bool, x int32, width int, dock dockSide, dockEdge, dockGap int32, tuck float64, sliver int, dpi uint32, work displayRect, hasWork bool) int32 {
	if dragActive {
		return x
	}
	untucked := x
	if dock != dockNone {
		untucked = dockedOrigin(dock, dockEdge, dockGap, width)
	}
	if dock == dockNone || tuck <= 0 {
		return clampFrameLeftToWork(untucked, dpi, work, hasWork)
	}
	tucked := tuckedOrigin(dock, dockEdge, width, sliver)
	return slideBetween(untucked, tucked, tuck)
}

func panelFrameY(dragActive bool, y int32, height int, dpi uint32, work displayRect, hasWork bool) int32 {
	if dragActive || !hasWork {
		return y
	}
	margin := int32(8 * dpi / 96)
	if maximum := work.Bottom - margin - int32(height); y > maximum {
		y = maximum
	}
	if minimum := work.Top + margin; y < minimum {
		y = minimum
	}
	return y
}

func clampPositionToWorkArea(x, y, width, height int32, dpi uint32, work displayRect) (int32, int32) {
	minimumX := work.Left + int32(8*dpi/96)
	minimumY := work.Top + int32(8*dpi/96)
	maximumX := work.Right - width - int32(8*dpi/96)
	maximumY := work.Bottom - height - int32(8*dpi/96)
	if maximumX < minimumX {
		x = work.Left
	} else {
		x = max32(minimumX, min32(x, maximumX))
	}
	if maximumY < minimumY {
		y = work.Top
	} else {
		y = max32(minimumY, min32(y, maximumY))
	}
	return x, y
}

func defaultWorkAreaOrigin(width, height int32, dpi uint32, work displayRect) (int32, int32) {
	x := work.Right - width - int32(24*dpi/96)
	y := work.Top + (work.Bottom-work.Top-height)/2
	return clampPositionToWorkArea(x, y, width, height, dpi, work)
}

func stepZoomValue(zoom, delta float64) float64 {
	if delta == 0 {
		return zoom
	}
	step := 0.1
	if delta < 0 {
		step = -step
	}
	return math.Max(0.65, math.Min(1.6, zoom+step))
}

func buildTapInputs(expanded, autoApprove bool, width, height, x, y int, sessions []session.Session, approvals *approvalTracker) tapInputs {
	if expanded {
		_, hasJump := mostUrgentJump(sessions)
		if row, hit := focusHit(x, y, width, height, true, hasJump, false, len(sessions)); hit {
			return tapInputs{expanded: true, jumpHit: true, jumpRow: row, jumpValid: true}
		}
		inputs := tapInputs{expanded: true}
		if key := approvals.urgentKey(sessions); autoApprove && key != "" {
			inputs.vetoArmed = true
		}
		if index, hit := expandedRowHit(x, y, width, height, len(sessions)); hit {
			if _, copyable := rowCopyText(sessions, index); copyable {
				inputs.copyRowHit = true
				inputs.copyRow = index
			}
		}
		return inputs
	}
	inputs := tapInputs{}
	_, hasJump := mostUrgentJump(sessions)
	_, stripCopyable := compactCopyText(sessions)
	inputs.copyStripValid = stripCopyable
	inputs.copyStripHit = compactCopyHit(x, y, width, height, stripCopyable)
	if row, hit := focusHit(x, y, width, height, false, hasJump, stripCopyable, len(sessions)); hit {
		return tapInputs{jumpHit: true, jumpRow: row, jumpValid: true}
	}
	return inputs
}
