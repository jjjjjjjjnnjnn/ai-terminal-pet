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

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	win "github.com/lxn/win"
)

const (
	timerIdle      = 1
	timerClock     = 2
	timerAnimation = 3

	menuShow          = 1001
	menuHide          = 1002
	menuExit          = 1003
	menuNotifications = 1004
	menuTopmost       = 1005
	menuReset         = 1006
	menuSound         = 1007
	menuAutoStart     = 1008
	menuEdgeRetract   = 1009
	menuLangAuto      = 1010
	menuLangEn        = 1011
	menuLangZh        = 1012
	menuAutoApprove   = 1013
	menuSoundVolume   = 1014
	menuNotifyFinish  = 1015
	menuSettings      = 1016

	// The pulse and the slide share one timer but want different frame rates.
	//
	// The pulse is a slow breath - about 1.1 s per cycle. Every animation
	// tick is a full GDI raster of the layered window, so a session parked in
	// waiting_permission used to burn CPU forever at 12.5 fps. At 160 ms
	// (6.25 fps) the breath still reads as a breath but costs half as much.
	// The slide is short and reads as motion, so it keeps the higher rate.
	pulseInterval = 160
	slideInterval = 33

	// windowClassName is also the rendezvous used by Control, so a
	// short-lived second process can find the live overlay.
	windowClassName = "AITerminalPetWindow"
	windowTitle     = "AI Terminal Pet"

	// trayTipMax is NOTIFYICONDATA.SzTip's capacity including the NUL.
	trayTipMax = 128
	// trayInfoMax is SzInfo/SzInfoTitle's capacity including the NUL.
	trayInfoMax  = 256
	trayTitleMax = 64
)

// Placement is the window geometry remembered between launches. The ui
// package deliberately does not import internal/config: the app layer owns the
// file paths and hands the UI a plain struct plus a persist callback.
type Placement struct {
	Placed   bool
	X        int32
	Y        int32
	Zoom     float64
	Expanded bool
}

// Settings is the subset of user preferences the tray menu can flip.
type Settings struct {
	Notifications bool
	Sound         bool
	// SoundVolume is the tone amplitude, 0..100. It is mirrored into the
	// audio layer on startup, hot reload and every tray change.
	SoundVolume int
	// SoundScheme picks one of the built-in timbres; see ui.SoundSchemes.
	SoundScheme string
	// SoundOnWaiting / SoundOnError gate the tone per event, so someone who
	// only wants to hear genuine blockers can mute the rest.
	SoundOnWaiting bool
	SoundOnError   bool
	// NotifyOnFinish makes a session settling into idle/closed an alert too.
	NotifyOnFinish bool
	AlwaysOnTop    bool
	// EdgeRetract slides the panel into the screen edge it is hugging when
	// nothing needs attention.
	EdgeRetract bool
	// Language is the persisted language preference: "auto" follows the
	// host OS locale, "en" and "zh-CN" pin one. See internal/i18n.
	Language string
	// AutoApprove lets a permission prompt auto-approve after the timeout
	// when the user says nothing. Off by default.
	AutoApprove bool
	// AutoApproveTimeoutSec is the countdown length; the settings window can
	// change it at runtime (it used to be read only at startup).
	AutoApproveTimeoutSec int
	// Opacity is the layered window's constant alpha, 1..255.
	Opacity int
	// Theme picks one of the named palettes (see ThemeNames). It is a name,
	// not a colour value: the ui package owns the hues.
	Theme string
	// PerSourceColor colours each expanded row's source name by which terminal
	// it came from. Off by default, because the state colour is the one signal
	// the strip exists for and provenance is secondary.
	PerSourceColor bool
	// HiddenAtLaunch starts with the panel in the tray: the reporters start the
	// overlay, and someone who wants it out of the way until a session needs
	// attention should not have to hide it every logon.
	HiddenAtLaunch bool
}

// DefaultSettings is what "restore defaults" puts back, and what the overlay
// starts with when no config file exists. It mirrors config.Default(): the ui
// package does not import internal/config (the app layer owns file paths), so
// a test in the app package pins the two together.
func DefaultSettings() Settings {
	return Settings{
		Notifications:         true,
		Sound:                 true,
		SoundVolume:           85,
		SoundScheme:           soundSchemeChime,
		SoundOnWaiting:        true,
		SoundOnError:          true,
		NotifyOnFinish:        true,
		AlwaysOnTop:           true,
		EdgeRetract:           true,
		Language:              "auto",
		AutoApprove:           false,
		AutoApproveTimeoutSec: 5,
		Opacity:               235,
		Theme:                 themeSlate,
		PerSourceColor:        false,
		HiddenAtLaunch:        false,
	}
}

// Options is everything the UI needs from outside. All callbacks are invoked
// on the UI thread and may be nil.
type Options struct {
	// Opacity is the layered window constant alpha, 1..255.
	Opacity uint8
	// Settings is the initial state of the tray switches.
	Settings Settings
	// AutoStart is the initial state of the run-at-logon switch. It lives in
	// the registry rather than config.json, so the UI only mirrors it.
	AutoStart bool
	// AutoApproveTimeoutSec is the countdown length for auto-approve.
	// Out-of-range values fall back to 5 seconds.
	AutoApproveTimeoutSec int
	// OpenSettings opens the settings window right after the first paint, so
	// `--settings` works even when it had to start the overlay itself.
	OpenSettings bool
	// Initial is the remembered geometry, or the zero value when the window
	// has never been placed.
	Initial Placement
	// OnPlacementChanged fires after the user moves, zooms or expands the
	// overlay so the app layer can persist it.
	OnPlacementChanged func(Placement)
	// OnSettingsChanged fires when a tray switch is flipped.
	OnSettingsChanged func(Settings)
	// OnAutoStartChanged fires when the run-at-logon switch is flipped. It
	// returns whether the change took effect; the UI keeps the old state when
	// it returns false, so the tick never lies about the registry.
	OnAutoStartChanged func(bool) bool
}

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
	event win.MSG
}

type App struct {
	values  *store.Store
	handle  atomic.Uintptr
	options Options
	// pendingSettings carries a config reload from the app layer's background
	// watcher to the UI thread. The UI never reads the config file itself.
	pendingSettings atomic.Pointer[Settings]
	dpi             uint32
	zoom            float64
	expanded        bool
	hidden          bool
	pressed         bool
	// hovering is true while the cursor is over the panel. It only brightens
	// the outline, but it is the difference between "a picture of a status"
	// and "something I can click".
	hovering bool
	x        int32
	y        int32
	width    int
	height   int
	// dock is the screen edge the panel hugs, dockEdge is that edge's x in
	// screen space, and dockGap is the distance it settled at from that edge.
	// All three are refreshed when the panel settles somewhere new. The gap is
	// what makes expand/collapse exact - see dock_windows.go.
	//
	// tuck is how far the panel has slid into that edge, 0 (fully out) to 1
	// (only a sliver left), and tuckTarget is where it is heading; the two
	// differ only while the slide animation is running.
	dock       dockSide
	dockEdge   int32
	dockGap    int32
	tuck       float64
	tuckTarget float64
	// lastFrame is the key of the frame currently on the window, so repeated
	// refreshes that change nothing cost nothing.
	lastFrame renderedKey
	// slideFrom and slideStart are where and when the current slide began, so
	// stepSlide can drive it by elapsed time instead of by tick count.
	slideFrom  float64
	slideStart time.Time
	// tuckHoldUntil suppresses the tuck until the given time, so an explicit
	// reveal (tray click, --show, --reset) actually shows something instead of
	// sliding back out of view before the user can look at it.
	tuckHoldUntil time.Time
	sessions      []session.Session
	state         session.State
	revision      uint64
	// digest fingerprints the session content the panel actually draws. The
	// store bumps revision on every push because updatedAt moves, so revision
	// alone would make every refresh look like a change and defeat the redraw
	// skip entirely; see sessionDigest.
	digest uint64
	hit    []byte
	// surface is the reused DIB canvas. Recreating it every frame meant a
	// CreateDIBSection/SelectObject/DeleteObject round trip at up to 20 FPS
	// (the alert-state animation timer), which is pure GDI churn on a
	// layered window. UI thread only.
	surface *canvas
	drag    dragState
	// animating tracks whether the pulse timer is currently installed, so
	// syncAnimation only calls SetTimer/KillTimer on a real transition.
	animating bool
	// timerInterval is the interval the installed timer is running at, so a
	// change between pulse and slide frame rates retunes it instead of being
	// ignored.
	timerInterval uint32
	// alertStates is the last observed state of every session. Tracking each
	// conversation separately matters: a single global "most urgent session"
	// key used to be reset whenever that session ended, so another still-
	// waiting session re-alerted later - or an unrelated new session looked
	// like the cause of a balloon. The first load only seeds this map, which
	// stops hot reload/restart from replaying old alerts.
	alertStates map[string]session.State
	// lastAvailable is the grace anchor for the idle self-destruct. With the
	// new existence model a live terminal keeps its session indefinitely, so
	// this only ever fires after every AI terminal has really been closed.
	// sessions the overlay removes itself, so the reporters' on-demand
	// launch path stays the only thing that keeps a process alive.
	lastAvailable time.Time
	// lang is the resolved display language. It is fixed at New() from the
	// settings plus the OS locale, and refreshed whenever the tray language
	// switches - drawing code reads it but never resolves it.
	lang i18n.Lang
	// approvals tracks the live auto-approve countdowns, keyed by session.
	// approveTimeout is how long each countdown runs. UI thread only.
	approvals       *approvalTracker
	approveTimeout  time.Duration
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
	// Settings window. It is a second top-level window on the same UI thread,
	// created lazily the first time it is opened: most runs never need it, and
	// a window plus a DIB is real memory for a panel nobody configured.
	settingsHandle     atomic.Uintptr
	settingsSurface    *canvas
	settingsDPI        uint32
	settingsGeometry   settingsGeometry
	settingsFrame      settingsFrameKey
	settingsTab        int
	settingsHover      settingsControl
	settingsHoverIndex int
	settingsPress      settingsControl
	// settingsDrag is the slider currently being dragged, settingsNone when
	// the pointer is not held. A drag repaints but does not persist; release
	// writes config.json once.
	settingsDrag settingsControl
	// settingsHoverTracking records whether WM_MOUSELEAVE is armed for the
	// settings window; TrackMouseEvent must be re-armed after every leave.
	settingsHoverTracking bool
	// settingsMoving is a header drag. settingsMoveDX/DY are the grab offset
	// inside the window, so the window does not jump to the cursor.
	settingsMoving bool
	settingsMoveDX int32
	settingsMoveDY int32
	settingsX      int32
	settingsY      int32
}

func New(values *store.Store, options Options) *App {
	if options.Opacity == 0 {
		options.Opacity = 235
	}
	if options.Settings.Opacity > 0 {
		options.Opacity = uint8(options.Settings.Opacity)
	}
	setSoundVolume(options.Settings.SoundVolume)
	setSoundScheme(options.Settings.SoundScheme)
	// Before anything is rasterised: a palette applied after the first paint
	// would mean the very first frame is drawn in the previous theme.
	applyAppearance(options.Settings.Theme, options.Settings.PerSourceColor)
	zoom := options.Initial.Zoom
	if math.IsNaN(zoom) || zoom < 0.65 || zoom > 1.6 {
		zoom = 1
	}
	timeout := options.AutoApproveTimeoutSec
	if options.Settings.AutoApproveTimeoutSec > 0 {
		timeout = options.Settings.AutoApproveTimeoutSec
	}
	if timeout < 1 || timeout > 60 {
		timeout = 5
	}
	return &App{
		values:         values,
		options:        options,
		dpi:            96,
		zoom:           zoom,
		expanded:       options.Initial.Expanded,
		state:          session.StateIdle,
		revision:       ^uint64(0),
		lastAvailable:  time.Now(),
		lang:           i18n.Resolve(i18n.Setting(options.Settings.Language), i18n.SystemLocale()),
		approvals:      newApprovalTracker(),
		approveTimeout: time.Duration(timeout) * time.Second,
	}
}

// ControlAction identifies a command one process can send to the live
// overlay. The second process is short lived: it finds the window by class
// name and posts the message, then exits. PostMessage is asynchronous and
// safe across threads, so a missing or busy window can never block it.
type ControlAction int

const (
	// ControlActivate reveals the overlay (same path as a tray click).
	ControlActivate ControlAction = iota
	// ControlToggle flips compact/expanded.
	ControlToggle
	// ControlQuit asks the overlay to close and exit.
	ControlQuit
	// ControlReset drops the remembered geometry and re-anchors.
	ControlReset
	// ControlSettings opens the settings window. It is a separate action so a
	// hotkey or a shortcut can reach it without going through the tray.
	ControlSettings
)

// Control sends action to an already-running instance and reports whether one
// was found. A missing window is a plain false, never an error: callers use it
// to decide whether to fall back to starting a fresh instance.
func Control(action ControlAction) bool {
	name, err := syscall.UTF16FromString(windowClassName)
	if err != nil {
		return false
	}
	handle := win.FindWindow(&name[0], nil)
	if handle == 0 {
		return false
	}
	var message uint32
	switch action {
	case ControlActivate:
		message = wmActivate
	case ControlToggle:
		message = wmToggle
	case ControlQuit:
		message = wmClose
	case ControlReset:
		message = wmReset
	case ControlSettings:
		message = wmSettings
	default:
		return false
	}
	return win.PostMessage(handle, message, 0, 0) != 0
}

func (a *App) Refresh() {
	a.postRefresh()
}

// QueueSettings applies a hot-reloaded configuration on the UI thread.
// The app layer owns config.json and does the file read on a background
// goroutine; all this does is stage the resulting plain struct and wake the
// message pump, so no file IO or cross-thread mutation happens here.
func (a *App) QueueSettings(value Settings) {
	staged := value
	a.pendingSettings.Store(&staged)
	a.postRefresh()
}

func (a *App) applyQueuedSettings(value Settings) {
	a.options.Settings = value
	setSoundVolume(value.SoundVolume)
	setSoundScheme(value.SoundScheme)
	if value.Opacity > 0 {
		a.options.Opacity = uint8(value.Opacity)
	}
	if value.AutoApproveTimeoutSec > 0 {
		a.approveTimeout = time.Duration(value.AutoApproveTimeoutSec) * time.Second
	}
	a.lang = i18n.Resolve(i18n.Setting(value.Language), i18n.SystemLocale())
	// Reading config.json again after the settings window wrote it must not
	// double-apply the palette; applyTheme is idempotent and cheap, and the
	// frame key below is what makes the repaint happen exactly once.
	applyAppearance(value.Theme, value.PerSourceColor)
	a.setTopmost()
	a.syncTuck()
	a.updateTooltip()
	// frameKey includes lang and the tuck state, so the refresh already in
	// flight repaints exactly once with the new settings.
	if a.settingsHandle.Load() != 0 {
		_ = a.paintSettings()
	}
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
	// The settings window has its own class and its own procedure, but lives on
	// this same thread and message pump. Registering it here means Run owns
	// both lifetimes, so neither can outlive the thread that created it.
	if settingsClass, ok := registerSettingsClass(instance); ok {
		defer unregisterClass(settingsClass, instance)
	}
	style := uint32(win.WS_EX_TOOLWINDOW | win.WS_EX_LAYERED)
	if a.options.Settings.AlwaysOnTop {
		style |= uint32(win.WS_EX_TOPMOST)
	}
	created := win.CreateWindowEx(
		style,
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
	// The window was created before load() knew how many sessions the
	// remembered state carries, and the expanded panel is sized from that
	// count, so publish the real frame before the first paint.
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.resize()
	a.restoreOrAnchor()
	// Which edge the restored position hugs decides how the panel grows and
	// whether it tucks, so settle that before the first paint. snapTuck
	// rather than syncTuck: starting with a slide animation would show the
	// panel fully open and then yank it away.
	a.refreshDock()
	a.snapTuck()
	if err := a.present(); err != nil {
		return fmt.Errorf("present: %w", err)
	}
	// HiddenAtLaunch is applied exactly here, once: after the first paint (so
	// the DIB exists and a later SW_SHOW has something to show), after
	// setTopmost, and before the message loop. `--show`, `--toggle` and the tray
	// entry point all stay authoritative afterwards - this is a startup
	// preference, not a mode.
	if a.options.Settings.HiddenAtLaunch && !a.options.OpenSettings {
		a.hideToTray()
	} else {
		win.ShowWindow(created, win.SW_SHOW)
	}
	// A hidden panel still needs its topmost flag ready for the moment it comes
	// back; hiding does not touch it, so set it either way.
	a.setTopmost()
	// The idle check used to fire every 500 ms and destroy the window after
	// just five quiet seconds. Combined with heartbeat-based TTLs that made
	// the panel blink out while the terminal was still working. Now it is a
	// slow safety net: 30 s polling, ten empty minutes before exit.
	win.SetTimer(created, timerIdle, 30000, 0)
	win.SetTimer(created, timerClock, 2000, 0)
	a.syncAnimation()
	if a.options.OpenSettings {
		a.openSettings()
	}
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
	case wmToggle:
		return "WM_APP+4(toggle)"
	case wmReset:
		return "WM_APP+5(reset)"
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
		if pending := a.pendingSettings.Swap(nil); pending != nil {
			a.applyQueuedSettings(*pending)
		}
		a.load()
		if len(a.sessions) > 0 {
			a.lastAvailable = time.Now()
		}
		a.syncAnimation()
		// Sessions coming and going change the expanded panel's content
		// height, and a state change can veto the tuck. Both are no-ops
		// unless something actually moved.
		a.syncTuck()
		a.applyGeometry()
		if !a.hidden {
			_ = a.present()
		}
		return 0
	case wmActivate:
		a.activatePending.Store(false)
		a.showFromTray()
		return 0
	case wmToggle:
		a.toggleExpanded()
		return 0
	case wmReset:
		a.resetPlacement()
		a.showFromTray()
		return 0
	case wmSettings:
		a.showFromTray()
		a.openSettings()
		return 0
	case wmTray:
		a.handleTrayMessage(message, lParam)
		return 0
	case wmTimer:
		return a.handleTimer(hwnd, wParam)
	case wmNCHitTest:
		return a.hitTest(lParam)
	case wmLButtonDown:
		a.beginDrag()
		return 0
	case wmMouseMove:
		if a.drag.active && wParam&mkLButton != 0 {
			a.moveDrag()
			return 0
		}
		a.trackHover()
		return 0
	case wmMouseLeave:
		a.clearHover()
		return 0
	case wmLButtonUp:
		// WM_LBUTTONUP carries client-space coordinates in lParam. Use them
		// instead of GetCursorPos: this desktop moves the cursor on its own,
		// so a re-read position can be hundreds of pixels off and would make
		// an aimed tap land on the wrong control.
		pointX := int(int16(uint16(lParam)))
		pointY := int(int16(uint16(lParam >> 16)))
		a.endDrag(pointX, pointY)
		return 0
	case wmFocusActivate:
		// Activation happens on the UI thread, which owns the foreground
		// because the user just clicked the overlay. Nothing here touches
		// any other window's input queue or z-order.
		if !platform.ActivateTarget(uintptr(wParam)) {
			a.focusFailure()
			return 0
		}
		// Keep the panel out for a moment: without this, losing hover as the
		// pointer leaves for the terminal can slide it into the screen edge,
		// which reads as "the overlay closed itself".
		a.holdTuck(2 * time.Second)
		return 0
	case wmFocusResult:
		a.focusFailure()
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
		a.syncAnimation()
		if !a.hidden {
			a.postRefresh()
		}
		return 0
	case wmClose:
		a.persistPlacement()
		// Tear the settings window down first: DestroyWindow on the overlay
		// ends the message loop, and an orphaned settings window would then
		// keep a dead class alive until the process exits.
		a.destroySettings()
		win.DestroyWindow(hwnd)
		return 0
	case wmDestroy:
		// The overlay also removes itself on idle, so this is the only
		// chance to remember where the user left it.
		a.persistPlacement()
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
		} else if time.Since(a.lastAvailable) >= 10*time.Minute {
			win.DestroyWindow(hwnd)
		}
	case timerClock:
		if a.expanded && !a.hidden {
			a.load()
			a.postRefresh()
		}
		// Reclaim handshake files orphaned by killed hooks. The sweep only
		// touches the approvals directory and only when the feature is on;
		// a live hook deletes its own files, so anything old here is trash.
		// Twice the timeout is the grace: a request can legitimately sit
		// for one full timeout plus polling slack.
		if a.options.Settings.AutoApprove {
			sweepApprovals(2*a.approveTimeout + 30*time.Second)
		}
		// A tuck hold expires on a wall-clock deadline, and the animation
		// timer is not running while the panel just sits there, so this is
		// what actually notices that the hold is over. It also re-checks the
		// hover in case a WM_MOUSELEAVE was dropped.
		a.verifyHover()
		a.syncTuck()
		a.syncAnimation()
	case timerAnimation:
		if a.hidden {
			return 0
		}
		a.load()
		// The state can go quiet without a wmRefresh (a session simply
		// expires), so re-check here and stop the timer when nothing is
		// moving or pulsing any more.
		a.stepSlide()
		a.syncTuck()
		a.syncAnimation()
		a.postRefresh()
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
	case wmLButtonUp, wmLButtonDblClk, ninBalloonUserClick:
		a.showFromTray()
	case wmRButtonUp, wmContextMenu:
		a.showContextMenu(true)
	}
}

// trackHover arms the one-shot WM_MOUSELEAVE notification the first time the
// cursor enters the panel. Without TrackMouseEvent there is no leave message,
// so the highlight would stay on forever after the first hover.
func (a *App) trackHover() {
	if a.hovering {
		return
	}
	a.hovering = true
	track := win.TRACKMOUSEEVENT{
		CbSize:    uint32(unsafe.Sizeof(win.TRACKMOUSEEVENT{})),
		DwFlags:   win.TME_LEAVE,
		HwndTrack: a.window(),
	}
	win.TrackMouseEvent(&track)
	a.syncTuck()
	a.postRefresh()
}

// verifyHover checks that the cursor really is still on the panel and clears
// the highlight if it is not.
//
// WM_MOUSELEAVE is delivered by TrackMouseEvent, but it is not guaranteed when
// the pointer is warped rather than moved - a remote desktop, a synthetic
// SetCursorPos, or a display change can all drop it. The cost of a missed one
// is that the panel stays out on screen indefinitely, which is exactly what
// the tuck is supposed to prevent. This is the cheap backstop, and it is why
// the tuck never depends on a single notification arriving.
func (a *App) verifyHover() {
	if !a.hovering || a.drag.active || a.hidden {
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
	if cursor.X < bounds.Left || cursor.X >= bounds.Right || cursor.Y < bounds.Top || cursor.Y >= bounds.Bottom {
		a.clearHover()
	}
}

func (a *App) clearHover() {
	if !a.hovering {
		return
	}
	a.hovering = false
	a.syncTuck()
	a.postRefresh()
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

// beginDrag arms a press. The whole panel is draggable, including the
// expanded body: the old code only allowed the top 40 units so that clicks on
// the session rows could be reserved for something else, but nothing was ever
// wired to them, and the result was a panel that looked grabbable and was not.
func (a *App) beginDrag() {
	// Complete any slide first: the drag maths works from the frame position,
	// and a frame that is deliberately hanging off the screen edge would give
	// it a nonsense origin.
	a.untuckNow()
	var cursor win.POINT
	if !win.GetCursorPos(&cursor) {
		return
	}
	var bounds win.RECT
	if !win.GetWindowRect(a.window(), &bounds) {
		return
	}
	a.pressed = true
	a.syncAnimation()
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
	win.SetWindowPos(a.window(), a.zOrder(), a.x, a.y, 0, 0, win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
}

// zOrder maps the AlwaysOnTop preference onto the SetWindowPos insert-after
// handle, so a user who turns the tray switch off gets a normal window that
// other apps can cover.
func (a *App) zOrder() win.HWND {
	if a.options.Settings.AlwaysOnTop {
		return win.HWND_TOPMOST
	}
	return win.HWND_NOTOPMOST
}

// releaseAction is what a finished drag should trigger.
type releaseAction int

const (
	releaseNothing releaseAction = iota
	releaseToggle
)

// decideRelease maps a completed press to the action it triggers. It is pure
// so the invariant "a drag that moved must never also toggle the panel" is
// locked down by a test.
//
// The rule is deliberately the same in both directions - a tap toggles, a drag
// moves - so there is nothing to learn and nothing to aim at. The old code
// read a.drag.moved AFTER calling ReleaseCapture, which dispatches
// WM_CAPTURECHANGED synchronously to this same window and clears a.drag, so
// moved was always observed as false: every drag additionally toggled the
// panel, which is what "cannot drag the pet" looked like from the outside. It
// then also required a tap to land on a 20-unit close button to collapse,
// which is what "the panel will not close" looked like.
func decideRelease(active, moved bool) releaseAction {
	if !active || moved {
		return releaseNothing
	}
	return releaseToggle
}

func (a *App) endDrag(x, y int) {
	if !a.drag.active {
		a.pressed = false
		return
	}
	// Decide before touching capture: see decideRelease.
	action := decideRelease(a.drag.active, a.drag.moved)
	moved := a.drag.moved
	a.drag = dragState{}
	a.pressed = false
	win.ReleaseCapture()
	a.syncAnimation()
	a.postRefresh()
	if moved {
		// The panel has a new home. Pull it back on screen first - a drag can
		// otherwise leave it somewhere unreachable - then work out which edge
		// it now hugs.
		a.clampHome()
		a.refreshDock()
		a.persistPlacement()
	}
	if action == releaseToggle {
		// An aimed tap on a jump arrow switches to that terminal instead of
		// collapsing the panel. It takes priority over the auto-approve veto
		// because it is an explicit action on a distinct control.
		_, hasJump := mostUrgentJump(a.sessions)
		if row, hit := focusHit(x, y, a.width, a.height, a.expanded, hasJump, len(a.sessions)); hit && a.jumpTo(row) {
			a.syncTuck()
			return
		}
		// With auto-approve on, a tap on the expanded panel is a veto, not
		// a collapse: the hook is about to approve on silence, so the only
		// thing a tap can usefully mean is "don't". Taps on the compact
		// strip still toggle, so the pending request stays one tap away.
		if a.expanded && a.options.Settings.AutoApprove {
			if key := a.approvals.urgentKey(a.sessions); key != "" {
				a.approvals.deny(key, a.sessions)
				a.postRefresh()
				a.syncTuck()
				return
			}
		}
		a.toggleExpanded()
	}
	a.syncTuck()
}

// jumpTo brings the terminal that owns one session to the foreground.
//
// Resolving the window walks other processes' windows (EnumWindows /
// GetWindowText), which can stall if any of them is hung. That must never
// happen on the UI thread, so the lookup runs in a goroutine and posts its
// result back. row < 0 means the compact one-tap bar (most urgent session).
// It returns false when there is no session to jump to, so the caller can
// fall through to the ordinary click behaviour.
func (a *App) jumpTo(row int) bool {
	var target session.Session
	if row < 0 {
		picked, ok := mostUrgentJump(a.sessions)
		if !ok {
			return false
		}
		target = picked
	} else {
		if row >= len(a.sessions) {
			return false
		}
		target = a.sessions[row]
	}
	request := platform.FocusTarget{
		PID:       target.PID,
		ParentPID: target.ParentPID,
		Title:     target.Title,
		Project:   target.Project,
		Source:    target.Source,
		Terminal:  target.Terminal,
	}
	handle := uintptr(a.window())
	go func() {
		// Only the lookup runs off-thread (EnumWindows touches other
		// processes). The resolved HWND is posted back so activation happens
		// on the UI thread, which holds the foreground after the click.
		target, _, ok := platform.ResolveTarget(request)
		if handle == 0 {
			return
		}
		if !ok || target == 0 {
			win.PostMessage(win.HWND(handle), wmFocusResult, 0, 0)
			return
		}
		win.PostMessage(win.HWND(handle), wmFocusActivate, target, 0)
	}()
	return true
}

// focusFailure tells the user we could not switch terminals. It never forces
// a foreground change: silently yanking focus away from whatever the user is
// doing would be worse than a small notification.
func (a *App) focusFailure() {
	beep(beepExclamation)
	a.balloonText(i18n.T(a.lang, i18n.KeyBalloonJumpTitle),
		i18n.T(a.lang, i18n.KeyBalloonJumpBody), win.NIIF_WARNING)
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
	a.applyGeometry()
	a.setTopmost()
	a.postRefresh()
	a.persistPlacement()
}

func (a *App) toggleExpanded() {
	a.expanded = !a.expanded
	a.load()
	// Publish the new geometry synchronously (resize/clamp only position
	// the frame; they never block), but paint asynchronously: present()
	// issues UpdateLayeredWindow which has stalled for seconds under
	// system load, and stalling an input handler is exactly what Windows
	// reports as AppHang (busy cursor, ghost window, click "crash").
	a.applyGeometry()
	a.setTopmost()
	a.postRefresh()
	a.persistPlacement()
	a.syncTuck()
}

// placement snapshots the geometry worth remembering across launches.
func (a *App) placement() Placement {
	return Placement{Placed: true, X: a.x, Y: a.y, Zoom: a.zoom, Expanded: a.expanded}
}

// persistPlacement hands the current geometry to the app layer. It is called
// only on settled transitions (drag release, zoom step, expand toggle) and
// once at shutdown, never per mouse-move, so a drag writes one file.
func (a *App) persistPlacement() {
	if a.options.OnPlacementChanged == nil || a.handle.Load() == 0 {
		return
	}
	a.options.OnPlacementChanged(a.placement())
}

// restoreOrAnchor puts the window back where the user last left it, or falls
// back to the right-edge anchor on a first run. The restored position is still
// clamped to the monitor, so a saved coordinate from a disconnected display
// cannot strand the overlay off-screen.
func (a *App) restoreOrAnchor() {
	initial := a.options.Initial
	if !initial.Placed {
		a.positionForWorkArea()
		return
	}
	a.x = initial.X
	a.y = initial.Y
	a.clampHome()
}

// clampHome pulls the remembered position back onto a monitor. This is the only
// place that writes a.x/a.y from a clamp, and it only runs for positions that
// cannot know the current display layout: a restored state file, or --reset.
func (a *App) clampHome() {
	work, ok := a.workArea()
	if !ok {
		return
	}
	a.clampToWorkArea(work, int32(a.logicalWidth()), int32(a.logicalHeight()))
}

// settleHome re-anchors the remembered position to where the frame visually is
// and recomputes the dock. Used when the panel's own size changed underneath it
// (a DPI change), where "stay where it looks like it is" is the right answer.
func (a *App) settleHome() {
	a.x, a.y = a.frameX(), a.frameY()
	a.refreshDock()
}

// syncAnimation installs the 50 ms pulse timer only while something is
// actually pulsing. Before this, the timer ran at 20 Hz forever, waking the UI
// thread 20 times a second to redraw an identical idle frame.
//
// A press is deliberately not an animation: the press feedback is a static
// inset and a dimmed orb, so spinning up a repaint loop to draw the same frame
// 20 times a second while the button is held was pure waste.
func (a *App) syncAnimation() {
	handle := a.handle.Load()
	if handle == 0 {
		return
	}
	want := !a.hidden && (a.state == session.StateWaitingPermission || a.state == session.StateError || a.tuck != a.tuckTarget)
	interval := uint32(pulseInterval)
	if a.tuck != a.tuckTarget {
		interval = slideInterval
	}
	if want && (interval != a.timerInterval || !a.animating) {
		a.timerInterval = interval
		a.animating = true
		win.SetTimer(win.HWND(handle), timerAnimation, interval, 0)
		return
	}
	if !want {
		a.animating = false
		win.KillTimer(win.HWND(handle), timerAnimation)
	}
}

func (a *App) showContextMenu(fromTray bool) {
	menu := win.CreatePopupMenu()
	if menu == 0 {
		return
	}
	defer win.DestroyMenu(menu)
	lang := a.lang
	position := uint32(0)
	if fromTray {
		a.appendMenu(menu, position, menuShow, i18n.T(lang, i18n.KeyTrayShow))
		position++
	} else {
		a.appendMenu(menu, position, menuHide, i18n.T(lang, i18n.KeyTrayHide))
		position++
	}
	a.appendMenu(menu, position, menuReset, i18n.T(lang, i18n.KeyTrayReset))
	position++
	a.appendMenu(menu, position, menuSettings, i18n.T(lang, i18n.KeyTraySettings))
	position++
	a.appendSeparator(menu, position)
	position++
	a.appendMenu(menu, position, menuNotifications, checkLabel(i18n.T(lang, i18n.KeyTrayNotify), a.options.Settings.Notifications))
	position++
	a.appendMenu(menu, position, menuSound, checkLabel(i18n.T(lang, i18n.KeyTraySound), a.options.Settings.Sound))
	position++
	a.appendMenu(menu, position, menuSoundVolume, fmt.Sprintf("%s · %d%%", i18n.T(lang, i18n.KeyTraySoundVolume), a.options.Settings.SoundVolume))
	position++
	a.appendMenu(menu, position, menuNotifyFinish, checkLabel(i18n.T(lang, i18n.KeyTrayNotifyFinish), a.options.Settings.NotifyOnFinish))
	position++
	a.appendMenu(menu, position, menuTopmost, checkLabel(i18n.T(lang, i18n.KeyTrayTopmost), a.options.Settings.AlwaysOnTop))
	position++
	a.appendMenu(menu, position, menuEdgeRetract, checkLabel(i18n.T(lang, i18n.KeyTrayRetract), a.options.Settings.EdgeRetract))
	position++
	a.appendMenu(menu, position, menuAutoApprove, checkLabel(i18n.T(lang, i18n.KeyTrayAutoApprove), a.options.Settings.AutoApprove))
	position++
	a.appendMenu(menu, position, menuLangAuto, langLabel(i18n.T(lang, i18n.KeyTrayLangTitle)+" · "+i18n.T(lang, i18n.KeyTrayLangAuto), a.options.Settings.Language == "" || i18n.NormalizeSetting(a.options.Settings.Language) == i18n.SettingAuto))
	position++
	a.appendMenu(menu, position, menuLangEn, langLabel("English", i18n.NormalizeSetting(a.options.Settings.Language) == i18n.SettingEn))
	position++
	a.appendMenu(menu, position, menuLangZh, langLabel("简体中文", i18n.NormalizeSetting(a.options.Settings.Language) == i18n.SettingZhCN))
	position++
	a.appendMenu(menu, position, menuAutoStart, checkLabel(i18n.T(lang, i18n.KeyTrayAutostart), a.options.AutoStart))
	position++
	a.appendMenu(menu, position, menuExit, i18n.T(lang, i18n.KeyTrayQuit))
	var cursor win.POINT
	win.GetCursorPos(&cursor)
	win.SetForegroundWindow(a.window())
	command := win.TrackPopupMenu(menu, win.TPM_RIGHTBUTTON|win.TPM_RETURNCMD|win.TPM_NONOTIFY, cursor.X, cursor.Y, 0, a.window(), nil)
	switch command {
	case menuShow:
		a.showFromTray()
	case menuHide:
		a.hideToTray()
	case menuReset:
		a.resetPlacement()
	case menuSettings:
		a.openSettings()
	case menuNotifications:
		a.options.Settings.Notifications = !a.options.Settings.Notifications
		a.publishSettings()
	case menuSound:
		a.options.Settings.Sound = !a.options.Settings.Sound
		// Play the tone immediately on enable: the switch is otherwise
		// silent until the next session happens to need attention.
		if a.options.Settings.Sound {
			beep(beepExclamation)
		}
		a.publishSettings()
	case menuSoundVolume:
		// Three useful steps rather than a free slider: clicking cycles
		// 60 -> 85 -> 100 -> 60 and previews the new loudness at once.
		switch {
		case a.options.Settings.SoundVolume < 70:
			a.options.Settings.SoundVolume = 85
		case a.options.Settings.SoundVolume < 95:
			a.options.Settings.SoundVolume = 100
		default:
			a.options.Settings.SoundVolume = 60
		}
		setSoundVolume(a.options.Settings.SoundVolume)
		beep(beepExclamation)
		a.publishSettings()
	case menuNotifyFinish:
		a.options.Settings.NotifyOnFinish = !a.options.Settings.NotifyOnFinish
		if a.options.Settings.NotifyOnFinish {
			beep(beepFinished)
		}
		a.publishSettings()
	case menuTopmost:
		a.options.Settings.AlwaysOnTop = !a.options.Settings.AlwaysOnTop
		a.applyTopmost()
		a.publishSettings()
	case menuEdgeRetract:
		a.options.Settings.EdgeRetract = !a.options.Settings.EdgeRetract
		// Turning it on while the cursor is elsewhere should tuck immediately;
		// turning it off should bring the panel straight back.
		a.syncTuck()
		a.publishSettings()
	case menuAutoApprove:
		a.options.Settings.AutoApprove = !a.options.Settings.AutoApprove
		if !a.options.Settings.AutoApprove {
			a.approvals.update(nil, time.Now(), false, 0)
		}
		a.publishSettings()
		a.postRefresh()
	case menuLangAuto, menuLangEn, menuLangZh:
		switch command {
		case menuLangEn:
			a.options.Settings.Language = string(i18n.SettingEn)
		case menuLangZh:
			a.options.Settings.Language = string(i18n.SettingZhCN)
		default:
			a.options.Settings.Language = string(i18n.SettingAuto)
		}
		a.lang = i18n.Resolve(i18n.Setting(a.options.Settings.Language), i18n.SystemLocale())
		a.publishSettings()
		a.updateTooltip()
		a.postRefresh()
	case menuAutoStart:
		a.toggleAutoStart()
	case menuExit:
		a.persistPlacement()
		win.DestroyWindow(a.window())
	}
}

// toggleAutoStart flips the run-at-logon entry. The registry is the source of
// truth, so the tick only moves when the app layer confirms the write - a
// checkmark that disagrees with the registry would be worse than no switch.
func (a *App) toggleAutoStart() {
	if a.options.OnAutoStartChanged == nil {
		return
	}
	want := !a.options.AutoStart
	if !a.options.OnAutoStartChanged(want) {
		return
	}
	a.options.AutoStart = want
}

// checkLabel prefixes a menu entry with a tick so the two tray switches read
// their own state without a separate checked-menu plumbing.
func checkLabel(label string, on bool) string {
	if on {
		return "✓ " + label
	}
	return "　 " + label
}

// langLabel marks the active language with a filled dot, the others with a
// hollow one. Same idea as checkLabel, but a tick would wrongly suggest the
// languages are independent switches rather than one choice.
func langLabel(label string, active bool) string {
	if active {
		return "● " + label
	}
	return "○ " + label
}

// applyTopmost pushes the current AlwaysOnTop preference onto the live window.
func (a *App) applyTopmost() {
	if a.handle.Load() == 0 {
		return
	}
	win.SetWindowPos(a.window(), a.zOrder(), 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE)
}

// resetPlacement drops the remembered geometry and re-anchors to the right
// edge of the work area at 100% zoom, compact. It is the escape hatch for a
// pet dragged somewhere awkward or onto a display that no longer exists.
func (a *App) resetPlacement() {
	a.zoom = 1
	a.expanded = false
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.resize()
	a.positionForWorkArea()
	a.refreshDock()
	a.holdTuck(3 * time.Second)
	a.setTopmost()
	a.postRefresh()
	a.persistPlacement()
	a.syncTuck()
}

func (a *App) publishSettings() {
	if a.options.OnSettingsChanged != nil {
		a.options.OnSettingsChanged(a.options.Settings)
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

// appendSeparator inserts a menu divider, so the settings entry reads as its own
// group rather than another toggle in the list.
func (a *App) appendSeparator(menu win.HMENU, position uint32) {
	info := win.MENUITEMINFO{
		CbSize: uint32(unsafe.Sizeof(win.MENUITEMINFO{})),
		FMask:  win.MIIM_FTYPE,
		FType:  win.MFT_SEPARATOR,
	}
	win.InsertMenuItem(menu, position, true, &info)
}

func (a *App) hideToTray() {
	if a.handle.Load() != 0 {
		win.ShowWindow(a.window(), win.SW_HIDE)
	}
	a.hidden = true
	// A hidden window gets no WM_MOUSELEAVE, so the highlight would still be
	// on the next time the panel appears under a cursor that has long moved.
	// Snap it fully out too: it should be where the user left it when it
	// comes back, not mid-slide.
	a.hovering = false
	a.tuck, a.tuckTarget = 0, 0
	a.applyFramePosition()
	a.syncAnimation()
}

func (a *App) showFromTray() {
	handle := a.handle.Load()
	if handle == 0 {
		return
	}
	a.hidden = false
	a.holdTuck(3 * time.Second)
	a.syncTuck()
	// Put the frame back where the geometry says before revealing it. Hiding
	// snapped the slide out, and nothing since then has moved the window.
	a.applyFramePosition()
	win.ShowWindow(win.HWND(handle), win.SW_SHOW)
	a.load()
	a.syncAnimation()
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
	tip, _ := syscall.UTF16FromString(truncateUTF16(a.trayTip(), trayTipMax))
	copy(data.SzTip[:], tip)
	return data
}

// trayTip is the hover text: the aggregate state plus how many sessions are
// behind it, so the tray icon is informative without opening the panel.
func (a *App) trayTip() string {
	if len(a.sessions) == 0 {
		if a.lang == i18n.LangZhCN {
			return windowTitle + " · 待机"
		}
		return windowTitle + " · Idle"
	}
	return fmt.Sprintf("%s · %d %s · %s", windowTitle, len(a.sessions), i18n.T(a.lang, i18n.KeyPanelSessions), stateLabel(a.state, a.lang))
}

// updateTooltip pushes a refreshed hover text to the shell. Cheap, but only
// worth doing when the tray icon actually exists.
func (a *App) updateTooltip() {
	if !a.trayAdded {
		return
	}
	data := a.notifyIcon()
	data.UFlags = win.NIF_TIP
	win.Shell_NotifyIcon(win.NIM_MODIFY, &data)
}

// alertLevel ranks what just happened across all sessions in one load.
type alertLevel uint8

const (
	alertNone alertLevel = iota
	alertFinished
	alertUrgent
)

// busyState is true for states that mean "the agent is doing something".
// Leaving one of these for idle/closed is a real completion, whereas an idle
// session that stays idle is not.
func busyState(state session.State) bool {
	switch state {
	case session.StateWorking, session.StateToolUse, session.StateThinking,
		session.StateWaitingPermission, session.StateError:
		return true
	}
	return false
}

// transitionLevel is the pure decision behind alertOnTransitions. A new
// session that opens already in an alert state does need announcing, but a
// plain new working session must not: that was the "it only pops up when I
// start a new session" complaint.
func transitionLevel(previous session.State, existed bool, current session.State) alertLevel {
	if !existed {
		if current == session.StateWaitingPermission || current == session.StateError {
			return alertUrgent
		}
		return alertNone
	}
	if previous == current {
		return alertNone
	}
	switch current {
	case session.StateWaitingPermission, session.StateError:
		return alertUrgent
	case session.StateIdle, session.StateClosed:
		if busyState(previous) {
			return alertFinished
		}
	}
	return alertNone
}

func sessionAlertKey(value session.Session) string {
	return fmt.Sprintf("%s\x00%s\x00%d", value.Source, value.SessionID, value.PID)
}

func currentAlertStates(items []session.Session) map[string]session.State {
	states := make(map[string]session.State, len(items))
	for _, value := range items {
		states[sessionAlertKey(value)] = value.State
	}
	return states
}

// decideAlerts is the pure half of alertOnTransitions. Keeping it separate is
// what lets a regression test reproduce the ghost-popup scenario: one session
// ending used to reset a single global key and re-announce a sibling that had
// been waiting the whole time.
func decideAlerts(previous map[string]session.State, items []session.Session) (alertLevel, *session.Session) {
	level := alertNone
	var subject *session.Session
	for index := range items {
		candidate := &items[index]
		was, existed := previous[sessionAlertKey(*candidate)]
		if found := transitionLevel(was, existed, candidate.State); found > level {
			level, subject = found, candidate
		}
	}
	// List() drops closed sessions entirely. A disappearance is the
	// strongest completion signal there is, but it must never masquerade
	// as an urgent alert from a different, still-waiting conversation.
	present := currentAlertStates(items)
	for key, was := range previous {
		if _, stillHere := present[key]; !stillHere && busyState(was) && level < alertFinished {
			level = alertFinished
		}
	}
	return level, subject
}

// alertOnTransitions fires the configured alerts on real state transitions.
// This is the point of the whole app: the overlay is glanceable, but an alert
// reaches the user when the panel is hidden, covered, or they are looking at
// another monitor.
//
// Every conversation is tracked independently. The old implementation kept a
// single global key for the *most urgent* session: when that session ended the
// key was cleared, so a still-waiting sibling re-alerted later and looked like
// it had been caused by an unrelated new session. The first load only seeds
// the map, so hot reload / restart never replays historical alerts. The sound
// and the balloon share one decision, so turning either off never changes when
// the other fires.
func (a *App) alertOnTransitions() {
	current := currentAlertStates(a.sessions)
	if a.alertStates == nil || (!a.options.Settings.Notifications && !a.options.Settings.Sound) {
		// Keep the baseline fresh even while muted, so re-enabling alerts
		// cannot replay everything that happened while they were off.
		a.alertStates = current
		return
	}

	level, subject := decideAlerts(a.alertStates, a.sessions)
	a.alertStates = current
	if level == alertNone || (level == alertFinished && !a.options.Settings.NotifyOnFinish) {
		return
	}
	if level == alertUrgent && subject != nil {
		// Per-event switches: someone who only wants to hear genuine blockers
		// can mute the "waiting for approval" tone without losing the error
		// one. The completion tone is gated by NotifyOnFinish above.
		if soundForState(a.options.Settings, subject.State) {
			beep(beepKind(subject.State))
		}
	} else if a.options.Settings.Sound {
		beep(beepFinished)
	}
	if a.options.Settings.Notifications {
		if level == alertUrgent && subject != nil {
			a.balloon(*subject)
		} else {
			a.balloonText(i18n.T(a.lang, i18n.KeyBalloonDoneTitle),
				i18n.T(a.lang, i18n.KeyBalloonDoneBody), niifInfo)
		}
	}
}

// soundForState reports whether this transition should make a noise, honouring
// both the master switch and the per-event one.
func soundForState(settings Settings, state session.State) bool {
	if !settings.Sound {
		return false
	}
	if state == session.StateError {
		return settings.SoundOnError
	}
	return settings.SoundOnWaiting
}

// beepKind maps a state onto one of the synthesised timbres: errors get the
// descending figure, everything else the rising one.
func beepKind(state session.State) uint32 {
	if state == session.StateError {
		return beepError
	}
	return beepExclamation
}

// balloon shows the tray balloon tip. It is a no-op when the tray icon is not
// registered, which happens briefly during startup and after an Explorer
// restart.
func (a *App) balloon(target session.Session) {
	title := i18n.T(a.lang, i18n.KeyBalloonWaitTitle)
	flags := uint32(win.NIIF_WARNING)
	if target.State == session.StateError {
		title = i18n.T(a.lang, i18n.KeyBalloonErrTitle)
		flags = win.NIIF_ERROR
	}
	body := projectLabel(target, a.lang)
	if reason := reasonLabel(target, a.lang); reason != "" {
		body += " · " + reason
	}
	a.balloonText(title, body, flags)
}

// balloonText shows a tray balloon with arbitrary copy. Alert balloons and
// "could not find that terminal" feedback share the same construction.
func (a *App) balloonText(title, body string, flags uint32) {
	if !a.trayAdded {
		return
	}
	data := a.notifyIcon()
	data.UFlags = win.NIF_INFO
	// Windows would otherwise play its own quiet scheme sound with the
	// balloon, and that sound is queued independently - which is how an
	// alert from a finished terminal could appear only after the next one
	// started. Our synthesised tone is the single, immediate audio cue.
	data.DwInfoFlags = flags | niifNoSound
	infoTitle, _ := syscall.UTF16FromString(truncateUTF16(title, trayTitleMax))
	infoBody, _ := syscall.UTF16FromString(truncateUTF16(body, trayInfoMax))
	copy(data.SzInfoTitle[:], infoTitle)
	copy(data.SzInfo[:], infoBody)
	win.Shell_NotifyIcon(win.NIM_MODIFY, &data)
}

// truncateUTF16 limits a string to capacity-1 UTF-16 code units so the NUL
// terminator the copy leaves behind is always in range.
func truncateUTF16(value string, capacity int) string {
	encoded, err := syscall.UTF16FromString(value)
	if err != nil || len(encoded) <= capacity {
		return value
	}
	return syscall.UTF16ToString(encoded[:capacity-1])
}

func (a *App) present() error {
	width, height := a.logicalWidth(), a.logicalHeight()
	now := time.Now()
	// Two frames with the same key are pixel-identical and in the same place,
	// so the whole raster can be skipped. This matters because reporters push
	// on every statusline render and most pushes repeat the same state - which
	// used to mean a full GDI repaint of the panel several times a second for
	// nothing.
	key := frameKey(width, height, a.digest, a.state, len(a.sessions), a.expanded, a.pressed, a.hovering, a.tuck, a.animating, now, a.lang, a.approvals.maxRemainingSeconds(now), a.options.Settings.Theme, a.options.Settings.PerSourceColor)
	if key == a.lastFrame && a.surface != nil {
		return nil
	}
	opacity := a.options.Opacity
	// A breathing alert changes no pixels: reuse the rendered bitmap and just
	// vary the layered window's constant alpha. This is what keeps a parked
	// permission prompt from burning a core on GDI work forever.
	reused := a.surface != nil && purePulseFrame(a.lastFrame, key, a.state, a.tuck, a.tuckTarget)
	if reused {
		opacity = pulseOpacity(a.options.Opacity, a.state, now)
	} else {
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
		drawVerticalPanel(target, panelState{
			Items:     a.sessions,
			Now:       time.Now(),
			Expanded:  a.expanded,
			Pressed:   a.pressed,
			Hovered:   a.hovering,
			Lang:      a.lang,
			Countdown: a.approvals.views(time.Now()),
			Tuck:      a.tuck,
			Sliver:    a.sliverWidth(),
			Dock:      a.dock,
		})
	}
	a.lastFrame = key
	screenDC := win.GetDC(0)
	if screenDC == 0 {
		return syscall.EINVAL
	}
	defer win.ReleaseDC(0, screenDC)
	destination := win.POINT{X: a.frameX(), Y: a.frameY()}
	dimensions := win.SIZE{CX: int32(width), CY: int32(height)}
	if err := updateLayeredWindow(a.window(), screenDC, &destination, dimensions, a.surface.dc, opacity); err != nil {
		return err
	}
	if !reused {
		a.width = width
		a.height = height
		if len(a.hit) != width*height {
			a.hit = make([]byte, width*height)
		}
		a.surface.hitMapInto(a.hit)
	}
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
	a.digest = sessionDigest(a.sessions)
	// Reconcile the auto-approve countdowns with what is actually waiting.
	// Sessions that left waiting (approved in the terminal, denied, closed,
	// expired) lose their countdowns here, on the only path that observes
	// store changes.
	a.approvals.update(a.sessions, time.Now(), a.options.Settings.AutoApprove, a.approveTimeout)
	// Both are cheap no-ops unless the tray icon exists, and both read only
	// the freshly loaded snapshot.
	a.updateTooltip()
	a.alertOnTransitions()
}

func (a *App) setTopmost() {
	if a.handle.Load() == 0 {
		return
	}
	win.SetWindowPos(a.window(), a.zOrder(), 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE)
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
	base := compactWidth
	if a.expanded {
		base = expandedWidth
	}
	return scaledDimension(base, a.dpi, a.zoom)
}

// logicalHeight is the compact strip's fixed height, or the expanded panel's
// content height. Sizing the expanded panel from the session count is what
// lets it actually list what it has instead of hiding everything past the
// second row.
func (a *App) logicalHeight() int {
	base := compactHeight
	if a.expanded {
		base = expandedPanelHeight(len(a.sessions))
	}
	return scaledDimension(base, a.dpi, a.zoom)
}

// applyGeometry publishes the current logical size to the frame.
//
// It does not move a.x/a.y. Where the frame ends up is derived on demand by
// frameX/frameY from the dock the panel settled at, which is what makes an
// expand/collapse round trip exact: the position is rebuilt from the screen
// edge and the remembered gap rather than adjusted in place, so a clamp that
// fires on the wide frame cannot accumulate into a drift.
func (a *App) applyGeometry() {
	width, height := a.logicalWidth(), a.logicalHeight()
	if width == a.width && height == a.height {
		return
	}
	a.width, a.height = width, height
	a.resize()
	a.applyFramePosition()
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
	a.applyGeometry()
	// The panel just changed physical size, so re-anchor it where it looks like
	// it is and work out the dock again.
	a.settleHome()
	a.setTopmost()
	a.postRefresh()
	a.persistPlacement()
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
