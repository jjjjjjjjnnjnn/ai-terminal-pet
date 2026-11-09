package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

func pulseKey(clock int64) renderedKey {
	return renderedKey{
		width: 84, height: 234, digest: 7, state: session.StateWaitingPermission,
		count: 2, expanded: false, pressed: false, hovered: false,
		tuck: 0, lang: i18n.LangZhCN, approveSec: -1, clock: clock,
	}
}

func TestPurePulseFrameOnlyReusesClockOnlyChanges(t *testing.T) {
	previous := pulseKey(100)
	current := pulseKey(200)
	if !purePulseFrame(previous, current, session.StateWaitingPermission, 0, 0) {
		t.Fatal("a clock-only frame should reuse the rendered bitmap")
	}
	if purePulseFrame(previous, previous, session.StateWaitingPermission, 0, 0) {
		t.Fatal("an identical frame must not be treated as a new pulse")
	}
	if purePulseFrame(renderedKey{}, current, session.StateWaitingPermission, 0, 0) {
		t.Fatal("the first frame has nothing to reuse")
	}

	mutations := map[string]func(*renderedKey){
		"digest":     func(k *renderedKey) { k.digest++ },
		"count":      func(k *renderedKey) { k.count++ },
		"expanded":   func(k *renderedKey) { k.expanded = true },
		"hovered":    func(k *renderedKey) { k.hovered = true },
		"pressed":    func(k *renderedKey) { k.pressed = true },
		"width":      func(k *renderedKey) { k.width++ },
		"height":     func(k *renderedKey) { k.height++ },
		"tuck":       func(k *renderedKey) { k.tuck = 500 },
		"lang":       func(k *renderedKey) { k.lang = i18n.LangEn },
		"approveSec": func(k *renderedKey) { k.approveSec = 3 },
	}
	for name, mutate := range mutations {
		changed := current
		mutate(&changed)
		if purePulseFrame(previous, changed, session.StateWaitingPermission, 0, 0) {
			t.Errorf("%s changed but the frame was still reused", name)
		}
	}

	if purePulseFrame(previous, current, session.StateWorking, 0, 0) {
		t.Fatal("a non-alert state must never take the pulse shortcut")
	}
	if purePulseFrame(previous, current, session.StateError, 0.4, 1) {
		t.Fatal("a frame mid-slide needs real geometry, not an alpha change")
	}
}

func TestPulseOpacityStaysWithinReadableBand(t *testing.T) {
	base := uint8(235)
	if got := pulseOpacity(base, session.StateIdle, time.Now()); got != base {
		t.Fatalf("idle opacity = %d, want the static %d", got, base)
	}
	seen := map[uint8]bool{}
	start := time.Unix(5000, 0)
	for step := 0; step < 40; step++ {
		got := pulseOpacity(base, session.StateWaitingPermission, start.Add(time.Duration(step)*50*time.Millisecond))
		if got < base-26 || got > base {
			t.Fatalf("opacity %d outside [%d,%d]", got, base-26, base)
		}
		seen[got] = true
	}
	if len(seen) < 3 {
		t.Fatalf("the breath did not vary: %v", seen)
	}

	// A very transparent panel must not be made brighter than requested.
	low := pulseOpacity(40, session.StateError, start.Add(37*time.Millisecond))
	if low > 40 {
		t.Fatalf("low opacity pulse = %d, want <= 40", low)
	}
}
