package ui

// The one place that decides whether a session change becomes an alert.
//
// What this replaced: an `if` chain inside alertOnTransitions over a map of
// last states, with a separate sound gate and a separate balloon gate next to
// it. Two real bugs followed from that shape:
//
//  1. The map keyed on state alone, so answering a permission prompt and
//     immediately being asked a question - same waiting_permission, different
//     reason - produced no alert at all. Approve-then-ask is the commonest
//     cadence an agent has.
//  2. Every other question (is the user even looking at that terminal? is this
//     session muted? which completion policy applies?) had nowhere to live, so
//     it either became one more `if` or was never asked.
//
// The policy is now a pure function of one AlertInput plus a settings snapshot.
// Nothing here reads a window, a file or a clock, so every rule below is pinned
// by a table test.
//
// The governing principle, the same one internal/platform's presence model
// uses: **when the facts are unknown, alert.** Announcing slightly too often is
// a small annoyance; staying silent while an agent is blocked is the product
// failing at its only job, and nobody ever finds out it did not happen.

import (
	"time"

	"ai-terminal-pet/internal/session"
)

// DonePolicy is the "reply finished" preference. Three values rather than a
// switch, because the useful middle ground is not a boolean: someone working in
// one window wants the completions of the *other* sessions and nothing else.
type DonePolicy string

const (
	// DonePolicyOff never announces a completion.
	DonePolicyOff DonePolicy = "off"
	// DonePolicyBackground announces a completion only when the session it
	// belongs to was not the one being looked at. Default: a session that
	// finishes in front of you needs no balloon.
	DonePolicyBackground DonePolicy = "background"
	// DonePolicyAlways announces every completion.
	DonePolicyAlways DonePolicy = "always"
)

// ValidDonePolicy normalises an unknown or empty value to the default, so a
// hand-edited config.json cannot silence every completion by accident.
func ValidDonePolicy(value string) DonePolicy {
	switch DonePolicy(value) {
	case DonePolicyOff:
		return DonePolicyOff
	case DonePolicyAlways:
		return DonePolicyAlways
	default:
		return DonePolicyBackground
	}
}

// donePolicyNames is the complete, user-visible set, in segment order. Order
// matters: the first entry is the default the segment starts on.
var donePolicyNames = []DonePolicy{DonePolicyBackground, DonePolicyOff, DonePolicyAlways}

// DonePolicyNames exposes the set to internal/app, which pins it against
// config.DonePolicies - the same deliberate duplication the theme names have,
// because ui does not import config.
func DonePolicyNames() []DonePolicy { return donePolicyNames }

// DefaultIdleSeconds is the shipped "the window is in front but nobody is
// there" threshold, exposed for the same cross-package pin.
func DefaultIdleSeconds() int { return defaultIdleSeconds }

// SessionIdentityFor exposes session.Identity's encoding to internal/app, so a
// test can prove config.SessionMuteKey and the ui lookup cannot drift apart.
func SessionIdentityFor(source, sessionID string) string {
	return session.Identity(session.Session{Source: source, SessionID: sessionID})
}

// SecondsToIdleMinutes converts the ui-side duration into the persisted
// minutes. Truncating is deliberate: config.json stores whole minutes, and
// rounding up would silently make the threshold longer than the user set.
func SecondsToIdleMinutes(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return seconds / 60
}

// AttentionSettings is the live preference snapshot the policy reads. A plain
// struct rather than Settings: the policy has no business seeing tray labels or
// audio volumes, and keeping it narrow is what lets its tests state only the
// facts that matter.
type AttentionSettings struct {
	Notifications  bool
	Sound          bool
	SoundOnWaiting bool
	SoundOnError   bool
	// DonePolicy decides when a finished turn is announced.
	DonePolicy DonePolicy
	// QuietCurrent suppresses the balloon for a session the user is demonstrably
	// looking at. It never suppresses the unread mark.
	QuietCurrent bool
	// ForegroundAware turns the "is the user looking" question on at all. When
	// it is off, every urgent ask announces - which is what the overlay did
	// before this existed.
	ForegroundAware bool
	// IdleAfter is how long without keyboard or mouse input counts as being
	// away even while a window is in front. Zero means "window focus only".
	IdleAfter time.Duration
	// QuietSubagents silences sessions that belong to another one.
	QuietSubagents bool
}

// AlertInput is everything the policy knows about one session at one scan.
type AlertInput struct {
	// Session is the report as the store holds it.
	Session session.Session
	// Kind is what it is asking for, derived from state and reason.
	Kind session.Kind
	// Previous and Existed describe the last scan.
	Previous session.State
	Existed  bool
	// ChangedKind is true when the state stood still while the ask changed -
	// the case the old map could not see.
	ChangedKind bool

	// Attending is true when the window that owns this session is the one in
	// front; Present when the user is at the machine. Together they answer
	// "is the user already looking at this very session".
	Attending bool
	Present   bool
	// PresenceKnown is false when the sample failed. Unknown is read as "not
	// looking", per the governing principle.
	PresenceKnown bool

	// Muted is the user's per-session silence; Subagent marks a child session.
	Muted    bool
	Subagent bool

	// Vanished marks a conversation the store dropped while it was busy.
	//
	// This is deliberately NOT a completion. A reporter that dies, a duplicate
	// pid that gets pruned, or a snapshot that ages out all look like "the busy
	// session went away", and announcing that as "finished" is a claim nothing
	// supports: the agent may have been killed, or it may still be running in a
	// terminal the overlay merely lost sight of. It is carried through so the
	// receipt can say what happened (reason "vanished") without any balloon,
	// tone or unread mark.
	Vanished bool
}

// AlertDecision is the answer for one session. Deliberately not a bool: a
// balloon, a tone and an unread mark are three channels, and forcing them to
// move together is what made "quiet in the foreground" and "mute this session"
// impossible to express.
type AlertDecision struct {
	// Balloon authorises the balloon layer to say something.
	Balloon bool
	// Tone is the synthesised timbre to play, or "" for silence.
	Tone string
	// Unread marks the row until the user looks at it, and is set even when the
	// balloon is suppressed - that is the point of suppressing it.
	Unread bool
	// Reason is a short untranslated token naming why the decision came out
	// this way. It is written to the alert receipt and printed by --json, so
	// "why was I not told" has an answer.
	Reason string
}

// Timbres, matching the synthesised kind constants in sound_windows.go.
const (
	toneWaiting  = "waiting"
	toneError    = "error"
	toneFinished = "finished"
)

// Decision reasons. Diagnostics, never copy: they reach the receipt file and
// --json, not the panel.
const (
	decisionNotAsking       = "not-asking"
	decisionUnchanged       = "unchanged"
	decisionMuted           = "muted"
	decisionSubagent        = "subagent"
	decisionDoneOff         = "done-off"
	decisionDoneCoalesced   = "done-coalesced"
	decisionDoneSeen        = "done-seen"
	decisionDoneAway        = "done-away"
	decisionAttending       = "attending"
	decisionAttendingQuiet  = "attending-quiet"
	decisionForeground      = "foreground"
	decisionIdleAway        = "idle-away"
	decisionPresenceUnknown = "presence-unknown"
	decisionElsewhere       = "elsewhere"
	decisionAwarenessOff    = "awareness-off"
	// decisionVanished names a conversation that disappeared while it was busy.
	// It is a diagnostic, not a signal: nothing is announced for it.
	decisionVanished = "vanished"
)

// sinceSpoke answers "how long ago did this conversation last make a balloon",
// which is what the coalescing window measures. A missing entry is zero: it has
// never spoken, so nothing can be folded into it.
func sinceSpoke(lastAlertAt map[string]time.Time, key string) time.Duration {
	at, ok := lastAlertAt[key]
	if !ok {
		return 0
	}
	return time.Since(at)
}

// completionCoalesceWindow folds two completions of one session into one noise.
// A turn can end in steps - a subagent settling, then the parent - and each of
// them looks like a fresh completion to a state-only scan.
const completionCoalesceWindow = 15 * time.Second

// Decide is the whole alert policy for one session.
//
// The order is the policy, and it is deliberate:
//
//  1. Muting wins over everything. A user who silenced a job asked for quiet,
//     and second-guessing that with "but this one is urgent" is how a mute
//     stops being believed. The one thing muting never suppresses is the
//     auto-approve countdown drawn on the panel - a rendering concern and a
//     safety control, not an alert.
//  2. A subagent session is silent by default: hosts split work into child
//     sessions that are steps of a single turn.
//  3. A completion obeys the three-way policy and is coalesced.
//  4. An urgent ask announces unless the user is demonstrably already looking
//     at that session's terminal, and even then it marks itself unread.
//  5. Unmeasurable facts count as "not looking".
func Decide(in AlertInput, settings AttentionSettings, sinceFinish time.Duration) AlertDecision {
	if in.Muted {
		return AlertDecision{Reason: decisionMuted}
	}
	if in.Subagent && settings.QuietSubagents {
		return AlertDecision{Reason: decisionSubagent}
	}
	// Nothing moved, so there is nothing to announce. This guard lives in the
	// pure function rather than in the caller on purpose: the caller already
	// filters unchanged sessions, and if that filter ever regressed, a session
	// parked in waiting_permission would re-balloon on every scan. Pinned here,
	// the duplicate is impossible from any caller.
	//
	// The rule is state + kind, never the free-text reason: reporters reword
	// their reasons ("permission" / "permission_dangerous"), and turning each
	// rewording into a new balloon is how alerts become noise.
	if in.Existed && in.Previous == in.Session.State && !in.ChangedKind {
		return AlertDecision{Reason: decisionUnchanged}
	}
	// A conversation that vanished while it was busy is not a completion. The
	// old rule announced it as one - `vanishedInputs` synthesized a closed/done
	// session for it - and that is the "it rings although no agent finished"
	// report: a reporter dying, a duplicate pid being pruned, or a snapshot
	// aging out all produce it, and none of them is evidence that work ended.
	//
	// The honest reading of a disappearance is "unknown". The receipt still
	// records it (see decisionVanished), because a session that vanished while
	// blocked is worth being able to explain later, but nothing is announced.
	if in.Vanished {
		return AlertDecision{Reason: decisionVanished}
	}
	if in.Kind == session.KindDone {
		return decideCompletion(in, settings, sinceFinish)
	}
	if !in.Kind.Urgent() {
		return AlertDecision{Reason: decisionNotAsking}
	}

	decision := AlertDecision{Balloon: true, Unread: true, Reason: decisionElsewhere}
	switch {
	case !settings.ForegroundAware:
		decision.Reason = decisionAwarenessOff
	case !in.PresenceKnown:
		decision.Reason = decisionPresenceUnknown
	case in.Present && in.Attending:
		// Looking at it, and awake in front of it, is the only combination
		// that can suppress anything.
		decision.Reason = decisionAttending
		if settings.QuietCurrent {
			decision.Balloon = false
			decision.Reason = decisionAttendingQuiet
		}
	case in.Present && !in.Attending:
		decision.Reason = decisionForeground
	default:
		decision.Reason = decisionIdleAway
	}
	if !settings.Notifications {
		decision.Balloon = false
	}
	if settings.Sound && soundWanted(in.Kind, settings) {
		decision.Tone = toneForKind(in.Kind)
	}
	return decision
}

// decideCompletion applies the three-way completion policy.
//
// A completion is something that *ended*. A session that arrives already idle
// has finished nothing the user was waiting for, and announcing it is exactly
// the "it only pops up when I start a new session" complaint - so the guard is
// on evidence, not on timing: it must have existed and it must have been busy.
func decideCompletion(in AlertInput, settings AttentionSettings, sinceFinish time.Duration) AlertDecision {
	if !in.Existed || !busyState(in.Previous) {
		return AlertDecision{Reason: decisionNotAsking}
	}
	if sinceFinish > 0 && sinceFinish < completionCoalesceWindow {
		return AlertDecision{Reason: decisionDoneCoalesced}
	}
	switch settings.DonePolicy {
	case DonePolicyOff:
		return AlertDecision{Reason: decisionDoneOff}
	case DonePolicyAlways:
		return AlertDecision{Balloon: settings.Notifications, Tone: toneIf(settings.Sound, toneFinished), Reason: decisionDoneAway}
	default:
		// Background: a session the user has been staring at has been seen.
		if settings.ForegroundAware && in.Present && in.Attending {
			return AlertDecision{Reason: decisionDoneSeen}
		}
		return AlertDecision{Balloon: settings.Notifications, Tone: toneIf(settings.Sound, toneFinished), Reason: decisionDoneAway}
	}
}

// soundWanted gates the tone per kind, the same split the tray menu already
// offered for waiting and error. Questions and plan reviews follow the waiting
// gate because they are the same interruption to the same person.
func soundWanted(kind session.Kind, settings AttentionSettings) bool {
	if kind == session.KindError {
		return settings.SoundOnError
	}
	return settings.SoundOnWaiting
}

func toneIf(enabled bool, tone string) string {
	if !enabled {
		return ""
	}
	return tone
}

func toneForKind(kind session.Kind) string {
	if kind == session.KindError {
		return toneError
	}
	return toneWaiting
}

// AlertPlan is the result of one scan: at most one balloon, one tone, and the
// list of rows to mark unread.
type AlertPlan struct {
	// Balloon is the session the one balloon should name, or nil.
	Balloon *session.Session
	// Tone is the timbre for this scan, or "".
	Tone string
	// Unread lists the session keys to mark.
	Unread []string
	// Reasons records why each session did or did not speak, keyed by
	// sessionAlertKey. It is the receipt --json prints.
	Reasons map[string]string
	// Level is the highest rank any input reached, reusing the alertLevel the
	// transition scan already speaks in.
	Level alertLevel
}

// PlanAlerts is the pure scan over one refresh. Inputs arrive in priority
// order from the store, so ties keep the store's ordering. At most one balloon
// leaves this function: two balloons for two sessions in one scan would stack
// in the tray and neither would be read.
func PlanAlerts(inputs []AlertInput, settings AttentionSettings, lastAlertAt map[string]time.Time) AlertPlan {
	plan := AlertPlan{Reasons: make(map[string]string, len(inputs))}
	bestUrgent, bestFinished := -1, -1
	decisions := make([]AlertDecision, len(inputs))
	for index := range inputs {
		in := inputs[index]
		key := sessionAlertKey(in.Session)
		decision := Decide(in, settings, sinceSpoke(lastAlertAt, key))
		decisions[index] = decision
		plan.Reasons[key] = decision.Reason
		if decision.Unread {
			plan.Unread = append(plan.Unread, key)
		}
		if !decision.Balloon {
			continue
		}
		if in.Kind.Urgent() {
			if bestUrgent < 0 {
				bestUrgent = index
			}
		} else if bestFinished < 0 {
			bestFinished = index
		}
	}
	switch {
	case bestUrgent >= 0:
		plan.Balloon = &inputs[bestUrgent].Session
		plan.Tone = decisions[bestUrgent].Tone
		plan.Level = alertUrgent
	case bestFinished >= 0:
		plan.Balloon = &inputs[bestFinished].Session
		plan.Tone = decisions[bestFinished].Tone
		plan.Level = alertFinished
	}
	return plan
}
