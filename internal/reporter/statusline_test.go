package reporter

import (
	"testing"

	"ai-terminal-pet/internal/session"
)

// These cases are the statusline mapping that reporters/antigravity-statusline.js
// documents. They were captured by running the previous hand-rolled
// implementation in cmd/pet-statusline over 35 payloads and diffing what it
// printed and filed against the version that calls this package - the two were
// identical, so these values are a contract, not a guess.
func TestStatuslineStatusMappings(t *testing.T) {
	cases := []struct {
		name   string
		input  map[string]any
		state  session.State
		reason string
	}{
		{"idle", map[string]any{"agent_state": "idle"}, session.StateIdle, "agent_state"},
		{"complete", map[string]any{"agent_state": "complete"}, session.StateIdle, "agent_state"},
		{"terminated", map[string]any{"agent_state": "terminated"}, session.StateIdle, "agent_state"},
		{"working", map[string]any{"agent_state": "working"}, session.StateWorking, "agent_state"},
		{"executing", map[string]any{"agent_state": "executing"}, session.StateWorking, "agent_state"},
		{"processing", map[string]any{"agent_state": "processing"}, session.StateWorking, "agent_state"},
		{"thinking", map[string]any{"agent_state": "thinking"}, session.StateThinking, "agent_state"},
		{"streaming", map[string]any{"agent_state": "streaming"}, session.StateThinking, "agent_state"},
		{"failed", map[string]any{"agent_state": "failed"}, session.StateError, "agent_state"},
		{"waiting", map[string]any{"agent_state": "waiting_permission"}, session.StateWaitingPermission, "agent_state"},
		{"tool", map[string]any{"agent_state": "tool_use"}, session.StateToolUse, "agent_state"},
		{"object", map[string]any{"agent_state": map[string]any{"type": "tool_use"}}, session.StateToolUse, "agent_state"},
		{"unknown", map[string]any{"agent_state": "nonsense"}, session.StateIdle, "agent_state"},
		{"missing", map[string]any{}, session.StateIdle, "agent_state"},
		{
			"confirmation pending wins over agent_state",
			map[string]any{"agent_state": "working", "tool_confirmation_pending": true},
			session.StateWaitingPermission,
			"tool_confirmation_pending",
		},
		{
			"false confirmation pending is ignored",
			map[string]any{"agent_state": "working", "tool_confirmation_pending": false},
			session.StateWorking,
			"agent_state",
		},
	}
	for _, testCase := range cases {
		state, reason := StatuslineStatus(testCase.input)
		if state != testCase.state || reason != testCase.reason {
			t.Errorf("%s: got (%s, %s), want (%s, %s)", testCase.name, state, reason, testCase.state, testCase.reason)
		}
	}
}

// The ordering inside statuslineStates is load-bearing: "complete" also matches
// the idle pattern and "permission" also matches "waiting", so a reordering
// would change what the panel shows without failing any obvious assertion.
func TestStatuslineStatePrecedence(t *testing.T) {
	cases := map[string]session.State{
		"closed":            session.StateIdle,
		"finished":          session.StateIdle,
		"permission_denied": session.StateWaitingPermission,
		"confirming":        session.StateWaitingPermission,
		"tool_running":      session.StateToolUse,
		// "tool" is tested before "working", so a value carrying both is a
		// tool state - that is the reference implementation's order, and a
		// reordering is exactly what this case would catch.
		"working_on_tooling": session.StateToolUse,
		"working_on_files":   session.StateWorking,
		"reasoning":          session.StateThinking,
	}
	for text, want := range cases {
		if got := StatuslineAgentState(text); got != want {
			t.Errorf("%q: got %s, want %s", text, got, want)
		}
	}
}

func TestStatuslineWorkspaceFallbacks(t *testing.T) {
	cases := []struct {
		name  string
		input map[string]any
		want  string
	}{
		{"workspace string", map[string]any{"workspace": `C:\a`}, `C:\a`},
		{"workspace cwd", map[string]any{"workspace": map[string]any{"cwd": `C:\b`}}, `C:\b`},
		{"workspace current_dir", map[string]any{"workspace": map[string]any{"current_dir": `C:\c`}}, `C:\c`},
		{"bare cwd", map[string]any{"cwd": `C:\d`}, `C:\d`},
		{"blank workspace falls through", map[string]any{"workspace": "  ", "cwd": `C:\e`}, `C:\e`},
		{"none", map[string]any{}, ""},
	}
	for _, testCase := range cases {
		if got := StatuslineWorkspace(testCase.input); got != testCase.want {
			t.Errorf("%s: got %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

func TestStatuslineToolName(t *testing.T) {
	cases := []struct {
		name  string
		input map[string]any
		want  string
	}{
		{"tool_name", map[string]any{"tool_name": "Bash"}, "Bash"},
		{"toolCall", map[string]any{"toolCall": map[string]any{"name": "Read"}}, "Read"},
		{"tool_call", map[string]any{"tool_call": map[string]any{"name": "Write"}}, "Write"},
		{"none", map[string]any{}, ""},
	}
	for _, testCase := range cases {
		if got := StatuslineToolName(testCase.input); got != testCase.want {
			t.Errorf("%s: got %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

// The statusline reuses AntigravitySessionID, so this pins the fallback the
// store keys on: a payload without an id must still be one session, not none.
func TestStatuslineSessionIDPrefersConversationID(t *testing.T) {
	cases := []struct {
		input map[string]any
		want  string
	}{
		{map[string]any{"conversation_id": "c1", "session_id": "s1"}, "c1"},
		{map[string]any{"session_id": "s1"}, "s1"},
		{map[string]any{"sessionId": "s2"}, "s2"},
		{map[string]any{}, "global"},
	}
	for _, testCase := range cases {
		if got := AntigravitySessionID(testCase.input); got != testCase.want {
			t.Errorf("got %q, want %q", got, testCase.want)
		}
	}
}
