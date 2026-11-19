//go:build darwin

package ui

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"github.com/ebitengine/purego/objc"
)

const (
	nsWindowStyleMaskBorderless         uintptr = 0
	nsBackingStoreBuffered              uintptr = 2
	nsApplicationActivationPolicyAccess uintptr = 1
	nsFloatingWindowLevel               int     = 3
	nsNormalWindowLevel                 int     = 0
	nsWindowCollectionJoinAllSpaces     uintptr = 1 << 0
	nsWindowCollectionFullScreenAux     uintptr = 1 << 8
	nsImageScaleAxesIndependently       uintptr = 1
	nsVariableStatusItemLength          float64 = -1.0
	nsEventMaskAny                      uint64  = ^uint64(0)
	nsEventTypeLeftMouseDown            int     = 1
	nsEventTypeLeftMouseUp              int     = 2
	nsEventTypeRightMouseDown           int     = 3
	nsEventTypeMouseMoved               int     = 5
	nsEventTypeLeftMouseDragged         int     = 6
	nsEventTypeKeyDown                  int     = 10
	nsEventTypeApplicationDefined       int     = 15
	nsEventTypeScrollWheel              int     = 22
	nsEventModifierFlagCommand          uint64  = 1 << 20
)

var (
	darwinClassesOnce    sync.Once
	darwinWindowClass    objc.Class
	darwinMenuClass      objc.Class
	darwinDefaultRunMode objc.ID

	darwinMenuMu     sync.Mutex
	darwinMenuSink   chan ControlAction
	darwinMenuWakeup func()
)

func ensureDarwinAppClasses() error {
	var initErr error
	darwinClassesOnce.Do(func() {
		if !ensureCocoaLoaded() {
			initErr = errors.New("appkit framework unavailable")
			return
		}
		darwinDefaultRunMode = nsStringFromGo("kCFRunLoopDefaultMode")

		nsWindow := objc.GetClass("NSWindow")
		if nsWindow == 0 {
			initErr = errors.New("NSWindow class unavailable")
			return
		}
		winCls, err := objc.RegisterClass(
			"AITerminalPetWindow",
			nsWindow,
			nil,
			nil,
			[]objc.MethodDef{
				{
					Cmd: objc.RegisterName("canBecomeKeyWindow"),
					Fn: func(self objc.ID, cmd objc.SEL) bool {
						return true
					},
				},
			},
		)
		if err != nil {
			initErr = err
			return
		}
		darwinWindowClass = winCls

		nsObject := objc.GetClass("NSObject")
		if nsObject == 0 {
			initErr = errors.New("NSObject class unavailable")
			return
		}
		sendMenuAction := func(action ControlAction) {
			darwinMenuMu.Lock()
			sink := darwinMenuSink
			wake := darwinMenuWakeup
			darwinMenuMu.Unlock()
			if sink != nil {
				select {
				case sink <- action:
				default:
				}
			}
			if wake != nil {
				wake()
			}
		}
		menuCls, err := objc.RegisterClass(
			"AITerminalPetMenuTarget",
			nsObject,
			nil,
			nil,
			[]objc.MethodDef{
				{
					Cmd: objc.RegisterName("togglePanel:"),
					Fn: func(self objc.ID, cmd objc.SEL, sender objc.ID) {
						sendMenuAction(ControlToggle)
					},
				},
				{
					Cmd: objc.RegisterName("resetPanel:"),
					Fn: func(self objc.ID, cmd objc.SEL, sender objc.ID) {
						sendMenuAction(ControlReset)
					},
				},
				{
					Cmd: objc.RegisterName("quitApp:"),
					Fn: func(self objc.ID, cmd objc.SEL, sender objc.ID) {
						sendMenuAction(ControlQuit)
					},
				},
			},
		)
		if err != nil {
			initErr = err
			return
		}
		darwinMenuClass = menuCls
	})
	return initErr
}

// App runs the floating status panel as a pure-Go Cocoa/AppKit floating overlay.
type App struct {
	values          *store.Store
	options         Options
	pendingSettings atomic.Pointer[Settings]

	refreshCh  chan struct{}
	activateCh chan struct{}
	closeCh    chan struct{}
	focusCh    chan focusRequest
	closeOnce  sync.Once
	closed     atomic.Bool
	cocoaReady atomic.Bool

	dpi      uint32
	zoom     float64
	expanded bool
	hidden   bool
	pressed  bool
	hovering bool
	x        int32
	y        int32
	width    int
	height   int

	dock       dockSide
	dockEdge   int32
	dockGap    int32
	tuck       float64
	tuckTarget float64

	lastFrame          renderedKey
	lastAppliedOpacity uint8
	slideFrom          float64
	slideStart         time.Time
	tuckHoldUntil      time.Time

	sessions []session.Session
	state    session.State
	revision uint64
	digest   uint64
	hit      []byte

	surface *canvas
	toast   copyToast
	drag    dragState
	layouts map[string]LayoutPosition

	animating bool

	alertStates     map[string]alertFacts
	lastAlertAt     map[string]time.Time
	lastAlertRecord alertRecord
	attendance      *attendanceSampler
	mutes           *muteSet
	unread          *unreadSet
	approvals       *approvalTracker
	approveTimeout  time.Duration
	lastAvailable   time.Time
	contextRow      int
	keyboard        keyboardNav
	lang            i18n.Lang

	nsApp         objc.ID
	window        objc.ID
	imageView     objc.ID
	statusItem    objc.ID
	statusMenu    objc.ID
	menuTarget    objc.ID
	wakeupEvent   objc.ID
	primaryTop    float64
	displaysCache []displayRect
	displaysAt    time.Time
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
		refreshCh:      make(chan struct{}, 1),
		activateCh:     make(chan struct{}, 1),
		closeCh:        make(chan struct{}),
		focusCh:        make(chan focusRequest, 4),
		dpi:            96,
		zoom:           zoom,
		expanded:       options.Initial.Expanded,
		layouts:        cloneLayoutsWithCurrent(options.Initial.Layouts, "", 0, 0),
		state:          session.StateIdle,
		revision:       ^uint64(0),
		lastAvailable:  time.Now(),
		lang:           i18n.Resolve(i18n.Setting(options.Settings.Language), i18n.SystemLocale()),
		approvals:      newApprovalTracker(),
		approveTimeout: time.Duration(timeout) * time.Second,
		mutes:          newMuteSet(options.Settings.MutedSessions),
		unread:         newUnreadSet(),
		lastAlertAt:    make(map[string]time.Time, 4),
		contextRow:     -1,
	}
}

func (a *App) Refresh() {
	a.postRefresh()
}

func (a *App) Activate() {
	if a.closed.Load() {
		return
	}
	select {
	case a.activateCh <- struct{}{}:
	default:
	}
	a.postCocoaWakeup()
}

func (a *App) Close() {
	a.closed.Store(true)
	a.closeOnce.Do(func() {
		close(a.closeCh)
	})
	a.postCocoaWakeup()
}

func (a *App) QueueSettings(value Settings) {
	staged := value
	a.pendingSettings.Store(&staged)
	a.postRefresh()
}

func (a *App) postRefresh() {
	if a.closed.Load() {
		return
	}
	select {
	case a.refreshCh <- struct{}{}:
	default:
	}
	a.postCocoaWakeup()
}

func (a *App) postCocoaWakeup() {
	if !a.cocoaReady.Load() || a.nsApp == 0 || a.wakeupEvent == 0 {
		return
	}
	a.nsApp.Send(objc.RegisterName("postEvent:atStart:"), a.wakeupEvent, false)
}

func (a *App) Run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ensureDarwinAppClasses(); err != nil {
		return fmt.Errorf("appkit init: %w", err)
	}

	outerPool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
	if outerPool != 0 {
		defer outerPool.Send(objc.RegisterName("drain"))
	}

	if err := a.initCocoa(); err != nil {
		return err
	}
	defer a.teardownCocoa()

	controlCh := make(chan ControlAction, 16)
	darwinMenuMu.Lock()
	darwinMenuSink = controlCh
	darwinMenuWakeup = a.postCocoaWakeup
	darwinMenuMu.Unlock()
	defer func() {
		darwinMenuMu.Lock()
		darwinMenuSink = nil
		darwinMenuWakeup = nil
		darwinMenuMu.Unlock()
	}()

	if ctl, ctlErr := startControlListener(func(action ControlAction) {
		select {
		case controlCh <- action:
		default:
		}
		a.postCocoaWakeup()
	}); ctlErr == nil && ctl != nil {
		defer ctl.Close()
	}

	a.load()
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.restoreOrAnchor()
	a.refreshDock()
	a.snapTuck()
	a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
	if !a.options.Settings.HiddenAtLaunch {
		a.window.Send(objc.RegisterName("orderFrontRegardless"))
		a.applyTopmost(a.options.Settings.AlwaysOnTop)
	} else {
		a.hidden = true
		a.window.Send(objc.RegisterName("orderOut:"), objc.ID(0))
	}
	if err := a.present(); err != nil {
		return fmt.Errorf("present: %w", err)
	}

	watchStop := make(chan struct{})
	defer close(watchStop)
	if a.attendance == nil && a.options.Presence != nil {
		a.attendance = newAttendanceSampler(a.options.Presence)
	}
	if a.attendance != nil {
		a.attendance.publish(a.sessions)
		go a.attendance.run(watchStop, func() time.Duration {
			return time.Duration(a.options.Settings.IdleSeconds) * time.Second
		}, a.postRefresh)
	}

	a.cocoaReady.Store(true)
	defer a.cocoaReady.Store(false)

	lastClockTick := time.Now()
	lastIdleTick := time.Now()

	selDateWithInterval := objc.RegisterName("dateWithTimeIntervalSinceNow:")
	selNextEvent := objc.RegisterName("nextEventMatchingMask:untilDate:inMode:dequeue:")
	selUpdateWindows := objc.RegisterName("updateWindows")
	clsNSDate := objc.GetClass("NSDate")
	pastDate := objc.ID(clsNSDate).Send(objc.RegisterName("distantPast"))

	for {
		select {
		case <-a.closeCh:
			a.persistPlacement()
			return nil
		default:
		}

		loopPool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))

		for drained := true; drained; {
			drained = false
			select {
			case action := <-controlCh:
				drained = true
				if !a.handleControl(action) {
					if loopPool != 0 {
						loopPool.Send(objc.RegisterName("drain"))
					}
					a.persistPlacement()
					return nil
				}
			case <-a.activateCh:
				drained = true
				a.showWindow()
			case request := <-a.focusCh:
				drained = true
				if request.handle != 0 && platform.ActivateTarget(request.handle) {
					if request.partial {
						a.focusPartial(request, time.Now())
					}
					a.holdTuck(2 * time.Second)
					a.syncTuck()
				} else {
					a.focusFailed(request, time.Now())
				}
			case <-a.refreshCh:
				drained = true
			default:
			}
		}

		if queued := a.pendingSettings.Swap(nil); queued != nil {
			a.applyQueuedSettings(*queued)
		}

		now := time.Now()
		if now.Sub(lastIdleTick) >= 30*time.Second {
			lastIdleTick = now
			quit, anchor := idleDecision(now, a.lastAvailable, a.values.HasAvailable(), false, idleGrace)
			a.lastAvailable = anchor
			if quit {
				if loopPool != 0 {
					loopPool.Send(objc.RegisterName("drain"))
				}
				a.persistPlacement()
				return nil
			}
		}
		if now.Sub(lastClockTick) >= 2*time.Second {
			lastClockTick = now
			a.approvals.update(a.sessions, now, a.options.Settings.AutoApprove, a.approveTimeout)
			if a.options.Settings.AutoApprove {
				sweepApprovals(2*a.approveTimeout + 30*time.Second)
			}
			_ = a.present()
		}

		a.load()
		a.stepSlide()
		a.syncAnimation()
		a.pollMouseHover()
		_ = a.present()

		waitSecs := a.nextWaitDuration().Seconds()
		untilDate := objc.ID(clsNSDate).Send(selDateWithInterval, waitSecs)
		ev := a.nsApp.Send(selNextEvent, nsEventMaskAny, untilDate, darwinDefaultRunMode, true)
		for ev != 0 {
			keepRunning := a.handleCocoaEvent(ev)
			if !keepRunning {
				if loopPool != 0 {
					loopPool.Send(objc.RegisterName("drain"))
				}
				return nil
			}
			ev = a.nsApp.Send(selNextEvent, nsEventMaskAny, pastDate, darwinDefaultRunMode, true)
		}
		a.nsApp.Send(selUpdateWindows)
		if loopPool != 0 {
			loopPool.Send(objc.RegisterName("drain"))
		}
	}
}

func (a *App) nextWaitDuration() time.Duration {
	if a.animating {
		if a.tuck != a.tuckTarget {
			return slideInterval
		}
		return pulseInterval
	}
	return 250 * time.Millisecond
}

func (a *App) initCocoa() error {
	clsApp := objc.GetClass("NSApplication")
	clsNSEvent := objc.GetClass("NSEvent")
	clsImageView := objc.GetClass("NSImageView")
	clsNSColor := objc.GetClass("NSColor")
	clsStatusBar := objc.GetClass("NSStatusBar")
	clsNSMenu := objc.GetClass("NSMenu")
	clsNSMenuItem := objc.GetClass("NSMenuItem")
	if clsApp == 0 || clsNSEvent == 0 || clsImageView == 0 || clsNSColor == 0 {
		return errors.New("required AppKit classes unavailable")
	}

	app := objc.ID(clsApp).Send(objc.RegisterName("sharedApplication"))
	if app == 0 {
		return errors.New("NSApplication.sharedApplication returned nil")
	}
	a.nsApp = app
	app.Send(objc.RegisterName("setActivationPolicy:"), nsApplicationActivationPolicyAccess)
	app.Send(objc.RegisterName("finishLaunching"))
	// Only now is there an AppKit session for NSSound to belong to. beep()
	// refuses to touch AppKit before this point: a headless `go test` on macOS
	// has no application, no run loop and no audio session, and the first time
	// the tone path was ever executed (CI, 2026-11-16) it died with SIGSEGV
	// inside objc_msgSend rather than doing anything useful.
	darwinAudioReady.Store(true)

	a.refreshDisplayCache()

	initW, initH := a.logicalWidth(), a.logicalHeight()
	initRect := nsRect{Origin: nsPoint{100, 100}, Size: nsSize{float64(initW), float64(initH)}}
	win := objc.ID(darwinWindowClass).Send(objc.RegisterName("alloc")).Send(
		objc.RegisterName("initWithContentRect:styleMask:backing:defer:"),
		initRect,
		nsWindowStyleMaskBorderless,
		nsBackingStoreBuffered,
		false,
	)
	if win == 0 {
		return errors.New("NSWindow creation failed")
	}
	a.window = win

	title := nsStringFromGo(windowTitle)
	if title != 0 {
		win.Send(objc.RegisterName("setTitle:"), title)
		title.Send(objc.RegisterName("release"))
	}
	win.Send(objc.RegisterName("setOpaque:"), false)
	clearColor := objc.ID(clsNSColor).Send(objc.RegisterName("clearColor"))
	win.Send(objc.RegisterName("setBackgroundColor:"), clearColor)
	win.Send(objc.RegisterName("setHasShadow:"), false)
	win.Send(objc.RegisterName("setAcceptsMouseMovedEvents:"), true)
	win.Send(objc.RegisterName("setIgnoresMouseEvents:"), false)
	win.Send(objc.RegisterName("setMovableByWindowBackground:"), false)
	win.Send(objc.RegisterName("setCollectionBehavior:"), nsWindowCollectionJoinAllSpaces|nsWindowCollectionFullScreenAux)
	a.applyTopmost(a.options.Settings.AlwaysOnTop)

	viewRect := nsRect{Origin: nsPoint{0, 0}, Size: nsSize{float64(initW), float64(initH)}}
	imgView := objc.ID(clsImageView).Send(objc.RegisterName("alloc")).Send(
		objc.RegisterName("initWithFrame:"),
		viewRect,
	)
	if imgView == 0 {
		return errors.New("NSImageView creation failed")
	}
	imgView.Send(objc.RegisterName("setImageScaling:"), nsImageScaleAxesIndependently)
	win.Send(objc.RegisterName("setContentView:"), imgView)
	a.imageView = imgView

	// Retained application-defined wakeup event for cross-goroutine postEvent:atStart:.
	wakeEv := objc.ID(clsNSEvent).Send(
		objc.RegisterName("otherEventWithType:location:modifierFlags:timestamp:windowNumber:context:subtype:data1:data2:"),
		nsEventTypeApplicationDefined,
		nsPoint{},
		uintptr(0),
		float64(0),
		uintptr(0),
		objc.ID(0),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)
	if wakeEv != 0 {
		a.wakeupEvent = wakeEv.Send(objc.RegisterName("retain"))
	}

	if clsStatusBar != 0 && clsNSMenu != 0 && clsNSMenuItem != 0 && darwinMenuClass != 0 {
		a.initStatusBar(clsStatusBar, clsNSMenu, clsNSMenuItem)
	}
	return nil
}

func (a *App) initStatusBar(clsStatusBar, clsNSMenu, clsNSMenuItem objc.Class) {
	bar := objc.ID(clsStatusBar).Send(objc.RegisterName("systemStatusBar"))
	if bar == 0 {
		return
	}
	item := bar.Send(objc.RegisterName("statusItemWithLength:"), nsVariableStatusItemLength)
	if item == 0 {
		return
	}
	a.statusItem = item.Send(objc.RegisterName("retain"))
	btn := item.Send(objc.RegisterName("button"))
	if btn != 0 {
		label := nsStringFromGo("Pet")
		if label != 0 {
			btn.Send(objc.RegisterName("setTitle:"), label)
			label.Send(objc.RegisterName("release"))
		}
	}

	target := objc.ID(darwinMenuClass).Send(objc.RegisterName("alloc")).Send(objc.RegisterName("init"))
	a.menuTarget = target

	menu := objc.ID(clsNSMenu).Send(objc.RegisterName("alloc")).Send(objc.RegisterName("init"))
	addItem := func(title string, selName string) {
		nsTitle := nsStringFromGo(title)
		emptyKey := nsStringFromGo("")
		mi := objc.ID(clsNSMenuItem).Send(objc.RegisterName("alloc")).Send(
			objc.RegisterName("initWithTitle:action:keyEquivalent:"),
			nsTitle,
			objc.RegisterName(selName),
			emptyKey,
		)
		if nsTitle != 0 {
			nsTitle.Send(objc.RegisterName("release"))
		}
		if emptyKey != 0 {
			emptyKey.Send(objc.RegisterName("release"))
		}
		if mi != 0 {
			mi.Send(objc.RegisterName("setTarget:"), target)
			menu.Send(objc.RegisterName("addItem:"), mi)
			mi.Send(objc.RegisterName("release"))
		}
	}
	addItem(i18n.T(a.lang, i18n.KeyTrayShow), "togglePanel:")
	addItem(i18n.T(a.lang, i18n.KeyTrayReset), "resetPanel:")
	menu.Send(objc.RegisterName("addItem:"), objc.ID(clsNSMenuItem).Send(objc.RegisterName("separatorItem")))
	addItem(i18n.T(a.lang, i18n.KeyTrayQuit), "quitApp:")

	item.Send(objc.RegisterName("setMenu:"), menu)
	a.statusMenu = menu
}

func (a *App) teardownCocoa() {
	if a.surface != nil {
		a.surface.close()
		a.surface = nil
	}
	if a.statusItem != 0 {
		clsStatusBar := objc.GetClass("NSStatusBar")
		if clsStatusBar != 0 {
			bar := objc.ID(clsStatusBar).Send(objc.RegisterName("systemStatusBar"))
			if bar != 0 {
				bar.Send(objc.RegisterName("removeStatusItem:"), a.statusItem)
			}
		}
		a.statusItem.Send(objc.RegisterName("release"))
		a.statusItem = 0
	}
	if a.statusMenu != 0 {
		a.statusMenu.Send(objc.RegisterName("release"))
		a.statusMenu = 0
	}
	if a.menuTarget != 0 {
		a.menuTarget.Send(objc.RegisterName("release"))
		a.menuTarget = 0
	}
	if a.wakeupEvent != 0 {
		a.wakeupEvent.Send(objc.RegisterName("release"))
		a.wakeupEvent = 0
	}
	if a.imageView != 0 {
		a.imageView.Send(objc.RegisterName("release"))
		a.imageView = 0
	}
	if a.window != 0 {
		a.window.Send(objc.RegisterName("orderOut:"), objc.ID(0))
		a.window.Send(objc.RegisterName("close"))
		a.window = 0
	}
}

func (a *App) refreshDisplayCache() []displayRect {
	clsScreen := objc.GetClass("NSScreen")
	if clsScreen == 0 {
		return nil
	}
	screens := objc.ID(clsScreen).Send(objc.RegisterName("screens"))
	if screens == 0 {
		return nil
	}
	count := objc.Send[int](screens, objc.RegisterName("count"))
	if count <= 0 {
		return nil
	}
	selObj := objc.RegisterName("objectAtIndex:")
	selFrame := objc.RegisterName("frame")
	selVisible := objc.RegisterName("visibleFrame")

	screen0 := screens.Send(selObj, uintptr(0))
	frame0 := objc.Send[nsRect](screen0, selFrame)
	a.primaryTop = frame0.Origin.Y + frame0.Size.Height

	out := make([]displayRect, 0, count)
	for i := 0; i < count; i++ {
		scr := screens.Send(selObj, uintptr(i))
		vis := objc.Send[nsRect](scr, selVisible)
		rect := cocoaScreenToDisplayRect(vis, a.primaryTop)
		if rect.Right > rect.Left && rect.Bottom > rect.Top {
			out = append(out, rect)
		}
	}
	if len(out) > 0 {
		a.displaysCache = out
		a.displaysAt = time.Now()
	}
	return out
}

func (a *App) displayWorkAreas() []displayRect {
	if len(a.displaysCache) > 0 && time.Since(a.displaysAt) < 2*time.Second {
		return a.displaysCache
	}
	return a.refreshDisplayCache()
}

func (a *App) workArea() (displayRect, bool) {
	displays := a.displayWorkAreas()
	if len(displays) == 0 {
		return displayRect{}, false
	}
	w := int32(a.width)
	h := int32(a.height)
	if w == 0 || h == 0 {
		w, h = int32(a.logicalWidth()), int32(a.logicalHeight())
	}
	return bestDisplayForPoint(displays, a.x+w/2, a.y+h/2)
}

func (a *App) applyTopmost(above bool) {
	if a.window == 0 {
		return
	}
	level := nsNormalWindowLevel
	if above {
		level = nsFloatingWindowLevel
	}
	a.window.Send(objc.RegisterName("setLevel:"), level)
}

func (a *App) applyOpacityProperty(opacity uint8) {
	if a.window == 0 || opacity == a.lastAppliedOpacity {
		return
	}
	a.lastAppliedOpacity = opacity
	a.window.Send(objc.RegisterName("setAlphaValue:"), float64(opacity)/255.0)
}

func (a *App) applyWindowBounds(x, y int32, width, height int) {
	if a.window == 0 {
		return
	}
	origin := topLeftToCocoaOrigin(x, y, height, a.primaryTop)
	rect := nsRect{
		Origin: origin,
		Size:   nsSize{Width: float64(width), Height: float64(height)},
	}
	a.window.Send(objc.RegisterName("setFrame:display:"), rect, false)
	if a.imageView != 0 {
		viewRect := nsRect{Origin: nsPoint{0, 0}, Size: nsSize{Width: float64(width), Height: float64(height)}}
		a.imageView.Send(objc.RegisterName("setFrame:"), viewRect)
	}
}

func (a *App) moveWindow(x, y int32) {
	if a.window == 0 {
		return
	}
	origin := topLeftToCocoaOrigin(x, y, a.height, a.primaryTop)
	a.window.Send(objc.RegisterName("setFrameOrigin:"), origin)
}

func (a *App) logicalWidth() int {
	if a.expanded {
		return scaledDimension(expandedWidth, a.dpi, a.zoom)
	}
	return scaledDimension(compactWidth, a.dpi, a.zoom)
}

func (a *App) logicalHeight() int {
	if a.expanded {
		return scaledDimension(expandedPanelHeight(len(a.sessions)), a.dpi, a.zoom)
	}
	return scaledDimension(compactHeight, a.dpi, a.zoom)
}

func (a *App) refreshDock() {
	work, ok := a.workArea()
	a.dock, a.dockEdge, a.dockGap = computeDock(a.x, a.logicalWidth(), a.dpi, work, ok)
}

func (a *App) sliverWidth() int {
	return panelSliverWidth(a.dpi, a.zoom)
}

func (a *App) frameX() int32 {
	work, ok := a.workArea()
	return panelFrameX(a.drag.active, a.x, a.logicalWidth(), a.dock, a.dockEdge, a.dockGap, a.tuck, a.sliverWidth(), a.dpi, work, ok)
}

func (a *App) frameY() int32 {
	work, ok := a.workArea()
	return panelFrameY(a.drag.active, a.y, a.height, a.dpi, work, ok)
}

func (a *App) wantsTuck() bool {
	if !a.options.Settings.EdgeRetract || a.dock == dockNone {
		return false
	}
	if a.hidden || a.expanded || a.pressed || a.hovering || a.drag.active {
		return false
	}
	if time.Now().Before(a.tuckHoldUntil) {
		return false
	}
	return a.state != session.StateWaitingPermission && a.state != session.StateError
}

func (a *App) holdTuck(d time.Duration) {
	a.tuckHoldUntil = time.Now().Add(d)
}

func (a *App) startSlide(target float64) {
	a.tuckTarget = target
	a.slideFrom = a.tuck
	a.slideStart = time.Now()
	a.syncAnimation()
}

func (a *App) syncTuck() {
	target := 0.0
	if a.wantsTuck() {
		target = 1.0
	}
	if target != a.tuckTarget {
		a.startSlide(target)
	}
}

func (a *App) snapTuck() {
	a.startSlide(0)
	if a.wantsTuck() {
		a.startSlide(1)
	}
	a.tuck = a.tuckTarget
	a.moveWindow(a.frameX(), a.frameY())
}

func (a *App) stepSlide() {
	if a.tuck == a.tuckTarget {
		return
	}
	a.tuck = slideTuck(a.slideFrom, a.tuckTarget, time.Since(a.slideStart))
	a.moveWindow(a.frameX(), a.frameY())
	if a.tuck == a.tuckTarget {
		a.syncAnimation()
	}
}

func (a *App) untuckNow() {
	a.holdTuck(250 * time.Millisecond)
	if a.tuck == 0 && a.tuckTarget == 0 {
		return
	}
	a.tuck = 0
	a.tuckTarget = 0
	a.moveWindow(a.frameX(), a.frameY())
}

func (a *App) restoreOrAnchor() {
	initial := a.options.Initial
	displays := a.displayWorkAreas()
	if x, y, ok := resolveInitialPlacement(initial, displays, int32(a.logicalWidth()), int32(a.logicalHeight()), a.dpi); ok {
		a.x = x
		a.y = y
		return
	}
	if !initial.Placed {
		a.positionForWorkArea()
		return
	}
	a.x = initial.X
	a.y = initial.Y
	a.clampHome()
}

func (a *App) positionForWorkArea() {
	displays := a.displayWorkAreas()
	work, ok := primaryDisplay(displays)
	if !ok {
		a.x, a.y = 40, 60
		return
	}
	a.x, a.y = defaultWorkAreaOrigin(int32(a.logicalWidth()), int32(a.logicalHeight()), a.dpi, work)
}

func (a *App) clampHome() {
	work, ok := a.workArea()
	if !ok {
		return
	}
	a.x, a.y = clampPositionToWorkArea(a.x, a.y, int32(a.logicalWidth()), int32(a.logicalHeight()), a.dpi, work)
}

func (a *App) placement() Placement {
	sig := layoutSignature(a.displayWorkAreas())
	layouts := cloneLayoutsWithCurrent(a.layouts, sig, a.x, a.y)
	a.layouts = cloneLayoutsWithCurrent(layouts, "", 0, 0)
	return Placement{
		Placed:   true,
		X:        a.x,
		Y:        a.y,
		Expanded: a.expanded,
		Zoom:     a.zoom,
		Layouts:  layouts,
	}
}

func (a *App) persistPlacement() {
	if a.options.OnPlacementChanged != nil {
		a.options.OnPlacementChanged(a.placement())
	}
}

func (a *App) publishSettings() {
	if a.options.OnSettingsChanged != nil {
		a.options.OnSettingsChanged(a.options.Settings)
	}
}

func (a *App) applyQueuedSettings(value Settings) {
	a.options.Settings = value
	if value.Opacity > 0 {
		a.options.Opacity = uint8(value.Opacity)
	}
	if value.AutoApproveTimeoutSec > 0 {
		a.approveTimeout = time.Duration(value.AutoApproveTimeoutSec) * time.Second
	}
	setSoundVolume(value.SoundVolume)
	setSoundScheme(value.SoundScheme)
	a.lang = i18n.Resolve(i18n.Setting(value.Language), i18n.SystemLocale())
	a.applyTopmost(value.AlwaysOnTop)
	applyAppearance(value.Theme, value.PerSourceColor)
	a.mutes = newMuteSet(value.MutedSessions)
	a.approvals.update(a.sessions, time.Now(), value.AutoApprove, a.approveTimeout)
	a.syncTuck()
	a.syncAnimation()
}

func (a *App) applyGeometry() {
	width, height := a.logicalWidth(), a.logicalHeight()
	if width == a.width && height == a.height {
		return
	}
	a.width, a.height = width, height
	a.applyWindowBounds(a.frameX(), a.frameY(), width, height)
}

func (a *App) load() {
	revision := a.values.Revision()
	if revision == a.revision {
		return
	}
	a.revision = revision
	a.sessions = panelRows(dedupeSessions(a.values.List()))
	a.state = session.Aggregate(a.sessions)
	a.digest = sessionDigest(a.sessions)
	a.approvals.update(a.sessions, time.Now(), a.options.Settings.AutoApprove, a.approveTimeout)
	a.keyboard.sync(a.expanded, a.sessions)
	if a.attendance != nil {
		a.attendance.publish(a.sessions)
	}
	a.alertOnTransitions()
	a.refreshDock()
	a.syncTuck()
	if !a.drag.active {
		a.clampHome()
		a.applyGeometry()
	}
}

func (a *App) alertOnTransitions() {
	now := time.Now()
	current := factsOf(a.sessions, now, a.alertStates)
	previous := a.alertStates
	a.alertStates = current
	if previous == nil {
		return
	}
	if a.unread.retain(a.sessions) {
		a.postRefresh()
	}
	var sample attendanceState
	var sampleKnown bool
	if a.attendance != nil {
		sample, sampleKnown = a.attendance.Current()
	}
	settings := attentionSettingsFrom(a.options.Settings)
	inputs := buildAlertInputs(a.sessions, a.mutes, previous, current, sample, sampleKnown)
	inputs = append(inputs, buildVanishedInputs(previous, current, sample)...)
	if len(inputs) == 0 {
		return
	}
	plan := PlanAlerts(inputs, settings, a.lastAlertAt)
	receipt := buildAlertRecord(now, inputs, plan, sample, sampleKnown)
	if !alertRecordsEqual(a.lastAlertRecord, receipt) {
		a.lastAlertRecord = receipt
		writeAlertRecord(receipt)
	}
	if plan.Balloon != nil {
		a.notify(composeAlert(*plan.Balloon, a.lang))
	}
	if plan.Tone != "" {
		beep(toneKind(plan.Tone))
	}
	a.lastAlertAt = recordRememberedAlerts(a.lastAlertAt, inputs, plan, now)
	if len(plan.Unread) > 0 {
		a.unread.mark(plan.Unread)
		a.postRefresh()
	}
}

func (a *App) syncAnimation() {
	now := time.Now()
	sliding := a.tuck != a.tuckTarget
	pulsing := a.options.Settings.Animations && (a.state == session.StateWaitingPermission || a.state == session.StateError)
	toasting := a.toast.frame(now, a.sessions).visible()
	a.animating = !a.hidden && !a.drag.active && (sliding || pulsing || a.approvals.maxRemainingSeconds(now) >= 0 || toasting)
}

func (a *App) present() error {
	if a.window == 0 || a.imageView == 0 {
		return nil
	}
	width, height := a.logicalWidth(), a.logicalHeight()
	now := time.Now()
	key := frameKey(width, height, a.digest, a.state, len(a.sessions), a.expanded, a.pressed, a.hovering, a.tuck, a.animating, now, a.lang, a.approvals.maxRemainingSeconds(now), a.options.Settings.Theme, a.options.Settings.PerSourceColor, unreadBits(a.sessions, a.unread), mutedBits(a.sessions, a.mutes), a.contextRow)
	toast := a.toast.frame(now, a.sessions)
	key.withToast(toast)
	selectedRow, hasSelected := a.keyboard.selected()
	key.withSelectedRow(selectedRow, hasSelected)
	if key == a.lastFrame && a.surface != nil {
		return nil
	}
	opacity := a.options.Opacity
	reused := a.surface != nil && purePulseFrame(a.lastFrame, key, a.state, a.tuck, a.tuckTarget)
	if reused {
		opacity = pulseOpacity(a.options.Opacity, !a.options.Settings.Animations, a.state, now)
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
			Items:       a.sessions,
			Now:         now,
			Expanded:    a.expanded,
			Pressed:     a.pressed,
			Hovered:     a.hovering,
			Lang:        a.lang,
			Countdown:   a.approvals.views(now),
			Unread:      a.unread,
			Muted:       a.mutes,
			HoverRow:    a.contextRow,
			HasSelected: hasSelected,
			SelectedRow: selectedRow,
			Toast:       toast,
			Tuck:        a.tuck,
			Still:       !a.options.Settings.Animations,
			Sliver:      a.sliverWidth(),
			Dock:        a.dock,
		})
		a.width = width
		a.height = height
		if len(a.hit) != width*height {
			a.hit = make([]byte, width*height)
		}
		a.surface.hitMapInto(a.hit)
	}
	a.lastFrame = key
	a.applyOpacityProperty(opacity)
	a.applyWindowBounds(a.frameX(), a.frameY(), width, height)
	if !reused && a.surface != nil {
		a.uploadSurface(a.surface.width, a.surface.height, a.surface.pixels)
	}
	return nil
}

func (a *App) uploadSurface(width, height int, pixels []byte) {
	if a.imageView == 0 || width <= 0 || height <= 0 || len(pixels) < width*height*4 {
		return
	}
	clsRep := objc.GetClass("NSBitmapImageRep")
	clsImg := objc.GetClass("NSImage")
	if clsRep == 0 || clsImg == 0 {
		return
	}
	csName := nsStringFromGo("NSDeviceRGBColorSpace")
	if csName == 0 {
		return
	}
	defer csName.Send(objc.RegisterName("release"))

	rep := objc.ID(clsRep).Send(objc.RegisterName("alloc")).Send(
		objc.RegisterName("initWithBitmapDataPlanes:pixelsWide:pixelsHigh:bitsPerSample:samplesPerPixel:hasAlpha:isPlanar:colorSpaceName:bytesPerRow:bitsPerPixel:"),
		(*byte)(nil),
		width,
		height,
		8,
		4,
		true,
		false,
		csName,
		width*4,
		32,
	)
	if rep == 0 {
		return
	}
	defer rep.Send(objc.RegisterName("release"))

	ptr := objc.Send[*byte](rep, objc.RegisterName("bitmapData"))
	if ptr == nil {
		return
	}
	rgba := bgraStraightToPremulRGBA(pixels[:width*height*4])
	dst := unsafe.Slice(ptr, len(rgba))
	copy(dst, rgba)

	img := objc.ID(clsImg).Send(objc.RegisterName("alloc")).Send(
		objc.RegisterName("initWithSize:"),
		nsSize{Width: float64(width), Height: float64(height)},
	)
	if img == 0 {
		return
	}
	defer img.Send(objc.RegisterName("release"))

	img.Send(objc.RegisterName("addRepresentation:"), rep)
	a.imageView.Send(objc.RegisterName("setImage:"), img)
}

func (a *App) handleControl(action ControlAction) bool {
	switch action {
	case ControlActivate, ControlSettings:
		a.showWindow()
	case ControlToggle:
		if a.hidden {
			a.showWindow()
		} else {
			a.hidden = true
			a.window.Send(objc.RegisterName("orderOut:"), objc.ID(0))
			a.syncAnimation()
		}
	case ControlReset:
		a.resetPlacement()
	case ControlQuit:
		return false
	}
	return true
}

func (a *App) showWindow() {
	if a.window == 0 {
		return
	}
	a.hidden = false
	a.untuckNow()
	a.window.Send(objc.RegisterName("orderFrontRegardless"))
	a.applyTopmost(a.options.Settings.AlwaysOnTop)
	a.syncAnimation()
	_ = a.present()
}

func (a *App) toggleExpanded() {
	a.expanded = !a.expanded
	a.contextRow = -1
	a.keyboard.sync(a.expanded, a.sessions)
	a.untuckNow()
	a.applyGeometry()
	a.clampHome()
	a.refreshDock()
	a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
	a.syncTuck()
	a.syncAnimation()
	a.persistPlacement()
	_ = a.present()
}

func (a *App) resetPlacement() {
	a.zoom = 1
	a.layouts = nil
	a.refreshDisplayCache()
	a.positionForWorkArea()
	a.refreshDock()
	a.untuckNow()
	a.applyGeometry()
	a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
	a.persistPlacement()
	_ = a.present()
}

func (a *App) zoomBy(delta float64) {
	if delta == 0 {
		return
	}
	next := stepZoomValue(a.zoom, delta)
	if next == a.zoom {
		return
	}
	a.zoom = next
	a.applyGeometry()
	a.clampHome()
	a.refreshDock()
	a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
	a.persistPlacement()
	_ = a.present()
}

func (a *App) toggleSessionMute(target session.Session) {
	if a.mutes.toggle(target) {
		a.unread.clearOne(target)
	}
	a.options.Settings.MutedSessions = a.mutes.list()
	a.publishSettings()
	_ = a.present()
}

func (a *App) setHover(hovered bool) {
	if a.hovering == hovered {
		return
	}
	a.hovering = hovered
	a.syncTuck()
	_ = a.present()
}

func (a *App) clearHover() {
	if !a.hovering && a.contextRow < 0 {
		return
	}
	a.hovering = false
	a.contextRow = -1
	a.syncTuck()
	_ = a.present()
}

func (a *App) pollMouseHover() {
	if a.hidden || a.drag.active || a.window == 0 {
		return
	}
	clsNSEvent := objc.GetClass("NSEvent")
	if clsNSEvent == 0 {
		return
	}
	pt := objc.Send[nsPoint](objc.ID(clsNSEvent), objc.RegisterName("mouseLocation"))
	winFrame := objc.Send[nsRect](a.window, objc.RegisterName("frame"))
	inside := pt.X >= winFrame.Origin.X &&
		pt.X < winFrame.Origin.X+winFrame.Size.Width &&
		pt.Y >= winFrame.Origin.Y &&
		pt.Y < winFrame.Origin.Y+winFrame.Size.Height
	if !inside {
		a.clearHover()
		return
	}
	a.setHover(true)
	cx, cy := cocoaWindowPointToCanvas(pt.X-winFrame.Origin.X, pt.Y-winFrame.Origin.Y, a.width, a.height)
	a.trackRowHover(cx, cy)
}

func (a *App) trackRowHover(x, y int) {
	row := -1
	if a.expanded {
		if index, ok := expandedRowHit(x, y, a.width, a.height, len(a.sessions)); ok {
			row = index
		}
	}
	if row == a.contextRow {
		return
	}
	a.contextRow = row
	_ = a.present()
}

func (a *App) beginDrag(rootX, rootY, eventX, eventY int32) {
	a.untuckNow()
	a.pressed = true
	a.syncAnimation()
	a.drag = dragState{
		active:  true,
		startX:  rootX,
		startY:  rootY,
		windowX: rootX - eventX,
		windowY: rootY - eventY,
	}
	_ = a.present()
}

func (a *App) moveDrag(rootX, rootY int32) {
	dx := rootX - a.drag.startX
	dy := rootY - a.drag.startY
	if abs32(dx)+abs32(dy) > 5 {
		a.drag.moved = true
	}
	a.x = a.drag.windowX + dx
	a.y = a.drag.windowY + dy
	a.moveWindow(a.x, a.y)
}

func (a *App) endDrag(x, y int) {
	if !a.drag.active {
		a.pressed = false
		return
	}
	action := decideRelease(a.drag.active, a.drag.moved)
	moved := a.drag.moved
	a.drag = dragState{}
	a.pressed = false
	a.syncAnimation()
	if moved {
		a.clampHome()
		a.refreshDock()
		a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
		a.persistPlacement()
	}
	if action == releaseToggle {
		a.handleTap(x, y)
	} else {
		_ = a.present()
	}
	a.syncTuck()
}

func (a *App) handleTap(x, y int) {
	inputs := buildTapInputs(a.expanded, a.options.Settings.AutoApprove, a.width, a.height, x, y, a.sessions, a.approvals)
	switch action, row := decideTap(inputs); action {
	case tapJump:
		if a.jumpTo(row) {
			a.syncTuck()
			return
		}
		a.toggleExpanded()
	case tapCopyStrip:
		if value, ok := compactCopyText(a.sessions); ok {
			a.copyTo(value, toastCompactRow)
			return
		}
		a.toggleExpanded()
	case tapVeto:
		if key := a.approvals.urgentKey(a.sessions); key != "" {
			a.approvals.deny(key, a.sessions)
			_ = a.present()
			return
		}
		a.toggleExpanded()
	case tapCopyRow:
		if value, ok := rowCopyText(a.sessions, row); ok {
			a.copyTo(value, row)
			return
		}
		a.toggleExpanded()
	default:
		a.toggleExpanded()
	}
}

func (a *App) copyTo(value string, row int) {
	identity := ""
	switch {
	case row == toastCompactRow:
		if picked, ok := mostUrgentJump(a.sessions); ok {
			identity = session.Identity(picked)
		}
	case row >= 0 && row < len(a.sessions):
		identity = session.Identity(a.sessions[row])
	}
	copied := copyToClipboard(value)
	a.toast.show(row, identity, copyToastKind(!copied), time.Now())
	if !copied {
		// Same cue as Windows: a clipboard that refused must be audible as well
		// as visible, because the pill can be missed while looking away.
		beep(beepExclamation)
	}
	a.syncAnimation()
	_ = a.present()
}

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
	if a.unread.clearOne(target) {
		a.postRefresh()
	}
	// The tap is timestamped before the resolver leaves the UI thread:
	// jumpStaleAfter measures how long the user has waited, not how long the
	// goroutine happened to run.
	requested := time.Now()
	identity := session.Identity(target)
	go func() {
		resolved, windowTitle, ok := platform.ResolveTarget(request)
		partial := ok && !platform.FocusMatchIsExact(windowTitle, target.Title)
		if !ok {
			resolved = 0
		}
		select {
		case a.focusCh <- focusRequest{handle: resolved, row: row, identity: identity, at: requested, partial: partial}:
		default:
		}
		a.postCocoaWakeup()
	}()
	return true
}

func (a *App) handlePanelKey(key panelKey) bool {
	if key == panelKeyNone {
		return false
	}
	outcome := a.keyboard.handle(key, a.expanded, a.sessions)
	if outcome.toggleExpanded {
		a.toggleExpanded()
		return true
	}
	if outcome.jump {
		if a.jumpTo(outcome.jumpRow) {
			a.syncTuck()
			return true
		}
		return false
	}
	if outcome.repainted {
		_ = a.present()
		return true
	}
	return false
}

func (a *App) handleCocoaEvent(ev objc.ID) bool {
	evType := objc.Send[int](ev, objc.RegisterName("type"))
	if evType == nsEventTypeApplicationDefined {
		return true
	}

	clsNSEvent := objc.GetClass("NSEvent")
	evWin := ev.Send(objc.RegisterName("window"))

	switch evType {
	case nsEventTypeLeftMouseDragged:
		if a.drag.active {
			pt := objc.Send[nsPoint](objc.ID(clsNSEvent), objc.RegisterName("mouseLocation"))
			rx, ry := cocoaPointToTopLeft(pt.X, pt.Y, a.primaryTop)
			a.moveDrag(rx, ry)
			return true
		}
	case nsEventTypeLeftMouseUp:
		if a.drag.active {
			loc := objc.Send[nsPoint](ev, objc.RegisterName("locationInWindow"))
			cx, cy := cocoaWindowPointToCanvas(loc.X, loc.Y, a.width, a.height)
			a.endDrag(cx, cy)
			return true
		}
	}

	if evWin != a.window {
		a.nsApp.Send(objc.RegisterName("sendEvent:"), ev)
		return true
	}

	switch evType {
	case nsEventTypeLeftMouseDown:
		a.window.Send(objc.RegisterName("makeKeyWindow"))
		loc := objc.Send[nsPoint](ev, objc.RegisterName("locationInWindow"))
		cx, cy := cocoaWindowPointToCanvas(loc.X, loc.Y, a.width, a.height)
		pt := objc.Send[nsPoint](objc.ID(clsNSEvent), objc.RegisterName("mouseLocation"))
		rx, ry := cocoaPointToTopLeft(pt.X, pt.Y, a.primaryTop)
		a.beginDrag(rx, ry, int32(cx), int32(cy))
		return true

	case nsEventTypeRightMouseDown:
		loc := objc.Send[nsPoint](ev, objc.RegisterName("locationInWindow"))
		cx, cy := cocoaWindowPointToCanvas(loc.X, loc.Y, a.width, a.height)
		if a.expanded {
			if row, ok := expandedRowHit(cx, cy, a.width, a.height, len(a.sessions)); ok && row < len(a.sessions) {
				a.toggleSessionMute(a.sessions[row])
				return true
			}
		}
		a.toggleExpanded()
		return true

	case nsEventTypeMouseMoved:
		loc := objc.Send[nsPoint](ev, objc.RegisterName("locationInWindow"))
		cx, cy := cocoaWindowPointToCanvas(loc.X, loc.Y, a.width, a.height)
		a.setHover(true)
		a.trackRowHover(cx, cy)
		return true

	case nsEventTypeScrollWheel:
		flags := objc.Send[uint64](ev, objc.RegisterName("modifierFlags"))
		if flags&nsEventModifierFlagCommand != 0 {
			delta := objc.Send[float64](ev, objc.RegisterName("scrollingDeltaY"))
			if delta > 0 {
				a.zoomBy(0.1)
			} else if delta < 0 {
				a.zoomBy(-0.1)
			}
			return true
		}

	case nsEventTypeKeyDown:
		code := objc.Send[uint16](ev, objc.RegisterName("keyCode"))
		flags := objc.Send[uint64](ev, objc.RegisterName("modifierFlags"))
		if delta, reset, ok := darwinZoomShortcut(code, flags&nsEventModifierFlagCommand != 0); ok {
			if reset {
				a.zoom = 1
				a.applyGeometry()
				a.clampHome()
				a.refreshDock()
				a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
				a.persistPlacement()
				_ = a.present()
			} else {
				a.zoomBy(delta)
			}
			return true
		}
		if a.handlePanelKey(darwinKeyCodeToPanelKey(code)) {
			return true
		}
	}

	a.nsApp.Send(objc.RegisterName("sendEvent:"), ev)
	return true
}
