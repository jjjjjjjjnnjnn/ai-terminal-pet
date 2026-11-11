package reporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"ai-terminal-pet/internal/session"
)

// Payload is one report, in the shape internal/session expects on the wire.
// The tags must stay identical to session.Session: the server decodes with
// DisallowUnknownFields, so an extra field is not a warning but a rejected
// report.
type Payload struct {
	SchemaVersion   string `json:"schemaVersion"`
	Source          string `json:"source"`
	SessionID       string `json:"sessionId"`
	ParentSessionID string `json:"parentSessionId"`
	PID             int    `json:"pid"`
	ParentPID       int    `json:"parentPid"`
	Project         string `json:"project"`
	Title           string `json:"title"`
	Model           string `json:"model"`
	State           session.State
	Reason          string    `json:"reason"`
	StartedAt       time.Time `json:"startedAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Tool            string    `json:"tool"`
	Terminal        string    `json:"terminal"`
}

// MarshalJSON keeps State in the same position protocol.js emits it, so a
// payload produced here and one produced by the Node reporter diff cleanly in
// a test or a bug report. Field order is cosmetic on the wire, but a stable
// order makes the two implementations comparable by eye.
func (p Payload) MarshalJSON() ([]byte, error) {
	type wire struct {
		SchemaVersion   string    `json:"schemaVersion"`
		Source          string    `json:"source"`
		SessionID       string    `json:"sessionId"`
		ParentSessionID string    `json:"parentSessionId"`
		PID             int       `json:"pid"`
		ParentPID       int       `json:"parentPid"`
		Project         string    `json:"project"`
		Title           string    `json:"title"`
		Model           string    `json:"model"`
		State           string    `json:"state"`
		Reason          string    `json:"reason"`
		StartedAt       time.Time `json:"startedAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
		Tool            string    `json:"tool"`
		Terminal        string    `json:"terminal"`
	}
	return json.Marshal(wire{
		SchemaVersion:   p.SchemaVersion,
		Source:          p.Source,
		SessionID:       p.SessionID,
		ParentSessionID: p.ParentSessionID,
		PID:             p.PID,
		ParentPID:       p.ParentPID,
		Project:         p.Project,
		Title:           p.Title,
		Model:           p.Model,
		State:           string(p.State),
		Reason:          p.Reason,
		StartedAt:       p.StartedAt.UTC(),
		UpdatedAt:       p.UpdatedAt.UTC(),
		Tool:            p.Tool,
		Terminal:        p.Terminal,
	})
}

// NewPayload builds a payload with the schema version, the empty-string
// defaults and both timestamps already normalized. Callers set the fields that
// identify the session and then call Finish.
func NewPayload(now time.Time) Payload {
	return Payload{
		SchemaVersion: session.SchemaVersion,
		Source:        "unknown",
		SessionID:     "unknown",
		State:         session.StateIdle,
		StartedAt:     now.UTC(),
		UpdatedAt:     now.UTC(),
	}
}

// Finish applies the protocol's last three rules and returns the payload:
// bounded ids, cleaned text fields, clamped timestamps. Everything the server
// would reject for being out of range is corrected here instead, because a
// reporter that fails validation files nothing at all.
func (p Payload) Finish(now time.Time) Payload {
	p.SchemaVersion = session.SchemaVersion
	p.Source = CleanToken(p.Source, 48)
	if p.Source == "" {
		p.Source = "unknown"
	}
	p.SessionID = BoundedID(p.SessionID, 200)
	if parent := CleanText(p.ParentSessionID, 200); parent != "" {
		p.ParentSessionID = BoundedID(parent, 200)
	} else {
		p.ParentSessionID = ""
	}
	if p.PID <= 0 || uint64(p.PID) > uint64(0xffffffff) {
		// Same seed the reporter scripts use, so a session reported once by
		// the Node hook and once by this binary lands on one store entry
		// instead of two.
		p.PID = StableProcessID(p.Source + ":" + p.SessionID)
	}
	if p.ParentPID < 0 || uint64(p.ParentPID) > uint64(0xffffffff) {
		p.ParentPID = 0
	}
	p.Project = CleanText(p.Project, 1024)
	p.Title = CleanText(p.Title, 200)
	p.Model = CleanText(p.Model, 200)
	if !p.State.Valid() {
		p.State = session.StateIdle
	}
	p.Reason = CleanToken(p.Reason, 120)
	p.Tool = CleanText(p.Tool, 120)
	p.Terminal = CleanText(p.Terminal, 120)
	p.StartedAt, p.UpdatedAt = NormalizeTimestamps(p.StartedAt, p.UpdatedAt, now)
	return p
}

// Validate reports whether the server would accept this payload. The native
// reporters run it before writing anything: a rejected payload must not leave
// a snapshot behind that the scanner would later keep re-reading.
func (p Payload) Validate(now time.Time) error {
	line, err := p.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = session.ParsePayload(append(line, '\n'), now)
	return err
}

// RuntimeDirectory is %LOCALAPPDATA%\AI-Terminal-Pet\runtime, the directory the
// overlay's snapshot scanner watches. An empty result means the environment is
// unusable and every disk-side operation becomes a no-op.
func RuntimeDirectory() string {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return ""
	}
	return filepath.Join(root, "AI-Terminal-Pet", "runtime")
}

// SnapshotPath is the file the overlay would read this payload back from. The
// name encodes source, pid and session id so a stale file for a dead session
// can be aged out individually.
func SnapshotPath(p Payload) string {
	directory := RuntimeDirectory()
	if directory == "" {
		return ""
	}
	name := FilenamePart(p.Source, "unknown") + "-" +
		itoa(p.PID) + "-" +
		FilenamePart(p.SessionID, "unknown") + ".json"
	return filepath.Join(directory, name)
}

// WriteSnapshot stores the payload atomically: temp file in the same
// directory, then rename. The scanner never sees a half-written JSON body, and
// a crash leaves only a ".pet-*.tmp" that the overlay sweeps on its next start.
func WriteSnapshot(p Payload) bool {
	target := SnapshotPath(p)
	if target == "" {
		return false
	}
	line, err := p.MarshalJSON()
	if err != nil {
		return false
	}
	directory := filepath.Dir(target)
	tmp, err := os.CreateTemp(directory, ".pet-*.tmp")
	if err != nil {
		return false
	}
	name := tmp.Name()
	_ = tmp.Chmod(0600)
	if _, err := tmp.Write(append(line, '\n')); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return false
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return false
	}
	if err := os.Rename(name, target); err != nil {
		_ = os.Remove(name)
		return false
	}
	return true
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var digits [12]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		index--
		digits[index] = '-'
	}
	return string(digits[index:])
}
