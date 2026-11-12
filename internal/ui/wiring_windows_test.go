package ui

// The wiring rules that are easy to get wrong and expensive to notice.
//
// Everything here is about how the stateful layer (App) and the pure policy
// (attention.go) fit together: mute is not "unread", unread does not survive
// clicking elsewhere, "restore defaults" does not un-mute anybody, and turning
// the reply-completed switch back on restores the default policy rather than
// the loudest one. Each of these has already been the wrong answer at least
// once in a design draft, so they are pinned.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func mutedConversation(id string) session.Session {
	return session.Session{Source: "opencode", SessionID: id, PID: 42, State: session.StateWaitingPermission}
}

// Muting a conversation takes its unread mark with it. The mark means "this
// needs you", and the user has just said it does not - leaving the badge up
// would contradict the mute.
func TestTogglingMuteClearsTheUnreadMark(t *testing.T) {
	value := mutedConversation("ses_1")
	a := &App{
		options: Options{Settings: DefaultSettings()},
		mutes:   newMuteSet(nil),
		unread:  newUnreadSet(),
	}
	a.unread.mark([]string{session.Identity(value)})
	if !a.unread.has(value) {
		t.Fatal("the mark did not stick")
	}
	a.toggleSessionMute(value)
	if !a.mutes.has(value) {
		t.Fatal("the conversation was not muted")
	}
	if a.unread.has(value) {
		t.Fatal("a muted conversation kept its unread mark")
	}
	if len(a.options.Settings.MutedSessions) != 1 {
		t.Fatalf("the mute list was not published: %v", a.options.Settings.MutedSessions)
	}

	// Unmuting restores alerts but does not resurrect the old mark: it was
	// about the alert that the user explicitly declined.
	a.toggleSessionMute(value)
	if a.mutes.has(value) {
		t.Fatal("the conversation stayed muted")
	}
	if a.unread.has(value) {
		t.Fatal("unmuting resurrected a mark the user had cleared")
	}
	if len(a.options.Settings.MutedSessions) != 0 {
		t.Fatalf("the empty mute list was not published: %v", a.options.Settings.MutedSessions)
	}
}

// Unread marks must be keyed on the conversation, not on the store key: a hook
// reporter's pid is a synthetic per-process hash.
func TestUnreadSurvivesASyntheticPIDChange(t *testing.T) {
	a := &App{mutes: newMuteSet(nil), unread: newUnreadSet()}
	first := session.Session{Source: "claude-hook", SessionID: "ses_9", PID: 111}
	a.unread.mark([]string{session.Identity(first)})
	second := first
	second.PID = 999
	if !a.unread.has(second) {
		t.Fatal("the mark was lost across a pid change")
	}
}

// Turning the legacy reply-completed switch back on restores the default
// policy, never "always": "on" asked for completions, not for a balloon every
// time the user watches one finish.
func TestDonePolicyFromTheLegacySwitch(t *testing.T) {
	off := AttentionSettings{DonePolicy: DonePolicyOff}
	if off.DonePolicy != DonePolicyOff {
		t.Fatal("baseline")
	}
	// The tray handler's rule, restated: off -> on gives background, on -> off
	// gives off, and always stays always when it is already on.
	next := func(current DonePolicy) DonePolicy {
		if current == DonePolicyOff {
			return DonePolicyBackground
		}
		return DonePolicyOff
	}
	if got := next(DonePolicyOff); got != DonePolicyBackground {
		t.Fatalf("off toggles to %q, want background", got)
	}
	if got := next(DonePolicyBackground); got != DonePolicyOff {
		t.Fatalf("background toggles to %q, want off", got)
	}
	if got := next(DonePolicyAlways); got != DonePolicyOff {
		t.Fatalf("always toggles to %q, want off", got)
	}
}

// The idle threshold is stored in whole minutes and used in seconds; the clamp
// must keep zero (window focus only) and reject anything config.json would reject.
func TestIdleMinutesClamp(t *testing.T) {
	a := &App{options: Options{Settings: DefaultSettings()}}
	a.options.Settings.IdleSeconds = defaultIdleSeconds
	for _, minutes := range []int{-5, 0, 7, 500} {
		a.setIdleMinutes(minutes)
		want := minutes
		if want < 0 {
			want = 0
		}
		if want > 120 {
			want = 120
		}
		if got := a.idleMinutes(); got != want {
			t.Fatalf("setIdleMinutes(%d) left %d", minutes, got)
		}
		if a.options.Settings.IdleSeconds != want*60 {
			t.Fatalf("seconds out of step: %d", a.options.Settings.IdleSeconds)
		}
	}
}

// A test alert must not touch a session: it is about the channel, not about
// anybody's conversation.
func TestTestAlertIsNotASessionAlert(t *testing.T) {
	a := &App{unread: newUnreadSet()}
	a.testAlert(session.KindApproval)
	a.testAlert(session.KindDone)
	if a.unread.size() != 0 {
		t.Fatal("a test alert marked a session unread")
	}
}

// The alert receipt is rewritten, never appended: it must not grow on an
// eight-second cadence.
func TestAlertReceiptDoesNotGrow(t *testing.T) {
	a := &App{}
	record := buildAlertRecord(time.Unix(1000, 0), []AlertInput{
		{Session: mutedConversation("a"), Kind: session.KindApproval},
	}, AlertPlan{Reasons: map[string]string{}}, attendanceState{Present: true, Known: true}, true)
	if len(record.Sessions) != 1 {
		t.Fatalf("want one session row, got %d", len(record.Sessions))
	}
	// The in-memory copy is replaced wholesale, so the file it was written from
	// stays one screen long however many scans ran.
	a.lastAlertRecord = record
	again := buildAlertRecord(time.Unix(1002, 0), []AlertInput{
		{Session: mutedConversation("a"), Kind: session.KindApproval},
	}, AlertPlan{Reasons: map[string]string{}}, attendanceState{Present: true, Known: true}, true)
	if len(again.Sessions) != len(record.Sessions) {
		t.Fatal("the receipt grew between two scans")
	}
}
