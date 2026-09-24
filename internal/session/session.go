package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const SchemaVersion = "ai-terminal-pet.v1"
const MaxPayloadBytes = 1 << 20

type State string

const (
	StateIdle              State = "idle"
	StateThinking          State = "thinking"
	StateWorking           State = "working"
	StateToolUse           State = "tool_use"
	StateWaitingPermission State = "waiting_permission"
	StateError             State = "error"
	StateClosed            State = "closed"
)

type Session struct {
	SchemaVersion   string    `json:"schemaVersion"`
	Source          string    `json:"source"`
	SessionID       string    `json:"sessionId"`
	ParentSessionID string    `json:"parentSessionId"`
	PID             int       `json:"pid"`
	ParentPID       int       `json:"parentPid"`
	Project         string    `json:"project"`
	Title           string    `json:"title"`
	Model           string    `json:"model"`
	State           State     `json:"state"`
	Reason          string    `json:"reason"`
	StartedAt       time.Time `json:"startedAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Tool            string    `json:"tool"`
	Terminal        string    `json:"terminal"`
}

func ParsePayload(line []byte, now time.Time) (Session, error) {
	if len(line) == 0 {
		return Session{}, errors.New("empty payload")
	}
	if len(line) > MaxPayloadBytes {
		return Session{}, errors.New("payload too large")
	}
	trimmed := bytes.TrimSpace(line)
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var value Session
	if err := decoder.Decode(&value); err != nil {
		return Session{}, fmt.Errorf("decode payload: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Session{}, errors.New("payload must contain one JSON value")
	}
	return validate(value, now)
}

func (s State) Valid() bool {
	switch s {
	case StateIdle, StateThinking, StateWorking, StateToolUse, StateWaitingPermission, StateError, StateClosed:
		return true
	default:
		return false
	}
}

func (s State) Priority() int {
	switch s {
	case StateWaitingPermission:
		return 6
	case StateError:
		return 5
	case StateWorking:
		return 4
	case StateToolUse:
		return 3
	case StateThinking:
		return 2
	case StateIdle:
		return 1
	default:
		return 0
	}
}

func Aggregate(items []Session) State {
	selected := StateIdle
	selectedPriority := -1
	for _, item := range items {
		if item.State == StateClosed || !item.State.Valid() {
			continue
		}
		if priority := item.State.Priority(); priority > selectedPriority {
			selected = item.State
			selectedPriority = priority
		}
	}
	return selected
}

func Sort(items []Session) {
	sort.SliceStable(items, func(i, j int) bool {
		left, right := items[i], items[j]
		if left.State.Priority() != right.State.Priority() {
			return left.State.Priority() > right.State.Priority()
		}
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.After(right.UpdatedAt)
		}
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		if left.SessionID != right.SessionID {
			return left.SessionID < right.SessionID
		}
		return left.PID < right.PID
	})
}

func validate(value Session, now time.Time) (Session, error) {
	var err error
	if value.SchemaVersion == "" {
		value.SchemaVersion = SchemaVersion
	}
	if value.SchemaVersion != SchemaVersion {
		return Session{}, errors.New("unsupported schema version")
	}
	if value.Source, err = text("source", value.Source, 48, true); err != nil {
		return Session{}, err
	}
	if value.SessionID, err = text("sessionId", value.SessionID, 200, true); err != nil {
		return Session{}, err
	}
	if value.ParentSessionID, err = text("parentSessionId", value.ParentSessionID, 200, false); err != nil {
		return Session{}, err
	}
	if value.PID <= 0 || uint64(value.PID) > uint64(0xffffffff) {
		return Session{}, errors.New("pid must be a positive 32-bit value")
	}
	if value.ParentPID < 0 || uint64(value.ParentPID) > uint64(0xffffffff) {
		return Session{}, errors.New("parentPid must be a 32-bit value")
	}
	if value.Project, err = text("project", value.Project, 1024, false); err != nil {
		return Session{}, err
	}
	if value.Title, err = text("title", value.Title, 200, false); err != nil {
		return Session{}, err
	}
	if value.Model, err = text("model", value.Model, 200, false); err != nil {
		return Session{}, err
	}
	if !value.State.Valid() {
		return Session{}, errors.New("invalid state")
	}
	if value.Reason, err = text("reason", value.Reason, 120, false); err != nil {
		return Session{}, err
	}
	if value.Tool, err = text("tool", value.Tool, 120, false); err != nil {
		return Session{}, err
	}
	if value.Terminal, err = text("terminal", value.Terminal, 120, false); err != nil {
		return Session{}, err
	}
	if value.StartedAt.IsZero() || value.UpdatedAt.IsZero() {
		return Session{}, errors.New("timestamps are required")
	}
	if now.IsZero() {
		now = time.Now()
	}
	if value.UpdatedAt.Before(value.StartedAt) {
		return Session{}, errors.New("updatedAt precedes startedAt")
	}
	if value.StartedAt.After(now.Add(5*time.Minute)) || value.UpdatedAt.After(now.Add(5*time.Minute)) {
		return Session{}, errors.New("timestamp is too far in the future")
	}
	value.StartedAt = value.StartedAt.UTC()
	value.UpdatedAt = value.UpdatedAt.UTC()
	return value, nil
}

func text(name, value string, limit int, required bool) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("%s is not valid UTF-8", name)
	}
	value = strings.TrimSpace(value)
	if required && value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	if utf8.RuneCountInString(value) > limit {
		return "", fmt.Errorf("%s is too long", name)
	}
	for _, current := range value {
		if unicode.IsControl(current) {
			return "", fmt.Errorf("%s contains control characters", name)
		}
	}
	return value, nil
}
