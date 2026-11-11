package reporter

import (
	"strings"
)

// HookPayloads are the two host dialects the native hook speaks. Everything
// below is a port of the statusFor/isPermissionGate/toolText helpers in
// reporters/claude-hook.js and reporters/antigravity-hook.js: those scripts
// stay in the repository as the reference implementation, and the golden
// fixtures in approver_test.go / hook_test.go pin the two against each other.

// ClaudeStatus maps a Claude Code hook event onto a state/reason pair.
// Mirrors claude-hook.js statusFor().
func ClaudeStatus(input map[string]any) (string, string) {
	hook := strings.ToLower(StringField(input, "hook_event_name", "hookEventName"))
	notification := strings.ToLower(StringField(input, "notification_type", "notificationType"))
	match := func(pattern string) bool { return strings.Contains(notification, pattern) }
	switch {
	case strings.Contains(hook, "sessionend") || match("session_end"):
		return "closed", "session_end"
	case Truthy(input["error"]):
		return "error", "error"
	case strings.Contains(hook, "permission") || match("permission") || match("waiting") || match("elicitation"):
		return "waiting_permission", "permission"
	case strings.Contains(hook, "pretooluse") || match("tool"):
		return "tool_use", "tool"
	case strings.Contains(hook, "posttooluse") || Truthy(input["stop_hook_active"]):
		return "working", "tool_result"
	case strings.Contains(hook, "userprompt"):
		return "thinking", "prompt_submitted"
	case strings.Contains(hook, "stop") || match("idle") || match("complete") || match("auth"):
		return "idle", "idle"
	case StringField(input, "tool_name") != "":
		return "tool_use", "tool"
	default:
		return "idle", "hook"
	}
}

// ClaudeToolText gathers what the blacklist sees: the tool name plus its
// input. The input shape varies by Claude version, so stringify the lot -
// a missed field could auto-approve a destructive command.
func ClaudeToolText(input map[string]any) string {
	parts := make([]string, 0, 2)
	if name := StringField(input, "tool_name"); name != "" {
		parts = append(parts, name)
	}
	if raw, ok := input["tool_input"]; ok && raw != nil {
		parts = append(parts, CompactJSON(raw))
	}
	return strings.Join(parts, "\n")
}

// ClaudeModelName mirrors modelName(): "provider/model" when the host sends a
// nested object, the plain string otherwise.
func ClaudeModelName(value any) string {
	text, _ := ModelName(value)
	return text
}

// ModelName is the shared implementation for both dialects.
func ModelName(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case map[string]any:
		model := StringField(typed, "model", "modelID", "modelId", "id", "name")
		provider := StringField(typed, "provider", "providerID", "providerId")
		parts := make([]string, 0, 2)
		if provider != "" {
			parts = append(parts, provider)
		}
		if model != "" {
			parts = append(parts, model)
		}
		if len(parts) == 0 {
			return "", false
		}
		return strings.Join(parts, "/"), true
	default:
		return "", false
	}
}

// AntigravityStatus maps an Antigravity hook event onto a state/reason pair.
// Mirrors antigravity-hook.js statusFor().
func AntigravityStatus(input map[string]any) (string, string) {
	if termination := StringField(input, "terminationReason", "termination_reason"); termination != "" {
		return "closed", "termination_reason"
	}
	if Truthy(input["error"]) {
		return "error", "error"
	}
	if Truthy(input["fullyIdle"]) || Truthy(input["fully_idle"]) {
		return "idle", "fully_idle"
	}
	if tool := ObjectField(input, "toolCall"); tool != nil && StringField(tool, "name") != "" {
		return "tool_use", "tool_call"
	}
	return "idle", "hook"
}

// AntigravityWorkspace mirrors workspacePath(): the value may be a string or
// an object carrying cwd/current_dir/project_dir/path.
func AntigravityWorkspace(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]any:
		return StringField(typed, "cwd", "current_dir", "project_dir", "path")
	default:
		return ""
	}
}

// AntigravityIsPermissionGate mirrors isPermissionGate(): only gated events
// enter the veto window; plain toolCall observations stay fire-and-forget.
func AntigravityIsPermissionGate(input map[string]any) bool {
	if Truthy(input["tool_confirmation_pending"]) {
		return true
	}
	if Truthy(input["permission_required"]) || Truthy(input["permissionRequired"]) {
		return true
	}
	if tool := ObjectField(input, "toolCall"); tool != nil {
		if Truthy(tool["requiresApproval"]) || Truthy(tool["requiresConfirmation"]) {
			return true
		}
	}
	for _, key := range []string{"decision_request", "decisionRequest", "prompt_type", "promptType", "event"} {
		if value, ok := input[key].(string); ok {
			lower := strings.ToLower(value)
			if strings.Contains(lower, "permission") || strings.Contains(lower, "confirm") ||
				strings.Contains(lower, "approval") || strings.Contains(lower, "waiting") {
				return true
			}
		}
	}
	return false
}

// AntigravityToolText mirrors toolText(): tool name, tool input and the raw
// command, because Antigravity sometimes sends the command outside toolCall.
func AntigravityToolText(input map[string]any) string {
	parts := make([]string, 0, 3)
	if tool := ObjectField(input, "toolCall"); tool != nil {
		if name := StringField(tool, "name"); name != "" {
			parts = append(parts, name)
		}
		if raw, ok := tool["input"]; ok && raw != nil {
			parts = append(parts, CompactJSON(raw))
		}
	}
	if command := StringField(input, "command"); command != "" {
		parts = append(parts, command)
	}
	return strings.Join(parts, "\n")
}

// ClaudeSessionID resolves the session id the store will key on.
func ClaudeSessionID(input map[string]any) string {
	if id := StringField(input, "session_id", "sessionId"); id != "" {
		return id
	}
	return "global"
}

// AntigravitySessionID resolves the same for Antigravity, which also uses
// conversation_id.
func AntigravitySessionID(input map[string]any) string {
	if id := StringField(input, "conversation_id", "session_id", "sessionId"); id != "" {
		return id
	}
	return "global"
}

// ParentSessionID reads the optional parent link both dialects may carry.
func ParentSessionID(input map[string]any) string {
	return StringField(input, "parent_session_id", "parentSessionId")
}

// ExplicitPID returns the payload pid only when the host really sent one.
// Zero means "synthesize from the session id", which is what the reporters do.
func ExplicitPID(input map[string]any) int {
	if raw, ok := input["pid"]; ok {
		return PositiveInt(raw, 0)
	}
	return 0
}

// ExplicitParentPID returns the payload parent pid, or 0 for "use mine".
func ExplicitParentPID(input map[string]any) int {
	if raw, ok := input["parentPid"]; ok {
		return PositiveInt(raw, 0)
	}
	return 0
}
