package ui

import (
	"testing"
	"time"
)

// The watchdog's false positives, pinned as a policy test.
//
// Every case below is one of the six entries in the machine's hang.log, or the
// rule that keeps them from coming back. The measurement behind it is in
// hang.go: the machine counted 394 minutes of standby as elapsed time since
// boot, and three dump headers land on the same second as a Kernel-Power 507
// ("system is exiting standby").
func TestHangDueIgnoresStandbyAndReportsRealStalls(t *testing.T) {
	cases := []struct {
		name          string
		elapsed       time.Duration
		sinceLastDump time.Duration
		want          bool
	}{
		{name: "awake 5s: too early", elapsed: 5 * time.Second, sinceLastDump: time.Hour},
		{name: "awake 6s: a stall", elapsed: 6 * time.Second, sinceLastDump: time.Hour, want: true},
		{name: "awake 40s: a stall", elapsed: 40 * time.Second, sinceLastDump: time.Hour, want: true},
		// A resume hands the UI thread messages that were queued before sleep.
		// The unbiased clock reports a few milliseconds; the wall clock would
		// have reported the whole sleep.
		{name: "just resumed: zero awake time", elapsed: 0, sinceLastDump: time.Hour},
		{name: "just resumed: 200 ms awake", elapsed: 200 * time.Millisecond, sinceLastDump: time.Hour},
		{name: "cooled down: one outage, one dump", elapsed: time.Minute, sinceLastDump: time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hangDue(tc.elapsed, tc.sinceLastDump); got != tc.want {
				t.Fatalf("hangDue(%v, %v) = %v, want %v", tc.elapsed, tc.sinceLastDump, got, tc.want)
			}
		})
	}
}

func TestAwakeElapsedReadsHundredNanosecondUnits(t *testing.T) {
	// 10,000,000 units of 100 ns is one second. Getting this factor wrong is
	// how a watchdog either never fires or fires every tick.
	if got := awakeElapsed(1000, 1000+10_000_000); got != time.Second {
		t.Fatalf("awakeElapsed = %v, want 1s", got)
	}
	if got := awakeElapsed(2000, 1000); got != 0 {
		t.Fatalf("a counter that went backwards produced %v, want 0", got)
	}
	if got := awakeElapsed(0, 0); got != 0 {
		t.Fatalf("a missing reading produced %v, want 0", got)
	}
}

func TestInterruptTicksIsAvailableOnThisWindows(t *testing.T) {
	// If this fails the watchdog silently stops reporting real stalls, which is
	// a worse outcome than a red test: the alternative clock is the one that
	// produced six false positives.
	first := interruptTicks()
	if first == 0 {
		t.Fatal("QueryUnbiasedInterruptTime returned nothing")
	}
	time.Sleep(20 * time.Millisecond)
	second := interruptTicks()
	if second <= first {
		t.Fatalf("the clock did not advance: %d then %d", first, second)
	}
	was := awakeElapsed(first, second)
	if was < 10*time.Millisecond || was > 200*time.Millisecond {
		t.Fatalf("a 20 ms sleep measured as %v", was)
	}
}
