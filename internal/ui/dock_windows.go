package ui

import (
	"math"
	"time"
	"unsafe"

	"ai-terminal-pet/internal/session"
	win "github.com/lxn/win"
)

// This file owns where the panel sits: which screen edge it hugs, how it grows
// when its content changes, and how it slides out of the way.
//
// It is separated from the window plumbing because this is the part that went
// wrong. The old code preserved the window centre across a size change, so
// expanding a panel parked on the right edge - where the clamp then pushed the
// wider frame back on screen - and collapsing it again landed well to the left
// of where it started. Anchoring on the frame's own edge was not enough either:
// the clamp can still nudge a wide frame inward, and the next collapse would
// then measure from the nudged position.
//
// So the panel remembers a gap, not a position. A docked frame is always
// rebuilt as "edge - gap - width", which means a wider frame simply reaches
// further from the same edge and the clamp's own margin can never accumulate.
// a.x/a.y stay exactly where the user put them and no geometry change writes
// to them.

// dockSide is the screen edge the panel hugs. Only left and right are
// supported: the overlay is a tall vertical strip, and sliding it into the top
// or bottom edge would hide the very content it exists to show.
type dockSide int

const (
	dockNone dockSide = iota
	dockLeft
	dockRight
)

// dockTolerance is how close to a work-area edge counts as hugging it. It is
// generous on purpose: the default anchor leaves a margin, and a user who
// drags the strip up against an edge means to dock it there.
const dockTolerance = 48

// dockedSliver is how much of the panel stays on screen when it is fully
// tucked, in logical units at 96 DPI. Narrow enough to be out of the way, wide
// enough to aim at without hunting for it.
const dockedSliver = 10

// dockedOrigin is where a docked frame sits when it is fully out: the gap it
// settled at, measured from the edge it hugs.
func dockedOrigin(dock dockSide, edge, gap int32, width int) int32 {
	switch dock {
	case dockRight:
		return edge - gap - int32(width)
	case dockLeft:
		return edge + gap
	}
	return 0
}

// tuckedOrigin is where a frame sits when it has slid all the way into its
// edge, leaving exactly sliver units on screen.
func tuckedOrigin(dock dockSide, edge int32, width, sliver int) int32 {
	switch dock {
	case dockRight:
		return edge - int32(sliver)
	case dockLeft:
		return edge - int32(width-sliver)
	}
	return 0
}

// slideBetween interpolates the frame origin partway through the dock slide.
func slideBetween(untucked, tucked int32, tuck float64) int32 {
	switch {
	case tuck <= 0:
		return untucked
	case tuck >= 1:
		return tucked
	}
	return untucked + int32(math.Round(float64(tucked-untucked)*tuck))
}

// workArea is the usable rectangle of the monitor the panel is on.
func (a *App) workArea() (win.RECT, bool) {
	width, height := int32(a.width), int32(a.height)
	if width == 0 || height == 0 {
		width, height = int32(a.logicalWidth()), int32(a.logicalHeight())
	}
	center := win.POINT{X: a.x + width/2, Y: a.y + height/2}
	monitor := monitorFromPoint(&center)
	if monitor == 0 {
		return win.RECT{}, false
	}
	info := win.MONITORINFO{CbSize: uint32(unsafe.Sizeof(win.MONITORINFO{}))}
	if !getMonitorInfo(monitor, &info) {
		return win.RECT{}, false
	}
	return info.RcWork, true
}

// refreshDock recomputes which screen edge the panel is hugging and how far it
// sits from it. It is called when the panel settles somewhere new - drag
// release, reset, DPI change, startup - never per frame, because the answer
// only changes when the user moves it.
func (a *App) refreshDock() {
	a.dock, a.dockEdge, a.dockGap = dockNone, 0, 0
	work, ok := a.workArea()
	if !ok {
		return
	}
	tolerance := int32(dockTolerance * a.dpi / 96)
	width := int32(a.logicalWidth())
	switch {
	case a.x <= work.Left+tolerance:
		a.dock, a.dockEdge, a.dockGap = dockLeft, work.Left, a.x-work.Left
	case a.x+width >= work.Right-tolerance:
		a.dock, a.dockEdge, a.dockGap = dockRight, work.Right, work.Right-(a.x+width)
	}
}

// sliverWidth is how many screen pixels of the panel stay visible when it is
// fully tucked. Scaled independently of the panel's own size because
// scaledDimension floors everything at 32 units, which would make a tucked
// strip wider than the panel it came from at low zoom.
func (a *App) sliverWidth() int {
	value := int(math.Round(float64(dockedSliver) * float64(a.dpi) / 96 * a.zoom))
	return maxInt(value, 6)
}

// untuckedX is the left edge the frame would have if it were fully out.
func (a *App) untuckedX() int32 {
	if a.dock == dockNone {
		return a.x
	}
	return dockedOrigin(a.dock, a.dockEdge, a.dockGap, a.logicalWidth())
}

// frameX is the left edge of the window right now.
//
// a.x is never consulted for a docked panel beyond deciding the dock: the
// position is rebuilt from the edge and the gap every time, so nothing about
// expanding, clamping or sliding can accumulate into a drift.
func (a *App) frameX() int32 {
	// While dragging the pointer owns the position. Deriving it - and
	// especially clamping it - would fight the pointer, because the drag
	// writes the window position directly.
	if a.drag.active {
		return a.x
	}
	untucked := a.untuckedX()
	if a.dock == dockNone || a.tuck <= 0 {
		return a.clampFrameX(untucked, a.logicalWidth())
	}
	// While sliding, interpolate straight to the tucked origin. Running the
	// clamp here would drag the panel back out of the edge it is sliding into.
	tucked := tuckedOrigin(a.dock, a.dockEdge, a.logicalWidth(), a.sliverWidth())
	return slideBetween(untucked, tucked, a.tuck)
}

// frameY is the top edge of the window right now. A tall session list near the
// bottom of the screen is pushed up so it fits, and the push is not written
// back: a.y stays where the user put it, so collapsing returns there.
func (a *App) frameY() int32 {
	if a.drag.active {
		return a.y
	}
	y := a.y
	work, ok := a.workArea()
	if !ok {
		return y
	}
	margin := int32(8 * a.dpi / 96)
	if maximum := work.Bottom - margin - int32(a.height); y > maximum {
		y = maximum
	}
	if minimum := work.Top + margin; y < minimum {
		y = minimum
	}
	return y
}

// clampFrameX keeps a frame from escaping off the left of the work area, which
// can only happen if the frame is wider than the screen.
//
// There is deliberately no right-hand margin. A docked panel is meant to sit
// against its edge, and forcing a margin there makes a flush position
// unrepresentable: the frame would be rebuilt flush on the way back and then
// pushed inward again, which is exactly the drift this model exists to avoid.
// The remembered position is kept fully on screen by clampHome instead.
func (a *App) clampFrameX(x int32, width int) int32 {
	work, ok := a.workArea()
	if !ok {
		return x
	}
	if minimum := work.Left + int32(8*a.dpi/96); x < minimum {
		return minimum
	}
	return x
}

// applyFramePosition moves the window to where the geometry says it belongs.
// One SetWindowPos, no repaint, so the slide animation can call it per tick.
func (a *App) applyFramePosition() {
	if a.handle.Load() == 0 {
		return
	}
	win.SetWindowPos(a.window(), a.zOrder(), a.frameX(), a.frameY(), 0, 0, win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_NOACTIVATE)
}

// wantsTuck reports whether the panel should be out of the way right now.
//
// Five things veto it, and the fourth is the point of the whole overlay: a
// session waiting for the user, or one that failed, keeps the panel fully on
// screen. Sliding the "approve this" signal into the screen edge would be the
// worst possible moment to hide it.
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

// holdTuck keeps the panel out for a moment. Explicit reveals use it: sliding
// straight back into the edge would make the tray's "show" item, and --show,
// look broken to anyone who just asked to see the panel.
func (a *App) holdTuck(d time.Duration) {
	a.tuckHoldUntil = time.Now().Add(d)
}

// slideDurations are how long the panel takes to slide out of, and into, the
// screen edge. Revealing is quicker than hiding: a hover should feel instant,
// while the panel settling out of the way can afford to be gentle.
const (
	slideOutDuration = 160 * time.Millisecond
	slideInDuration  = 240 * time.Millisecond
)

// startSlide begins a move to target, remembering where the slide started and
// when, so the animation can be driven by elapsed time rather than by how many
// timer ticks happen to arrive.
func (a *App) startSlide(target float64) {
	a.tuckTarget = target
	a.slideFrom = a.tuck
	a.slideStart = time.Now()
}

// syncTuck points the slide at where it should end up and keeps the animation
// timer alive while it travels.
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

// snapTuck jumps straight to the target instead of animating. Used once at
// startup, where an animation would read as a glitch: the panel would appear
// fully open and immediately slide away.
func (a *App) snapTuck() {
	a.startSlide(0)
	if a.wantsTuck() {
		a.startSlide(1)
	}
	a.tuck = a.tuckTarget
	a.applyFramePosition()
}

// slideTuck is the tuck value after elapsed time of a slide from -> to, easing
// out so the panel leaves promptly and settles gently.
func slideTuck(from, to float64, elapsed time.Duration) float64 {
	duration := slideOutDuration
	if to > from {
		duration = slideInDuration
	}
	if elapsed <= 0 {
		return from
	}
	if elapsed >= duration {
		return to
	}
	progress := float64(elapsed) / float64(duration)
	eased := 1 - math.Pow(1-progress, 3)
	return from + (to-from)*eased
}

// stepSlide advances the tuck animation.
//
// Progress comes from elapsed time rather than a fixed fraction per tick. A
// fixed fraction makes the slide take longer whenever frames are dropped -
// which is exactly when the desktop is busy and the animation most needs to
// stay predictable.
func (a *App) stepSlide() {
	if a.tuck == a.tuckTarget {
		return
	}
	// The position rides along with the next paint: UpdateLayeredWindow moves
	// and repaints in one step, so a separate SetWindowPos here would show a
	// frame of old content at the new position before the paint landed.
	a.tuck = slideTuck(a.slideFrom, a.tuckTarget, time.Since(a.slideStart))
}

// untuckNow completes the slide immediately, in the direction it was already
// heading. Used before a drag, which does its arithmetic on the frame position
// and must not start from a frame that is deliberately hanging off screen.
func (a *App) untuckNow() {
	if a.tuck == 0 && a.tuckTarget == 0 {
		return
	}
	a.tuck, a.tuckTarget = 0, 0
	a.applyFramePosition()
}
