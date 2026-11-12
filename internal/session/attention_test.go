package session

import "testing"

func TestKindOfDerivesWhatIsBeingAsked(t *testing.T) {
	cases := []struct {
		name   string
		state  State
		reason string
		want   Kind
	}{
		// Approvals: the default for a blocked session, whatever the wording.
		{"permission", StateWaitingPermission, "permission", KindApproval},
		{"dangerous", StateWaitingPermission, "permission_dangerous", KindApproval},
		{"tool confirmation", StateWaitingPermission, "tool_confirmation_pending", KindApproval},
		{"empty reason still blocks", StateWaitingPermission, "", KindApproval},
		{"unrecognised wording still blocks", StateWaitingPermission, "?????", KindApproval},
		// Questions and plan reviews are worded differently by the reporters.
		{"question", StateWaitingPermission, "question", KindQuestion},
		{"ask user question", StateWaitingPermission, "ask_user_question", KindQuestion},
		{"plan review", StateWaitingPermission, "plan_review", KindPlan},
		{"plan", StateWaitingPermission, "plan", KindPlan},
		// Capitals do not change the answer.
		{"question uppercase", StateWaitingPermission, "QUESTION", KindQuestion},
		// Other states.
		{"error", StateError, "exit status 1", KindError},
		{"idle", StateIdle, "", KindDone},
		{"closed", StateClosed, "harness exited", KindDone},
		{"working", StateWorking, "", KindNone},
		{"thinking", StateThinking, "", KindNone},
		{"tool use", StateToolUse, "Bash", KindNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := KindOf(tc.state, tc.reason); got != tc.want {
				t.Fatalf("KindOf(%q, %q) = %q, want %q", tc.state, tc.reason, got, tc.want)
			}
		})
	}
}

func TestKindUrgent(t *testing.T) {
	for _, kind := range []Kind{KindApproval, KindQuestion, KindPlan, KindError} {
		if !kind.Urgent() {
			t.Errorf("%q should be urgent: the user has to act on it", kind)
		}
	}
	for _, kind := range []Kind{KindNone, KindDone} {
		if kind.Urgent() {
			t.Errorf("%q must not be urgent", kind)
		}
	}
}

// A waiting session must never be classified as "nothing", because that is the
// classification that silences the alert. This is the invariant the whole file
// exists for.
func TestWaitingPermissionIsNeverSilent(t *testing.T) {
	for _, reason := range []string{"", "permission", "question", "plan", "anything at all"} {
		if kind := KindOf(StateWaitingPermission, reason); kind == KindNone {
			t.Fatalf("waiting_permission with reason %q became KindNone", reason)
		}
	}
}

func TestKindOfSession(t *testing.T) {
	value := Session{State: StateWaitingPermission, Reason: "question"}
	if got := KindOfSession(value); got != KindQuestion {
		t.Fatalf("KindOfSession = %q, want %q", got, KindQuestion)
	}
}

// Identity is what a mute and an unread mark are remembered by, so it must
// ignore the pid: hook reporters hand out a synthetic pid per process, and a
// mute keyed on it would expire at the next event.
func TestIdentityIgnoresTheProcess(t *testing.T) {
	first := Session{Source: "claude-hook", SessionID: "ses_1", PID: 111}
	second := first
	second.PID = 999
	if Identity(first) != Identity(second) {
		t.Fatalf("the same conversation must key the same: %q vs %q", Identity(first), Identity(second))
	}
	other := first
	other.Source = "opencode"
	if Identity(other) == Identity(first) {
		t.Fatal("two terminals can share a session id and must not collide")
	}
	if got := Identity(Session{Source: "claude-hook"}); got != "" {
		t.Fatalf("a session without an id has no identity to store, got %q", got)
	}
	if Identity(Session{Source: "  claude-hook  ", SessionID: " ses_1 "}) != Identity(first) {
		t.Fatal("identity must trim both parts")
	}
}
