package ui

// The copy receipt and the compact strip's two controls, pinned.
//
// Three claims are under test here, each of which has already been the wrong
// answer once in this project's history:
//
//   - the receipt is drawn, so it must be in the frame key, and being
//     time-driven it cannot ride on clock (the breathing path ignores clock
//     on purpose - that is exactly how 1.0.1's duration froze);
//   - the compact strip is one pill with two targets, so their rectangles must
//     not overlap at any zoom the panel can take;
//   - a receipt is a claim about a specific conversation, so it must not
//     survive that conversation leaving the row it was drawn on.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func toastFixture() []session.Session {
	return []session.Session{
		{Source: "claude-hook", SessionID: "a", Project: `C:\work\app-a`, State: session.StateWaitingPermission},
		{Source: "opencode", SessionID: "b", Project: `C:\work\app-b`, State: session.StateWorking},
		{Source: "opencode", SessionID: "c", Project: `C:\work\app-c`, State: session.StateIdle},
	}
}

func TestCopyToastIsDeadUntilSomethingIsCopied(t *testing.T) {
	var toast copyToast
	if toast.active(time.Unix(1000, 0)) {
		t.Fatal("a fresh receipt must be inert")
	}
	if p := toast.frame(time.Unix(1000, 0), toastFixture()); p.visible() {
		t.Fatal("an inert receipt must contribute nothing to the frame")
	}
}

// The receipt expires on its own, and the frame that removes it has to be a
// repaint: a key that never stopped carrying the pill would leave it on screen
// until some unrelated refresh.
func TestCopyToastEntersTheFrameKeyAndLeavesItAgain(t *testing.T) {
	items := toastFixture()
	start := time.Unix(1000, 0)
	var toast copyToast
	toast.show(1, session.Identity(items[1]), false, start)

	live := toast.frame(start.Add(100*time.Millisecond), items)
	if !live.visible() {
		t.Fatal("a live receipt must be visible")
	}
	key := renderedKey{}
	key.withToast(live)
	blank := renderedKey{}
	if key == blank {
		t.Fatal("the receipt must move the frame key")
	}
	if live.tick <= 0 {
		t.Fatal("a visible receipt needs a positive tick, or it cannot be told from an absent one")
	}

	// Same receipt, later instant: the tick is smaller, so the key differs
	// again - that is what makes the pill leave the screen without waiting
	// for a session update.
	older := toast.frame(start.Add(900*time.Millisecond), items)
	if !older.visible() || older.tick >= live.tick {
		t.Fatalf("tick did not advance: %d -> %d", live.tick, older.tick)
	}

	// Past its deadline nothing is drawn, and the key says so with tick 0.
	gone := toast.frame(start.Add(copyToastDuration), items)
	if gone.visible() {
		t.Fatal("an expired receipt is still visible")
	}
	expired := renderedKey{}
	expired.withToast(gone)
	if expired != blank {
		t.Fatal("an expired receipt must leave the key exactly as it found it")
	}
}

// Row 0 with a receipt and no receipt at all must be different keys: tick is
// the discriminator, which is why it is never 0 for a live pill (see frame).
func TestCopyToastRowZeroIsNotTheSameKeyAsNoToast(t *testing.T) {
	var key renderedKey
	key.withToast(panelToast{row: 0, tick: 5})
	if key == (renderedKey{}) {
		t.Fatal("a receipt on row 0 collapsed into the empty key")
	}
	var other renderedKey
	other.withToast(panelToast{row: 1, tick: 5})
	if key == other {
		t.Fatal("the row is part of the key")
	}
	var failed renderedKey
	failed.withToast(panelToast{row: 0, tick: 5, failed: true})
	if key == failed {
		t.Fatal("the failure flag changes what is drawn, so it changes the key")
	}
}

// A pill over a row whose session has been replaced claims "copied" about
// somebody else's project. The receipt dies instead.
func TestCopyToastDiesWhenItsRowChangesOwner(t *testing.T) {
	items := toastFixture()
	start := time.Unix(1000, 0)
	var toast copyToast
	toast.show(0, session.Identity(items[0]), false, start)

	if p := toast.frame(start.Add(time.Second), items); !p.visible() {
		t.Fatal("the receipt should be visible while its own session is on that row")
	}
	// The waiting session ended and the busy ones moved up: row 0 now belongs
	// to somebody else.
	shuffled := []session.Session{items[1], items[2]}
	if p := toast.frame(start.Add(time.Second), shuffled); p.visible() {
		t.Fatal("the receipt survived its session leaving the row")
	}
}

// The compact receipt follows the most urgent session, not a row index.
func TestCopyToastCompactFollowsTheMostUrgentSession(t *testing.T) {
	items := toastFixture()
	start := time.Unix(1000, 0)
	most, ok := mostUrgentJump(items)
	if !ok {
		t.Fatal("the fixture has a jumpable session")
	}
	var toast copyToast
	toast.show(toastCompactRow, session.Identity(most), false, start)
	p := toast.frame(start.Add(200*time.Millisecond), items)
	if !p.compact() {
		t.Fatalf("the compact receipt = %+v, want a compact toast", p)
	}
	// The strip's subject changed, so the receipt about the old subject goes.
	if toast.frame(start.Add(200*time.Millisecond), []session.Session{items[1]}).visible() {
		t.Fatal("a receipt whose session stopped being the subject must disappear")
	}
}

// One pill, two targets: the copy half and the jump half must never overlap,
// at any width the panel can actually take (65%..160% zoom plus DPI).
func TestCompactActionRectsDoNotOverlap(t *testing.T) {
	for _, width := range []int{37, 44, 56, 84, 120, 200, 300} {
		copyRect, jumpRect := compactActionRects(width, true)
		left, top, w, h := compactFocusRect(width)
		if jumpRect.left+jumpRect.width > left+w {
			t.Fatalf("width %d: jump half overflows the pill", width)
		}
		if jumpRect.top != top || jumpRect.height != h {
			t.Fatalf("width %d: jump half does not match the pill's vertical extent", width)
		}
		if copyRect.width == 0 {
			// Narrow strips keep the unsplit jump bar, starting at the pill.
			if jumpRect.left != left {
				t.Fatalf("width %d: unsplit jump half no longer starts at the pill", width)
			}
			continue
		}
		if copyRect.left+copyRect.width > jumpRect.left {
			t.Fatalf("width %d: copy half overlaps the jump half", width)
		}
		if copyRect.left != left {
			t.Fatalf("width %d: copy half is not flush with the pill's left edge", width)
		}
		if copyRect.top != top || copyRect.height != h {
			t.Fatalf("width %d: copy half does not match the pill's vertical extent", width)
		}
	}
}

// A tap in the copy half copies and does not jump; a tap in the jump half
// jumps and does not copy. Both come from the same rectangles the rasterizer
// uses, so "drawn there, clicked here" cannot happen.
func TestCompactStripHalvesAreDistinctTargets(t *testing.T) {
	width := compactWidth
	height := compactHeight
	copyRect, jumpRect := compactActionRects(width, true)
	if copyRect.width == 0 {
		t.Fatal("the shipped strip is wide enough for two controls")
	}
	x := copyRect.left + copyRect.width/2
	y := copyRect.top + copyRect.height/2
	if !compactCopyHit(x, y, width, height, true) {
		t.Fatal("the copy glyph is not a copy target")
	}
	if _, ok := focusHit(x, y, width, height, false, true, true, 3); ok {
		t.Fatal("a tap on the copy glyph also jumped")
	}
	if compactCopyHit(x, y, width, height, false) {
		t.Fatal("a strip with nothing to copy must not offer the control")
	}
	jx := jumpRect.left + jumpRect.width/2
	jy := jumpRect.top + jumpRect.height/2
	if compactCopyHit(jx, jy, width, height, true) {
		t.Fatal("the jump half is also a copy target")
	}
	if row, ok := focusHit(jx, jy, width, height, false, true, true, 3); !ok || row != -1 {
		t.Fatal("the jump half is not a jump target")
	}
}

// When the strip is too narrow to split, there is no copy control at all: two
// four-pixel targets would be worse than the single jump bar 1.0.4 shipped.
func TestCompactStripDropsTheCopyTargetWhenNarrow(t *testing.T) {
	for _, width := range []int{20, 24, 30} {
		copyRect, jumpRect := compactActionRects(width, true)
		if copyRect.width != 0 {
			t.Fatalf("width %d: a copy target was drawn in %d units", width, copyRect.width)
		}
		if jumpRect.width <= 0 {
			t.Fatalf("width %d: the jump bar disappeared", width)
		}
		if compactCopyHit(jumpRect.left+1, jumpRect.top+1, width, 200, true) {
			t.Fatalf("width %d: the whole strip must stay the jump target", width)
		}
	}
}

// The compact copy and the compact jump must always address the same
// conversation: two notions of "the current session" in one 48-unit pill is
// indistinguishable from a bug.
func TestCompactCopyTextUsesTheJumpTarget(t *testing.T) {
	items := toastFixture()
	most, ok := mostUrgentJump(items)
	if !ok {
		t.Fatal("the fixture has a jumpable session")
	}
	got, copyable := compactCopyText(items)
	if !copyable || got != most.Project {
		t.Fatalf("compactCopyText = (%q,%v), want the jump target's %q", got, copyable, most.Project)
	}
	if _, copyable := compactCopyText(nil); copyable {
		t.Fatal("no sessions means nothing to copy")
	}
	if _, copyable := compactCopyText([]session.Session{{Source: "opencode", SessionID: "z", State: session.StateWorking}}); copyable {
		t.Fatal("a strip with no jumpable session has no copy control to show")
	}
	blank := most
	blank.Project = "  "
	if _, copyable := compactCopyText([]session.Session{blank}); copyable {
		t.Fatal("an empty project is not a copy target")
	}
}
