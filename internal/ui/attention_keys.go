package ui

// The alert policy's keys, kinds and constants, in a file with no build tag.
//
// This is handover §55.6 step 1. internal/ui could not be compiled for darwin
// because `go build ./internal/ui` stopped at
//
//	attention.go:72: undefined: defaultIdleSeconds
//	attention.go:305: undefined: busyState
//	attention.go:363: undefined: alertLevel
//	attention.go:376: undefined: sessionAlertKey
//	attention.go:398: undefined: alertUrgent
//
// and every one of those is pure: a threshold, a state test, a ranking, a key.
// They lived in app_windows.go only because that is where the scan that uses
// them lives, not because they touch a window. Moving them here means the
// darwin/linux build can type-check the 400-line policy in attention.go before
// the window layer exists - which is what makes the rest of the port possible
// to review in small steps.

import (
	"fmt"
	"strings"
	"time"

	"ai-terminal-pet/internal/session"
)

// defaultIdleSeconds mirrors config.DefaultIdleMinutes. The ui package
// deliberately does not import config (the app layer owns the file paths), so
// internal/app has a test that the two values agree.
const defaultIdleSeconds = 3 * 60

// alertLevel ranks what just happened across all sessions in one load.
type alertLevel uint8

const (
	alertNone alertLevel = iota
	alertFinished
	alertUrgent
)

// alertFacts is what the alert policy remembers about one conversation between
// scans: the state it was in and the kind it was asking for. The kind is the
// half that makes "answered a permission prompt, then got asked a question"
// visible - the state did not move, but the ask did.
type alertFacts struct {
	State session.State
	Kind  session.Kind
	// At is when this state first appeared, which is what a completion's
	// coalescing window measures against.
	At time.Time
}

// factsOf builds the snapshot the next scan compares against. When neither the
// state nor the kind changed, the original timestamp is kept: coalescing
// measures "how long has this been settled", not "when did the reporter last
// repeat itself".
func factsOf(items []session.Session, now time.Time, previous map[string]alertFacts) map[string]alertFacts {
	facts := make(map[string]alertFacts, len(items))
	for _, value := range items {
		key := sessionAlertKey(value)
		kind := session.KindOfSession(value)
		fact := alertFacts{State: value.State, Kind: kind, At: now}
		if old, ok := previous[key]; ok && old.State == value.State && old.Kind == kind {
			fact.At = old.At
		}
		facts[key] = fact
	}
	return facts
}

// busyState is true for states that mean "the agent is doing something".
// Leaving one of these for idle/closed is a real completion, whereas an idle
// session that stays idle is not.
func busyState(state session.State) bool {
	switch state {
	case session.StateWorking, session.StateToolUse, session.StateThinking,
		session.StateWaitingPermission, session.StateError:
		return true
	}
	return false
}

// sessionAlertKey is the ledger key for one conversation.
//
// It is session.Identity (source + sessionId) rather than the store key, and
// that is a deliberate change: the store key carries the reporter's pid, so an
// OpenCode restart used to look like a brand-new session to the alert ledger -
// a waiting conversation could be announced twice, once under each pid. The
// mute list and the unread marks were already keyed on identity; the ledger is
// the last place that still keyed on the process.
//
// The pid is only used as a fallback when the reporter sent no session id at
// all, which the protocol allows but no real reporter does.
func sessionAlertKey(value session.Session) string {
	if identity := session.Identity(value); identity != "" {
		return identity
	}
	return fmt.Sprintf("%s\x00\x00%d", value.Source, value.PID)
}

// splitAlertKey reads the first two parts of sessionAlertKey back. The pid is
// not needed here, and the two that are needed identify whose completion this
// was.
func splitAlertKey(key string) (source, sessionID string) {
	parts := strings.SplitN(key, "\x00", 3)
	if len(parts) > 0 {
		source = parts[0]
	}
	if len(parts) > 1 {
		sessionID = parts[1]
	}
	return source, sessionID
}
