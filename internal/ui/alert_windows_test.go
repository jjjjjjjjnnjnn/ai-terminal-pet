package ui

// The regression tests that shaped this area, kept and re-pointed at the pure
// policy. The old code answered these three questions with four functions
// (currentAlertStates / transitionLevel / decideAlerts / soundForState); the
// questions are the same, so the tests are the same questions asked of the
// replacement, and every one of them is a bug that actually shipped.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// TestFactsOfKeysArePerConversation: the ledger must be keyed on the
// conversation, not on a single global "most urgent session" and not on the
// reporter process. A single global key used to be cleared when that session
// ended, which made a still-waiting sibling look like it had been caused by an
// unrelated new session; a pid-keyed ledger instead made an OpenCode restart
// look like a brand-new conversation, so a waiting ask could be announced
// twice - once per pid.
func TestFactsOfKeysArePerConversation(t *testing.T) {
	a := session.Session{Source: "deepseek-harness", SessionID: "ses_1", PID: 11, State: session.StateWorking}
	b := session.Session{Source: "deepseek-harness", SessionID: "ses_1", PID: 22, State: session.StateWaitingPermission}
	c := session.Session{Source: "deepseek-harness", SessionID: "ses_2", PID: 11, State: session.StateIdle}
	if sessionAlertKey(a) != sessionAlertKey(b) {
		t.Fatal("one conversation reported by two pids must share a ledger key")
	}
	if sessionAlertKey(a) == sessionAlertKey(c) {
		t.Fatal("two conversations must not share a ledger key")
	}
	// Two conversations, three reports: the map holds two entries, and the
	// second conversation keeps its own state.
	facts := factsOf([]session.Session{c, a}, time.Now(), nil)
	if len(facts) != 2 {
		t.Fatalf("want 2 conversations, got %d", len(facts))
	}
	if facts[sessionAlertKey(c)].Kind != session.KindDone {
		t.Fatal("facts were mixed between conversations")
	}
	if facts[sessionAlertKey(a)].State != session.StateWorking {
		t.Fatal("the wrong report won the conversation")
	}
}

func TestBusyState(t *testing.T) {
	for _, state := range []session.State{
		session.StateWorking, session.StateToolUse, session.StateThinking,
		session.StateWaitingPermission, session.StateError,
	} {
		if !busyState(state) {
			t.Errorf("%q should count as busy", state)
		}
	}
	for _, state := range []session.State{session.StateIdle, session.StateClosed} {
		if busyState(state) {
			t.Errorf("%q should not count as busy", state)
		}
	}
}

// Every conversation is tracked separately, so a sibling ending never
// re-announces one that has been waiting the whole time. This is the ghost
// popup, and it is why the baseline is a map rather than one key.
func TestPlanAlertsDoesNotReannounceASibling(t *testing.T) {
	settings := attentionSettings()
	waiting := session.Session{Source: "deepseek-harness", SessionID: "A", PID: 1, State: session.StateWaitingPermission, Reason: "permission"}
	busy := session.Session{Source: "claude-ipc", SessionID: "B", PID: 2, State: session.StateWorking}
	fresh := session.Session{Source: "opencode", SessionID: "C", PID: 3, State: session.StateWorking}

	previous := factsOf([]session.Session{waiting, busy}, time.Now(), nil)
	now := time.Now()

	// B ends (its snapshot disappears) while A is still waiting. The only alert
	// is a soft "finished", attributed to B itself - A must not speak.
	plan := PlanAlerts([]AlertInput{
		{Session: waiting, Kind: session.KindApproval, Previous: session.StateWaitingPermission,
			Existed: true, PresenceKnown: true},
		{Session: session.Session{Source: "claude-ipc", SessionID: "B", PID: 2, State: session.StateClosed},
			Kind: session.KindDone, Previous: session.StateWorking, Existed: true, PresenceKnown: true},
	}, settings, nil)
	if plan.Balloon == nil || plan.Balloon.SessionID != "B" {
		t.Fatalf("the balloon should name the session that finished, got %+v", plan.Balloon)
	}
	if plan.Reasons[sessionAlertKey(waiting)] != decisionUnchanged {
		t.Fatalf("the waiting sibling spoke: %q", plan.Reasons[sessionAlertKey(waiting)])
	}

	// Starting an unrelated session while A waits must be completely silent too.
	plan = PlanAlerts([]AlertInput{
		{Session: waiting, Kind: session.KindApproval, Previous: session.StateWaitingPermission,
			Existed: true, PresenceKnown: true},
		{Session: fresh, Kind: session.KindNone, Previous: session.StateIdle,
			Existed: true, Present: true, PresenceKnown: true},
	}, settings, nil)
	if plan.Balloon != nil {
		t.Fatalf("a new working session produced a balloon: %+v", plan.Balloon)
	}
	_ = previous
	_ = now

	// A real new urgent session still gets announced, with itself as subject.
	plan = PlanAlerts([]AlertInput{
		{Session: waiting, Kind: session.KindApproval, Previous: session.StateIdle,
			Existed: true, PresenceKnown: true},
	}, settings, nil)
	if plan.Balloon == nil || plan.Balloon.SessionID != "A" {
		t.Fatalf("a session that just started waiting was missed: %+v", plan.Balloon)
	}
}

// A session that opens already blocked does need announcing; a plain new busy
// one must not. That was the "it only pops up when I start a new session"
// complaint, and the rule now lives in Decide's Existed branch.
func TestNewSessionAnnouncedOnlyWhenBlocked(t *testing.T) {
	settings := attentionSettings()
	cases := []struct {
		name  string
		state session.State
		want  bool
	}{
		{"new working session is silent", session.StateWorking, false},
		{"new idle session is silent", session.StateIdle, false},
		{"new session already waiting", session.StateWaitingPermission, true},
		{"new session already failed", session.StateError, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := AlertInput{
				Session: session.Session{Source: "opencode", SessionID: "n", PID: 5, State: tc.state},
				Kind:    session.KindOf(tc.state, ""), Existed: false, PresenceKnown: true,
			}
			got := Decide(in, settings, 0)
			if got.Balloon != tc.want {
				t.Fatalf("balloon=%v, want %v (%+v)", got.Balloon, tc.want, got)
			}
		})
	}
}

// The per-event switches must actually gate the tone: muting "waiting" must not
// silence errors, and the master switch must silence everything. The gate moved
// from state to kind when questions and plan reviews joined waiting_permission,
// and this is the test that keeps both halves honest.
func TestSoundWantedHonoursPerEventSwitches(t *testing.T) {
	on := attentionSettings()
	if !soundWanted(session.KindApproval, on) || !soundWanted(session.KindError, on) {
		t.Fatal("everything should be audible by default")
	}
	quietWaiting := on
	quietWaiting.SoundOnWaiting = false
	if soundWanted(session.KindApproval, quietWaiting) {
		t.Fatal("muting waiting did not silence the waiting tone")
	}
	if !soundWanted(session.KindError, quietWaiting) {
		t.Fatal("muting waiting also silenced errors")
	}
	// A question is the same interruption as an approval, so it follows the
	// same gate - not a third switch nobody asked for.
	if soundWanted(session.KindQuestion, quietWaiting) {
		t.Fatal("muting waiting must also silence a question")
	}
	quietError := on
	quietError.SoundOnError = false
	if soundWanted(session.KindError, quietError) {
		t.Fatal("muting errors did not silence the error tone")
	}
	master := on
	master.Sound = false
	in := AlertInput{
		Session: session.Session{Source: "opencode", SessionID: "m", PID: 9, State: session.StateWaitingPermission},
		Kind:    session.KindApproval, Previous: session.StateWorking, Existed: true, PresenceKnown: true,
	}
	if got := Decide(in, master, 0); got.Tone != "" {
		t.Fatalf("the master switch must silence every tone, got %q", got.Tone)
	}
}
