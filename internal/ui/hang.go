package ui

// The hang watchdog's policy, and the half of its clock that has nothing to do
// with the operating system.
//
// hang.log carried six "stalls" since 2026-10-06, and they were not stalls: the
// machine had just come out of standby, and the watchdog measured the gap with
// time.Now, which keeps counting while the machine sleeps. Measured on the
// machine that produced them (handover §54.1): Kernel-Power 507 ("system is
// exiting standby") at 06:59:26.560 on 11-06, 08:47:22.789 on 11-13 and
// 03:07:53.061 on 11-14, against dump headers stamped in the same second, three
// for three - and the goroutine dump showed the UI thread sitting in GetMessage,
// not inside our code.
//
//	GetTickCount64             102.73 h since boot   (counts standby)
//	QueryUnbiasedInterruptTime  96.17 h since boot   (counts awake time only)
//
// So the fix is the clock, not the threshold: measure "how long has the UI
// thread failed to dispatch *while the machine was running*". Suspend time must
// not be able to reach that number, and that has to be a property of the clock
// rather than a filter somebody has to remember to apply.
//
// The platform half - what "awake time" means on each OS - lives in
// awclock_<goos>.go. This file keeps the arithmetic and the policy, so both
// can be tested without a window, a clock or a sleep, and so the darwin/linux
// build can type-check the rule before its clock exists (handover §55.6 step 1).

import (
	"time"
)

const (
	// hangStallThreshold is how long the UI thread may go without dispatching
	// before the watchdog believes it. Six seconds of *awake* time: a resume
	// from standby no longer counts, a real block still does.
	hangStallThreshold = 6 * time.Second
	// hangDumpCooldown and hangMaxDumps exist so one outage produces one
	// diagnosis. The previous unbounded, 60s-throttled version appended 370 KB
	// of identical stacks over a single 80-minute outage without ever naming
	// the culprit.
	hangDumpCooldown = 2 * time.Minute
	hangMaxDumps     = 3
	hangLogMaxBytes  = 256 << 10
)

// awakeElapsed converts two clock readings into elapsed awake time.
//
// A reading that did not advance yields zero rather than a wrapped duration
// measured in centuries: the counter cannot decrease, so this only guards a
// stale or missing value.
func awakeElapsed(last, now uint64) time.Duration {
	if now <= last {
		return 0
	}
	// One unit is 100 ns on every platform's clock (see awclock_*.go), and
	// time.Duration counts nanoseconds.
	return time.Duration(now-last) * 100
}

// hangDue is the watchdog's policy, kept pure so the rule can be tested without
// a window, a clock or a sleep.
func hangDue(elapsed, sinceLastDump time.Duration) bool {
	return elapsed >= hangStallThreshold && sinceLastDump >= hangDumpCooldown
}
