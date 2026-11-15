//go:build linux

package ui

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/bigreq"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
)

const defaultX11MaxRequestBytes = 262140

// motifWMHints matches the MWM _MOTIF_WM_HINTS 5-word property so X11 /
// XWayland compositors omit title bars and borders around the floating panel.
type motifWMHints struct {
	flags       uint32
	functions   uint32
	decorations uint32
	inputMode   int32
	status      uint32
}

type x11Monitor struct {
	Left    int32
	Top     int32
	Width   uint16
	Height  uint16
	Primary bool
}

// App runs the floating status panel as a pure-Go X11 top-level utility
// window with an ARGB 32-bit surface and zero cgo / toolkit dependencies.
type App struct {
	values          *store.Store
	options         Options
	pendingSettings atomic.Pointer[Settings]

	refreshCh  chan struct{}
	activateCh chan struct{}
	closeCh    chan struct{}
	focusCh    chan uintptr
	closeOnce  sync.Once
	closed     atomic.Bool

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

	animating     bool
	timerInterval uint32

	alertStates     map[string]alertFacts
	lastAvailable   time.Time
	lang            i18n.Lang
	approvals       *approvalTracker
	approveTimeout  time.Duration
	mutes           *muteSet
	unread          *unreadSet
	attendance      *attendanceSampler
	lastAlertAt     map[string]time.Time
	lastAlertRecord alertRecord
	contextRow      int
	keyboard        keyboardNav

	conn        *xgb.Conn
	screen      *xproto.ScreenInfo
	win         xproto.Window
	gc          xproto.Gcontext
	winDepth    byte
	winVisual   xproto.Visualid
	winColormap xproto.Colormap
	hasARGB     bool
	hasRandr    bool
	maxReqBytes int

	atomClipboard             xproto.Atom
	atomTargets               xproto.Atom
	atomUTF8String            xproto.Atom
	atomString                xproto.Atom
	atomText                  xproto.Atom
	atomWMProtocols           xproto.Atom
	atomWMDelete              xproto.Atom
	atomNetWMName             xproto.Atom
	atomUTF8Type              xproto.Atom
	atomNetWMPid              xproto.Atom
	atomNetWMWinType          xproto.Atom
	atomNetWMWTypeUtil        xproto.Atom
	atomNetWMState            xproto.Atom
	atomNetWMStateAbove       xproto.Atom
	atomNetWMStateSkipTaskbar xproto.Atom
	atomNetWMStateSkipPager   xproto.Atom
	atomNetWMOpacity          xproto.Atom
	atomNetWorkarea           xproto.Atom
	atomMotifHints            xproto.Atom

	minKeycode     xproto.Keycode
	maxKeycode     xproto.Keycode
	keysymsPerCode byte
	keysyms        []xproto.Keysym

	displaysCache   []displayRect
	displaysCacheAt time.Time
}

func New(values *store.Store, options Options) *App {
	if options.Opacity == 0 {
		options.Opacity = 235
	}
	if options.Settings.Opacity > 0 {
		options.Opacity = uint8(options.Settings.Opacity)
	}
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
		focusCh:        make(chan uintptr, 4),
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
}

func (a *App) Close() {
	a.closed.Store(true)
	a.closeOnce.Do(func() {
		close(a.closeCh)
	})
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
}

func (a *App) Run() error {
	// WSLg and many minimal X11 sessions omit ~/.Xauthority; silence xgb's
	// stderr warning before dialing DISPLAY.
	xgb.Logger = log.New(io.Discard, "", 0)

	conn, err := xgb.NewConn()
	if err != nil {
		return fmt.Errorf("x11 connect: %w", err)
	}
	a.conn = conn
	defer func() {
		x11ClipboardMu.Lock()
		if x11ClipboardOwner == a {
			x11ClipboardOwner = nil
		}
		x11ClipboardMu.Unlock()
		a.releaseSurface()
		if a.gc != 0 {
			_ = xproto.FreeGCChecked(conn, a.gc).Check()
		}
		if a.win != 0 {
			_ = xproto.DestroyWindowChecked(conn, a.win).Check()
		}
		if a.hasARGB && a.winColormap != 0 {
			_ = xproto.FreeColormapChecked(conn, a.winColormap).Check()
		}
		a.win = 0
		a.gc = 0
		a.conn = nil
		conn.Close()
	}()

	if err := a.initX11(); err != nil {
		return err
	}

	controlCh := make(chan ControlAction, 8)
	if ctl, ctlErr := startControlListener(func(action ControlAction) {
		select {
		case controlCh <- action:
		default:
		}
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
		xproto.MapWindow(a.conn, a.win)
		a.applyTopmost(a.options.Settings.AlwaysOnTop)
	} else {
		a.hidden = true
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

	type x11Msg struct {
		ev  xgb.Event
		err xgb.Error
	}
	eventsCh := make(chan x11Msg, 64)
	go func() {
		defer close(eventsCh)
		for {
			ev, xErr := conn.WaitForEvent()
			if ev == nil && xErr == nil {
				return
			}
			select {
			case eventsCh <- x11Msg{ev: ev, err: xErr}:
			case <-watchStop:
				return
			}
		}
	}()

	idleTicker := time.NewTicker(30 * time.Second)
	defer idleTicker.Stop()
	clockTicker := time.NewTicker(2 * time.Second)
	defer clockTicker.Stop()
	animTimer := time.NewTimer(time.Hour)
	if !animTimer.Stop() {
		select {
		case <-animTimer.C:
		default:
		}
	}
	defer animTimer.Stop()

	resetAnimTimer := func() {
		if !a.animating || a.timerInterval == 0 {
			return
		}
		animTimer.Reset(time.Duration(a.timerInterval) * time.Millisecond)
	}
	a.syncAnimation()
	resetAnimTimer()

	for {
		select {
		case <-a.closeCh:
			a.persistPlacement()
			return nil

		case <-a.activateCh:
			a.showWindow()

		case action := <-controlCh:
			if quit := a.handleControl(action); quit {
				a.persistPlacement()
				return nil
			}
			resetAnimTimer()

		case target := <-a.focusCh:
			if target != 0 && platform.ActivateTarget(target) {
				a.holdTuck(2 * time.Second)
				a.syncTuck()
				resetAnimTimer()
			}

		case <-a.refreshCh:
			if staged := a.pendingSettings.Swap(nil); staged != nil {
				a.applyQueuedSettings(*staged)
			}
			a.load()
			a.applyGeometry()
			a.syncAnimation()
			resetAnimTimer()
			if !a.hidden {
				_ = a.present()
			}

		case <-idleTicker.C:
			quit, anchor := idleDecision(time.Now(), a.lastAvailable, a.values.HasAvailable(), false, idleGrace)
			a.lastAvailable = anchor
			if quit {
				traceExit("idle")
				a.persistPlacement()
				return nil
			}

		case <-clockTicker.C:
			if a.expanded && !a.hidden {
				a.load()
				_ = a.present()
			}
			if a.options.Settings.AutoApprove {
				sweepApprovals(2*a.approveTimeout + 30*time.Second)
			}
			a.syncTuck()
			a.syncAnimation()
			resetAnimTimer()

		case <-animTimer.C:
			if !a.animating {
				continue
			}
			if !a.hidden {
				a.load()
				a.stepSlide()
				a.syncTuck()
				a.syncAnimation()
				_ = a.present()
			} else {
				a.syncAnimation()
			}
			resetAnimTimer()

		case msg, ok := <-eventsCh:
			if !ok {
				a.persistPlacement()
				return nil
			}
			if msg.ev != nil {
				if quit := a.handleX11Event(msg.ev); quit {
					a.persistPlacement()
					return nil
				}
				resetAnimTimer()
			}
		}
	}
}

func (a *App) initX11() error {
	setup := xproto.Setup(a.conn)
	a.screen = setup.DefaultScreen(a.conn)
	if a.screen == nil {
		return errors.New("no X11 screen")
	}
	a.maxReqBytes = defaultX11MaxRequestBytes
	if err := bigreq.Init(a.conn); err == nil {
		if reply, bErr := bigreq.Enable(a.conn).Reply(); bErr == nil && reply.MaximumRequestLength > 0 {
			a.maxReqBytes = int(reply.MaximumRequestLength) * 4
		}
	}
	if err := randr.Init(a.conn); err == nil {
		if _, rErr := randr.QueryVersion(a.conn, 1, 5).Reply(); rErr == nil {
			a.hasRandr = true
		}
	}

	a.selectVisualAndColormap()
	a.dpi = detectX11DPI(a.screen)

	winID, err := xproto.NewWindowId(a.conn)
	if err != nil {
		return err
	}
	a.win = winID

	eventMask := uint32(
		xproto.EventMaskExposure |
			xproto.EventMaskStructureNotify |
			xproto.EventMaskButtonPress |
			xproto.EventMaskButtonRelease |
			xproto.EventMaskPointerMotion |
			xproto.EventMaskEnterWindow |
			xproto.EventMaskLeaveWindow |
			xproto.EventMaskKeyPress |
			xproto.EventMaskFocusChange,
	)
	valueMask := uint32(xproto.CwBackPixel | xproto.CwBorderPixel | xproto.CwEventMask | xproto.CwColormap)
	valueList := []uint32{0, 0, eventMask, uint32(a.winColormap)}

	w := uint16(maxInt(a.logicalWidth(), 1))
	h := uint16(maxInt(a.logicalHeight(), 1))
	if err := xproto.CreateWindowChecked(
		a.conn,
		a.winDepth,
		a.win,
		a.screen.Root,
		0, 0, w, h, 0,
		xproto.WindowClassInputOutput,
		a.winVisual,
		valueMask,
		valueList,
	).Check(); err != nil {
		return fmt.Errorf("create window: %w", err)
	}

	gcID, err := xproto.NewGcontextId(a.conn)
	if err != nil {
		return err
	}
	a.gc = gcID
	if err := xproto.CreateGCChecked(a.conn, a.gc, xproto.Drawable(a.win), 0, nil).Check(); err != nil {
		return fmt.Errorf("create gc: %w", err)
	}

	a.internAtoms()
	a.configureWindowProperties()
	a.refreshKeymap()

	x11ClipboardMu.Lock()
	x11ClipboardOwner = a
	x11ClipboardMu.Unlock()
	return nil
}

func (a *App) selectVisualAndColormap() {
	a.winDepth = a.screen.RootDepth
	a.winVisual = a.screen.RootVisual
	a.winColormap = a.screen.DefaultColormap
	a.hasARGB = false

	for _, d := range a.screen.AllowedDepths {
		if d.Depth != 32 {
			continue
		}
		for _, v := range d.Visuals {
			if v.Class != xproto.VisualClassTrueColor {
				continue
			}
			cmap, err := xproto.NewColormapId(a.conn)
			if err != nil {
				continue
			}
			if err := xproto.CreateColormapChecked(a.conn, xproto.ColormapAllocNone, cmap, a.screen.Root, v.VisualId).Check(); err != nil {
				continue
			}
			a.winDepth = 32
			a.winVisual = v.VisualId
			a.winColormap = cmap
			a.hasARGB = true
			return
		}
	}
}

func (a *App) internAtoms() {
	intern := func(name string) xproto.Atom {
		reply, err := xproto.InternAtom(a.conn, false, uint16(len(name)), name).Reply()
		if err != nil || reply == nil {
			return 0
		}
		return reply.Atom
	}
	a.atomClipboard = intern("CLIPBOARD")
	a.atomTargets = intern("TARGETS")
	a.atomUTF8String = intern("UTF8_STRING")
	a.atomString = intern("STRING")
	a.atomText = intern("TEXT")
	a.atomWMProtocols = intern("WM_PROTOCOLS")
	a.atomWMDelete = intern("WM_DELETE_WINDOW")
	a.atomNetWMName = intern("_NET_WM_NAME")
	a.atomUTF8Type = a.atomUTF8String
	a.atomNetWMPid = intern("_NET_WM_PID")
	a.atomNetWMWinType = intern("_NET_WM_WINDOW_TYPE")
	a.atomNetWMWTypeUtil = intern("_NET_WM_WINDOW_TYPE_UTILITY")
	a.atomNetWMState = intern("_NET_WM_STATE")
	a.atomNetWMStateAbove = intern("_NET_WM_STATE_ABOVE")
	a.atomNetWMStateSkipTaskbar = intern("_NET_WM_STATE_SKIP_TASKBAR")
	a.atomNetWMStateSkipPager = intern("_NET_WM_STATE_SKIP_PAGER")
	a.atomNetWMOpacity = intern("_NET_WM_WINDOW_OPACITY")
	a.atomNetWorkarea = intern("_NET_WORKAREA")
	a.atomMotifHints = intern("_MOTIF_WM_HINTS")
}

func (a *App) configureWindowProperties() {
	setBytes := func(prop, typ xproto.Atom, format byte, data []byte) {
		if prop == 0 || typ == 0 {
			return
		}
		count := uint32(len(data))
		if format == 32 {
			count /= 4
		}
		_ = xproto.ChangeProperty(a.conn, xproto.PropModeReplace, a.win, prop, typ, format, count, data)
	}
	setBytes(xproto.AtomWmName, xproto.AtomString, 8, []byte(windowTitle))
	setBytes(a.atomNetWMName, a.atomUTF8Type, 8, []byte(windowTitle))

	classVal := []byte("ai-terminal-pet\x00" + windowClassName + "\x00")
	setBytes(xproto.AtomWmClass, xproto.AtomString, 8, classVal)

	if a.atomNetWMPid != 0 {
		setBytes(a.atomNetWMPid, xproto.AtomCardinal, 32, encodeCardinals(uint32(os.Getpid())))
	}
	if a.atomWMProtocols != 0 && a.atomWMDelete != 0 {
		setBytes(a.atomWMProtocols, xproto.AtomAtom, 32, encodeCardinals(uint32(a.atomWMDelete)))
	}
	if a.atomMotifHints != 0 {
		// flags=2 (MWM_HINTS_DECORATIONS), decorations=0 -> borderless.
		setBytes(a.atomMotifHints, a.atomMotifHints, 32, encodeCardinals(2, 0, 0, 0, 0))
	}
	if a.atomNetWMWinType != 0 && a.atomNetWMWTypeUtil != 0 {
		setBytes(a.atomNetWMWinType, xproto.AtomAtom, 32, encodeCardinals(uint32(a.atomNetWMWTypeUtil)))
	}
	if a.atomNetWMState != 0 {
		states := make([]uint32, 0, 3)
		if a.options.Settings.AlwaysOnTop && a.atomNetWMStateAbove != 0 {
			states = append(states, uint32(a.atomNetWMStateAbove))
		}
		if a.atomNetWMStateSkipTaskbar != 0 {
			states = append(states, uint32(a.atomNetWMStateSkipTaskbar))
		}
		if a.atomNetWMStateSkipPager != 0 {
			states = append(states, uint32(a.atomNetWMStateSkipPager))
		}
		if len(states) > 0 {
			setBytes(a.atomNetWMState, xproto.AtomAtom, 32, encodeCardinals(states...))
		}
	}
	a.applyOpacityProperty(a.options.Opacity)
}

func (a *App) applyOpacityProperty(opacity uint8) {
	if a.conn == nil || a.win == 0 || a.atomNetWMOpacity == 0 {
		return
	}
	if opacity == a.lastAppliedOpacity && a.lastAppliedOpacity != 0 {
		return
	}
	a.lastAppliedOpacity = opacity
	val := windowOpacityCardinal(opacity)
	_ = xproto.ChangeProperty(
		a.conn,
		xproto.PropModeReplace,
		a.win,
		a.atomNetWMOpacity,
		xproto.AtomCardinal,
		32,
		1,
		encodeCardinals(val),
	)
}

func (a *App) applyTopmost(above bool) {
	if a.conn == nil || a.win == 0 || a.atomNetWMState == 0 || a.atomNetWMStateAbove == 0 {
		return
	}
	action := uint32(0) // _NET_WM_STATE_REMOVE
	if above {
		action = 1 // _NET_WM_STATE_ADD
	}
	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: a.win,
		Type:   a.atomNetWMState,
		Data: xproto.ClientMessageDataUnionData32New([]uint32{
			action,
			uint32(a.atomNetWMStateAbove),
			0,
			1, // source indication: normal application
			0,
		}),
	}
	_ = xproto.SendEvent(
		a.conn,
		false,
		a.screen.Root,
		xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify,
		string(ev.Bytes()),
	)
	if above {
		_ = xproto.ConfigureWindow(a.conn, a.win, xproto.ConfigWindowStackMode, []uint32{xproto.StackModeAbove})
	}
}

func (a *App) refreshKeymap() {
	if a.conn == nil {
		return
	}
	setup := xproto.Setup(a.conn)
	minCode := setup.MinKeycode
	maxCode := setup.MaxKeycode
	if maxCode < minCode {
		return
	}
	count := byte(maxCode - minCode + 1)
	reply, err := xproto.GetKeyboardMapping(a.conn, minCode, count).Reply()
	if err != nil || reply == nil || reply.KeysymsPerKeycode == 0 {
		return
	}
	a.minKeycode = minCode
	a.maxKeycode = maxCode
	a.keysymsPerCode = reply.KeysymsPerKeycode
	a.keysyms = reply.Keysyms
}

func (a *App) keysymForKeycode(code xproto.Keycode) xproto.Keysym {
	if a.keysymsPerCode == 0 || code < a.minKeycode || code > a.maxKeycode {
		return 0
	}
	idx := int(code-a.minKeycode) * int(a.keysymsPerCode)
	if idx < 0 || idx >= len(a.keysyms) {
		return 0
	}
	return a.keysyms[idx]
}

func (a *App) displayWorkAreas() []displayRect {
	if a.screen == nil {
		return nil
	}
	now := time.Now()
	if len(a.displaysCache) > 0 && now.Sub(a.displaysCacheAt) < 500*time.Millisecond {
		return a.displaysCache
	}
	var monitors []x11Monitor
	if a.hasRandr && a.conn != nil {
		if reply, err := randr.GetMonitors(a.conn, a.screen.Root, true).Reply(); err == nil && reply != nil {
			for _, m := range reply.Monitors {
				monitors = append(monitors, x11Monitor{
					Left:    int32(m.X),
					Top:     int32(m.Y),
					Width:   m.Width,
					Height:  m.Height,
					Primary: m.Primary,
				})
			}
		}
	}
	work, hasWork := a.queryNetWorkarea()
	displays := displaysFromMonitors(a.screen.WidthInPixels, a.screen.HeightInPixels, monitors, hasWork, work)
	a.displaysCache = displays
	a.displaysCacheAt = now
	return displays
}

func (a *App) queryNetWorkarea() (displayRect, bool) {
	if a.conn == nil || a.screen == nil || a.atomNetWorkarea == 0 {
		return displayRect{}, false
	}
	reply, err := xproto.GetProperty(a.conn, false, a.screen.Root, a.atomNetWorkarea, xproto.AtomCardinal, 0, 16).Reply()
	if err != nil || reply == nil {
		return displayRect{}, false
	}
	return parseWorkareaCardinal(reply.Value)
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

func (a *App) applyWindowBounds(x, y int32, width, height int) {
	if a.conn == nil || a.win == 0 {
		return
	}
	mask := uint16(xproto.ConfigWindowX | xproto.ConfigWindowY | xproto.ConfigWindowWidth | xproto.ConfigWindowHeight)
	vals := []uint32{uint32(x), uint32(y), uint32(maxInt(width, 1)), uint32(maxInt(height, 1))}
	_ = xproto.ConfigureWindow(a.conn, a.win, mask, vals)
}

func (a *App) moveWindow(x, y int32) {
	if a.conn == nil || a.win == 0 {
		return
	}
	mask := uint16(xproto.ConfigWindowX | xproto.ConfigWindowY)
	_ = xproto.ConfigureWindow(a.conn, a.win, mask, []uint32{uint32(x), uint32(y)})
}

func (a *App) applyFramePosition() {
	a.moveWindow(a.frameX(), a.frameY())
}

func (a *App) logicalWidth() int {
	base := compactWidth
	if a.expanded {
		base = expandedWidth
	}
	return scaledDimension(base, a.dpi, a.zoom)
}

func (a *App) logicalHeight() int {
	base := compactHeight
	if a.expanded {
		base = expandedPanelHeight(len(a.sessions))
	}
	return scaledDimension(base, a.dpi, a.zoom)
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
	if a.hidden || a.expanded || a.drag.active || a.hovering {
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
}

func (a *App) syncTuck() {
	target := 0.0
	if a.wantsTuck() {
		target = 1
	}
	if target == a.tuckTarget {
		return
	}
	a.startSlide(target)
	a.syncAnimation()
}

func (a *App) snapTuck() {
	a.startSlide(0)
	if a.wantsTuck() {
		a.startSlide(1)
	}
	a.tuck = a.tuckTarget
	a.applyFramePosition()
}

func (a *App) stepSlide() {
	if a.tuck == a.tuckTarget {
		return
	}
	a.tuck = slideTuck(a.slideFrom, a.tuckTarget, time.Since(a.slideStart))
}

func (a *App) untuckNow() {
	if a.tuck == 0 && a.tuckTarget == 0 {
		return
	}
	a.tuck, a.tuckTarget = 0, 0
	a.applyFramePosition()
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
		a.x, a.y = 40, 40
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
	return Placement{Placed: true, X: a.x, Y: a.y, Zoom: a.zoom, Expanded: a.expanded, Layouts: layouts}
}

func (a *App) persistPlacement() {
	if a.options.OnPlacementChanged == nil || a.win == 0 {
		return
	}
	a.options.OnPlacementChanged(a.placement())
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
	a.sessions = dedupeSessions(a.values.List())
	a.state = session.Aggregate(a.sessions)
	a.digest = sessionDigest(a.sessions)
	a.approvals.update(a.sessions, time.Now(), a.options.Settings.AutoApprove, a.approveTimeout)
	a.keyboard.sync(a.expanded, a.sessions)
	if a.attendance != nil {
		a.attendance.publish(a.sessions)
	}
	a.alertOnTransitions()
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
	a.lastAlertAt = recordRememberedAlerts(a.lastAlertAt, inputs, plan, now)
	if len(plan.Unread) > 0 {
		a.unread.mark(plan.Unread)
		a.postRefresh()
	}
}

func (a *App) syncAnimation() {
	now := time.Now()
	sliding := a.tuck != a.tuckTarget
	pulsing := a.state == session.StateWaitingPermission || a.state == session.StateError
	toasting := a.toast.frame(now, a.sessions).visible()
	shouldAnimate := !a.hidden && !a.drag.active && (sliding || pulsing || a.approvals.maxRemainingSeconds(now) >= 0 || toasting)
	desiredInterval := uint32(pulseInterval)
	if sliding {
		desiredInterval = slideInterval
	}
	if a.animating == shouldAnimate && (!shouldAnimate || a.timerInterval == desiredInterval) {
		return
	}
	a.animating = shouldAnimate
	a.timerInterval = desiredInterval
}

func (a *App) present() error {
	if a.conn == nil || a.win == 0 {
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
	if a.conn == nil || a.win == 0 || a.gc == 0 || width <= 0 || height <= 0 {
		return
	}
	expected := width * height * 4
	if len(pixels) < expected {
		return
	}
	buf := pixels[:expected]
	if !a.hasARGB {
		buf = opaqueBGRA(buf)
	}
	chunkRows := putImageChunkRows(width, a.maxReqBytes)
	for y := 0; y < height; y += chunkRows {
		rows := minInt(chunkRows, height-y)
		start := y * width * 4
		end := (y + rows) * width * 4
		_ = xproto.PutImage(
			a.conn,
			xproto.ImageFormatZPixmap,
			xproto.Drawable(a.win),
			a.gc,
			uint16(width),
			uint16(rows),
			0,
			int16(y),
			0,
			a.winDepth,
			buf[start:end],
		)
	}
}

func (a *App) releaseSurface() {
	if a.surface != nil {
		a.surface.close()
		a.surface = nil
	}
}

func (a *App) handleControl(action ControlAction) bool {
	switch action {
	case ControlActivate:
		a.showWindow()
	case ControlToggle:
		if a.hidden {
			a.showWindow()
		} else {
			a.toggleExpanded()
		}
	case ControlReset:
		a.resetPlacement()
	case ControlSettings:
		a.showWindow()
	case ControlQuit:
		traceExit("quit-command")
		return true
	}
	return false
}

func (a *App) showWindow() {
	if a.conn == nil || a.win == 0 {
		return
	}
	a.hidden = false
	a.holdTuck(3 * time.Second)
	a.syncTuck()
	a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
	xproto.MapWindow(a.conn, a.win)
	a.applyTopmost(a.options.Settings.AlwaysOnTop)
	a.load()
	a.syncAnimation()
	_ = a.present()
}

func (a *App) toggleExpanded() {
	a.expanded = !a.expanded
	if !a.expanded {
		a.keyboard.clear()
	}
	a.toast.clear()
	a.load()
	a.applyGeometry()
	a.applyTopmost(a.options.Settings.AlwaysOnTop)
	_ = a.present()
	a.persistPlacement()
	a.syncTuck()
}

func (a *App) resetPlacement() {
	a.zoom = 1
	a.expanded = false
	a.layouts = nil
	a.width, a.height = a.logicalWidth(), a.logicalHeight()
	a.positionForWorkArea()
	a.refreshDock()
	a.holdTuck(3 * time.Second)
	a.applyWindowBounds(a.frameX(), a.frameY(), a.width, a.height)
	a.applyTopmost(a.options.Settings.AlwaysOnTop)
	_ = a.present()
	a.persistPlacement()
	a.syncTuck()
}

func (a *App) zoomBy(delta float64) {
	if delta == 0 {
		return
	}
	a.zoom = stepZoomValue(a.zoom, delta)
	a.load()
	a.applyGeometry()
	_ = a.present()
	a.persistPlacement()
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

func (a *App) beginDrag(rootX, rootY, eventX, eventY int16) {
	a.untuckNow()
	a.pressed = true
	a.syncAnimation()
	a.drag = dragState{
		active:  true,
		startX:  int32(rootX),
		startY:  int32(rootY),
		windowX: int32(rootX) - int32(eventX),
		windowY: int32(rootY) - int32(eventY),
	}
	_ = a.present()
}

func (a *App) moveDrag(rootX, rootY int16) {
	dx := int32(rootX) - a.drag.startX
	dy := int32(rootY) - a.drag.startY
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
	a.toast.show(row, identity, !copied, time.Now())
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
	go func() {
		resolved, _, ok := platform.ResolveTarget(request)
		if !ok {
			resolved = 0
		}
		select {
		case a.focusCh <- resolved:
		default:
		}
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

func (a *App) handleX11Event(ev xgb.Event) bool {
	switch e := ev.(type) {
	case xproto.ExposeEvent:
		if e.Count == 0 && a.surface != nil && !a.hidden {
			a.uploadSurface(a.surface.width, a.surface.height, a.surface.pixels)
		}
	case xproto.ButtonPressEvent:
		switch e.Detail {
		case 1:
			_ = xproto.SetInputFocus(a.conn, xproto.InputFocusParent, a.win, e.Time)
			if classifyHit(int(e.EventX), int(e.EventY), a.width, a.height, a.hit) {
				a.beginDrag(e.RootX, e.RootY, e.EventX, e.EventY)
			} else {
				a.handleTap(int(e.EventX), int(e.EventY))
			}
		case 3:
			if a.expanded {
				if row, ok := expandedRowHit(int(e.EventX), int(e.EventY), a.width, a.height, len(a.sessions)); ok && row < len(a.sessions) {
					a.toggleSessionMute(a.sessions[row])
				}
			}
		case 4:
			a.zoomBy(120)
		case 5:
			a.zoomBy(-120)
		}
	case xproto.MotionNotifyEvent:
		if a.drag.active {
			a.moveDrag(e.RootX, e.RootY)
		} else {
			a.setHover(true)
			a.trackRowHover(int(e.EventX), int(e.EventY))
		}
	case xproto.ButtonReleaseEvent:
		if e.Detail == 1 && a.drag.active {
			a.endDrag(int(e.EventX), int(e.EventY))
		}
	case xproto.EnterNotifyEvent:
		a.setHover(true)
		a.trackRowHover(int(e.EventX), int(e.EventY))
	case xproto.LeaveNotifyEvent:
		if e.Mode == xproto.NotifyModeNormal && !a.drag.active {
			a.clearHover()
		}
	case xproto.KeyPressEvent:
		a.handlePanelKey(keySymToPanelKey(a.keysymForKeycode(e.Detail)))
	case xproto.FocusOutEvent:
		if a.keyboard.clear() {
			_ = a.present()
		}
	case xproto.MappingNotifyEvent:
		a.refreshKeymap()
	case xproto.SelectionRequestEvent:
		a.handleSelectionRequest(e)
	case xproto.ClientMessageEvent:
		if len(e.Data.Data32) > 0 && a.atomWMDelete != 0 && xproto.Atom(e.Data.Data32[0]) == a.atomWMDelete {
			traceExit("wm-delete")
			return true
		}
	}
	return false
}

func (a *App) handleSelectionRequest(req xproto.SelectionRequestEvent) {
	if a.conn == nil {
		return
	}
	notify := xproto.SelectionNotifyEvent{
		Time:      req.Time,
		Requestor: req.Requestor,
		Selection: req.Selection,
		Target:    req.Target,
		Property:  0,
	}
	prop := req.Property
	if prop == 0 {
		prop = req.Target
	}
	text := currentX11ClipboardText()
	if req.Selection == a.atomClipboard && a.atomClipboard != 0 && text != "" {
		switch req.Target {
		case a.atomTargets:
			atoms := encodeCardinals(
				uint32(a.atomTargets),
				uint32(a.atomUTF8String),
				uint32(a.atomString),
				uint32(a.atomText),
			)
			_ = xproto.ChangeProperty(a.conn, xproto.PropModeReplace, req.Requestor, prop, xproto.AtomAtom, 32, 4, atoms)
			notify.Property = prop
		case a.atomUTF8String, a.atomString, a.atomText:
			data := []byte(text)
			_ = xproto.ChangeProperty(a.conn, xproto.PropModeReplace, req.Requestor, prop, req.Target, 8, uint32(len(data)), data)
			notify.Property = prop
		}
	}
	_ = xproto.SendEvent(a.conn, false, req.Requestor, 0, string(notify.Bytes()))
}

func displaysFromMonitors(screenW, screenH uint16, monitors []x11Monitor, hasWorkarea bool, work displayRect) []displayRect {
	displays := make([]displayRect, 0, len(monitors))
	for _, m := range monitors {
		if m.Width == 0 || m.Height == 0 {
			continue
		}
		rect := displayRect{
			Left:    m.Left,
			Top:     m.Top,
			Right:   m.Left + int32(m.Width),
			Bottom:  m.Top + int32(m.Height),
			Primary: m.Primary,
		}
		displays = append(displays, rect)
	}
	if len(displays) == 1 && hasWorkarea && work.valid() {
		clipped := intersectDisplayRect(displays[0], work)
		if clipped.valid() {
			displays[0] = clipped
		}
	}
	if len(displays) > 0 {
		return displays
	}
	if hasWorkarea && work.valid() {
		work.Primary = true
		return []displayRect{work}
	}
	if screenW > 0 && screenH > 0 {
		return []displayRect{{Left: 0, Top: 0, Right: int32(screenW), Bottom: int32(screenH), Primary: true}}
	}
	return nil
}

func intersectDisplayRect(a, b displayRect) displayRect {
	return displayRect{
		Left:    max32(a.Left, b.Left),
		Top:     max32(a.Top, b.Top),
		Right:   min32(a.Right, b.Right),
		Bottom:  min32(a.Bottom, b.Bottom),
		Primary: a.Primary,
	}
}

func parseWorkareaCardinal(data []byte) (displayRect, bool) {
	if len(data) < 16 {
		return displayRect{}, false
	}
	x := int32(binary.LittleEndian.Uint32(data[0:4]))
	y := int32(binary.LittleEndian.Uint32(data[4:8]))
	w := int32(binary.LittleEndian.Uint32(data[8:12]))
	h := int32(binary.LittleEndian.Uint32(data[12:16]))
	if w <= 0 || h <= 0 {
		return displayRect{}, false
	}
	return displayRect{Left: x, Top: y, Right: x + w, Bottom: y + h}, true
}

func keySymToPanelKey(sym xproto.Keysym) panelKey {
	switch sym {
	case 0xff1b:
		return panelKeyEscape
	case 0xff52, 0xff97:
		return panelKeyUp
	case 0xff54, 0xff99:
		return panelKeyDown
	case 0xff50, 0xff95:
		return panelKeyHome
	case 0xff57, 0xff9c:
		return panelKeyEnd
	case 0xff0d, 0xff8d:
		return panelKeyEnter
	default:
		return panelKeyNone
	}
}

func putImageChunkRows(width, maxReqBytes int) int {
	if width <= 0 {
		return 1
	}
	if maxReqBytes <= 64 {
		maxReqBytes = defaultX11MaxRequestBytes
	}
	const putImageHeaderBytes = 24
	rowBytes := width * 4
	payloadBudget := maxReqBytes - putImageHeaderBytes - 64
	if payloadBudget < rowBytes {
		return 1
	}
	rows := payloadBudget / rowBytes
	if rows < 1 {
		return 1
	}
	return rows
}

func windowOpacityCardinal(opacity uint8) uint32 {
	if opacity == 0 || opacity == 255 {
		if opacity == 0 {
			return 0
		}
		return 0xFFFFFFFF
	}
	return uint32(math.Round(float64(opacity) / 255.0 * float64(0xFFFFFFFF)))
}

func encodeCardinals(values ...uint32) []byte {
	buf := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(buf[i*4:], v)
	}
	return buf
}

func opaqueBGRA(src []byte) []byte {
	out := make([]byte, len(src))
	copy(out, src)
	for i := 3; i < len(out); i += 4 {
		out[i] = 255
	}
	return out
}

func detectX11DPI(screen *xproto.ScreenInfo) uint32 {
	if raw := strings.TrimSpace(os.Getenv("GDK_SCALE")); raw != "" {
		if scale, err := strconv.Atoi(raw); err == nil && scale >= 1 && scale <= 4 {
			return uint32(scale * 96)
		}
	}
	if raw := strings.TrimSpace(os.Getenv("QT_SCALE_FACTOR")); raw != "" {
		if scale, err := strconv.ParseFloat(raw, 64); err == nil && scale >= 1.0 && scale <= 4.0 {
			return uint32(math.Round(scale * 96))
		}
	}
	if screen != nil && screen.WidthInMillimeters > 0 && screen.WidthInPixels > 0 {
		dpi := uint32(math.Round(float64(screen.WidthInPixels) * 25.4 / float64(screen.WidthInMillimeters)))
		if dpi >= 120 && dpi <= 480 {
			// Snap high-DPI XWayland / physical densities smoothly while keeping
			// standard ~96 DPI screens untouched.
			return dpi
		}
	}
	return 96
}
