package ui

// The three alert-layer bugs that produced phantom balloons and tones, pinned
// so they cannot come back.
//
// All three were reproduced on the shipped 1.0.5 binary in a sandbox before
// they were fixed, and each one is a different mistake:
//
//   - a session *disappearing* was announced as a completion;
//   - the "is this session still alive" question could be answered by another
//     process's window title;
//   - one conversation could hold several rows because the store keys on the
//     reporter's pid.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// A conversation that vanished while busy must not make a sound. It is the
// shape of "reporter died", "duplicate pid pruned" and "snapshot aged out",
// none of which proves anything finished.
func TestVanishedSessionIsSilent(t *testing.T) {
	settings := attentionSettings()
	settings.DonePolicy = DonePolicyAlways // the loudest setting: even then, silent

	in := AlertInput{
		Session:  session.Session{Source: "opencode", SessionID: "ses_gone", State: session.StateClosed},
		Kind:     session.KindNone,
		Previous: session.StateWorking,
		Existed:  true,
		Present:  true,
		Vanished: true,
	}
	got := Decide(in, settings, 0)
	if got.Balloon || got.Tone != "" || got.Unread {
		t.Fatalf("a vanished session spoke: %+v", got)
	}
	if got.Reason != decisionVanished {
		t.Fatalf("reason = %q, want %q", got.Reason, decisionVanished)
	}
}

// The same input without the Vanished flag is a genuine completion, so the test
// above is proving the flag matters rather than proving completions are off.
func TestVanishedFlagIsWhatSilencesIt(t *testing.T) {
	settings := attentionSettings()
	settings.DonePolicy = DonePolicyAlways
	in := AlertInput{
		Session:  session.Session{Source: "opencode", SessionID: "ses_done", State: session.StateClosed},
		Kind:     session.KindDone,
		Previous: session.StateWorking,
		Existed:  true,
		Present:  true,
	}
	got := Decide(in, settings, 0)
	if !got.Balloon || got.Tone != toneFinished {
		t.Fatalf("a reported completion should still speak: %+v", got)
	}
}

// vanishedInputs must mark what it emits; it used to synthesize KindDone.
func TestVanishedInputsAreMarkedNotFinished(t *testing.T) {
	a := &App{}
	previous := map[string]alertFacts{
		"opencode\x00ses_a\x001": {State: session.StateWorking, Kind: session.KindNone},
		"opencode\x00ses_b\x001": {State: session.StateWaitingPermission, Kind: session.KindApproval},
		"opencode\x00ses_c\x001": {State: session.StateIdle, Kind: session.KindDone},
	}
	inputs := a.vanishedInputs(previous, map[string]alertFacts{}, attendanceState{})
	if len(inputs) != 2 {
		t.Fatalf("want the two busy sessions, got %d", len(inputs))
	}
	for _, in := range inputs {
		if !in.Vanished {
			t.Fatalf("input for %q was not marked vanished", in.Session.SessionID)
		}
		if in.Kind == session.KindDone {
			t.Fatalf("input for %q still claims a completion", in.Session.SessionID)
		}
	}
}

// End to end through the plan: one vanished busy session, nothing else. No
// balloon, no tone, no unread mark - but the receipt still explains it.
func TestPlanAlertsStaysSilentForVanished(t *testing.T) {
	settings := attentionSettings()
	settings.DonePolicy = DonePolicyAlways
	plan := PlanAlerts([]AlertInput{{
		Session:  session.Session{Source: "opencode", SessionID: "ses_x", State: session.StateClosed},
		Kind:     session.KindNone,
		Previous: session.StateToolUse,
		Existed:  true,
		Vanished: true,
	}}, settings, nil)
	if plan.Balloon != nil || plan.Tone != "" || len(plan.Unread) != 0 {
		t.Fatalf("a vanished session produced an alert: balloon=%+v tone=%q unread=%v",
			plan.Balloon, plan.Tone, plan.Unread)
	}
	key := sessionAlertKey(session.Session{Source: "opencode", SessionID: "ses_x"})
	if plan.Reasons[key] != decisionVanished {
		t.Fatalf("the receipt should say %q, got %q", decisionVanished, plan.Reasons[key])
	}
}

// The receipt field must survive a round trip, because "why was I not told"
// is answered from this file rather than from memory.
func TestVanishedReachesTheReceipt(t *testing.T) {
	record := buildAlertRecord(time.Unix(1000, 0), []AlertInput{{
		Session:  session.Session{Source: "opencode", SessionID: "ses_r", State: session.StateClosed},
		Kind:     session.KindNone,
		Previous: session.StateWorking,
		Existed:  true,
		Vanished: true,
	}}, AlertPlan{Reasons: map[string]string{
		sessionAlertKey(session.Session{Source: "opencode", SessionID: "ses_r"}): decisionVanished,
	}}, attendanceState{}, false)
	if len(record.Sessions) != 1 {
		t.Fatalf("want one recorded session, got %d", len(record.Sessions))
	}
	if record.Sessions[0].Decision != decisionVanished {
		t.Fatalf("decision = %q, want %q", record.Sessions[0].Decision, decisionVanished)
	}
	if record.Sessions[0].Kind != "none" {
		t.Fatalf("kind = %q, want none: a vanished session is not asking for anything", record.Sessions[0].Kind)
	}
}
