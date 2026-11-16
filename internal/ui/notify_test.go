package ui

// What the overlay says, and what the receipt admits. Delivery differs per
// platform, so these tests stay on the part every platform shares: the copy,
// its urgency, and the AppleScript escaping the macOS delivery depends on.

import (
	"strings"
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

func TestComposeAlertWaitingNamesTheProject(t *testing.T) {
	target := session.Session{
		Source:    "claude-hook",
		SessionID: "ses_1",
		Project:   "ai-terminal-pet",
		State:     session.StateWaitingPermission,
		Reason:    "permission",
	}
	got := composeAlert(target, i18n.LangEn)
	if want := i18n.T(i18n.LangEn, i18n.KeyBalloonWaitTitle); got.Title != want {
		t.Fatalf("title = %q, want %q", got.Title, want)
	}
	if got.Level != notifyWarning {
		t.Fatalf("level = %v, want warning", got.Level)
	}
	if !strings.Contains(got.Body, "ai-terminal-pet") {
		t.Fatalf("body = %q, want it to name the project", got.Body)
	}
}

func TestComposeAlertErrorIsTheOnlyCriticalOne(t *testing.T) {
	target := session.Session{
		Source:    "opencode",
		SessionID: "ses_2",
		Project:   "open-notebook",
		State:     session.StateError,
	}
	got := composeAlert(target, i18n.LangEn)
	if want := i18n.T(i18n.LangEn, i18n.KeyBalloonErrTitle); got.Title != want {
		t.Fatalf("title = %q, want %q", got.Title, want)
	}
	if got.Level != notifyError {
		t.Fatalf("level = %v, want error", got.Level)
	}
}

func TestComposeAlertFinishedUsesTheDoneCopy(t *testing.T) {
	target := session.Session{
		Source:    "antigravity-hook",
		SessionID: "ses_3",
		Project:   "game-dev",
		State:     session.StateIdle,
	}
	got := composeAlert(target, i18n.LangEn)
	if want := i18n.T(i18n.LangEn, i18n.KeyBalloonDoneTitle); got.Title != want {
		t.Fatalf("title = %q, want %q", got.Title, want)
	}
	if got.Level != notifyInfo {
		t.Fatalf("level = %v, want info", got.Level)
	}
	if want := doneBalloonBody(target, i18n.LangEn); got.Body != want {
		t.Fatalf("body = %q, want %q", got.Body, want)
	}
}

func TestComposeAlertClosedIsAlsoACompletion(t *testing.T) {
	target := session.Session{Source: "opencode", SessionID: "ses_4", Project: "x", State: session.StateClosed}
	got := composeAlert(target, i18n.LangEn)
	if want := i18n.T(i18n.LangEn, i18n.KeyBalloonDoneTitle); got.Title != want {
		t.Fatalf("a closed session is a completion: title = %q, want %q", got.Title, want)
	}
}

func TestComposeJumpFailureIsTheSharedCopy(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangEn, i18n.LangZhCN} {
		got := composeJumpFailure(lang)
		if want := i18n.T(lang, i18n.KeyBalloonJumpTitle); got.Title != want {
			t.Fatalf("%s title = %q, want %q", lang, got.Title, want)
		}
		if want := i18n.T(lang, i18n.KeyBalloonJumpBody); got.Body != want {
			t.Fatalf("%s body = %q, want %q", lang, got.Body, want)
		}
		if got.Level != notifyWarning {
			t.Fatalf("%s level = %v, want warning", lang, got.Level)
		}
	}
}

func TestAppleScriptLiteralSurvivesAProjectName(t *testing.T) {
	cases := []struct{ in, want string }{
		{`plain`, `"plain"`},
		{`say "hi"`, `"say \"hi\""`},
		{`back\slash`, `"back\\slash"`},
		{"line\nbreak", `"line break"`},
		{"tab\there", `"tab here"`},
		{"carriage\rreturn", `"carriage return"`},
	}
	for _, tc := range cases {
		if got := appleScriptLiteral(tc.in); got != tc.want {
			t.Fatalf("appleScriptLiteral(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}
	// The escaping has to be complete: one unescaped quote and osascript sees a
	// different program than the one we wrote. A lone quote becomes " + \" + ".
	if got, want := appleScriptLiteral(`"`), `"\"`+`"`; got != want {
		t.Fatalf("a lone quote produced %q, want %q", got, want)
	}
}

// The receipt must not claim a balloon no surface could show.
func TestAlertReceiptRecordsAMissingNotificationSurface(t *testing.T) {
	original := alertSurfaceAvailable
	t.Cleanup(func() { alertSurfaceAvailable = original })

	target := session.Session{Source: "claude-hook", SessionID: "ses_1", Project: "p", State: session.StateWaitingPermission}
	inputs := []AlertInput{{Session: target, Kind: session.KindApproval}}
	plan := AlertPlan{Balloon: &target, Reasons: map[string]string{sessionAlertKey(target): decisionElsewhere}}
	at := time.Date(2026, 11, 16, 22, 0, 0, 0, time.UTC)

	alertSurfaceAvailable = func() bool { return true }
	shown := buildAlertRecord(at, inputs, plan, attendanceState{}, false)
	if shown.Balloon == "" || shown.Skipped != "" {
		t.Fatalf("with a surface the receipt should name the balloon: %+v", shown)
	}

	alertSurfaceAvailable = func() bool { return false }
	hidden := buildAlertRecord(at, inputs, plan, attendanceState{}, false)
	if hidden.Balloon != "" {
		t.Fatalf("without a surface the receipt must not name a balloon: %+v", hidden)
	}
	if hidden.Skipped != skipNoNotificationSurface {
		t.Fatalf("skipped = %q, want %q", hidden.Skipped, skipNoNotificationSurface)
	}
	if hidden.Tone != plan.Tone {
		t.Fatalf("the tone decision must survive: %q", hidden.Tone)
	}
}

func TestJoinSkippedKeepsBothReasons(t *testing.T) {
	if got := joinSkipped("", skipNoNotificationSurface); got != skipNoNotificationSurface {
		t.Fatalf("joinSkipped(\"\") = %q", got)
	}
	if got := joinSkipped(skipNoSessions, skipNoNotificationSurface); got != skipNoSessions+"+"+skipNoNotificationSurface {
		t.Fatalf("joinSkipped kept %q", got)
	}
}
