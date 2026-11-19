package ui

// The alert surface: what the overlay says when it wants the user's attention,
// and how that sentence leaves the process.
//
// Windows answers with a tray balloon, Linux with the freedesktop notification
// daemon (through notify-send), macOS with osascript. All three take a title and
// a body, so the *text* lives here - untagged, tested on every platform - and
// only the delivery is per-platform (notify_windows.go, notify_linux.go,
// notify_darwin.go).
//
// The line protocol does not move: the alert policy still decides "one balloon
// about this session" (attention.go), and this file only says it in words.
//
// Nothing here runs per frame: calls come from alertOnTransitions, which is
// driven by PlanAlerts and rate-limited, and from a failed jump.

import (
	"strings"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

// notifyLevel is how urgent the text is, independent of the word a platform
// uses: Win32 picks NIIF_*, freedesktop picks low/normal/critical.
type notifyLevel int

const (
	notifyInfo notifyLevel = iota
	notifyWarning
	notifyError
)

// notification is one thing to say.
type notification struct {
	Title string
	Body  string
	Level notifyLevel
}

// composeAlert turns one session into alert copy.
//
// It used to be two branches inside app_windows.go (one for the finish
// balloon, one for waiting/error). The text was the only part that could work
// unchanged on the other platforms, so it moved here - one composition, three
// deliveries, and no chance of the three platforms drifting into three
// different sentences for the same event.
func composeAlert(target session.Session, lang i18n.Lang) notification {
	if target.State == session.StateIdle || target.State == session.StateClosed {
		return notification{
			Title: i18n.T(lang, i18n.KeyBalloonDoneTitle),
			Body:  doneBalloonBody(target, lang),
			Level: notifyInfo,
		}
	}
	title := i18n.T(lang, i18n.KeyBalloonWaitTitle)
	level := notifyWarning
	if target.State == session.StateError {
		title = i18n.T(lang, i18n.KeyBalloonErrTitle)
		level = notifyError
	}
	body := projectLabel(target, lang)
	if reason := reasonLabel(target, lang); reason != "" {
		body += " · " + reason
	}
	return notification{Title: title, Body: body, Level: level}
}

// composeJumpFailure is the "could not switch terminal" copy. Windows shows it
// as a balloon and the POSIX ports show it as a notification, so the sentence
// has to be built once.
func composeJumpFailure(lang i18n.Lang) notification {
	return notification{
		Title: i18n.T(lang, i18n.KeyBalloonJumpTitle),
		Body:  i18n.T(lang, i18n.KeyBalloonJumpBody),
		Level: notifyWarning,
	}
}

// composeJumpPartial is the "window came forward, conversation did not" copy.
// It is deliberately not a failure tone: the click did something useful, and
// the user only needs to know which half of the jump OpenCode can perform.
func composeJumpPartial(lang i18n.Lang) notification {
	return notification{
		Title: i18n.T(lang, i18n.KeyBalloonJumpPartialTitle),
		Body:  i18n.T(lang, i18n.KeyBalloonJumpPartialBody),
		Level: notifyInfo,
	}
}

// alertSurfaceAvailable reports whether this machine can show a notification at
// all. It is a variable rather than a direct call so the receipt's tests can pin
// both answers on a machine where the tool happens to be installed.
var alertSurfaceAvailable = notificationSurfaceAvailable

// appleScriptLiteral renders value as an AppleScript string literal.
//
// It lives in this untagged file on purpose: only notify_darwin.go calls it,
// but a quote or a backslash inside a project name must not end the literal
// early, and that rule deserves tests on machines that can run them. Newlines
// and tabs become spaces because "display notification" would turn them into a
// line break the notification daemon cannot lay out.
func appleScriptLiteral(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"', '\\':
			builder.WriteByte('\\')
			builder.WriteRune(r)
		case '\n', '\r', '\t':
			builder.WriteByte(' ')
		default:
			builder.WriteRune(r)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
