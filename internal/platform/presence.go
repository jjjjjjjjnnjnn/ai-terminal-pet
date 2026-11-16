package platform

// Presence is the authoritative answer to "is this session's terminal still
// there?".
//
// Existence and state are separate concerns:
//
//   existence -> observed from the OS (process table + terminal windows)
//   state     -> whatever the reporter last said, however long ago

import (
	"strings"
	"sync"
	"time"

	"ai-terminal-pet/internal/session"
)

const (
	DefaultPresenceTTL = 10 * time.Second
	DefaultOrphanAfter = 45 * time.Minute
)

type Presence struct {
	mutex       sync.Mutex
	ttl         time.Duration
	orphanAfter time.Duration
	now         func() time.Time
	load        func() (map[int]focusProcess, []focusWindow)
	processes   map[int]focusProcess
	windows     []focusWindow
	expires     time.Time
	captures    int
}

func NewPresence(ttl time.Duration) *Presence {
	if ttl <= 0 {
		ttl = DefaultPresenceTTL
	}
	presence := &Presence{ttl: ttl, orphanAfter: DefaultOrphanAfter, now: time.Now}
	presence.load = presence.capture
	return presence
}

func (p *Presence) capture() (map[int]focusProcess, []focusWindow) {
	processes := snapshotProcesses()
	return processes, visibleWindows(processes)
}

func (p *Presence) view() (map[int]focusProcess, []focusWindow) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	current := p.now()
	if p.processes == nil || !current.Before(p.expires) {
		p.processes, p.windows = p.load()
		p.expires = current.Add(p.ttl)
		p.captures++
	}
	return p.processes, p.windows
}

func (p *Presence) Captures() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.captures
}

func (p *Presence) ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	processes, _ := p.view()
	_, ok := processes[pid]
	return ok
}

func (p *Presence) SessionAlive(value session.Session) bool {
	if value.State == session.StateClosed {
		return false
	}
	processes, windows := p.view()

	if trustedProcessID(value) {
		if _, alive := processes[value.PID]; !alive {
			return false
		}
	}

	if windowIdentifiesSession(windows, FocusTarget{
		PID:       value.PID,
		ParentPID: value.ParentPID,
		Title:     value.Title,
		Project:   value.Project,
		Source:    value.Source,
		Terminal:  value.Terminal,
	}) {
		return true
	}

	hostAlive := false
	if trustedProcessID(value) {
		_, hostAlive = processes[value.PID]
	} else if value.ParentPID > 0 {
		_, hostAlive = processes[value.ParentPID]
	}
	if !hostAlive {
		return false
	}
	if value.UpdatedAt.IsZero() {
		return true
	}
	orphanAfter := p.orphanAfter
	if orphanAfter <= 0 {
		orphanAfter = DefaultOrphanAfter
	}
	return value.UpdatedAt.After(p.now().Add(-orphanAfter))
}

func trustedProcessID(value session.Session) bool {
	if value.PID <= 0 {
		return false
	}
	switch value.Source {
	case "opencode", "powershell-ipc", "pwsh-session":
		return true
	}
	return false
}

var bareWindowTitle = map[string]bool{
	"opencode": true, "oc": true,
	"claude": true, "claude code": true, "claude-code": true,
	"agy": true, "antigravity": true, "antigravity cli": true,
	"codex": true, "codex cli": true,
	"deepseek harness": true, "deepseek-harness": true, "dsh": true,
	"terminal": true, "windows terminal": true, "powershell": true,
	"pwsh": true, "cmd": true, "command prompt": true, "bash": true,
	"wsl": true, "git bash": true,
}

func titleIdentifiesConversation(title string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(title))
	if len([]rune(trimmed)) < 3 {
		return false
	}
	return !bareWindowTitle[trimmed]
}

func windowIdentifiesSession(windows []focusWindow, target FocusTarget) bool {
	title := strings.ToLower(strings.TrimSpace(target.Title))
	base := strings.ToLower(projectBase(target.Project))
	if !titleIdentifiesConversation(title) {
		title = ""
	}
	if title == "" && base == "" {
		return false
	}
	for _, candidate := range windows {
		if candidate.handle == 0 || strings.EqualFold(candidate.class, "AITerminalPetWindow") {
			continue
		}
		if strings.TrimSpace(candidate.title) == "" || terminalWindowScore(candidate.class, candidate.exe) < 0 {
			continue
		}
		windowTitle := strings.ToLower(candidate.title)
		// A title that IS the window's title verbatim names the WINDOW, and one
		// window hosts every conversation in that tab - so it is not evidence
		// that this conversation still exists. The Antigravity statusline started
		// sending the console title (which is exactly the Windows Terminal tab
		// title) in 1.0.11 so jump-to-terminal could find the tab; reading that
		// same string as identity here kept finished conversations on the panel
		// for hours after their host had exited (handover §63). A conversation
		// title, by contrast, is a proper part of a tab title ("OC | <title>",
		// "agy · <project>"), so it still counts.
		if title != "" && windowTitle != title && strings.Contains(windowTitle, title) {
			return true
		}
		if base != "" && strings.Contains(windowTitle, base) {
			return true
		}
	}
	return false
}
