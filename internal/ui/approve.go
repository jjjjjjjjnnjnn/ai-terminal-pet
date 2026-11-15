package ui

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ai-terminal-pet/internal/session"
)

// Approval handshake between the overlay and the reporter hooks.
//
// The overlay is a read-only panel: it cannot reach into a terminal and
// click "allow". So auto-approve works the other way round - the hook
// blocks for up to AutoApproveTimeoutSec waiting for a veto from the
// overlay, and silence means approval:
//
//	reporters/*.js -- hook sees a permission event, writes
//	  %LOCALAPPDATA%\AI-Terminal-Pet\approvals\pending-<source>-<session>-<pid>-<epochms>.json
//	  then polls for a sibling <same>.deny every 100 ms.
//	overlay (this file) -- while a waiting_permission session is on screen
//	  with auto-approve on, runs a countdown per session. A tap on the
//	  expanded panel denies the most urgent waiting session by writing the
//	  .deny file; the countdown reaching zero needs no write at all, the
//	  hook treats the absence of a veto as approval.
//	hook -- sees .deny: replies deny. Sees nothing for the whole timeout:
//	  replies allow. Either way it deletes its pending file on exit.
//
// File names must survive session ids containing anything (paths, colons,
// unicode), so names carry only a sanitised stem; the exact identity lives
// inside the JSON body, which is what the overlay matches on.

// approvalPendingPrefix marks hook-owned request files in the approvals dir.
const approvalPendingPrefix = "pending-"

// approvalPending is the JSON body a hook writes next to its pending file.
type approvalPending struct {
	Source    string `json:"source"`
	SessionID string `json:"sessionId"`
	PID       int    `json:"pid"`
	Tool      string `json:"tool"`
	CreatedAt string `json:"createdAt"`
}

// approvalDeny is the JSON body the overlay writes to veto one request.
type approvalDeny struct {
	Decision string `json:"decision"`
	At       string `json:"at"`
}

// approvalsDir is where hooks and the overlay exchange pending/deny files.
// Missing LOCALAPPDATA means "" and every operation below becomes a no-op.
func approvalsDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return ""
	}
	return filepath.Join(base, "AI-Terminal-Pet", "approvals")
}

// sanitizeApprovalStem maps arbitrary text onto [A-Za-z0-9_-], truncated to
// 64 characters. It must match sanitizeStem in reporters/protocol.js - the
// two sides never compare stems for identity (the JSON body is the truth),
// but keeping them identical keeps the directory human-readable.
func sanitizeApprovalStem(value string) string {
	var out []byte
	for index := 0; index < len(value) && len(out) < 64; index++ {
		char := value[index]
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' {
			out = append(out, char)
		} else {
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "session"
	}
	return string(out)
}

// approvalSessionKey binds a countdown to one store entry. Source, session
// id and PID are exactly the store's own key parts (see internal/store),
// so two agents sharing a session id can never share a countdown.
func approvalSessionKey(source, sessionID string, pid int) string {
	return source + "\x00" + sessionID + "\x00" + strconv.Itoa(pid)
}

// approval is one live countdown.
type approval struct {
	// observedAt is the session UpdatedAt the deadline was derived from. A
	// newer UpdatedAt means a new permission turn, which restarts the
	// countdown instead of inheriting whatever was left.
	observedAt time.Time
	deadline   time.Time
	timeout    time.Duration
}

// approvalView is what the rasterizer needs: whole seconds left plus the
// fraction for the progress bar.
type approvalView struct {
	RemainingSec int
	Fraction     float64
}

// approvalTracker owns the live countdowns. UI thread only.
type approvalTracker struct {
	pending map[string]approval
}

func newApprovalTracker() *approvalTracker {
	return &approvalTracker{pending: map[string]approval{}}
}

// update reconciles the countdowns with the current session list. A new
// waiting turn (newer UpdatedAt) restarts its deadline; anything that
// stopped waiting - approved in the terminal, denied, closed, expired -
// loses its entry immediately, so a timer can never fire for a request
// that is already gone (no ghost approvals).
func (t *approvalTracker) update(items []session.Session, now time.Time, enabled bool, timeout time.Duration) {
	if !enabled || timeout <= 0 {
		if len(t.pending) > 0 {
			t.pending = map[string]approval{}
		}
		return
	}
	live := map[string]approval{}
	for _, current := range items {
		if current.State != session.StateWaitingPermission {
			continue
		}
		key := approvalSessionKey(current.Source, current.SessionID, current.PID)
		if existing, ok := t.pending[key]; ok && existing.observedAt.Equal(current.UpdatedAt) {
			live[key] = existing
			continue
		}
		live[key] = approval{observedAt: current.UpdatedAt, deadline: now.Add(timeout), timeout: timeout}
	}
	t.pending = live
}

// views returns the drawable state for countdowns still running. Expired
// entries read as zero seconds (the hook is approving right now) and are
// dropped on the next update once the session leaves waiting.
func (t *approvalTracker) views(now time.Time) map[string]approvalView {
	if len(t.pending) == 0 {
		return nil
	}
	out := make(map[string]approvalView, len(t.pending))
	for key, entry := range t.pending {
		remaining := entry.deadline.Sub(now)
		seconds := 0
		fraction := 0.0
		if remaining > 0 {
			seconds = int(math.Ceil(remaining.Seconds()))
			if entry.timeout > 0 {
				fraction = float64(remaining) / float64(entry.timeout)
				fraction = math.Max(0, math.Min(1, fraction))
			}
		}
		out[key] = approvalView{RemainingSec: seconds, Fraction: fraction}
	}
	return out
}

// maxRemainingSeconds is the hottest countdown in whole seconds, or -1
// when nothing is pending. It feeds the frame key so the panel repaints
// as the numbers tick.
func (t *approvalTracker) maxRemainingSeconds(now time.Time) int64 {
	best := int64(-1)
	for _, entry := range t.pending {
		remaining := entry.deadline.Sub(now)
		seconds := int64(0)
		if remaining > 0 {
			seconds = int64(math.Ceil(remaining.Seconds()))
		}
		if seconds > best {
			best = seconds
		}
	}
	return best
}

// urgentKey returns the session key of the most urgent waiting session -
// items arrive priority-sorted, so the first waiting one wins - or "".
// A tap on the expanded panel denies exactly this request.
func (t *approvalTracker) urgentKey(items []session.Session) string {
	for _, current := range items {
		if current.State != session.StateWaitingPermission {
			continue
		}
		key := approvalSessionKey(current.Source, current.SessionID, current.PID)
		if _, ok := t.pending[key]; ok {
			return key
		}
	}
	return ""
}

// deny writes a veto file for every hook request matching the session key
// and drops the countdown at once, so a second tap can never double-deny
// and an expired timer can never approve afterwards. It returns whether a
// veto file was actually written.
func (t *approvalTracker) deny(key string, items []session.Session) bool {
	var target *session.Session
	for index := range items {
		current := &items[index]
		if current.State != session.StateWaitingPermission {
			continue
		}
		if approvalSessionKey(current.Source, current.SessionID, current.PID) != key {
			continue
		}
		target = current
		break
	}
	delete(t.pending, key)
	if target == nil {
		return false
	}
	return writeApprovalDeny(*target)
}

// writeApprovalDeny vetoes all hook requests currently filed for the
// session. Hooks poll, so several pending files for one session (retries,
// concurrent turns) are all vetoed, not just the first found.
func writeApprovalDeny(target session.Session) bool {
	directory := approvalsDir()
	if directory == "" {
		return false
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return false
	}
	body, err := json.Marshal(approvalDeny{Decision: "deny", At: time.Now().UTC().Format(time.RFC3339)})
	if err != nil {
		return false
	}
	wrote := false
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, approvalPendingPrefix) || !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".deny") {
			continue
		}
		pending, err := readApprovalPending(filepath.Join(directory, name))
		if err != nil {
			continue
		}
		if pending.Source != target.Source || pending.SessionID != target.SessionID || pending.PID != target.PID {
			continue
		}
		if writeFileAtomicSmall(filepath.Join(directory, name+".deny"), body) == nil {
			wrote = true
		}
	}
	return wrote
}

// readApprovalPending parses one hook request file. A torn write (the hook
// was killed mid-rename... actually hooks write atomically too, but a
// half-written file is still possible under AV filters) is an error and
// the caller simply tries the next file on the following poll.
func readApprovalPending(path string) (approvalPending, error) {
	var pending approvalPending
	data, err := os.ReadFile(path)
	if err != nil {
		return pending, err
	}
	if err := json.Unmarshal(data, &pending); err != nil {
		return pending, err
	}
	if pending.Source == "" || pending.SessionID == "" {
		return pending, errors.New("approval pending file is missing its identity")
	}
	return pending, nil
}

// writeFileAtomicSmall writes a small file via temp-plus-rename in the same
// directory, so a polling hook never reads a half-written veto.
func writeFileAtomicSmall(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".pet-deny-*.tmp")
	if err != nil {
		return err
	}
	name := temporary.Name()
	_ = temporary.Chmod(0600)
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		_ = os.Remove(name)
		return err
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}

// sweepApprovals removes hook/overlay handshake files older than maxAge.
// The hook deletes its own pending file on exit; this only reclaims the
// ones orphaned by a hard kill (taskkill, power loss). Veto files are
// removed with them - a stale veto must never deny a future request that
// happens to reuse a file name.
func sweepApprovals(maxAge time.Duration) int {
	directory := approvalsDir()
	if directory == "" {
		return 0
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0
	}
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		handshake := (strings.HasPrefix(name, approvalPendingPrefix) && strings.HasSuffix(name, ".json")) ||
			(strings.HasPrefix(name, approvalPendingPrefix) && strings.HasSuffix(name, ".deny")) ||
			(strings.HasPrefix(name, ".pet-deny-") && strings.HasSuffix(name, ".tmp"))
		if !handshake {
			continue
		}
		info, err := entry.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(directory, name)); err == nil {
			removed++
		}
	}
	return removed
}
