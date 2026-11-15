//go:build windows

package platform

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	attendUser32 = syscall.NewLazyDLL("user32.dll")
	attendKernel = syscall.NewLazyDLL("kernel32.dll")

	getLastInputInfoProc = attendUser32.NewProc("GetLastInputInfo")
	getTickCountProc     = attendKernel.NewProc("GetTickCount")
	foregroundWindowProc = attendUser32.NewProc("GetForegroundWindow")
)

type lastInputInfo struct {
	Size uint32
	Time uint32
}

// SampleForeground reads the foreground window handle and system input idle time.
func SampleForeground() ForegroundSample {
	sample := ForegroundSample{}
	if handle, _, _ := foregroundWindowProc.Call(); handle != 0 {
		sample.Handle = handle
		sample.Known = true
	}
	var info lastInputInfo
	info.Size = uint32(unsafe.Sizeof(info))
	if result, _, _ := getLastInputInfoProc.Call(uintptr(unsafe.Pointer(&info))); result != 0 {
		now, _, _ := getTickCountProc.Call()
		elapsed := uint32(uint32(now) - info.Time)
		sample.Idle = time.Duration(elapsed) * time.Millisecond
		sample.IdleKnown = true
	}
	return sample
}
