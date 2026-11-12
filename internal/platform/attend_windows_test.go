package platform

// The two facts the alert policy asks about the desktop, pinned where they can
// be pinned without a desktop. The probe at the bottom is the only part that
// needs a real session, and it Skips unless PET_ATTEND_PROBE is set - the same
// convention as PET_PRESENCE_PROBE and PET_SOUND_PROBE.

import (
	"os"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// IdleAtLeast is the one judgement that turns a duration into a decision, and
// every branch of it has a consequence: a false "still here" silences an alert
// the user needed.
func TestIdleAtLeast(t *testing.T) {
	sample := ForegroundSample{Idle: 5 * time.Minute, IdleKnown: true}
	if !sample.IdleAtLeast(3 * time.Minute) {
		t.Fatal("five idle minutes is past a three-minute threshold")
	}
	if sample.IdleAtLeast(5*time.Minute + time.Second) {
		t.Fatal("the threshold is inclusive at the boundary, not beyond it")
	}
	if !sample.IdleAtLeast(5 * time.Minute) {
		t.Fatal("exactly at the threshold counts as away")
	}

	// Zero disables the idle half: "window focus only" must not become "always
	// away", which would announce on a machine that was simply never touched.
	if sample.IdleAtLeast(0) {
		t.Fatal("a zero threshold means window focus only, not idle-forever")
	}

	// An unknown idle time is not evidence of presence. The policy reads it as
	// unknown and announces; this function must not claim otherwise.
	unknown := ForegroundSample{IdleKnown: false}
	if unknown.IdleAtLeast(time.Second) {
		t.Fatal("an unknown sample must never report the user as away")
	}
}

// Attends must be conservative. A zero foreground handle, or a session whose
// window cannot be resolved, both mean "not attending" - which the policy reads
// as "announce", so the cheap answer here is also the safe one.
func TestAttendsIsFalseWithoutEvidence(t *testing.T) {
	presence := fakePresence(time.Minute, clockPointer(), map[int]focusProcess{}, nil)
	busy := session.Session{Source: "opencode", Project: `C:\work\project`}
	if presence.Attends(busy, 0) {
		t.Fatal("a zero foreground handle cannot attend anything")
	}
	if presence.Attends(session.Session{Source: "opencode", Project: `C:\work\project`}, 0x9999) {
		t.Fatal("an unmatched window handle must not attend")
	}
}

// Attends is true when the foreground handle is the window that owns the
// session, which is the same ranking jump-to-terminal uses. This is the pairing
// that makes "quiet in the current session" possible at all.
func TestAttendsMatchesTheOwningWindow(t *testing.T) {
	procs := map[int]focusProcess{
		42: {parent: 1, name: "WindowsTerminal.exe"},
		1:  {parent: 0, name: "explorer.exe"},
	}
	window := terminalWindow(0x1001, 42, "OC | project")
	presence := fakePresence(time.Minute, clockPointer(), procs, []focusWindow{window})
	value := session.Session{
		Source: "opencode", PID: 42, ParentPID: 1,
		Title: "OC | project", Project: `C:\work\project`,
	}
	if !presence.Attends(value, window.handle) {
		t.Fatal("the foreground window owns this session and must count as attended")
	}
	// A different window in front is not this session.
	other := terminalWindow(0x1002, 7, "claude code - other")
	if presence.Attends(value, other.handle) {
		t.Fatal("another terminal's window must not count as attending")
	}
}

// clockPointer gives fakePresence the fixed clock it dereferences: a shared
// helper rather than nil, which would panic on the first view().
func clockPointer() *time.Time {
	fixed := time.Unix(5000, 0)
	return &fixed
}

func TestAttendanceProbe(t *testing.T) {
	if os.Getenv("PET_ATTEND_PROBE") == "" {
		t.Skip("set PET_ATTEND_PROBE=1 to sample the real desktop")
	}
	sample := SampleForeground()
	t.Logf("foreground=%#x known=%v idle=%s idleKnown=%v", sample.Handle, sample.Known, sample.Idle, sample.IdleKnown)
	if !sample.Known {
		t.Fatal("GetForegroundWindow answered nothing on a live desktop")
	}
	if !sample.IdleKnown {
		t.Fatal("GetLastInputInfo answered nothing on a live desktop")
	}
	if sample.Idle < 0 || sample.Idle > 7*24*time.Hour {
		t.Fatalf("idle duration is implausible: %s", sample.Idle)
	}
}
