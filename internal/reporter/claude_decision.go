package reporter

import "encoding/json"

// ClaudeDeniedMessage is what a denied permission gate says in the terminal.
// It has to travel in the answer itself: the host does not echo the hook's
// stderr, and a silent deny reads like a bug to the person in the terminal.
const ClaudeDeniedMessage = "Denied from the AI Terminal Pet panel."

// ClaudeApprovalAnswer is the JSON a Claude Code permission hook prints to
// decide one gate.
//
// Claude Code 2.1.312 does not accept the legacy {"decision":"allow"} /
// {"decision":"deny"} shape on PermissionRequest. Measured 2026-11-18 with a
// local Messages API stub: the same hook, same payload, printed the legacy
// allow and the tool was still denied as if no hook had answered; the shape
// below allowed it, and its deny twin surfaced ClaudeDeniedMessage. Mixing
// both shapes in one object is also rejected, so this emits only the accepted
// one. PreToolUse is a different dialect ({"decision":"block"}) that this
// project does not emit anywhere.
//
// eventName is the incoming hook_event_name, so the answer names the event the
// host actually asked about; an empty value means PermissionRequest.
//
// reporters/protocol.js has a byte-for-byte mirror; the same literal is pinned
// on both sides (claude_decision_test.go and reporters/claude-approval.test.mjs).
func ClaudeApprovalAnswer(eventName, behavior string) string {
	if eventName == "" {
		eventName = "PermissionRequest"
	}
	type claudeDecision struct {
		Behavior string `json:"behavior"`
		Message  string `json:"message,omitempty"`
	}
	type hookSpecificOutput struct {
		HookEventName string         `json:"hookEventName"`
		Decision      claudeDecision `json:"decision"`
	}
	answer := struct {
		HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
	}{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName: eventName,
			Decision:      claudeDecision{Behavior: behavior},
		},
	}
	if behavior == "deny" {
		answer.HookSpecificOutput.Decision.Message = ClaudeDeniedMessage
	}
	encoded, err := json.Marshal(answer)
	if err != nil {
		// Unreachable for these types, and a hook that cannot answer must not
		// invent a decision: an empty object defers to the host's own prompt.
		return "{}"
	}
	return string(encoded)
}
