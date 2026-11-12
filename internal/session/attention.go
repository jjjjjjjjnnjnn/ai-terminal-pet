package session

// What a waiting session is actually waiting for.
//
// The wire protocol has one state for "the agent needs you" -
// waiting_permission - and it is deliberately not being extended with a second
// field: adding a key would make every already-installed reporter's payload
// invalid (internal/session rejects unknown fields), which is a protocol
// break for a cosmetic distinction. The distinction is instead **derived**
// from the free-text `reason` that reporters already send, so today's hooks
// need no change and tomorrow's can carry a better one.
//
// Why it matters at all: a permission prompt and an open question ask the
// human for different things. An approval is "may I run this?", a question is
// "what did you mean?", a plan review is "is this the right plan?". They arrive
// as the same state and used to produce the same generic alert, which also made
// a second request after the first was answered look like no change at all -
// see decideAlert below.

import "strings"

// Kind is what the user is being asked to do. The zero value means "nothing".
type Kind string

const (
	KindNone     Kind = ""
	KindApproval Kind = "approval"
	KindQuestion Kind = "question"
	KindPlan     Kind = "plan"
	KindError    Kind = "error"
	KindDone     Kind = "done"
)

// KindOf derives the kind of one session from what it reports.
//
// Anything unknown inside waiting_permission falls back to KindApproval rather
// than to KindNone: the state itself already proves the agent is blocked, and
// dropping the alert because its wording was unrecognised is the one failure
// this whole file exists to prevent. Silence on a real blocker is worse than a
// slightly imprecise label.
func KindOf(state State, reason string) Kind {
	switch state {
	case StateWaitingPermission:
		lower := strings.ToLower(reason)
		switch {
		case strings.Contains(lower, "question") || strings.Contains(lower, "ask"):
			return KindQuestion
		case strings.Contains(lower, "plan"):
			return KindPlan
		default:
			return KindApproval
		}
	case StateError:
		return KindError
	case StateIdle, StateClosed:
		return KindDone
	default:
		return KindNone
	}
}

// Urgent reports whether this kind is something the user has to act on, as
// opposed to something that simply finished. Busy states are neither: there is
// nothing to decide or read while the agent is still working.
func (k Kind) Urgent() bool {
	switch k {
	case KindApproval, KindQuestion, KindPlan, KindError:
		return true
	}
	return false
}

// Label is the stable, untranslated name used in diagnostics and `--json`.
// It is not drawn on the panel: the panel shows the reporter's own reason
// text, which is more specific than any of these.
func (k Kind) Label() string {
	return string(k)
}

// KindOfSession is the shorthand every caller in internal/ui actually wants.
func KindOfSession(value Session) Kind {
	return KindOf(value.State, value.Reason)
}

// Identity is the stable per-conversation key: source plus session id, and
// deliberately NOT the pid.
//
// The store keys on (source, sessionId, pid) because a reporter process can
// report for two conversations at once. Everything that remembers a *user
// decision* about a conversation - it is muted, it is unread, its last alert
// was a minute ago - must not use that key: hook reporters hand out a synthetic
// pid per process, so the pid changes on every hook invocation while the
// conversation does not. Keying on it would lose every mute and every unread
// mark at the next event.
//
// config.SessionMuteKey builds the same string from the two parts it has, and
// internal/app has a test that the two agree - they are a cross-package
// contract, not a coincidence.
func Identity(value Session) string {
	source := strings.TrimSpace(value.Source)
	sessionID := strings.TrimSpace(value.SessionID)
	if source == "" || sessionID == "" {
		return ""
	}
	return source + "\x00" + sessionID
}
