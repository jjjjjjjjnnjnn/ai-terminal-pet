package platform

// Is the user actually looking at this terminal?
//
// The overlay used to alert on every state transition, which meant a balloon
// for a permission prompt the user was already reading. Two facts decide that
// question, and neither is available from the session report:
//
//   - **Presence**: has there been any keyboard or mouse input recently? A
//     window left in the foreground while its owner is at lunch looks attended
//     to every window-based rule, which is why GetLastInputInfo exists and why
//     it is checked before the window is trusted.
//   - **Focus**: is the window in front the one that owns this session? That is
//     the same question internal/platform already answers for jump-to-terminal
//     (process ancestry + window title), so it is answered by the same ranking
//     rather than by a second, subtly different rule.
//
// Everything here is read-only and cheap, but EnumWindows touches other
// processes: a hung window can stall the caller, so the sampler runs on a
// background goroutine and the UI reads its last snapshot (see ui.Options
// Attendance). Nothing in this file ever moves a window.

import (
	"syscall"
	"time"
	"unsafe"

	"ai-terminal-pet/internal/session"
)

var (
	attendUser32 = syscall.NewLazyDLL("user32.dll")
	attendKernel = syscall.NewLazyDLL("kernel32.dll")

	getLastInputInfoProc = attendUser32.NewProc("GetLastInputInfo")
	getTickCountProc     = attendKernel.NewProc("GetTickCount")
	foregroundWindowProc = attendUser32.NewProc("GetForegroundWindow")
)

// lastInputInfo mirrors LASTINPUTINFO. cbSize must be set by the caller.
type lastInputInfo struct {
	Size uint32
	Time uint32
}

// ForegroundSample is the desktop state at one moment: which window is in front
// and how long since any human touched the machine.
//
// Known is false when the sample failed. Callers must read that as "cannot
// tell", never as "the user is here" - staying silent while an agent waits is
// the failure this whole feature must not introduce.
type ForegroundSample struct {
	// Handle is the foreground window, or 0 when nothing is in front.
	Handle uintptr
	// Idle is how long since the last keyboard or mouse input, across the
	// whole session (not just this window).
	Idle time.Duration
	// IdleKnown is false when GetLastInputInfo did not answer or the tick
	// counters wrapped in a way this code cannot resolve.
	IdleKnown bool
	// Known is false when the foreground window could not be read at all.
	Known bool
}

// SampleForeground reads the desktop state. It does no enumeration: focus and
// idle only. Both calls are cheap and neither blocks on another process.
func SampleForeground() ForegroundSample {
	sample := ForegroundSample{}
	if handle, _, _ := foregroundWindowProc.Call(); handle != 0 {
		sample.Handle = handle
		sample.Known = true
	}
	var info lastInputInfo
	info.Size = uint32(unsafe.Sizeof(info))
	if result, _, _ := getLastInputInfoProc.Call(uintptr(unsafe.Pointer(&info))); result != 0 {
		now, _, _ := getTickCountProc.Call()
		// GetTickCount is 32-bit and wraps every 49.7 days; unsigned subtraction
		// is correct across a wrap, which is why both sides are uint32.
		elapsed := uint32(uint32(now) - info.Time)
		sample.Idle = time.Duration(elapsed) * time.Millisecond
		sample.IdleKnown = true
	}
	return sample
}

// IdleAtLeast reports whether the machine has been untouched for at least
// threshold. A known-false sample answers false, and the caller decides what
// "unknown" means - SampleForeground is deliberately not the place that guesses.
func (s ForegroundSample) IdleAtLeast(threshold time.Duration) bool {
	if threshold <= 0 || !s.IdleKnown {
		return false
	}
	return s.Idle >= threshold
}

// Attends reports whether the foreground window is the terminal that owns this
// session, using the same ranking as jump-to-terminal. The window list comes
// from Presence's cache, so calling this for every session in one scan costs no
// extra enumeration.
//
// It returns false for an unknown foreground handle. That is the conservative
// answer at the call site that reads it (see ui.Decide: unknown means announce),
// so a failed sample can never silence an alert.
func (p *Presence) Attends(value session.Session, foreground uintptr) bool {
	if foreground == 0 {
		return false
	}
	processes, windows := p.view()
	target := FocusTarget{
		PID:       value.PID,
		ParentPID: value.ParentPID,
		Title:     value.Title,
		Project:   value.Project,
		Source:    value.Source,
		Terminal:  value.Terminal,
	}
	best, ok := selectFocusWindow(windows, target, ancestorDepths(target, func(pid int) (int, bool) {
		process, present := processes[pid]
		return process.parent, present
	}, maxFocusDepth), 0)
	if !ok || best.handle == 0 {
		return false
	}
	return best.handle == foreground
}
