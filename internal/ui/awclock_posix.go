//go:build !windows

package ui

// The POSIX half of the watchdog clock, and the honest answer to the question
// the Windows clock answers exactly.
//
// §55.4 said "先测再选" for this and it is worth writing down why the choice is
// not obvious: on the machines this project targets, a suspend differs from a
// stall only in who was asleep.
//
//   - time.Now / time.Since counts suspend time, so a resume after six seconds
//     of sleep reproduces §54's false positive verbatim.
//   - Go's runtime already keeps a monotonic clock that does *not* count
//     suspend: on linux it is clock_gettime(CLOCK_MONOTONIC), on darwin
//     CLOCK_MONOTONIC_RAW (mach_absolute_time, which stops while the system
//     sleeps - mach_continuous_time is the one that keeps counting). Both are
//     what time.Since uses, so this platform gets the right semantics from
//     time.Duration arithmetic rather than from a platform-specific clock read.
//
// So the conversion here is "elapsed awake nanoseconds -> the 100 ns unit the
// shared arithmetic in hang.go speaks". Keeping the same unit is what lets one
// hangDue and one awakeElapsed serve three platforms, and it is why this file
// only has to translate.

import (
	"time"
)

// posixClockEpoch is the origin this platform's readings count from. It is
// initialised at package load, not on the first call: a lazily-set epoch makes
// the first reading exactly zero, and interruptTicks uses zero to mean "this
// clock is unavailable" (so the watchdog would stay quiet for the first scan).
var posixClockEpoch = time.Now()

// interruptTicks returns elapsed awake time since this package started, in
// 100 ns units, matching interruptTicks on Windows.
//
// time.Since is used deliberately, including its monotonic reading: see the
// header. A platform where the monotonic clock counted suspend time would need
// a different reader here, and that is the one thing a future port has to
// re-measure rather than copy.
func interruptTicks() uint64 {
	elapsed := time.Since(posixClockEpoch)
	if elapsed <= 0 {
		return 0
	}
	return uint64(elapsed / (100 * time.Nanosecond))
}
