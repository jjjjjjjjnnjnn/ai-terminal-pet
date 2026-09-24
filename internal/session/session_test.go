package session

import (
	"strings"
	"testing"
	"time"
)

const validPayload = `{"schemaVersion":"ai-terminal-pet.v1","source":"opencode","sessionId":"child-1","parentSessionId":"root-1","pid":1200,"parentPid":900,"project":"C:\\work","title":"Review","model":"provider/model","state":"thinking","reason":"prompt","startedAt":"2026-09-24T10:00:00Z","updatedAt":"2026-09-24T10:01:00Z","tool":"read","terminal":"Windows Terminal"}`

func TestParsePayload(t *testing.T) {
	now := time.Date(2026, 9, 24, 10, 2, 0, 0, time.UTC)
	item, err := ParsePayload([]byte(validPayload+"\n"), now)
	if err != nil {
		t.Fatal(err)
	}
	if item.Source != "opencode" || item.SessionID != "child-1" || item.ParentSessionID != "root-1" {
		t.Fatalf("unexpected identity: %#v", item)
	}
	if item.PID != 1200 || item.ParentPID != 900 || item.State != StateThinking {
		t.Fatalf("unexpected runtime fields: %#v", item)
	}
	if item.Project != `C:\work` || item.Tool != "read" || item.Terminal != "Windows Terminal" {
		t.Fatalf("unexpected text fields: %#v", item)
	}
}

func TestParsePayloadValidation(t *testing.T) {
	now := time.Date(2026, 9, 24, 10, 2, 0, 0, time.UTC)
	tests := map[string]string{
		"empty source":     strings.Replace(validPayload, `"source":"opencode"`, `"source":""`, 1),
		"invalid state":    strings.Replace(validPayload, `"state":"thinking"`, `"state":"busy"`, 1),
		"invalid pid":      strings.Replace(validPayload, `"pid":1200`, `"pid":0`, 1),
		"oversized pid":    strings.Replace(validPayload, `"pid":1200`, `"pid":4294967296`, 1),
		"oversized parent": strings.Replace(validPayload, `"parentPid":900`, `"parentPid":4294967296`, 1),
		"invalid schema":   strings.Replace(validPayload, `"ai-terminal-pet.v1"`, `"ai-terminal-pet.v2"`, 1),
		"future timestamp": strings.Replace(validPayload, `"updatedAt":"2026-09-24T10:01:00Z"`, `"updatedAt":"2026-09-24T10:10:00Z"`, 1),
		"reversed times":   strings.Replace(validPayload, `"startedAt":"2026-09-24T10:00:00Z"`, `"startedAt":"2026-09-24T10:03:00Z"`, 1),
		"unknown field":    strings.TrimSuffix(validPayload, "}") + `,"extra":true}`,
		"trailing object":  validPayload + `{}`,
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePayload([]byte(payload), now); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestStatePriorityOrder(t *testing.T) {
	// Locks the documented order: waiting(6) > error(5) > working(4) >
	// tool_use(3) > thinking(2) > idle(1) > closed/invalid(0).
	want := map[State]int{
		StateWaitingPermission: 6,
		StateError:             5,
		StateWorking:           4,
		StateToolUse:           3,
		StateThinking:          2,
		StateIdle:              1,
		StateClosed:            0,
	}
	for state, priority := range want {
		if got := state.Priority(); got != priority {
			t.Fatalf("state %q priority = %d, want %d", state, got, priority)
		}
	}
}

func TestStateAggregation(t *testing.T) {
	items := []Session{
		{State: StateIdle},
		{State: StateThinking},
		{State: StateToolUse},
		{State: StateWorking},
		{State: StateError},
		{State: StateWaitingPermission},
		{State: StateClosed},
	}
	if got := Aggregate(items); got != StateWaitingPermission {
		t.Fatalf("got %s", got)
	}
	items = items[:5]
	if got := Aggregate(items); got != StateError {
		t.Fatalf("got %s", got)
	}
	items = items[:3]
	if got := Aggregate(items); got != StateToolUse {
		t.Fatalf("got %s", got)
	}
}

func TestSortSessions(t *testing.T) {
	base := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	items := []Session{
		{Source: "z", SessionID: "old", State: StateIdle, UpdatedAt: base},
		{Source: "a", SessionID: "new", State: StateIdle, UpdatedAt: base.Add(time.Minute)},
		{Source: "b", SessionID: "tool", State: StateToolUse, UpdatedAt: base},
		{Source: "c", SessionID: "wait", State: StateWaitingPermission, UpdatedAt: base},
	}
	Sort(items)
	if items[0].State != StateWaitingPermission || items[1].State != StateToolUse {
		t.Fatalf("priority order is wrong: %#v", items)
	}
	if items[2].SessionID != "new" || items[3].SessionID != "old" {
		t.Fatalf("time order is wrong: %#v", items)
	}
}
