package ui

import (
	"testing"

	"ai-terminal-pet/internal/session"
)

// The old code kept one global "most urgent session" key. Clearing it when
// that session ended made a still-waiting sibling re-alert later, so a balloon
// appeared to be caused by an unrelated new session. These cases pin the
// per-conversation transition model.
func TestTransitionLevel(t *testing.T) {
	cases := []struct {
		name     string
		previous session.State
		existed  bool
		current  session.State
		want     alertLevel
	}{
		{"new busy session is silent", session.StateIdle, false, session.StateWorking, alertNone},
		{"new idle session is silent", session.StateIdle, false, session.StateIdle, alertNone},
		{"new session already waiting", session.StateIdle, false, session.StateWaitingPermission, alertUrgent},
		{"new session already failed", session.StateIdle, false, session.StateError, alertUrgent},
		{"unchanged waiting does not repeat", session.StateWaitingPermission, true, session.StateWaitingPermission, alertNone},
		{"unchanged busy does not repeat", session.StateWorking, true, session.StateWorking, alertNone},
		{"busy to waiting is urgent", session.StateWorking, true, session.StateWaitingPermission, alertUrgent},
		{"busy to error is urgent", session.StateToolUse, true, session.StateError, alertUrgent},
		{"waiting to idle is finished", session.StateWaitingPermission, true, session.StateIdle, alertFinished},
		{"working to idle is finished", session.StateWorking, true, session.StateIdle, alertFinished},
		{"working to closed is finished", session.StateWorking, true, session.StateClosed, alertFinished},
		{"error to idle is finished", session.StateError, true, session.StateIdle, alertFinished},
		{"idle to working is not an alert", session.StateIdle, true, session.StateWorking, alertNone},
		{"idle to idle is not an alert", session.StateIdle, true, session.StateIdle, alertNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := transitionLevel(tc.previous, tc.existed, tc.current); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCurrentAlertStatesKeysArePerConversation(t *testing.T) {
	a := session.Session{Source: "deepseek-harness", SessionID: "ses_1", PID: 11, State: session.StateWorking}
	b := session.Session{Source: "deepseek-harness", SessionID: "ses_1", PID: 22, State: session.StateWaitingPermission}
	c := session.Session{Source: "deepseek-harness", SessionID: "ses_2", PID: 11, State: session.StateIdle}
	states := currentAlertStates([]session.Session{a, b, c})
	if len(states) != 3 {
		t.Fatalf("want 3 distinct identities, got %d", len(states))
	}
	if states[sessionAlertKey(a)] != session.StateWorking ||
		states[sessionAlertKey(b)] != session.StateWaitingPermission ||
		states[sessionAlertKey(c)] != session.StateIdle {
		t.Fatal("states were mixed between conversations")
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

func TestDecideAlertsDoesNotReannounceASibling(t *testing.T) {
	waiting := session.Session{Source: "deepseek-harness", SessionID: "A", PID: 1, State: session.StateWaitingPermission}
	busy := session.Session{Source: "claude-ipc", SessionID: "B", PID: 2, State: session.StateWorking}
	fresh := session.Session{Source: "opencode", SessionID: "C", PID: 3, State: session.StateWorking}

	// B ends (its snapshot disappears) while A is still waiting. The only
	// alert is a soft "finished" - A must NOT be announced again.
	previous := currentAlertStates([]session.Session{waiting, busy})
	level, subject := decideAlerts(previous, []session.Session{waiting})
	if level != alertFinished || subject != nil {
		t.Fatalf("sibling ending produced level=%d subject=%v, want finished/nil", level, subject)
	}

	// Starting an unrelated session while A waits must be completely silent.
	level, subject = decideAlerts(currentAlertStates([]session.Session{waiting}), []session.Session{waiting, fresh})
	if level != alertNone || subject != nil {
		t.Fatalf("starting a new session produced an alert: level=%d", level)
	}

	// A real new urgent session still gets announced, with itself as subject.
	level, subject = decideAlerts(currentAlertStates([]session.Session{busy}),
		[]session.Session{busy, waiting})
	if level != alertUrgent || subject == nil || subject.SessionID != "A" {
		t.Fatalf("a new waiting session was missed: level=%d subject=%v", level, subject)
	}

	// A settling into idle is a completion naming A itself.
	settled := waiting
	settled.State = session.StateIdle
	level, subject = decideAlerts(currentAlertStates([]session.Session{waiting}), []session.Session{settled})
	if level != alertFinished || subject == nil || subject.SessionID != "A" {
		t.Fatalf("completion was not attributed to its session: level=%d subject=%v", level, subject)
	}
}

// The per-event switches must actually gate the tone: muting "waiting" must not
// silence errors, and the master switch must silence everything.
func TestSoundForStateHonoursPerEventSwitches(t *testing.T) {
	on := Settings{Sound: true, SoundOnWaiting: true, SoundOnError: true}
	if !soundForState(on, session.StateWaitingPermission) || !soundForState(on, session.StateError) {
		t.Fatal("everything should be audible by default")
	}
	quietWaiting := on
	quietWaiting.SoundOnWaiting = false
	if soundForState(quietWaiting, session.StateWaitingPermission) {
		t.Fatal("muting waiting did not silence the waiting tone")
	}
	if !soundForState(quietWaiting, session.StateError) {
		t.Fatal("muting waiting also silenced errors")
	}
	quietError := on
	quietError.SoundOnError = false
	if soundForState(quietError, session.StateError) {
		t.Fatal("muting errors did not silence the error tone")
	}
	master := on
	master.Sound = false
	if soundForState(master, session.StateWaitingPermission) || soundForState(master, session.StateError) {
		t.Fatal("the master switch must silence every event")
	}
}
