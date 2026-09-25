package ui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	win "github.com/lxn/win"
)

const (
	timerIdle      = 1
	timerClock     = 2
	timerAnimation = 3
	menuShow       = 1001
	menuHide       = 1002
	menuExit       = 1003

	// windowClassName is also the rendezvous used by ActivateExisting, so
	// the short-lived second process can find the live overlay.
	windowClassName = "AITerminalPetWindow"
	windowTitle     = "AI Terminal Pet"
)

var activeWindow atomic.Pointer[App]

type dragState struct {
	active  bool
	moved   bool
	startX  int32
	startY  int32
	windowX int32
	windowY int32
}

type messageEnvelope struct {
	event   win.MSG
	reserve [32]byte
}

type App struct {
	values   *store.Store
	handle   atomic.Uintptr
	dpi      uint32
	zoom     float64
	expanded bool
	hidden   bool
	pressed  bool
	x        int32
	y        int32
	width    int
	height   int
	sessions []session.Session
	state    session.State
	revision uint64
	hit      []byte
	// surface is the reused DIB canvas. Recreating it every frame meant a
	// CreateDIBSection/SelectObject/DeleteObject round trip at up to 20 FPS
	// (the alert-state animation timer), which is pure GDI churn on a
	// layered window. UI thread only.
	surface *canvas
	drag    dragState
	// lastAvailable is the grace anchor for the idle self-destruct: with no
	// sessions the overlay removes itself, so the reporters' on-demand
	// launch path stays the only thing that keeps a process alive.
	lastAvailable   time.Time
	taskbarMessage  uint32
	trayIcon        win.HICON
	trayAdded       bool
	refreshPending  atomic.Bool
	activatePending atomic.Bool
	// lastDispatchNs records the last time windowProcedure was entered
	// (any nesting level, including modal menu loops). A watchdog dumps
	// goroutine stacks when this goes stale while a window exists.
	lastDispatchNs atomic.Int64
	lastHangDumpNs atomic.Int64
	// lastMessage is the most recent window message id, recorded so a hang
	// dump can name the message the UI thread was last asked to handle.
	lastMessage atomic.Uint32
}

func New(values *store.Store) *App {
	return &App{values: values, dpi: 96, zoom: 1, state: session.StateIdle, revision: ^uint64(0), lastAvailable: time.Now()}
}

// ActivateExisting asks an already-running instance to reveal itself and
// reports whether one was found.
//
// The overlay is single-instance and on demand: reporters launch it, and it
// removes itself once no session is left. A user who double-clicks the
// executable while an instance is alive used to hit the single-instance
// mutex and exit silently, which is indistinguishable from "the program is
// broken" - the reported "open anomaly". The second process now pokes the
// live window with wmActivate instead, running the same show path a tray
// click would.
//
// PostMessage is asynchronous and safe across threads, so this can never
// block the short-lived second process; a missing window is a plain false.
func ActivateExisting() bool {
	name, err := syscall.UTF16FromString(windowClassName)
	if err != nil {
		return false
	}
	handle := win.FindWindow(&name[0], nil)
	if handle == 0 {
		return false
	}
	return win.PostMessage(handle, wmActivate, 0, 0) != 0
}

func (a *App) Refresh() {
	a.postRefresh()
}

// postRefresh coalesces paint requests onto the message queue so input
// handlers never block inside present()/UpdateLayeredWindow.
func (a *App) postRefresh() {
	handle := a.handle.Load()
	if handle == 0 || !a.refreshPending.CompareAndSwap(false, true) {
		return
	}
	if win.PostMessage(win.HWND(handle), wmRefresh, 0, 0) == 0 {
		a.refreshPending.Store(false)
	}
}

func (a *App) Activate() {
	handle := a.handle.Load()
	if handle == 0 || !a.activatePending.CompareAndSwap(false, true) {
		return
	}
	if win.PostMessage(win.HWND(handle), wmActivate, 0, 0) == 0 {
		a.activatePending.Store(false)
	}
}

func (a *App) Close() {
	handle := a.handle.Load()
	if handle != 0 {
		win.PostMessage(win.HWND(handle), wmClose, 0, 0)
	}
}

func (a *App) Run() error {
	// Win32 demands that window creation, the first ShowWindow/SetWindowPos
	// and the message pump all happen on ONE OS thread. Go's scheduler is
	// free to migrate an unlocked goroutine to a different M between any two
	// calls, and window messages are delivered to the queue of the thread
	// that created the window. When that migration happened here, the
	// creating thread ended up with a queue nobody pumped, so the very first
	// SetWindowPos(HWND_TOPMOST) blocked forever inside a cross-thread
	// SendMessage: WS_VISIBLE never got set, no message was ever dispatched,
	// and the process became a ghost window that Windows reported as hung.
	// That single omission produced every reported symptom - invisible
	// overlay, click "crash", dead timers, no drag - and is verified by
	// instrumenting Run() with and without this pin.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	enableDPIAwareness()
	className, err := syscall.UTF16FromString(windowClassName)
	if err != nil {
		return err
	}
	windowName, err := syscall.UTF16FromString(windowTitle)
	if err != nil {
		return err
	}
	instance := win.GetModuleHandle(nil)
	cursor := win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	class := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		Style:         win.CS_HREDRAW | win.CS_VREDRAW | win.CS_DBLCLKS,
		LpfnWndProc:   windowProcedure,
		HInstance:     instance,
		HCursor:       cursor,
		LpszClassName: &className[0],
	}
	activeWindow.Store(a)
	defer activeWindow.Store(nil)
	if atom := win.RegisterClassEx(&class); atom == 0 {
		return fmt.Errorf("register class: %w", syscall.GetLastError())
	}
	defer unregisterClass(&className[0], instance)
	created := win.CreateWindowEx(
		uint32(win.WS_EX_TOPMOST|win.WS_EX_TOOLWINDOW|win.WS_EX_LAYERED),
		&className[0], &windowName[0], uint32(win.WS_POPUP),
		0, 0, int32(a.logicalWidth()), int32(a.logicalHeight()),
		0, 0, instance, nil,
	)
	runtime.KeepAlive(className)
	runtime.KeepAlive(windowName)
	if created == 0 {
		return fmt.Errorf("create window: %w", syscall.GetLastError())
	}
	a.handle.Store(uintptr(created))
	a.dpi = a.windowDpi()
	taskbarName, _ := syscall.UTF16FromString("TaskbarCreated")
	a.taskbarMessage = win.RegisterWindowMessage(&taskbarName[0])
	a.addTray()
	a.load()
	a.positionForWorkArea()
	if err := a.present(); err != nil {
		return fmt.Errorf("present: %w", err)
	}
	win.ShowWindow(created, win.SW_SHOW)
	a.setTopmost()
	win.SetTimer(created, timerIdle, 500, 0)
	win.SetTimer(created, timerClock, 2000, 0)
	win.SetTimer(created, timerAnimation, 50, 0)
	a.lastDispatchNs.Store(time.Now().UnixNano())
	watchStop := make(chan struct{})
	defer close(watchStop)
	go a.watchHang(watchStop)
	var envelope messageEnvelope
	for {
		result := win.GetMessage(&envelope.event, 0, 0, 0)
		if int32(result) == -1 {
			return syscall.GetLastError()
		}
		if result == 0 {
			return nil
		}
		win.TranslateMessage(&envelope.event)
		win.DispatchMessage(&envelope.event)
	}
}

func procedure(hwnd win.HWND, message uint32, wParam, lParam uintptr) uintptr {
	window := activeWindow.Load()
	if window != nil {
		window.lastMessage.Store(message)
		window.lastDispatchNs.Store(time.Now().UnixNano())
		return window.handleMessage(hwnd, message, wParam, lParam)
	}
	return win.DefWindowProc(hwnd, message, wParam, lParam)
}

// watchHang dumps all goroutine stacks to hang.log when the UI thread
// stops dispatching for longer than stallThreshold while the window
// exists. Idle blocking inside GetMessage does NOT trigger this because
// GetMessage itself returns (timers fire at 50ms) and every dispatch
// refreshes lastDispatchNs.
//
// Dumps are bounded on purpose. The previous unbounded, 60s-throttled
// version appended 370 KB of identical stacks over a single 80-minute
// outage without ever naming the culprit, which made the log useless for
// diagnosis. A stall that repeats after maxDumps is the same stall; the
// first dump already carries the stack that matters.
func (a *App) watchHang(stop <-chan struct{}) {
	const stallThreshold = 6 * time.Second
	const dumpCooldown = 2 * time.Minute
	const maxDumps = 3
	const maxHangLog = 256 << 10
	dumps := 0
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case now := <-ticker.C:
			if dumps >= maxDumps || a.handle.Load() == 0 {
				continue
			}
			last := a.lastDispatchNs.Load()
			if last == 0 || now.Sub(time.Unix(0, last)) < stallThreshold {
				continue
			}
			if now.Sub(time.Unix(0, a.lastHangDumpNs.Load())) < dumpCooldown {
				continue
			}
			a.lastHangDumpNs.Store(now.UnixNano())
			dumps++
			writeHangLog(now, maxHangLog, a.lastMessage.Load(), dumps, maxDumps)
		}
	}
}

func writeHangLog(now time.Time, maxBytes int64, lastMessage uint32, dump, maxDumps int) {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return
	}
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return
	}
	path := filepath.Join(directory, "hang.log")
	if info, err := os.Stat(path); err == nil && info.Size() > maxBytes {
		_ = os.Remove(path)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "=== stall %d/%d at %s lastMessage=%s goroutines=%d ===\n",
		dump, maxDumps, now.Format(time.RFC3339), messageName(lastMessage), runtime.NumGoroutine())
	_ = pprof.Lookup("goroutine").WriteTo(file, 2)
}

// messageName labels the handful of window messages the UI thread handles
// so a stall dump is readable without a message-number table at hand.
func messageName(message uint32) string {
	switch message {
	case wmDestroy:
		return "WM_DESTROY"
	case wmClose:
		return "WM_CLOSE"
	case wmEraseBackground:
		return "WM_ERASEBKGND"
	case wmNCHitTest:
		return "WM_NCHITTEST"
	case wmContextMenu:
		return "WM_CONTEXTMENU"
	case wmKeyDown:
		return "WM_KEYDOWN"
	case wmTimer:
		return "WM_TIMER"
	case wmLButtonDown:
		return "WM_LBUTTONDOWN"
	case wmLButtonUp:
		return "WM_LBUTTONUP"
	case wmLButtonDblClk:
		return "WM_LBUTTONDBLCLK"
	case wmMouseMove:
		return "WM_MOUSEMOVE"
	case wmMouseWheel:
		return "WM_MOUSEWHEEL"
	case wmRButtonUp:
		return "WM_RBUTTONUP"
	case wmRefresh:
		return "WM_APP+1(refresh)"
	case wmTray:
		return "WM_APP+2(tray)"
	case wmActivate:
		return "WM_APP+3(activate)"
	case wmDPIChanged:
		return "WM_DPICHANGED"
	case wmCaptureChanged:
		return "WM_CAPTURECHANGED"
	case wmCancelMode:
		return "WM_CANCELMODE"
	default:
		return fmt.Sprintf("%#x", message)
	}
}

func (a *App) handleMessage(hwnd win.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmEraseBackground:
		return 1
	case wmRefresh:
		a.refreshPending.Store(false)
		a.load()
		if len(a.sessions) > 0 {
			a.lastAvailable = time.Now()
		}
		if !a.hidden {
			_ = a.present()
		}
		return 0
	case wmActivate:
		a.activatePending.Store(false)
		a.showFromTray()
		return 0
	case wmTray:
		a.handleTrayMessage(message, lParam)
		return 0
	case wmTimer:
		return a.handleTimer(hwnd, wParam)
	case wmNCHitTest:
		return a.hitTest(lParam)
	case wmLButtonDown:
		a.beginDrag(lParam)
		return 0
	case wmMouseMove:
		if a.drag.active && wParam&mkLButton != 0 {
			a.moveDrag()
		}
		return 0
	case wmLButtonUp:
		a.endDrag(lParam)
		return 0
	case wmMouseWheel:
		a.zoomBy(float64(int16(uint16(wParam >> 16))))
		return 0
	case wmLButtonDblClk:
		if !a.expanded {
			a.toggleExpanded()
		}
		return 0
	case wmContextMenu:
		a.showContextMenu(false)
		return 0
	case wmKeyDown:
		if wParam == wmKeyEscape && a.expanded {
			a.toggleExpanded()
			return 0
		}
	case wmDPIChanged:
		a.handleDPIChanged(wParam)
		return 0
	case wmCaptureChanged, wmCancelMode:
		a.drag = dragState{}
		a.pressed = false
		if !a.hidden {
			a.postRefresh()
		}
		return 0
	case wmClose:
		win.DestroyWindow(hwnd)
		return 0
	case wmDestroy:
		a.removeTray()
		a.releaseSurface()
		a.handle.Store(0)
		win.PostQuitMessage(0)
		return 0
	}
	if a.taskbarMessage != 0 && message == a.taskbarMessage {
		a.addTray()
		return 0
	}
	return win.DefWindowProc(hwnd, message, wParam, lParam)
}

func (a *App) handleTimer(hwnd win.HWND, timer uintptr) uintptr {
	switch timer {
	case timerIdle:
		if a.values.HasAvailable() {
			a.lastAvailable = time.Now()
		} else if time.Since(a.lastAvailable) >= 5*time.Second {
			win.DestroyWindow(hwnd)
		}
	case timerClock:
		if a.expanded && !a.hidden {
			a.load()
			a.postRefresh()
		}
	case timerAnimation:
		if a.hidden {
			return 0
		}
		a.load()
		if a.pressed || a.state == session.StateWaitingPermission || a.state == session.StateError {
			a.postRefresh()
		}
	}
	return 0
}

func (a *App) handleTrayMessage(message uint32, lParam uintptr) {
	if message == a.taskbarMessage && message != 0 {
		a.addTray()
		return
	}
	event := uint32(uint16(lParam))
	switch event {
	case wmLButtonUp, wmLButtonDblClk:
		a.showFromTray()
	case wmRButtonUp, wmContextMenu:
		a.showContextMenu(true)
	}
}

func (a *App) hitTest(lParam uintptr) uintptr {
	location := win.POINT{X: int32(int16(uint16(lParam))), Y: int32(int16(uint16(lParam >> 16)))}
	if !win.ScreenToClient(a.window(), &location) {
		return htTransparent
	}
	if classifyHit(int(location.X), int(location.Y), a.width, a.height, a.hit) {
		return htClient
	}
	return htTransparent
}

// classifyHit is the pure, unit-tested core of hitTest: true means the
// point lands on an opaque pixel. X must be checked against the canvas
// width (NOT len(hit) which equals width*height); the old check allowed
// X in [width, width*height) and Y*width+X then escaped the slice and
// panicked ("index out of range") the moment the window was wider than
// the canvas (zoom-out / DPI-down + margin hover), killing the process
// with no error.log entry.
func classifyHit(x, y, width, height int, hit []byte) bool {
	if width <= 0 || height <= 0 || len(hit) < width*height {
		return false
	}
	if x < 0 || y < 0 || x >= width || y >= height {
		return false
	}
	return hit[y*width+x] >= 20
}

func (a *App) beginDrag(lParam uintptr) {
	y := int32(int16(uint16(lParam >> 16)))
	if a.expanded && y > int32(float64(40)*float64(a.width)/200) {
		return
	}
	var cursor win.POINT
	if !win.GetCursorPos(&cursor) {
		return
	}
	var bounds win.RECT
	if !win.GetWindowRect(a.window(), &bounds) {
		return
	}
	a.pressed = true
	a.postRefresh()
	a.drag = dragState{active: true, startX: cursor.X, startY: cursor.Y, windowX: bounds.Left, windowY: bounds.Top}
	win.SetCapture(a.window())
}

func (a *App) moveDrag() {
	var cursor win.POINT
	if !win.GetCursorPos(&cursor) {
		return
	}
	dx := cursor.X - a.drag.startX
	dy := cursor.Y - a.drag.startY
	if abs32(dx)+abs32(dy) > 5 {
		a.drag.moved = true
	}
	a.x = a.drag.windowX + dx
	a.y = a.drag.windowY + dy
	win.SetWindowPos(a.window(), win.HWND_TOPMOST, a.x, a.y, 0, 0, win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
}

// releaseAction is what a finished drag should trigger.
type releaseAction int

const (
	releaseNothing releaseAction = iota
	releaseToggle
)

// decideRelease maps a completed drag to the action it triggers. It is pure
// so the invariant "a drag that moved must never also toggle the panel" is
// locked down by a test.
//
// The old code read a.drag.moved AFTER calling ReleaseCapture. ReleaseCapture
// dispatches WM_CAPTURECHANGED synchronously to this same window, and that
// handler clears a.drag, so moved was always observed as false: every drag
// additionally toggled the panel between compact and expanded, which is what
// "cannot drag the pet" looked like from the outside.
func decideRelease(active, moved, expanded bool, x, y, width int) releaseAction {
	if !active || moved {
		return releaseNothing
	}
	if !expanded {
		return releaseToggle
	}
	if closeButtonHit(x, y, width) {
		return releaseToggle
	}
	return releaseNothing
}

func (a *App) endDrag(lParam uintptr) {
	if !a.drag.active {
		a.pressed = false
		return
	}
	x := int(int16(uint16(lParam)))
	y := int(int16(uint16(lParam >> 16)))
	// Decide before touching capture: see decideRelease.
	action := decideRelease(a.drag.active, a.drag.moved, a.expanded, x, y, a.width)
	a.drag = dragState{}
	a.pressed = false
	win.ReleaseCapture()
	a.postRefresh()
	if action == releaseToggle {
		a.toggleExpanded()
	}
}

func (a *App) zoomBy(delta float64) {
	if delta == 0 {
		return
	}
	step := 0.1
	if delta < 0 {
		step = -step
	}
	a.zoom = math.Max(0.65, math.Min(1.6, a.zoom+step))
	a.load()
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.resize()
	a.clampToMonitor()
	a.setTopmost()
	a.postRefresh()
}

func (a *App) toggleExpanded() {
	a.expanded = !a.expanded
	a.load()
	// Publish the new geometry synchronously (resize/clamp only position
	// the frame; they never block), but paint asynchronously: present()
	// issues UpdateLayeredWindow which has stalled for seconds under
	// system load, and stalling an input handler is exactly what Windows
	// reports as AppHang (busy cursor, ghost window, click "crash").
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.resize()
	a.clampToMonitor()
	a.setTopmost()
	a.postRefresh()
}

func (a *App) showContextMenu(fromTray bool) {
	menu := win.CreatePopupMenu()
	if menu == 0 {
		return
	}
	defer win.DestroyMenu(menu)
	position := uint32(0)
	if fromTray {
		a.appendMenu(menu, position, menuShow, "显示浮窗")
		position++
	} else {
		a.appendMenu(menu, position, menuHide, "隐藏到托盘")
		position++
	}
	a.appendMenu(menu, position, menuExit, "退出")
	var cursor win.POINT
	win.GetCursorPos(&cursor)
	win.SetForegroundWindow(a.window())
	command := win.TrackPopupMenu(menu, win.TPM_RIGHTBUTTON|win.TPM_RETURNCMD|win.TPM_NONOTIFY, cursor.X, cursor.Y, 0, a.window(), nil)
	switch command {
	case menuShow:
		a.showFromTray()
	case menuHide:
		a.hideToTray()
	case menuExit:
		win.DestroyWindow(a.window())
	}
}

func (a *App) appendMenu(menu win.HMENU, position uint32, id uintptr, label string) {
	encoded, _ := syscall.UTF16FromString(label)
	if len(encoded) == 0 {
		return
	}
	info := win.MENUITEMINFO{
		CbSize:     uint32(unsafe.Sizeof(win.MENUITEMINFO{})),
		FMask:      win.MIIM_STRING | win.MIIM_ID,
		WID:        uint32(id),
		DwTypeData: &encoded[0],
	}
	win.InsertMenuItem(menu, position, true, &info)
	runtime.KeepAlive(encoded)
}

func (a *App) hideToTray() {
	if a.handle.Load() != 0 {
		win.ShowWindow(a.window(), win.SW_HIDE)
	}
	a.hidden = true
}

func (a *App) showFromTray() {
	handle := a.handle.Load()
	if handle == 0 {
		return
	}
	a.hidden = false
	win.ShowWindow(win.HWND(handle), win.SW_SHOW)
	a.load()
	win.PostMessage(win.HWND(handle), wmRefresh, 0, 0)
	a.setTopmost()
}

func (a *App) addTray() {
	handle := a.handle.Load()
	if handle == 0 {
		return
	}
	a.trayIcon = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))
	data := a.notifyIcon()
	if !win.Shell_NotifyIcon(win.NIM_ADD, &data) {
		return
	}
	a.trayAdded = true
	data.UVersion = win.NOTIFYICON_VERSION_4
	win.Shell_NotifyIcon(win.NIM_SETVERSION, &data)
}

func (a *App) removeTray() {
	if !a.trayAdded {
		return
	}
	data := a.notifyIcon()
	win.Shell_NotifyIcon(win.NIM_DELETE, &data)
	a.trayAdded = false
}

func (a *App) notifyIcon() win.NOTIFYICONDATA {
	data := win.NOTIFYICONDATA{
		CbSize:           uint32(unsafe.Sizeof(win.NOTIFYICONDATA{})),
		HWnd:             a.window(),
		UID:              1,
		UFlags:           win.NIF_MESSAGE | win.NIF_ICON | win.NIF_TIP,
		UCallbackMessage: wmTray,
		HIcon:            a.trayIcon,
	}
	tip, _ := syscall.UTF16FromString("AI Terminal Brief")
	copy(data.SzTip[:], tip)
	return data
}

func (a *App) present() error {
	width, height := a.logicalWidth(), a.logicalHeight()
	if a.surface != nil && (a.surface.width != width || a.surface.height != height) {
		a.surface.close()
		a.surface = nil
	}
	if a.surface == nil {
		target, err := newCanvas(width, height)
		if err != nil {
			return err
		}
		a.surface = target
	}
	target := a.surface
	target.clear()
	drawVerticalPanel(target, a.sessions, time.Now(), a.expanded, a.pressed)
	screenDC := win.GetDC(0)
	if screenDC == 0 {
		return syscall.EINVAL
	}
	defer win.ReleaseDC(0, screenDC)
	destination := win.POINT{X: a.x, Y: a.y}
	dimensions := win.SIZE{CX: int32(width), CY: int32(height)}
	if err := updateLayeredWindow(a.window(), screenDC, &destination, dimensions, target.dc); err != nil {
		return err
	}
	a.width = width
	a.height = height
	if len(a.hit) != width*height {
		a.hit = make([]byte, width*height)
	}
	target.hitMapInto(a.hit)
	return nil
}

func (a *App) releaseSurface() {
	if a.surface != nil {
		a.surface.close()
		a.surface = nil
	}
}

func (a *App) load() {
	revision := a.values.Revision()
	if revision == a.revision {
		return
	}
	a.revision = revision
	a.sessions = a.values.List()
	a.state = session.Aggregate(a.sessions)
}

func (a *App) setTopmost() {
	if a.handle.Load() == 0 {
		return
	}
	win.SetWindowPos(a.window(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE)
}

// resize syncs the window frame to the current logical canvas size.
// UpdateLayeredWindow never changes the window size on its own, so without
// this expand/zoom/DPI only repainted a clipped canvas and clicks appeared
// to do nothing.
func (a *App) resize() {
	if a.handle.Load() == 0 {
		return
	}
	width, height := int32(a.logicalWidth()), int32(a.logicalHeight())
	win.SetWindowPos(a.window(), 0, 0, 0, width, height, win.SWP_NOMOVE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
}

func (a *App) logicalWidth() int {
	base := 40
	if a.expanded {
		base = 200
	}
	return scaledDimension(base, a.dpi, a.zoom)
}

func (a *App) logicalHeight() int {
	base := 190
	if a.expanded {
		base = 250
	}
	return scaledDimension(base, a.dpi, a.zoom)
}

func (a *App) positionForWorkArea() {
	width, height := int32(a.logicalWidth()), int32(a.logicalHeight())
	probe := win.POINT{X: win.GetSystemMetrics(win.SM_CXSCREEN) - 1, Y: win.GetSystemMetrics(win.SM_CYSCREEN) - 1}
	monitor := monitorFromPoint(&probe)
	if monitor == 0 {
		a.x = 40
		a.y = 40
		return
	}
	info := win.MONITORINFO{CbSize: uint32(unsafe.Sizeof(win.MONITORINFO{}))}
	if !getMonitorInfo(monitor, &info) {
		a.x = 40
		a.y = 40
		return
	}
	a.x = info.RcWork.Right - width - int32(24*a.dpi/96)
	a.y = info.RcWork.Top + (info.RcWork.Bottom-info.RcWork.Top-height)/2
	a.clampToWorkArea(info.RcWork, width, height)
}

func (a *App) clampToMonitor() {
	width, height := int32(a.width), int32(a.height)
	if width == 0 || height == 0 {
		width, height = int32(a.logicalWidth()), int32(a.logicalHeight())
	}
	center := win.POINT{X: a.x + width/2, Y: a.y + height/2}
	monitor := monitorFromPoint(&center)
	if monitor == 0 {
		return
	}
	info := win.MONITORINFO{CbSize: uint32(unsafe.Sizeof(win.MONITORINFO{}))}
	if getMonitorInfo(monitor, &info) {
		a.clampToWorkArea(info.RcWork, width, height)
		win.SetWindowPos(a.window(), win.HWND_TOPMOST, a.x, a.y, 0, 0, win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
	}
}

func (a *App) clampToWorkArea(work win.RECT, width, height int32) {
	minimumX := work.Left + int32(8*a.dpi/96)
	minimumY := work.Top + int32(8*a.dpi/96)
	maximumX := work.Right - width - int32(8*a.dpi/96)
	maximumY := work.Bottom - height - int32(8*a.dpi/96)
	if maximumX < minimumX {
		a.x = work.Left
	} else {
		a.x = max32(minimumX, min32(a.x, maximumX))
	}
	if maximumY < minimumY {
		a.y = work.Top
	} else {
		a.y = max32(minimumY, min32(a.y, maximumY))
	}
}

func (a *App) handleDPIChanged(wParam uintptr) {
	dpi := uint32(wParam >> 16)
	if dpi < 72 {
		dpi = 72
	}
	if dpi > 600 {
		dpi = 600
	}
	a.dpi = dpi
	a.load()
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.resize()
	a.clampToMonitor()
	a.setTopmost()
	a.postRefresh()
}

func (a *App) windowDpi() uint32 {
	dpi := win.GetDpiForWindow(a.window())
	if dpi == 0 {
		return 96
	}
	return dpi
}

func (a *App) window() win.HWND {
	return win.HWND(a.handle.Load())
}

func scaledDimension(base int, dpi uint32, zoom float64) int {
	value := int(math.Round(float64(base) * float64(dpi) / 96 * zoom))
	return maxInt(value, 32)
}

func abs32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

func min32(left, right int32) int32 {
	if left < right {
		return left
	}
	return right
}

func max32(left, right int32) int32 {
	if left > right {
		return left
	}
	return right
}
