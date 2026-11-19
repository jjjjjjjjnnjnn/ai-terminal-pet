package ui

// Which reports deserve a row on the panel.
//
// The overlay is an attention strip, not a session browser. OpenCode reports a
// row for every conversation its long-lived service has ever seen, including
// child sessions that are steps of somebody else's turn. Those children never
// have a terminal tab of their own - they have no `time_viewed` in the host's
// database - so drawing them inflates the count and gives the user arrows that
// cannot lead anywhere.
//
// This is deliberately a *display* filter, not a store rule: --json keeps every
// reported row so a bug report can still show what the reporter sent. The same
// reasoning lives in sessions_dedupe.go.

import (
	"strings"

	"ai-terminal-pet/internal/session"
)

// panelRows drops settled child sessions. A child that is actually asking for
// the user is kept: a permission prompt from a subagent still needs a human,
// even though it cannot be mapped to a tab.
func panelRows(items []session.Session) []session.Session {
	visible := make([]session.Session, 0, len(items))
	for _, current := range items {
		if strings.TrimSpace(current.ParentSessionID) != "" &&
			!session.KindOfSession(current).Urgent() {
			continue
		}
		visible = append(visible, current)
	}
	return visible
}
