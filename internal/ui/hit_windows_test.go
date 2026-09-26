package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestClassifyHit(t *testing.T) {
	// 4x3 canvas: opaque block at x in [1,3), y in [1,2).
	hit := make([]byte, 4*3)
	hit[1*4+1] = 255
	hit[1*4+2] = 19 // below threshold
	if !classifyHit(1, 1, 4, 3, hit) {
		t.Fatal("opaque pixel must hit")
	}
	if classifyHit(2, 1, 4, 3, hit) {
		t.Fatal("sub-threshold alpha must miss")
	}
	if classifyHit(0, 0, 4, 3, hit) {
		t.Fatal("transparent pixel must miss")
	}
	// Regression: X must be checked against width, not len(hit).
	// Window wider than canvas (zoom-out/DPI-down) used to escape the
	// slice (Y*width+X >= len) and panic the process with no error.log.
	wide, high := 28, 133
	wideHit := make([]byte, wide*high)
	if classifyHit(39, 132, wide, high, wideHit) {
		t.Fatal("out-of-canvas X must miss, not panic")
	}
	if classifyHit(39, 170, wide, high, wideHit) {
		t.Fatal("out-of-canvas corner must miss, not panic")
	}
	if classifyHit(-1, 0, wide, high, wideHit) || classifyHit(0, -1, wide, high, wideHit) {
		t.Fatal("negative coords must miss")
	}
	if classifyHit(0, 0, 0, 0, nil) || classifyHit(0, 0, 4, 3, nil) {
		t.Fatal("empty geometry must miss")
	}
}

// TestExpandedCloseButtonRect keeps the drawn close button inside the header:
// it must not escape the panel and must not run into the state pill, which
// ends 40 units from the right edge.
func TestExpandedCloseButtonRect(t *testing.T) {
	for _, width := range []int{200, 220, 250, 300} {
		left, top, size := expandedCloseButtonRect(width)
		if size < 1 {
			t.Fatalf("width %d: size must be positive, got %d", width, size)
		}
		if left < 0 || left+size > width {
			t.Fatalf("width %d: button %d..%d escapes the panel", width, left, left+size)
		}
		unit := float64(width) / expandedWidth
		pillRight := width - int(40*unit+0.5)
		if left < pillRight {
			t.Fatalf("width %d: button starts at %d, overlapping the state pill which ends at %d",
				width, left, pillRight)
		}
		if top < 0 || top+size > int(float64(expandedHeaderHeight)*unit+0.5) {
			t.Fatalf("width %d: button %d..%d escapes the header", width, top, top+size)
		}
	}
}

// TestDecideRelease locks the rule that made dragging feel broken: a press
// that actually travelled must move the window only, never toggle it.
func TestDecideRelease(t *testing.T) {
	cases := []struct {
		name          string
		active, moved bool
		want          releaseAction
	}{
		{name: "drag that moved does nothing", active: true, moved: true, want: releaseNothing},
		{name: "tap toggles", active: true, moved: false, want: releaseToggle},
		{name: "inactive tap does nothing", active: false, moved: false, want: releaseNothing},
		{name: "inactive drag does nothing", active: false, moved: true, want: releaseNothing},
	}
	for _, tc := range cases {
		if got := decideRelease(tc.active, tc.moved); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

// TestExpandedPanelHeight pins the sizing rule that replaced the fixed
// two-row panel: the height has to follow the session count, but stay bounded
// so a burst of sessions cannot produce a window taller than the screen.
func TestExpandedPanelHeight(t *testing.T) {
	if got := expandedPanelHeight(0); got != expandedHeaderHeight+expandedSummaryHeight+expandedEmptyHeight+expandedFooterHeight {
		t.Fatalf("empty panel height = %d", got)
	}
	previous := 0
	for count := 1; count <= expandedMaxRows; count++ {
		height := expandedPanelHeight(count)
		if height <= previous {
			t.Fatalf("height must grow with the session count: %d sessions -> %d", count, height)
		}
		previous = height
	}
	// Past the visible-row cap only the overflow line is added.
	capped := expandedPanelHeight(expandedMaxRows)
	if got := expandedPanelHeight(expandedMaxRows + 1); got != capped+expandedOverflowHeight {
		t.Fatalf("overflow height = %d, want %d", got, capped+expandedOverflowHeight)
	}
	if got := expandedPanelHeight(500); got != capped+expandedOverflowHeight {
		t.Fatalf("height must not grow without bound, got %d", got)
	}
	if expandedPanelHeight(500) <= compactHeight {
		t.Fatal("expanded panel must be taller than the compact strip")
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		ago  time.Duration
		want string
	}{
		{ago: 0, want: "刚刚"},
		{ago: 30 * time.Second, want: "刚刚"},
		{ago: 60 * time.Second, want: "1 分钟前"},
		{ago: 9 * time.Minute, want: "9 分钟前"},
		{ago: 59 * time.Minute, want: "59 分钟前"},
		{ago: 90 * time.Minute, want: "1 小时前"},
		{ago: 5 * time.Hour, want: "5 小时前"},
	}
	for _, tc := range cases {
		if got := relativeTime(now.Add(-tc.ago), now); got != tc.want {
			t.Fatalf("%v ago: got %q want %q", tc.ago, got, tc.want)
		}
	}
	// A clock skew that puts updatedAt slightly ahead must not render
	// "0 分钟前" or a negative count.
	if got := relativeTime(now.Add(time.Minute), now); got != "刚刚" {
		t.Fatalf("future timestamp: got %q", got)
	}
	if got := relativeTime(time.Time{}, now); got != "" {
		t.Fatalf("zero timestamp: got %q", got)
	}
}

// TestSummaryLineSeparatesRunningFromWaiting pins the fix for the metric rows
// that double-counted: a session waiting on the user is not "running", and
// showing it in both counters made the census add up to more than the total.
func TestSummaryLineSeparatesRunningFromWaiting(t *testing.T) {
	items := []session.Session{
		{State: session.StateWaitingPermission},
		{State: session.StateWorking},
		{State: session.StateThinking},
		{State: session.StateIdle},
	}
	if got := summaryLine(items); got != "4 个会话 · 2 运行 · 1 等待" {
		t.Fatalf("summary = %q", got)
	}
	if got := summaryLine(nil); got != "暂无会话" {
		t.Fatalf("empty summary = %q", got)
	}
	if got := summaryLine([]session.Session{{State: session.StateIdle}}); got != "1 个会话" {
		t.Fatalf("idle-only summary = %q", got)
	}
}

// TestCompactSubjectPrefersTheUrgentSession locks the point of the strip:
// when something is waiting on the user, the strip names it rather than
// showing an arbitrary session.
func TestCompactSubjectPrefersTheUrgentSession(t *testing.T) {
	items := []session.Session{
		{State: session.StateWaitingPermission, Project: `C:\work\blocked`},
		{State: session.StateWorking, Project: `C:\work\busy`},
	}
	label, _ := compactSubject(items, session.Aggregate(items))
	if label != "blocked" {
		t.Fatalf("subject = %q, want the waiting session", label)
	}
	label, _ = compactSubject(items, session.StateWorking)
	if label != "blocked" {
		t.Fatalf("subject = %q, want the highest-priority session", label)
	}
	label, _ = compactSubject(nil, session.StateIdle)
	if label != "等待 AI CLI" {
		t.Fatalf("empty subject = %q", label)
	}
}
