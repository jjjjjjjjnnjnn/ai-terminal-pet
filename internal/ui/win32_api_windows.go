package ui

import (
	"sync"
	"syscall"
	"unsafe"

	win "github.com/lxn/win"
)

const (
	ulwAlpha                = 0x00000002
	acSrcOver               = 0x00000000
	acSrcAlpha              = 0x00000001
	monitorDefaultToPrimary = 0x00000001
	monitorInfoPrimary      = 0x00000001

	wmDestroy         = 0x0002
	wmKillFocus       = 0x0008
	wmClose           = 0x0010
	wmEraseBackground = 0x0014
	wmNCHitTest       = 0x0084
	wmContextMenu     = 0x007B
	wmKeyDown         = 0x0100
	wmTimer           = 0x0113
	wmLButtonDown     = 0x0201
	wmLButtonUp       = 0x0202
	wmLButtonDblClk   = 0x0203
	wmMouseMove       = 0x0200
	wmMouseLeave      = 0x02A3
	wmMouseWheel      = 0x020A
	wmRButtonUp       = 0x0205
	wmApp             = 0x8000
	wmRefresh         = wmApp + 1
	wmTray            = wmApp + 2
	wmActivate        = wmApp + 3
	wmToggle          = wmApp + 4
	wmReset           = wmApp + 5
	// wmFocusResult reports a failed off-thread lookup (no matching window).
	wmFocusResult = wmApp + 6
	// wmFocusActivate carries the resolved HWND in wParam and asks the UI
	// thread to activate it. The UI thread owns the foreground right after
	// the click, so no AttachThreadInput / z-order tricks are needed.
	wmFocusActivate = wmApp + 7
	// focusPartialFlag is the lParam of wmFocusActivate: the window came
	// forward but its title did not name this conversation.
	focusPartialFlag = 1
	// wmSettings opens the settings window (tray, --settings, hotkey).
	wmSettings       = wmApp + 8
	wmDPIChanged     = 0x02E0
	wmCaptureChanged = 0x0215
	wmCancelMode     = 0x001F
	wmKeyEscape      = 0x1B
	wmKeyTab         = 0x09
	wmKeyReturn      = 0x0D
	wmKeySpace       = 0x20
	wmKeyEnd         = 0x23
	wmKeyHome        = 0x24
	wmKeyUp          = 0x26
	wmKeyDownArrow   = 0x28
	mkLButton        = 0x0001

	// ninBalloonUserClick is the tray callback event raised when the user
	// clicks a balloon tip, so a notification can double as a shortcut back
	// to the overlay.
	ninBalloonUserClick = 0x0405

	htClient      = 1
	htTransparent = ^uintptr(0)
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	winmm  = syscall.NewLazyDLL("winmm.dll")

	setProcessDpiAwareness  = user32.NewProc("SetProcessDpiAwarenessContext")
	unregisterClassProc     = user32.NewProc("UnregisterClassW")
	monitorFromPointProc    = user32.NewProc("MonitorFromPoint")
	getMonitorInfoProc      = user32.NewProc("GetMonitorInfoW")
	enumDisplayMonitorsProc = user32.NewProc("EnumDisplayMonitors")
	updateLayeredProc       = user32.NewProc("UpdateLayeredWindow")
	messageBeepProc         = user32.NewProc("MessageBeep")
	playSoundProc           = winmm.NewProc("PlaySoundW")
)

var windowProcedure = syscall.NewCallback(procedure)

func enableDPIAwareness() {
	setProcessDpiAwareness.Call(^uintptr(3))
}

// playSoundW plays a complete WAV image held in Go memory. SND_ASYNC returns
// immediately, so a slow or absent audio device can never stall the UI pump.
// The caller keeps the backing slice alive until playback finishes.
func playSoundW(data *byte, flags uint32) bool {
	result, _, _ := playSoundProc.Call(uintptr(unsafe.Pointer(data)), 0, uintptr(flags))
	return result != 0
}

func unregisterClass(name *uint16, instance win.HINSTANCE) bool {
	result, _, _ := unregisterClassProc.Call(uintptr(unsafe.Pointer(name)), uintptr(instance))
	return result != 0
}

func monitorFromPoint(point *win.POINT) win.HMONITOR {
	result, _, _ := monitorFromPointProc.Call(uintptr(unsafe.Pointer(point)), monitorDefaultToPrimary)
	return win.HMONITOR(result)
}

func getMonitorInfo(monitor win.HMONITOR, info *win.MONITORINFO) bool {
	result, _, _ := getMonitorInfoProc.Call(uintptr(monitor), uintptr(unsafe.Pointer(info)))
	return result != 0
}

var (
	monitorEnumMu       sync.Mutex
	monitorEnumOut      []displayRect
	monitorEnumCallback = syscall.NewCallback(func(monitor, _, _, _ uintptr) uintptr {
		info := win.MONITORINFO{CbSize: uint32(unsafe.Sizeof(win.MONITORINFO{}))}
		if getMonitorInfo(win.HMONITOR(monitor), &info) {
			monitorEnumOut = append(monitorEnumOut, displayRect{
				Left:    info.RcWork.Left,
				Top:     info.RcWork.Top,
				Right:   info.RcWork.Right,
				Bottom:  info.RcWork.Bottom,
				Primary: info.DwFlags&monitorInfoPrimary != 0,
			})
		}
		return 1
	})
)

func enumerateDisplayWorkAreas() []displayRect {
	monitorEnumMu.Lock()
	defer monitorEnumMu.Unlock()
	monitorEnumOut = monitorEnumOut[:0]
	_, _, _ = enumDisplayMonitorsProc.Call(0, 0, monitorEnumCallback, 0)
	if len(monitorEnumOut) == 0 {
		return nil
	}
	out := make([]displayRect, len(monitorEnumOut))
	copy(out, monitorEnumOut)
	return out
}

// updateLayeredWindow pushes a 32-bit premultiplied BGRA frame to the layered
// window. opacity is the constant alpha applied to the whole surface
// (0..255); the per-pixel alpha in sourceDC still controls the panel's shape.
func updateLayeredWindow(hwnd win.HWND, destinationDC win.HDC, destination *win.POINT, dimensions win.SIZE, sourceDC win.HDC, opacity uint8) error {
	source := win.POINT{}
	blend := win.BLENDFUNCTION{
		BlendOp:             acSrcOver,
		BlendFlags:          0,
		SourceConstantAlpha: opacity,
		AlphaFormat:         acSrcAlpha,
	}
	result, _, callErr := syscall.Syscall9(
		updateLayeredProc.Addr(), 9,
		uintptr(hwnd),
		uintptr(destinationDC),
		uintptr(unsafe.Pointer(destination)),
		uintptr(unsafe.Pointer(&dimensions)),
		uintptr(sourceDC),
		uintptr(unsafe.Pointer(&source)),
		0,
		uintptr(unsafe.Pointer(&blend)),
		ulwAlpha,
	)
	if result == 0 {
		return callErr
	}
	return nil
}
