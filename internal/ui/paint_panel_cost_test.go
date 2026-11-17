package ui

// Measuring the overlay panel's full-frame paint, the way paint_cost_test.go
// measures the settings window: a report, not a gate.
//
// docs/ROADMAP.md §3 L4 has said "dirty-region redraw, or vectorise the inner
// loops of fillRounded/fillEllipse - but the panel is only 56x156 and 200x<rows>,
// so measure first and do not add complexity for 0.3 ms" since the project was
// a few rounds old, and never with a number behind it. This file is that number.
//
// It is untagged on purpose: the Linux and macOS panels rasterise text with the
// pure-Go OpenType path, which is the one that could plausibly be slow, and the
// point of a measurement is to be wrong out loud on the platform that is.
//
// It never fails on a slow machine. A threshold here would be a flaky test on a
// busy CI runner, which is worse than no test at all; the decision belongs in
// docs/ROADMAP.md, made from these numbers, not in an assertion.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestMeasurePanelPaint(t *testing.T) {
	if testing.Short() {
		t.Skip("measurement runs only when asked")
	}
	now := previewNow
	waiting := []session.Session{
		{Source: "claude-hook", SessionID: "a", Project: `C:\work\ai-terminal-pet`, State: session.StateWaitingPermission, Reason: "Bash(go test)", UpdatedAt: now.Add(-2 * time.Minute)},
		{Source: "opencode", SessionID: "b", Project: `C:\work\webapp`, State: session.StateWorking, Reason: "editing src/app.tsx", UpdatedAt: now.Add(-6 * time.Second)},
		{Source: "antigravity-statusline", SessionID: "c", Project: `C:\work\docs`, State: session.StateIdle, UpdatedAt: now.Add(-40 * time.Second)},
	}
	cases := []struct {
		name  string
		state panelState
	}{
		{"compact waiting", panelState{Items: waiting}},
		{"compact empty", panelState{}},
		{"expanded 3 rows", panelState{Items: waiting, Expanded: true}},
		{"expanded 6 rows", panelState{Items: repeatSessions(waiting, 2), Expanded: true}},
		// The alert layer animates at 20 FPS, so this is the frame that has to
		// fit in a 50 ms budget rather than the idle one.
		{"expanded 3 rows hovered", panelState{Items: waiting, Expanded: true, Hovered: true, HoverRow: 0}},
	}
	for _, tc := range cases {
		state := tc.state
		state.Now = now
		state.Lang = "zh-CN"
		state.Sliver = dockedSliver
		state.Dock = dockRight
		width, height := compactWidth, compactHeight
		if state.Expanded {
			width, height = expandedWidth, expandedPanelHeight(len(state.Items))
		}
		target, err := newCanvas(width, height)
		if err != nil {
			t.Fatalf("%s: canvas: %v", tc.name, err)
		}
		drawVerticalPanel(target, state) // warm-up: font cache, any lazy setup
		const runs = 50
		start := time.Now()
		for index := 0; index < runs; index++ {
			drawVerticalPanel(target, state)
		}
		perFrame := time.Since(start) / runs
		t.Logf("%-22s %3dx%-3d %8.3f ms/frame", tc.name, width, height, float64(perFrame.Microseconds())/1000)
		target.close()
	}
}

// repeatSessions is how the six-row case is built without a second literal: the
// panel's row count is what changes the paint cost, not the session contents.
func repeatSessions(base []session.Session, times int) []session.Session {
	result := make([]session.Session, 0, len(base)*times)
	for index := 0; index < times; index++ {
		for _, item := range base {
			copyOf := item
			copyOf.SessionID = item.SessionID + string(rune('0'+index))
			result = append(result, copyOf)
		}
	}
	return result
}
