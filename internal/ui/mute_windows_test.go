package ui

// The mute and unread rules, pinned. Both sets are keyed on session.Identity,
// which excludes the pid - so the first test here is the one that stops every
// mute from evaporating at the next hook invocation.

import (
	"testing"

	"ai-terminal-pet/internal/session"
)

func conversation(source, id string, pid int) session.Session {
	return session.Session{Source: source, SessionID: id, PID: pid}
}

func TestMuteSetSurvivesASyntheticPIDChange(t *testing.T) {
	set := newMuteSet(nil)
	first := conversation("claude-hook", "ses_1", 111)
	if !set.toggle(first) {
		t.Fatal("toggling an unmuted session should mute it")
	}
	// The same conversation reported by a different hook process: the pid is
	// a synthetic per-process hash, so this is the same conversation.
	second := conversation("claude-hook", "ses_1", 999)
	if !set.has(second) {
		t.Fatal("a mute keyed on the pid would have been lost here")
	}
	if set.has(conversation("opencode", "ses_1", 999)) {
		t.Fatal("another terminal's session with the same id must stay unmuted")
	}
	if !set.toggle(second) {
		// toggle returns the new state; unmuting returns false.
		_ = 0
	}
	if set.has(first) {
		t.Fatal("toggling twice must unmute")
	}
}

func TestMuteSetIgnoresUnkeyableReports(t *testing.T) {
	set := newMuteSet(nil)
	if set.toggle(session.Session{Source: "claude-hook"}) {
		t.Fatal("a report with no session id cannot be muted")
	}
	if set.size() != 0 {
		t.Fatal("nothing should have been stored")
	}
	if set.has(session.Session{Source: "claude-hook"}) {
		t.Fatal("an unkeyable report was reported as muted")
	}
}

func TestMuteSetCarriesTheConfigList(t *testing.T) {
	set := newMuteSet([]string{
		"claude-hook\x00ses_1",
		"opencode\x00ses_2",
		"",
	})
	if set.size() != 2 {
		t.Fatalf("want 2 keys, got %d", set.size())
	}
	keys := set.list()
	if len(keys) != 2 || keys[0] != "claude-hook\x00ses_1" || keys[1] != "opencode\x00ses_2" {
		t.Fatalf("list must be stable and sorted, got %v", keys)
	}
	if removed := set.clear(); removed != 2 {
		t.Fatalf("clear should report 2 removals, got %d", removed)
	}
	if set.size() != 0 {
		t.Fatal("clear did not empty the set")
	}
}

func TestDescribeMuteKey(t *testing.T) {
	if got := describeMuteKey("claude-hook\x00ses_1"); got != "claude-hook · ses_1" {
		t.Fatalf("got %q", got)
	}
	// A key without the separator is shown as-is rather than dropped: the
	// settings window lists stored keys, and hiding one would be worse.
	if got := describeMuteKey("junk"); got != "junk" {
		t.Fatalf("got %q", got)
	}
}

func TestUnreadSetMarksAndClears(t *testing.T) {
	set := newUnreadSet()
	value := conversation("claude-hook", "ses_1", 111)
	set.mark([]string{session.Identity(value)})
	if !set.has(value) {
		t.Fatal("the mark did not stick")
	}
	// The same conversation from a new hook process still counts as read.
	report := conversation("claude-hook", "ses_1", 999)
	if !set.clearOne(report) {
		t.Fatal("clearing must find the mark through a pid change")
	}
	if set.has(value) {
		t.Fatal("the mark should be gone")
	}
	if set.clearOne(report) {
		t.Fatal("clearing twice should report no change")
	}
}

// A conversation that disappears from the store cannot leave a badge behind
// forever, and retain has to report the change so the panel repaints.
func TestUnreadSetRetainsOnlyLiveSessions(t *testing.T) {
	set := newUnreadSet()
	live := conversation("claude-hook", "ses_1", 1)
	gone := conversation("opencode", "ses_2", 2)
	set.mark([]string{session.Identity(live), session.Identity(gone)})
	if !set.retain([]session.Session{live}) {
		t.Fatal("retain should report that it removed a mark")
	}
	if set.size() != 1 || !set.has(live) || set.has(gone) {
		t.Fatalf("retain left %d marks: %v", set.size(), set.keys)
	}
	if set.retain([]session.Session{live}) {
		t.Fatal("retaining the same list twice must be a no-op")
	}
}

// The frame key must change when a mark or a mute appears or disappears, or
// the "same key, same bitmap" guard swallows the change - the exact bug that
// froze the wait duration in 1.0.1 and the theme in 1.0.2.
func TestUnreadAndMutedBitsEnterTheFrameKey(t *testing.T) {
	items := []session.Session{
		conversation("claude-hook", "a", 1),
		conversation("opencode", "b", 2),
		conversation("antigravity", "c", 3),
	}
	unread := newUnreadSet()
	mutes := newMuteSet(nil)
	if got := unreadBits(items, unread); got != 0 {
		t.Fatalf("nothing marked, want 0 bits, got %d", got)
	}
	if got := mutedBits(items, mutes); got != 0 {
		t.Fatalf("nothing muted, want 0 bits, got %d", got)
	}

	unread.mark([]string{session.Identity(items[1])})
	got := unreadBits(items, unread)
	if got != 1<<1 {
		t.Fatalf("the second row should be bit 1, got %#b", got)
	}

	mutes.set(items[2], true)
	if got := mutedBits(items, mutes); got != 1<<2 {
		t.Fatalf("the third row should be bit 2, got %#b", got)
	}

	// A mark on a conversation that is not in the drawn list must not bleed
	// into a different row's bit.
	unread.mark([]string{session.Identity(conversation("ghost", "x", 9))})
	if got := unreadBits(items, unread); got != 1<<1 {
		t.Fatalf("an off-list mark changed the row bits: %#b", got)
	}
}
