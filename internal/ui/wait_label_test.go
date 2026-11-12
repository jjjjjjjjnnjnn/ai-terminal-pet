package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

func waitSession(state session.State, ago time.Duration, now time.Time) session.Session {
	return session.Session{
		Source:    "claude-hook",
		SessionID: "ses-" + string(state) + "-" + ago.String(),
		Project:   `C:\work\blocked`,
		State:     state,
		UpdatedAt: now.Add(-ago),
	}
}

// TestCompactWaitOnlyCoversStatesThatWaitOnTheUser is the policy test: a busy
// panel has nothing to wait for, and an idle one has finished - showing a
// duration there would be a number that means nothing.
func TestCompactWaitOnlyCoversStatesThatWaitOnTheUser(t *testing.T) {
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	for _, state := range []session.State{session.StateWorking, session.StateToolUse, session.StateThinking, session.StateIdle, session.StateClosed} {
		if label, ok := compactWait([]session.Session{waitSession(state, 12*time.Minute, now)}, now, i18n.LangEn); ok {
			t.Fatalf("%s produced %q; only waiting_permission and error may show a duration", state, label)
		}
	}
	for _, state := range []session.State{session.StateWaitingPermission, session.StateError} {
		if _, ok := compactWait([]session.Session{waitSession(state, 12*time.Minute, now)}, now, i18n.LangEn); !ok {
			t.Fatalf("%s must produce a duration", state)
		}
	}
	if _, ok := compactWait(nil, now, i18n.LangEn); ok {
		t.Fatal("an empty panel must not produce a duration")
	}
}

// TestCompactWaitPrefersTheOldestWait pins the one judgement call in the feature:
// a permission prompt that has been blocking for 12 minutes outranks a build that
// failed 30 seconds ago, even though StateError sorts first in the store.
func TestCompactWaitPrefersTheOldestWait(t *testing.T) {
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	items := []session.Session{
		waitSession(session.StateError, 30*time.Second, now),
		waitSession(session.StateWaitingPermission, 12*time.Minute, now),
	}
	label, ok := compactWait(items, now, i18n.LangEn)
	if !ok || label != "12m" {
		t.Fatalf("label = %q (%v), want 12m from the older permission prompt", label, ok)
	}
	label, _ = compactWait(items, now, i18n.LangZhCN)
	if label != "12分" {
		t.Fatalf("zh label = %q, want 12分", label)
	}
}

// TestShortElapsedUsesTheSameBucketsAsRelativeTime keeps the two formatters from
// drifting: the expanded row and the compact strip must agree about what "1
// minute" means.
func TestShortElapsedUsesTheSameBucketsAsRelativeTime(t *testing.T) {
	cases := []struct {
		delta time.Duration
		zh    string
		en    string
	}{
		{0, "1分", "1m"},
		{10 * time.Second, "1分", "1m"},
		{59 * time.Second, "1分", "1m"},
		{time.Minute, "1分", "1m"},
		{9 * time.Minute, "9分", "9m"},
		{59 * time.Minute, "59分", "59m"},
		{90 * time.Minute, "1小时", "1h"},
		{5 * time.Hour, "5小时", "5h"},
		{20 * time.Hour, "20小时", "20h"},
		// The ceiling: the long form in the expanded list still prints the
		// exact number, so the strip may saturate.
		{500 * time.Hour, "99小时", "99h"},
		// Clock skew must not render a negative duration.
		{-time.Minute, "1分", "1m"},
	}
	for _, tc := range cases {
		if got := shortElapsed(tc.delta, i18n.LangZhCN); got != tc.zh {
			t.Fatalf("%v zh: got %q want %q", tc.delta, got, tc.zh)
		}
		if got := shortElapsed(tc.delta, i18n.LangEn); got != tc.en {
			t.Fatalf("%v en: got %q want %q", tc.delta, got, tc.en)
		}
	}
}

// TestWaitLabelFitsTheStripIsAMeasuredClaim measures every label the strip can
// draw with the shipped font, at the shipped size and box. The widest real case
// ("99小时") is 44 px in a 48 px box at 16-bold; this test fails if a later font
// or string change pushes it over, instead of relying on somebody noticing.
func TestWaitLabelFitsTheStripIsAMeasuredClaim(t *testing.T) {
	canvas, err := newCanvas(compactWidth, compactHeight)
	if err != nil {
		t.Fatalf("canvas: %v", err)
	}
	defer canvas.close()
	box := compactWidth - 8
	sizes := []int{16}
	for _, size := range sizes {
		for _, value := range []string{
			"1分", "9分", "59分", "1小时", "23小时", "99小时",
			"1m", "9m", "59m", "1h", "23h", "99h",
		} {
			if width := canvas.measure(value, size, 700); width > box {
				t.Fatalf("%q is %d px at %d-bold, box is %d", value, width, size, box)
			}
		}
	}
	// The count path still exists and must keep fitting too: the caption is
	// drawn at 9-bold ("3 sessions" / "3 个会话").
	for _, value := range []string{"3 sessions", "12 sessions", "3 个会话", "12 个会话"} {
		if width := canvas.measure(value, 9, 400); width > box {
			t.Fatalf("caption %q is %d px in a %d px box at 9-bold", value, width, box)
		}
	}
}

// TestFrameKeyAdvancesTheDurationPerMinute pins the bug this feature would
// otherwise have: purePulseFrame ignores clock on purpose, so a duration riding
// on clock would freeze while an alert breathes.
func TestFrameKeyAdvancesTheDurationPerMinute(t *testing.T) {
	// Start on a minute boundary: with an arbitrary offset, "20 seconds later"
	// can cross the boundary and the check would be wrong, not the code.
	start := time.Unix(1_000_020, 0)
	first := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 3, false, false, false, 0, true, start, i18n.LangEn, -1, themeSlate, false)
	sameBucket := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 3, false, false, false, 0, true, start.Add(20*time.Second), i18n.LangEn, -1, themeSlate, false)
	nextBucket := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 3, false, false, false, 0, true, start.Add(70*time.Second), i18n.LangEn, -1, themeSlate, false)
	if first.waitMinute != sameBucket.waitMinute {
		t.Fatalf("20 s later the minute bucket moved: %d vs %d", first.waitMinute, sameBucket.waitMinute)
	}
	if first.waitMinute == nextBucket.waitMinute {
		t.Fatal("70 s later the minute bucket must have advanced, or the strip would freeze on 12m")
	}
	// ... and the alert frame must still be reusable in between, or this fix
	// would cost a full GDI repaint per pulse tick.
	if !purePulseFrame(first, sameBucket, session.StateWaitingPermission, 0, 0) {
		t.Fatal("the pulse shortcut must still apply between minute boundaries")
	}
	if purePulseFrame(first, nextBucket, session.StateWaitingPermission, 0, 0) {
		t.Fatal("a new minute must force a real repaint")
	}
	// A busy panel shows no duration, so it must not carry a minute bucket at
	// all - otherwise every busy compact panel would repaint once a minute for
	// nothing.
	busy := frameKey(compactWidth, compactHeight, 7, session.StateWorking, 3, false, false, false, 0, false, start, i18n.LangEn, -1, themeSlate, false)
	if busy.waitMinute != 0 {
		t.Fatalf("busy compact frame has waitMinute %d, want 0", busy.waitMinute)
	}
	// The expanded panel keeps its own 15 s clock and must not gain this one.
	expanded := frameKey(expandedWidth, 300, 7, session.StateError, 3, true, false, false, 0, false, start, i18n.LangEn, -1, themeSlate, false)
	if expanded.waitMinute != 0 {
		t.Fatalf("expanded frame has waitMinute %d, want 0", expanded.waitMinute)
	}
}
