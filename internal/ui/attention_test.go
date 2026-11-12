package ui

// The alert policy, pinned. Every rule that can silence a balloon has a test
// that proves it silences *only* what it should, and the two bugs this file
// exists to fix have their own regression tests.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func attentionSettings() AttentionSettings {
	return AttentionSettings{
		Notifications:   true,
		Sound:           true,
		SoundOnWaiting:  true,
		SoundOnError:    true,
		DonePolicy:      DonePolicyBackground,
		QuietCurrent:    true,
		ForegroundAware: true,
		IdleAfter:       3 * time.Minute,
		QuietSubagents:  true,
	}
}

func blocked(reason string) session.Session {
	return session.Session{
		Source: "claude-hook", SessionID: "s", PID: 11,
		Project: `C:\work\proj`, State: session.StateWaitingPermission, Reason: reason,
	}
}

// waiting is a small helper so the cases read as facts rather than as structs.
func waiting(state session.State, reason string) session.Session {
	return session.Session{Source: "src", SessionID: "id", PID: 1, State: state, Reason: reason}
}

func asking(kind session.Kind) AlertInput {
	return AlertInput{Session: waiting(stateForKind(kind), ""), Kind: kind, Existed: true, Present: true, PresenceKnown: true}
}

func stateForKind(kind session.Kind) session.State {
	switch kind {
	case session.KindError:
		return session.StateError
	case session.KindDone:
		return session.StateIdle
	case session.KindApproval, session.KindQuestion, session.KindPlan:
		return session.StateWaitingPermission
	}
	return session.StateWorking
}

// Rule 4 and 5: a balloon is suppressed only when the user is looking at that
// very session, awake. Everything else announces.
func TestDecideAnnouncesUnlessTheUserIsLookingAtIt(t *testing.T) {
	settings := attentionSettings()
	cases := []struct {
		name          string
		attending     bool
		present       bool
		presenceKnown bool
		wantBalloon   bool
		wantUnread    bool
		wantReason    string
	}{
		{"user is somewhere else", false, true, true, true, true, decisionForeground},
		{"user left the machine", false, false, true, true, true, decisionIdleAway},
		{"user is looking right at it", true, true, true, false, true, decisionAttendingQuiet},
		{"sample failed: announce", false, false, false, true, true, decisionPresenceUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := asking(session.KindApproval)
			in.Attending, in.Present, in.PresenceKnown = tc.attending, tc.present, tc.presenceKnown
			got := Decide(in, settings, 0)
			if got.Balloon != tc.wantBalloon || got.Unread != tc.wantUnread || got.Reason != tc.wantReason {
				t.Fatalf("got balloon=%v unread=%v reason=%q, want %v/%v/%q",
					got.Balloon, got.Unread, got.Reason, tc.wantBalloon, tc.wantUnread, tc.wantReason)
			}
		})
	}
}

// The unread mark survives the suppression. Without it, "quiet in the current
// session" would be indistinguishable from "this session was never noticed".
func TestDecideMarksUnreadEvenWhenQuiet(t *testing.T) {
	settings := attentionSettings()
	in := asking(session.KindApproval)
	in.Attending, in.Present = true, true
	if got := Decide(in, settings, 0); got.Balloon || !got.Unread {
		t.Fatalf("quiet-in-place must still mark unread: balloon=%v unread=%v", got.Balloon, got.Unread)
	}
	// And the unread mark is independent of the balloon preference.
	settings.Notifications = false
	if got := Decide(in, settings, 0); got.Unread != true || got.Balloon != false {
		t.Fatalf("balloons off must keep the unread mark: %+v", got)
	}
}

// Foreground awareness off is the "always tell me" mode, and it must announce
// even when the user is looking straight at the panel.
func TestDecideWithoutForegroundAwarenessAlwaysAnnounces(t *testing.T) {
	settings := attentionSettings()
	settings.ForegroundAware = false
	in := asking(session.KindApproval)
	in.Attending, in.Present = true, true
	got := Decide(in, settings, 0)
	if !got.Balloon || got.Reason != decisionAwarenessOff {
		t.Fatalf("awareness off must announce: %+v", got)
	}
}

// Muting wins over everything - including an error, and including the
// "quiet current session" path.
func TestDecideMuteSilencesEveryKind(t *testing.T) {
	settings := attentionSettings()
	for _, kind := range []session.Kind{session.KindApproval, session.KindQuestion, session.KindPlan, session.KindError} {
		in := asking(kind)
		in.Muted = true
		in.Attending = false
		in.Present = false
		got := Decide(in, settings, 0)
		if got.Balloon || got.Tone != "" || got.Unread || got.Reason != decisionMuted {
			t.Fatalf("%s while muted produced %+v", kind, got)
		}
	}
}

// A subagent session is silent unless the preference is turned off.
func TestDecideSubagentsAreSilentByDefault(t *testing.T) {
	settings := attentionSettings()
	in := asking(session.KindApproval)
	in.Subagent = true
	if got := Decide(in, settings, 0); got.Balloon || got.Reason != decisionSubagent {
		t.Fatalf("subagent should be silent: %+v", got)
	}
	settings.QuietSubagents = false
	if got := Decide(in, settings, 0); !got.Balloon {
		t.Fatalf("subagent should announce once quieting is off: %+v", got)
	}
}

// The three-way completion policy.
func TestDecideCompletionPolicy(t *testing.T) {
	cases := []struct {
		name      string
		policy    DonePolicy
		attending bool
		present   bool
		want      bool
	}{
		{"off, user elsewhere", DonePolicyOff, false, true, false},
		{"background, user elsewhere", DonePolicyBackground, false, true, true},
		{"background, user watching it", DonePolicyBackground, true, true, false},
		{"always, user watching it", DonePolicyAlways, true, true, true},
		{"always, user elsewhere", DonePolicyAlways, false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			settings := attentionSettings()
			settings.DonePolicy = tc.policy
			in := completed()
			in.Present, in.Attending = tc.present, tc.attending
			if got := Decide(in, settings, 0).Balloon; got != tc.want {
				t.Fatalf("balloon=%v, want %v", got, tc.want)
			}
		})
	}
}

// A completion never marks a row unread: the row already shows idle, and an
// unread badge on a finished session would be a second, contradictory claim.
func TestDecideCompletionIsNeverUnread(t *testing.T) {
	settings := attentionSettings()
	in := completed()
	if got := Decide(in, settings, 0); got.Unread {
		t.Fatalf("a completion must not mark unread: %+v", got)
	}
}

// Two completions of the same session inside the window are one event, which
// is what keeps a multi-step turn from ringing twice.
func TestDecideCompletionCoalesces(t *testing.T) {
	settings := attentionSettings()
	in := completed()
	if got := Decide(in, settings, 2*time.Second); got.Balloon || got.Reason != decisionDoneCoalesced {
		t.Fatalf("a repeat inside the window should fold: %+v", got)
	}
	if got := Decide(in, settings, completionCoalesceWindow+time.Second); !got.Balloon {
		t.Fatalf("a completion outside the window should speak: %+v", got)
	}
}

// completed is a session that really finished: it existed, and it was busy
// before. A session that is merely idle is not a completion, and the policy
// refuses to announce one.
func completed() AlertInput {
	return AlertInput{
		Session:       waiting(session.StateIdle, ""),
		Kind:          session.KindDone,
		Previous:      session.StateWorking,
		Existed:       true,
		PresenceKnown: true,
	}
}

// Busy work never alerts: there is nothing to decide or read while the agent
// is still working.
func TestDecideBusyStatesAreSilent(t *testing.T) {
	settings := attentionSettings()
	for _, state := range []session.State{session.StateWorking, session.StateToolUse, session.StateThinking} {
		in := AlertInput{Session: waiting(state, ""), Kind: session.KindOf(state, ""), Present: false, PresenceKnown: true}
		if got := Decide(in, settings, 0); got.Balloon || got.Tone != "" || got.Unread {
			t.Fatalf("%s should be silent: %+v", state, got)
		}
	}
}

// Tones are per-kind and respect the two existing event gates.
func TestDecideTones(t *testing.T) {
	settings := attentionSettings()
	in := asking(session.KindError)
	if got := Decide(in, settings, 0); got.Tone != toneError {
		t.Fatalf("an error should use the error timbre, got %q", got.Tone)
	}
	in = asking(session.KindQuestion)
	if got := Decide(in, settings, 0); got.Tone != toneWaiting {
		t.Fatalf("a question should use the waiting timbre, got %q", got.Tone)
	}
	settings.SoundOnWaiting = false
	if got := Decide(in, settings, 0); got.Tone != "" {
		t.Fatalf("muting waiting must silence a question too, got %q", got.Tone)
	}
	settings.Sound = false
	in = asking(session.KindError)
	if got := Decide(in, settings, 0); got.Tone != "" {
		t.Fatalf("the master switch must silence every tone, got %q", got.Tone)
	}
}

// The bug this whole file was written for: a permission prompt answered and a
// question asked is the SAME state, so a state-keyed map saw no change. It has
// to alert, and it has to be visible as a different reason.
//
// The sequence matters and is what the inputs encode: the first event is a
// transition *into* waiting_permission (from working), and the second is a
// change of kind while the state stands still. A scan that finds the state and
// the kind both where they already were is not a transition at all - see
// TestDecideIgnoresUnchangedSessions.
func TestDecideSeesAFollowUpAskInTheSameState(t *testing.T) {
	settings := attentionSettings()
	first := AlertInput{Session: waiting(session.StateWaitingPermission, "permission"),
		Kind: session.KindApproval, Existed: true, Previous: session.StateWorking,
		Present: false, PresenceKnown: true}
	second := AlertInput{Session: waiting(session.StateWaitingPermission, "question"),
		Kind: session.KindQuestion, Existed: true, Previous: session.StateWaitingPermission,
		ChangedKind: true, Present: false, PresenceKnown: true}
	for name, in := range map[string]AlertInput{"approval": first, "follow-up question": second} {
		got := Decide(in, settings, 0)
		if !got.Balloon || !got.Unread {
			t.Fatalf("%s must alert: %+v", name, got)
		}
	}
}

// A session that is exactly where it already was is not news, however many
// times the scanner looks at it. This is the guard that keeps an 8-second
// sample cadence from turning a parked permission prompt into a metronome.
func TestDecideIgnoresUnchangedSessions(t *testing.T) {
	settings := attentionSettings()
	in := AlertInput{Session: waiting(session.StateWaitingPermission, "permission"),
		Kind: session.KindApproval, Existed: true, Previous: session.StateWaitingPermission,
		Present: false, PresenceKnown: true}
	got := Decide(in, settings, 0)
	if got.Balloon || got.Unread || got.Tone != "" || got.Reason != decisionUnchanged {
		t.Fatalf("an unchanged session must be silent, got %+v", got)
	}
}

// PlanAlerts sends exactly one balloon per scan, to the most urgent session,
// and records why each session did or did not speak.
func TestPlanAlertsSendsOneBalloonToTheMostUrgent(t *testing.T) {
	settings := attentionSettings()
	error := AlertInput{Session: waiting(session.StateError, "boom"), Kind: session.KindError,
		Present: false, PresenceKnown: true}
	question := AlertInput{Session: waiting(session.StateWaitingPermission, "question"), Kind: session.KindQuestion,
		Present: false, PresenceKnown: true}
	muted := AlertInput{Session: waiting(session.StateWaitingPermission, "permission"), Kind: session.KindApproval,
		Present: false, PresenceKnown: true, Muted: true}
	// Inputs arrive in the store's priority order: waiting_permission first.
	plan := PlanAlerts([]AlertInput{question, error, muted}, settings, nil)
	if plan.Balloon == nil || plan.Balloon.State != session.StateWaitingPermission {
		t.Fatalf("the balloon should name the waiting session, got %+v", plan.Balloon)
	}
	if len(plan.Unread) != 2 {
		t.Fatalf("two unmuted urgent sessions should be unread, got %d: %v", len(plan.Unread), plan.Unread)
	}
	if plan.Reasons[sessionAlertKey(muted.Session)] != decisionMuted {
		t.Fatalf("a muted session's reason should be recorded, got %q", plan.Reasons[sessionAlertKey(muted.Session)])
	}
	if plan.Level != alertUrgent {
		t.Fatalf("level should be urgent, got %d", plan.Level)
	}
}

// A scan with nothing urgent still records the reason for every session, which
// is what makes "--json why not" answerable.
func TestPlanAlertsRecordsEveryReason(t *testing.T) {
	settings := attentionSettings()
	busy := waiting(session.StateWorking, "")
	busy.SessionID = "busy"
	idle := waiting(session.StateIdle, "")
	idle.SessionID = "idle"
	inputs := []AlertInput{
		{Session: busy, Kind: session.KindNone, Present: false, PresenceKnown: true},
		{Session: idle, Kind: session.KindDone, Present: true, Attending: true, PresenceKnown: true},
	}
	plan := PlanAlerts(inputs, settings, nil)
	if plan.Balloon != nil {
		t.Fatalf("nothing here should balloon: %+v", plan.Balloon)
	}
	if len(plan.Reasons) != 2 {
		t.Fatalf("every input needs a reason, got %v", plan.Reasons)
	}
	if plan.Level != alertNone {
		t.Fatalf("level should be none, got %d", plan.Level)
	}
}

func TestValidDonePolicy(t *testing.T) {
	cases := map[string]DonePolicy{
		"":           DonePolicyBackground,
		"off":        DonePolicyOff,
		"background": DonePolicyBackground,
		"always":     DonePolicyAlways,
		"junk":       DonePolicyBackground,
	}
	for input, want := range cases {
		if got := ValidDonePolicy(input); got != want {
			t.Errorf("ValidDonePolicy(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDonePolicyNamesStartWithTheDefault(t *testing.T) {
	if len(donePolicyNames) != 3 || donePolicyNames[0] != DonePolicyBackground {
		t.Fatalf("the segment must offer three options and start on the default, got %v", donePolicyNames)
	}
}
