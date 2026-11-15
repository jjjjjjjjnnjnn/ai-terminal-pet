//go:build windows

package ui

// The Windows half of the watchdog clock: QueryUnbiasedInterruptTime, in 100 ns
// units since boot, excluding the time the system spent in standby.
//
// It is documented since Windows 7, so a zero return is treated as "this clock
// is unavailable" rather than as an error: the caller then stays quiet instead
// of falling back to a clock that would resurrect the false positives this
// function exists to remove (see hang.go for the three resume timestamps).

import (
	"syscall"
	"unsafe"
)

var queryUnbiasedInterruptTime = syscall.NewLazyDLL("kernel32.dll").NewProc("QueryUnbiasedInterruptTime")

// interruptTicks reads the system's unbiased interrupt time, in 100 ns units.
func interruptTicks() uint64 {
	var value uint64
	result, _, _ := queryUnbiasedInterruptTime.Call(uintptr(unsafe.Pointer(&value)))
	if result == 0 {
		return 0
	}
	return value
}
