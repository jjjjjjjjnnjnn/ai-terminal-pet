//go:build windows

package ui

// What happens when code inside a window procedure is the problem.
//
// A Go panic that unwinds out of a syscall.NewCallback trampoline is fatal by
// design: the runtime cannot return through Windows frames, so it prints the
// panic and exits. That is the correct default for a crash, but it produced the
// worst possible outcome here, recorded in handover §79: the overlay is started
// with stdio ignored, so the panic text was discarded, the fatal printer then
// deadlocked on the world it had just frozen, and the process stayed alive as a
// window that never answered another message.
//
// Two bounds make that failure mode impossible to reach again without a record:
//
//   - callbackPanicked catches the panic at the callback boundary, writes the
//     value and its stack to error.log, and lets the overlay keep running (or
//     close after three, because a repeat means the state is not trustworthy).
//   - maxDispatchDepth refuses to nest further than a real modal loop ever
//     goes. The 2026-11-17 hang had a 2 MB thread stack almost entirely filled
//     with nested user32 send-message frames (handover §79.2); a cap turns a
//     stack-exhaustion fatal into one log line naming the message that asked
//     for the sixty-fifth level.

import (
	"fmt"
	"runtime/debug"

	win "github.com/lxn/win"
)

const (
	// maxDispatchDepth is far above what a modal loop needs (a context menu
	// dispatches one message at a time, so the depth stays at 2) and far below
	// what would exhaust the 2 MB stack the PE header reserves.
	maxDispatchDepth = 64
	// maxRecoveredPanics is the point where a contained fault stops being
	// contained: the same handler fails three times, so the process closes and
	// lets the next reporter event start a clean one.
	maxRecoveredPanics = 3
	// faultStackBytes bounds what a single record can add to error.log. The
	// stack that motivated this was megabytes deep.
	faultStackBytes = 64 << 10
)

// guardedDispatch is the only way a window procedure runs a handler: it counts
// the nesting, contains a panic, and falls back to DefWindowProc when the
// nesting bound is reached. Both procedures call it, so "a callback cannot die
// silently" has exactly one implementation and one test.
func (a *App) guardedDispatch(hwnd win.HWND, message uint32, wParam, lParam uintptr, surface string, run func() uintptr) (result uintptr) {
	if !a.beginDispatch() {
		return win.DefWindowProc(hwnd, message, wParam, lParam)
	}
	defer a.endDispatch()
	defer func() {
		if recovered := recover(); recovered != nil {
			a.callbackPanicked(hwnd, message, surface, recovered)
			result = 0
		}
	}()
	return run()
}

// beginDispatch records a window-procedure entry and reports whether the
// handler may run. The count is only touched by the UI thread (both window
// procedures run there), so it is a plain int; the timestamp stays atomic
// because the hang watchdog reads it from another goroutine.
func (a *App) beginDispatch() bool {
	a.dispatchDepth++
	a.lastDispatchTicks.Store(interruptTicks())
	if a.dispatchDepth > maxDispatchDepth {
		a.dispatchDepth--
		a.dispatchOverflowed()
		return false
	}
	return true
}

func (a *App) endDispatch() {
	if a.dispatchDepth > 0 {
		a.dispatchDepth--
	}
}

// dispatchOverflowed is the one-line record of the refusal. It is written once
// per process: at this point the stack is already pathological, and repeating
// the write on every refused message would turn a hang into a slow hang.
func (a *App) dispatchOverflowed() {
	if a.dispatchOverflowedOnce {
		return
	}
	a.dispatchOverflowedOnce = true
	writeFault("dispatch-overflow", fmt.Sprintf(
		"message=%s depth=%d: nested window messages were refused (see maxDispatchDepth in dispatch_windows.go)",
		messageName(a.lastMessage.Load()), a.dispatchDepth+1))
}

// callbackPanicked contains one panic that escaped a message handler.
//
// messageName labels the overlay's own messages; a settings-window message is
// printed as its number, which is enough to tell WM_PAINT from WM_MOUSEMOVE in
// the record.
func (a *App) callbackPanicked(hwnd win.HWND, message uint32, surface string, recovered any) {
	a.recoveredPanics++
	stack := debug.Stack()
	if len(stack) > faultStackBytes {
		stack = stack[:faultStackBytes]
	}
	writeFault("panic", fmt.Sprintf("%s message=%s depth=%d count=%d panic=%v\n%s",
		surface, messageName(message), a.dispatchDepth, a.recoveredPanics, recovered, stack))
	if a.recoveredPanics < maxRecoveredPanics {
		return
	}
	// Close the overlay, not just the window that failed: a handler that keeps
	// panicking leaves the state it reads in an unknown shape, and a clean exit
	// is what makes the next reporter event start a fresh process.
	target := hwnd
	if handle := a.window(); handle != 0 {
		target = handle
	}
	if target != 0 {
		win.PostMessage(target, wmClose, 0, 0)
	}
}
