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
	wmMouseWheel      = 0x020A
	wmRButtonUp       = 0x0205
	wmApp             = 0x8000
	wmRefresh         = wmApp + 1
	wmTray            = wmApp + 2
	wmActivate        = wmApp + 3
	wmDPIChanged      = 0x02E0
	wmCaptureChanged  = 0x0215
	wmCancelMode      = 0x001F
	wmKeyEscape       = 0x1B
	mkLButton         = 0x0001

	htClient      = 1
	htTransparent = ^uintptr(0)
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	setProcessDpiAwareness = user32.NewProc("SetProcessDpiAwarenessContext")
	unregisterClassProc    = user32.NewProc("UnregisterClassW")
	monitorFromPointProc   = user32.NewProc("MonitorFromPoint")
	getMonitorInfoProc     = user32.NewProc("GetMonitorInfoW")
	updateLayeredProc      = user32.NewProc("UpdateLayeredWindow")
)

var windowProcedure = syscall.NewCallback(procedure)

func enableDPIAwareness() {
	setProcessDpiAwareness.Call(^uintptr(3))
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

func updateLayeredWindow(hwnd win.HWND, destinationDC win.HDC, destination *win.POINT, dimensions win.SIZE, sourceDC win.HDC) error {
	source := win.POINT{}
	blend := win.BLENDFUNCTION{
		BlendOp:             acSrcOver,
		BlendFlags:          0,
		SourceConstantAlpha: 255,
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
