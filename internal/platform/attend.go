package platform

import (
	"time"

	"ai-terminal-pet/internal/session"
)

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
	// IdleKnown is false when idle duration could not be queried.
	IdleKnown bool
	// Known is false when the foreground window could not be read at all.
	Known bool
}

// IdleAtLeast reports whether the machine has been untouched for at least
// threshold. A known-false sample answers false, and the caller decides what
// "unknown" means.
func (s ForegroundSample) IdleAtLeast(threshold time.Duration) bool {
	if threshold <= 0 || !s.IdleKnown {
		return false
	}
	return s.Idle >= threshold
}

// Attends reports whether the foreground window is the terminal that owns this
// session, using the same ranking as jump-to-terminal.
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
