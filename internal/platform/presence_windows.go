package platform

// Presence is the authoritative answer to "is this session's terminal still
// there?".
//
// The old model treated a reporter push as a heartbeat: silence for one TTL
// meant the session was gone. Reporters are deliberately event-driven and
// de-duplicated (a session that stays "working" for ten minutes sends
// nothing), so a perfectly live terminal was repeatedly evicted, tombstoned
// and then unable to come back with its unchanged timestamp. That is the
// flicker this file fixes.
//
// Existence and state are now separate concerns:
//
//   existence -> observed from the OS (process table + terminal windows)
//   state     -> whatever the reporter last said, however long ago
//
// One Toolhelp snapshot plus one EnumWindows pass per cache window replaces a
// per-session OpenProcess probe, which is both cheaper and far more stable.

import (
	"strings"
	"sync"
	"time"

	"ai-terminal-pet/internal/session"
)

const (
	DefaultPresenceTTL = 10 * time.Second
	// DefaultOrphanAfter bounds how long a session may survive with *no*
	// identity on the desktop at all. This is deliberately not a heartbeat
	// timeout: a session whose terminal tab names it is kept forever, however
	// quiet, and one reclaimed here comes straight back on its next report
	// (tombstoneExpired has no timestamp gate).
	//
	// What it catches is a snapshot left behind by a reporter that died
	// without sending `closed`. Hosts such as opencode.exe stay up for days,
	// so "parentPid is alive" by itself pinned every conversation that host
	// had ever seen - the panel accumulated ghosts that no longer existed.
	// Forty-five minutes is far beyond any normal quiet stretch, and a real
	// tab always wins via windowIdentifiesSession.
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

// view returns the cached OS snapshot, refreshing it only when the cache
// window has elapsed. Every caller in a scan cycle therefore shares one
// enumeration instead of each doing its own.
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

// Captures reports how many OS snapshots were taken. It is a test/debug aid
// for proving the cache actually collapses repeated checks into one scan.
func (p *Presence) Captures() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.captures
}

// ProcessAlive answers from the cached process table. A single Toolhelp
// snapshot covers every PID, so probing 256 sessions costs the same as one.
func (p *Presence) ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	processes, _ := p.view()
	_, ok := processes[pid]
	return ok
}

// SessionAlive reports whether *this particular session* is still there.
//
// This is the subtle part. "The host process is alive" only proves the host
// is alive - opencode.exe can run for a week while dozens of conversations
// come and go. Treating the parent pid as unconditional proof is what made
// the panel pile up sessions that no longer existed.
//
// Evidence is now ordered by how specific it is:
//  1. A terminal tab whose title names this session (title or project).
//     This outranks every clock: a real but quiet tab is kept indefinitely.
//  2. The owning process is still in the process table: the reporter's own
//     real pid where it can be trusted, otherwise the parentPid hook
//     reporters carry. Its disappearance ends the session even when a
//     longer-lived ancestor terminal is still open.
//  3. Recent activity. A live host proves the host, not this conversation -
//     a reporter that stays up for days but hands out a new conversation id
//     per event is exactly the ghost-pile case this guards against.
//
// A live host with an old timestamp and no tab identity is an orphan - a
// snapshot left behind by a reporter that never sent `closed`. It is dropped,
// but tombstoneExpired lets it resurrect immediately on its next report, so
// this can never turn into the blink-out-and-never-return bug of the old TTL.
func (p *Presence) SessionAlive(value session.Session) bool {
	if value.State == session.StateClosed {
		return false
	}
	processes, windows := p.view()

	// A dead reporter process is the end of its conversation, and no window
	// title may overrule that.
	//
	// This is the "ghost row" rule. Before it, a snapshot left behind by
	// `opencode.exe` (a trustworthy pid) whose process had exited was kept alive
	// forever by *any* open tab whose title contained the session title or the
	// project name - including a tab belonging to a completely different
	// process. Measured on the machine that reported it: three of six rows came
	// from pids that no longer existed, kept alive by one "OC | <title>" tab.
	//
	// The title stays evidence for reporters whose pid is a synthetic hash
	// (hook and statusline reporters), because for them the pid proves nothing
	// either way; that is the case the title rule was written for.
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

	// Which process actually owns the conversation? For reporters with a
	// trustworthy pid it is the reporter itself - a still-open parent terminal
	// is not a reason to keep its sessions. Hook reporters only have a real
	// parentPid (their own pid is a synthetic hash), so that is the owner.
	hostAlive := false
	if trustedProcessID(value) {
		_, hostAlive = processes[value.PID]
	} else if value.ParentPID > 0 {
		_, hostAlive = processes[value.ParentPID]
	}
	if !hostAlive {
		return false
	}
	// The host is up, but that alone does not prove *this* conversation is
	// still one of its tabs (opencode.exe can host hundreds over a week).
	// Without a named tab, only recent activity keeps an unidentified session.
	// No desktop identity, but the host is up. Reporters that predate
	// timestamps (or hand-written probes) are given the benefit of the doubt.
	if value.UpdatedAt.IsZero() {
		return true
	}
	orphanAfter := p.orphanAfter
	if orphanAfter <= 0 {
		orphanAfter = DefaultOrphanAfter
	}
	return value.UpdatedAt.After(p.now().Add(-orphanAfter))
}

// trustedProcessID lists reporters whose PID is a real OS process. Hook and
// generic reporters key sessions on stableProcessId (a large FNV hash);
// probing that number could accidentally match an unrelated live process and
// pin a dead session forever, so they are never trusted here.
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

// bareWindowTitle is the set of titles that name a *tool*, not a conversation.
//
// An OpenCode plugin that has not learned the session title yet writes the
// literal "opencode", and the terminal tab is often titled just "OpenCode" or
// "OC" too. Matching those against each other is how one open tab kept every
// historical conversation of that tool alive: the title was 8 runes long, so
// the old "at least 3 characters" guard let it through.
//
// Matching is on the whole title, not a substring: a tab called
// "OC | refactor the parser" is real identity even though it starts with a
// bare tool name.
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

// titleIdentifiesConversation reports whether a session title can be used as
// evidence that a window belongs to that session.
//
// It lowercases and trims on the way in rather than trusting the caller: the
// window comparison already lowercases everything, and a helper that only
// recognised the exact casing it was written with would silently let
// "OpenCode" through.
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
	// A CLI keyword alone is deliberately NOT identity. With one Antigravity
	// window open, "agy" would otherwise keep every historical conversation
	// alive forever and the panel would accumulate ghosts. Those sessions
	// have a real parentPid, which is both stronger and per-conversation.
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
		if title != "" && strings.Contains(windowTitle, title) {
			return true
		}
		if base != "" && strings.Contains(windowTitle, base) {
			return true
		}
	}
	return false
}
