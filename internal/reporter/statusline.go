package reporter

import (
	"regexp"
	"strings"

	"ai-terminal-pet/internal/session"
)

// The Antigravity *statusline* dialect, which is not the same thing as the
// Antigravity *hook* dialect in hook.go. The statusline is fed one JSON blob per
// refresh with an agent_state field, and it must produce a state even when that
// field is missing or shaped differently between host versions. The hook
// dialect instead speaks in events.
//
// This is a port of statusFor() in reporters/antigravity-statusline.js, which
// stays in the repository as the reference implementation.

// statuslineStates maps the agent_state vocabulary onto session states. The
// order matters: "complete" must be tested before the generic idle bucket and
// "permission" before "tool", so the list is evaluated top to bottom rather
// than as a map.
var statuslineStates = []struct {
	pattern *regexp.Regexp
	state   session.State
}{
	{regexp.MustCompile(`(closed|terminated|finished|complete|done|idle)`), session.StateIdle},
	{regexp.MustCompile(`(error|failed|failure)`), session.StateError},
	{regexp.MustCompile(`(permission|confirm|waiting)`), session.StateWaitingPermission},
	{regexp.MustCompile(`(tool)`), session.StateToolUse},
	{regexp.MustCompile(`(working|executing|busy|running|processing)`), session.StateWorking},
	{regexp.MustCompile(`(thinking|reasoning|generat|stream)`), session.StateThinking},
}

// StatuslineStatus maps one statusline payload onto a state/reason pair.
// Mirrors statusFor() in reporters/antigravity-statusline.js.
func StatuslineStatus(input map[string]any) (session.State, string) {
	if pending, ok := input["tool_confirmation_pending"].(bool); ok && pending {
		return session.StateWaitingPermission, "tool_confirmation_pending"
	}
	return StatuslineAgentState(input["agent_state"]), "agent_state"
}

// StatuslineAgentState interprets the agent_state value on its own. It is
// exported separately because the reason string differs ("tool_confirmation_
// pending" vs "agent_state") and only the caller knows which one applies.
func StatuslineAgentState(raw any) session.State {
	var text string
	switch value := raw.(type) {
	case string:
		text = value
	case map[string]any:
		text = StringField(value, "type", "state", "status", "name")
	case nil:
		text = ""
	default:
		text = ScalarString(value)
	}
	lower := strings.ToLower(text)
	for _, entry := range statuslineStates {
		if entry.pattern.MatchString(lower) {
			return entry.state
		}
	}
	return session.StateIdle
}

// StatuslineWorkspace mirrors workspacePath(): the project may arrive as a
// workspace string, a workspace object, or a bare cwd/current_dir field, and
// the order of those fallbacks is what decides which one wins.
func StatuslineWorkspace(input map[string]any) string {
	if project := AntigravityWorkspace(input["workspace"]); strings.TrimSpace(project) != "" {
		return project
	}
	return StringField(input, "cwd", "current_dir", "project_dir", "path")
}

// StatuslineToolName mirrors toolName(): a bare tool_name wins, otherwise the
// name inside toolCall / tool_call.
func StatuslineToolName(input map[string]any) string {
	if name := StringField(input, "tool_name"); name != "" {
		return name
	}
	if call := ObjectField(input, "toolCall", "tool_call"); call != nil {
		return StringField(call, "name")
	}
	return ""
}
