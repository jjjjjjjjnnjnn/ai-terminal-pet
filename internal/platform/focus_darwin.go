//go:build darwin

package platform

import (
	"bytes"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"golang.org/x/sys/unix"
)

const (
	cgWindowListOnScreenOnly          uint32 = 1 << 0
	cgWindowListExcludeDesktopElem    uint32 = 1 << 4
	cgNullWindowID                    uint32 = 0
	cgEventSourceCombinedSessionState int32  = 0
	cgAnyInputEventType               uint32 = ^uint32(0)

	cfStringEncodingUTF8 uint32 = 0x08000100
	cfNumberSInt64Type   int32  = 4
	cfNumberDoubleType   int32  = 13

	nsAppActivateAllWindows        uintptr = 1 << 0
	nsAppActivateIgnoringOtherApps uintptr = 1 << 1
)

var (
	darwinFocusOnce sync.Once
	darwinFocusOK   bool

	cfRelease                        func(cf uintptr)
	cfGetTypeID                      func(cf uintptr) uintptr
	cfStringGetTypeID                func() uintptr
	cfNumberGetTypeID                func() uintptr
	cfArrayGetCount                  func(array uintptr) int
	cfArrayGetValueAtIndex           func(array uintptr, idx int) uintptr
	cfDictionaryGetValue             func(dict uintptr, key uintptr) uintptr
	cfStringCreateWithCString        func(alloc uintptr, cStr string, encoding uint32) uintptr
	cfStringGetLength                func(str uintptr) int
	cfStringGetMaximumSizeForEnc     func(length int, encoding uint32) int
	cfStringGetCString               func(str uintptr, buf *byte, bufSize int, encoding uint32) uint8
	cfNumberGetValue                 func(num uintptr, typ int32, value unsafe.Pointer) uint8
	cgWindowListCopyWindowInfo       func(option uint32, relativeTo uint32) uintptr
	cgEventSourceSecondsSinceLastEvt func(stateID int32, eventType uint32) float64

	cfKeyWindowNumber    uintptr
	cfKeyWindowOwnerPID  uintptr
	cfKeyWindowOwnerName uintptr
	cfKeyWindowName      uintptr
	cfKeyWindowLayer     uintptr
	cfKeyWindowAlpha     uintptr
)

func ensureDarwinFocusLoaded() bool {
	darwinFocusOnce.Do(func() {
		cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_GLOBAL|purego.RTLD_LAZY)
		if err != nil {
			return
		}
		cg, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_GLOBAL|purego.RTLD_LAZY)
		if err != nil {
			return
		}
		if _, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_GLOBAL|purego.RTLD_LAZY); err != nil {
			return
		}
		purego.RegisterLibFunc(&cfRelease, cf, "CFRelease")
		purego.RegisterLibFunc(&cfGetTypeID, cf, "CFGetTypeID")
		purego.RegisterLibFunc(&cfStringGetTypeID, cf, "CFStringGetTypeID")
		purego.RegisterLibFunc(&cfNumberGetTypeID, cf, "CFNumberGetTypeID")
		purego.RegisterLibFunc(&cfArrayGetCount, cf, "CFArrayGetCount")
		purego.RegisterLibFunc(&cfArrayGetValueAtIndex, cf, "CFArrayGetValueAtIndex")
		purego.RegisterLibFunc(&cfDictionaryGetValue, cf, "CFDictionaryGetValue")
		purego.RegisterLibFunc(&cfStringCreateWithCString, cf, "CFStringCreateWithCString")
		purego.RegisterLibFunc(&cfStringGetLength, cf, "CFStringGetLength")
		purego.RegisterLibFunc(&cfStringGetMaximumSizeForEnc, cf, "CFStringGetMaximumSizeForEncoding")
		purego.RegisterLibFunc(&cfStringGetCString, cf, "CFStringGetCString")
		purego.RegisterLibFunc(&cfNumberGetValue, cf, "CFNumberGetValue")
		purego.RegisterLibFunc(&cgWindowListCopyWindowInfo, cg, "CGWindowListCopyWindowInfo")
		purego.RegisterLibFunc(&cgEventSourceSecondsSinceLastEvt, cg, "CGEventSourceSecondsSinceLastEventType")

		cfKeyWindowNumber = cfStringCreateWithCString(0, "kCGWindowNumber", cfStringEncodingUTF8)
		cfKeyWindowOwnerPID = cfStringCreateWithCString(0, "kCGWindowOwnerPID", cfStringEncodingUTF8)
		cfKeyWindowOwnerName = cfStringCreateWithCString(0, "kCGWindowOwnerName", cfStringEncodingUTF8)
		cfKeyWindowName = cfStringCreateWithCString(0, "kCGWindowName", cfStringEncodingUTF8)
		cfKeyWindowLayer = cfStringCreateWithCString(0, "kCGWindowLayer", cfStringEncodingUTF8)
		cfKeyWindowAlpha = cfStringCreateWithCString(0, "kCGWindowAlpha", cfStringEncodingUTF8)
		darwinFocusOK = cfKeyWindowNumber != 0 && cfKeyWindowOwnerPID != 0
	})
	return darwinFocusOK
}

func snapshotProcesses() map[int]focusProcess {
	result := make(map[int]focusProcess, 128)
	list, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return result
	}
	for _, entry := range list {
		pid := int(entry.Proc.P_pid)
		if pid <= 0 {
			continue
		}
		comm := entry.Proc.P_comm[:]
		if nul := bytes.IndexByte(comm, 0); nul >= 0 {
			comm = comm[:nul]
		}
		result[pid] = focusProcess{
			parent: int(entry.Eproc.Ppid),
			name:   string(comm),
		}
	}
	return result
}

func cfDictInt64(dict, key uintptr) (int64, bool) {
	if dict == 0 || key == 0 {
		return 0, false
	}
	val := cfDictionaryGetValue(dict, key)
	if val == 0 || cfGetTypeID(val) != cfNumberGetTypeID() {
		return 0, false
	}
	var out int64
	if cfNumberGetValue(val, cfNumberSInt64Type, unsafe.Pointer(&out)) == 0 {
		return 0, false
	}
	return out, true
}

func cfDictFloat64(dict, key uintptr) (float64, bool) {
	if dict == 0 || key == 0 {
		return 0, false
	}
	val := cfDictionaryGetValue(dict, key)
	if val == 0 || cfGetTypeID(val) != cfNumberGetTypeID() {
		return 0, false
	}
	var out float64
	if cfNumberGetValue(val, cfNumberDoubleType, unsafe.Pointer(&out)) == 0 {
		return 0, false
	}
	return out, true
}

func cfDictString(dict, key uintptr) string {
	if dict == 0 || key == 0 {
		return ""
	}
	val := cfDictionaryGetValue(dict, key)
	if val == 0 || cfGetTypeID(val) != cfStringGetTypeID() {
		return ""
	}
	length := cfStringGetLength(val)
	if length <= 0 {
		return ""
	}
	maxSize := cfStringGetMaximumSizeForEnc(length, cfStringEncodingUTF8)
	if maxSize <= 0 {
		return ""
	}
	buf := make([]byte, maxSize+1)
	if cfStringGetCString(val, &buf[0], len(buf), cfStringEncodingUTF8) == 0 {
		return ""
	}
	if nul := bytes.IndexByte(buf, 0); nul >= 0 {
		buf = buf[:nul]
	}
	return string(buf)
}

func visibleWindows(processes map[int]focusProcess) []focusWindow {
	if !ensureDarwinFocusLoaded() {
		return nil
	}
	array := cgWindowListCopyWindowInfo(cgWindowListOnScreenOnly|cgWindowListExcludeDesktopElem, cgNullWindowID)
	if array == 0 {
		return nil
	}
	defer cfRelease(array)

	count := cfArrayGetCount(array)
	if count <= 0 {
		return nil
	}
	windows := make([]focusWindow, 0, count)
	for i := 0; i < count; i++ {
		dict := cfArrayGetValueAtIndex(array, i)
		if dict == 0 {
			continue
		}
		if layer, ok := cfDictInt64(dict, cfKeyWindowLayer); !ok || layer != 0 {
			continue
		}
		if alpha, ok := cfDictFloat64(dict, cfKeyWindowAlpha); ok && alpha <= 0 {
			continue
		}
		winNum, okNum := cfDictInt64(dict, cfKeyWindowNumber)
		ownerPID, okPID := cfDictInt64(dict, cfKeyWindowOwnerPID)
		if !okNum || !okPID || winNum <= 0 || ownerPID <= 0 {
			continue
		}
		ownerName := cfDictString(dict, cfKeyWindowOwnerName)
		title := cfDictString(dict, cfKeyWindowName)
		if title == "" {
			title = ownerName
		}
		exe := processes[int(ownerPID)].name
		if exe == "" {
			exe = ownerName
		}
		windows = append(windows, focusWindow{
			handle: uintptr(winNum),
			pid:    int(ownerPID),
			class:  ownerName,
			title:  title,
			exe:    exe,
		})
	}
	return windows
}

func frontmostAppPID() int32 {
	clsPool := objc.GetClass("NSAutoreleasePool")
	if clsPool != 0 {
		pool := objc.ID(clsPool).Send(objc.RegisterName("new"))
		if pool != 0 {
			defer pool.Send(objc.RegisterName("drain"))
		}
	}
	clsWS := objc.GetClass("NSWorkspace")
	if clsWS == 0 {
		return 0
	}
	ws := objc.ID(clsWS).Send(objc.RegisterName("sharedWorkspace"))
	if ws == 0 {
		return 0
	}
	app := ws.Send(objc.RegisterName("frontmostApplication"))
	if app == 0 {
		return 0
	}
	return objc.Send[int32](app, objc.RegisterName("processIdentifier"))
}

// SampleForeground reads the active window handle and macOS HID idle duration.
func SampleForeground() ForegroundSample {
	if !ensureDarwinFocusLoaded() {
		return ForegroundSample{}
	}
	sample := ForegroundSample{}

	if pid := frontmostAppPID(); pid > 0 {
		windows := visibleWindows(nil)
		for _, w := range windows {
			if w.pid == int(pid) && w.handle != 0 {
				sample.Handle = w.handle
				sample.Known = true
				break
			}
		}
	}

	secs := cgEventSourceSecondsSinceLastEvt(cgEventSourceCombinedSessionState, cgAnyInputEventType)
	if !math.IsNaN(secs) && !math.IsInf(secs, 0) && secs >= 0 && secs < 365*24*3600 {
		sample.Idle = time.Duration(secs * float64(time.Second))
		sample.IdleKnown = true
	}
	return sample
}

// ActivateTarget brings the macOS application owning handle forward on the UI thread.
func ActivateTarget(handle uintptr) bool {
	if handle == 0 || !ensureDarwinFocusLoaded() {
		return false
	}
	var targetPID int32
	for _, w := range visibleWindows(nil) {
		if w.handle == handle && w.pid > 0 {
			targetPID = int32(w.pid)
			break
		}
	}
	if targetPID <= 0 {
		return false
	}
	clsPool := objc.GetClass("NSAutoreleasePool")
	if clsPool != 0 {
		pool := objc.ID(clsPool).Send(objc.RegisterName("new"))
		if pool != 0 {
			defer pool.Send(objc.RegisterName("drain"))
		}
	}
	clsApp := objc.GetClass("NSRunningApplication")
	if clsApp == 0 {
		return false
	}
	app := objc.ID(clsApp).Send(objc.RegisterName("runningApplicationWithProcessIdentifier:"), targetPID)
	if app == 0 {
		return false
	}
	return objc.Send[bool](app, objc.RegisterName("activateWithOptions:"), nsAppActivateAllWindows|nsAppActivateIgnoringOtherApps)
}
