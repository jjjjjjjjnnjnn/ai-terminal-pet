package ui

import (
	"syscall"
	"unsafe"

	win "github.com/lxn/win"
)

const (
	ulwAlpha                = 0x00000002
	acSrcOver               = 0x00000000
	acSrcAlpha              = 0x00000001
	monitorDefaultToPrimary = 0x00000001

	wmDestroy         = 0x0002
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
	// wmSettings opens the settings window (tray, --settings, hotkey).
	wmSettings       = wmApp + 8
	wmDPIChanged     = 0x02E0
	wmCaptureChanged = 0x0215
	wmCancelMode     = 0x001F
	wmKeyEscape      = 0x1B
	mkLButton        = 0x0001

	// ninBalloonUserClick is the tray callback event raised when the user
	// clicks a balloon tip, so a notification can double as a shortcut back
	// to the overlay.
	ninBalloonUserClick = 0x0405

	// MessageBeep kinds are kept only as the last-resort fallback in beep()
	// (see sound_windows.go). The real alert tones are synthesised WAV data
	// played through winmm, whose volume the user controls.
	beepError       = 0x00000010 // MB_ICONHAND
	beepExclamation = 0x00000030 // MB_ICONEXCLAMATION

	htClient      = 1
	htTransparent = ^uintptr(0)
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	winmm  = syscall.NewLazyDLL("winmm.dll")

	setProcessDpiAwareness = user32.NewProc("SetProcessDpiAwarenessContext")
	unregisterClassProc    = user32.NewProc("UnregisterClassW")
	monitorFromPointProc   = user32.NewProc("MonitorFromPoint")
	getMonitorInfoProc     = user32.NewProc("GetMonitorInfoW")
	updateLayeredProc      = user32.NewProc("UpdateLayeredWindow")
	messageBeepProc        = user32.NewProc("MessageBeep")
	playSoundProc          = winmm.NewProc("PlaySoundW")
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
