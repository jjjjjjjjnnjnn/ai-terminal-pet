package ui

// The hang watchdog's clock.
//
// hang.log has carried six "stalls" since 2026-10-06, and they were not stalls.
// The one thing they have in common is not a message, a window or a paint: it is
// that the machine had just come out of standby.
//
// Measured on the machine that produced them (2026-11-15): Kernel-Power 507
// ("system is exiting standby") at 06:59:26.560 on 11-06, 08:47:22.789 on 11-13
// and 03:07:53.061 on 11-14, against dump headers stamped 06:59:26, 08:47:22 and
// 03:07:53 - the same second, three for three. The goroutine dump agrees too: the
// UI thread is sitting in GetMessage, waiting for a message, not inside our code.
// The watchdog fired immediately because it measured the gap with time.Now, which
// keeps counting while the machine sleeps:
//
//	GetTickCount64             102.73 h since boot   (counts standby)
//	QueryUnbiasedInterruptTime  96.17 h since boot   (counts awake time only)
//
// So a resume after any sleep longer than six seconds looks exactly like a
// six-second freeze. That matters beyond tidiness: HANDOVER §51.3 listed "the
// message pump stalls" as a candidate for "the settings window goes away", and
// these dumps were the evidence offered for it. They were evidence of standby.
//
// The fix is the clock, not the threshold: QueryUnbiasedInterruptTime excludes
// the time the system spends in standby, so the watchdog measures "how long has
// the UI thread failed to dispatch while the machine was running". Suspend time
// cannot reach the number - that is a property of the API, not a filter this
// code has to remember to apply. The thrice-measured coincidence with resume is
// then a fact about the machine, not a guess: hang.log gains a line saying the
// previous dispatch was N seconds of awake time ago.

import (
	"syscall"
	"time"
	"unsafe"
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

var queryUnbiasedInterruptTime = syscall.NewLazyDLL("kernel32.dll").NewProc("QueryUnbiasedInterruptTime")

// interruptTicks reads the system's unbiased interrupt time, in 100 ns units.
//
// It has been documented since Windows 7, so a zero is treated as "this clock is
// unavailable" and the caller stays quiet rather than falling back to a clock
// that would resurrect the false positives this function exists to remove.
func interruptTicks() uint64 {
	var value uint64
	result, _, _ := queryUnbiasedInterruptTime.Call(uintptr(unsafe.Pointer(&value)))
	if result == 0 {
		return 0
	}
	return value
}

// awakeElapsed converts two interrupt-time readings into elapsed awake time.
//
// A reading that did not advance yields zero rather than a wrapped duration
// measured in centuries: the counter cannot decrease, so this only guards a
// stale or missing value.
func awakeElapsed(last, now uint64) time.Duration {
	if now <= last {
		return 0
	}
	// One unit is 100 ns, and time.Duration counts nanoseconds.
	return time.Duration(now-last) * 100
}

// hangDue is the watchdog's policy, kept pure so the rule can be tested without
// a window, a clock or a sleep.
func hangDue(elapsed, sinceLastDump time.Duration) bool {
	return elapsed >= hangStallThreshold && sinceLastDump >= hangDumpCooldown
}
