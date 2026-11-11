// Command pet-hook is the native (no Node.js) replacement for
// reporters/claude-hook.js and reporters/antigravity-hook.js.
//
// Why it exists: on a machine with active AV hooks every process creation
// costs about a second (measured here: `node -e exit` 0.9-1.2 s), and these
// hooks run on *every* event a host CLI emits - a single tool call is
// PreToolUse plus PostToolUse. cmd/pet-statusline already showed the native
// route is an order of magnitude cheaper, so the two hook dialects move here
// while the JavaScript files stay in the repository as the reference
// implementation and as the fallback for hosts that cannot run an .exe.
//
// Contract (mirrors the two scripts, fail-fast, always exit 0):
//   - read stdin (<=1MiB, 2s), parse the host's hook JSON loosely
//   - map the event onto state/reason for the --mode dialect
//   - on a permission gate: alarm the panel, then either fall through to the
//     host's own prompt (Claude, dangerous command), deny (Antigravity,
//     dangerous command), or wait out the veto window and answer allow/deny
//   - write the runtime snapshot, dial the per-user pipe, expect exact "OK"
//   - print the one JSON line the host expects, never anything on stderr
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"ai-terminal-pet/internal/reporter"
	"ai-terminal-pet/internal/session"
)

const (
	modeClaude      = "claude"
	modeAntigravity = "antigravity"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout))
}

// sender is the delivery function run() uses. Tests replace it so the command
// can be exercised without a pipe, without the 750 ms retry budget and without
// touching the real runtime directory.
var sender = func(payload reporter.Payload) bool {
	return reporter.Send(payload, reporter.SendOptions{
		Retry:  reporter.DefaultRetryBudget,
		Launch: payload.State != session.StateClosed,
	})
}

// run is the whole command, with its streams injected so tests can drive it
// without a process boundary.
func run(args []string, stdin io.Reader, stdout io.Writer) int {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printUsage(stdout)
			return 0
		}
	}
	mode, ok := parseMode(args)
	if !ok {
		// Never fail a host CLI: an unknown flag means "report nothing and
		// say allow", which is what a hook that cannot decide must say.
		fmt.Fprintln(stdout, defaultAnswer(modeAntigravity))
		return 0
	}
	raw := reporter.ReadStdin(stdin)
	input := reporter.DecodeStdin(raw)
	now := time.Now().UTC()

	switch mode {
	case modeClaude:
		state, reason := reporter.ClaudeStatus(input)
		payload := claudePayload(input, state, reason, now)
		if state == string(session.StateWaitingPermission) {
			if answer, done := decide(payload, reporter.ClaudeToolText(input), mode); done {
				fmt.Fprintln(stdout, answer)
				return 0
			}
		}
		_ = sender(payload)
	default:
		state, reason := reporter.AntigravityStatus(input)
		payload := antigravityPayload(input, state, reason, now)
		if reporter.AntigravityIsPermissionGate(input) {
			payload.State = session.StateWaitingPermission
			payload.Reason = "permission"
			if answer, done := decide(payload, reporter.AntigravityToolText(input), mode); done {
				fmt.Fprintln(stdout, answer)
				return 0
			}
		}
		_ = sender(payload)
	}

	fmt.Fprintln(stdout, defaultAnswer(mode))
	return 0
}

// decide handles one permission gate. done is true when the hook has already
// printed its answer and must exit; done is false when auto-approve is off and
// the host's own prompt should take over (nothing has been printed yet).
//
// The two dialects differ in what a dangerous command means, and that
// difference is load-bearing: Claude Code re-asks natively, while Antigravity
// treats this hook as the decider, so "manual" there has to deny.
func decide(payload reporter.Payload, toolText, mode string) (string, bool) {
	plan, timeout := reporter.ApprovalPlan(toolText)
	source := "claude-hook"
	fallback := "{}"
	if mode == modeAntigravity {
		source = "antigravity-hook"
		fallback = `{"decision":"deny"}`
	}
	request := reporter.ApprovalRequest{
		Source:    source,
		SessionID: payload.SessionID,
		PID:       payload.PID,
		Tool:      payload.Tool,
	}
	switch plan {
	case reporter.ApprovalAuto:
		// Report first, unawaited in spirit: the overlay needs the session on
		// screen to start its countdown while this hook waits out the window.
		_ = sender(payload)
		if reporter.WaitForApprovalDecision(request, timeout) == "deny" {
			return `{"decision":"deny"}`, true
		}
		return `{"decision":"allow"}`, true
	case reporter.ApprovalManual:
		// Dangerous command with auto-approve on: alarm the panel, then get
		// out of the way. Claude falls through to its native prompt; the
		// overlay's countdown expires harmlessly because no request was filed.
		payload.Reason = "permission_dangerous"
		_ = sender(payload)
		return fallback, true
	default:
		return "", false
	}
}

// defaultAnswer is what a hook says when it has no opinion: Claude accepts an
// empty object, Antigravity requires an explicit decision.
func defaultAnswer(mode string) string {
	if mode == modeClaude {
		return "{}"
	}
	return `{"decision":"allow"}`
}

func parseMode(args []string) (string, bool) {
	mode := modeClaude
	for _, arg := range args {
		switch {
		case arg == "--help" || arg == "-h":
			// run() already answered --help before reaching here; this branch
			// exists so an unusual combination ("--mode=claude --help") still
			// resolves instead of failing.
		case strings.HasPrefix(arg, "--mode="):
			mode = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(arg, "--mode=")))
		case arg == "--claude":
			mode = modeClaude
		case arg == "--antigravity":
			mode = modeAntigravity
		default:
			return mode, false
		}
	}
	if mode != modeClaude && mode != modeAntigravity {
		return modeClaude, false
	}
	return mode, true
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "pet-hook - native reporter for Claude Code and Antigravity hooks")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  pet-hook --mode=claude        read one hook event on stdin, print its verdict")
	fmt.Fprintln(out, "  pet-hook --mode=antigravity   same, in the Antigravity decision dialect")
	fmt.Fprintln(out, "  pet-hook --help               this text")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Exit code is always 0: a reporter must never break its host CLI.")
}

func claudePayload(input map[string]any, state, reason string, now time.Time) reporter.Payload {
	sessionID := reporter.ClaudeSessionID(input)
	payload := reporter.NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = sessionID
	payload.PID = reporter.ExplicitPID(input)
	if payload.PID == 0 {
		payload.PID = reporter.StableProcessID("claude-hook:" + sessionID)
	}
	payload.ParentPID = reporter.ExplicitParentPID(input)
	if payload.ParentPID == 0 {
		payload.ParentPID = os.Getppid()
	}
	payload.ParentSessionID = reporter.ParentSessionID(input)
	payload.Project = reporter.StringField(input, "cwd")
	if model, ok := reporter.ModelName(input["model"]); ok {
		payload.Model = model
	}
	payload.State = session.State(state)
	payload.Reason = reason
	payload.StartedAt = reporter.Timestamp(firstPresent(input, "started_at", "startedAt"), now)
	payload.UpdatedAt = now
	payload.Tool = reporter.StringField(input, "tool_name")
	payload.Terminal = "claude"
	return payload.Finish(now)
}

func antigravityPayload(input map[string]any, state, reason string, now time.Time) reporter.Payload {
	sessionID := reporter.AntigravitySessionID(input)
	payload := reporter.NewPayload(now)
	payload.Source = "antigravity-hook"
	payload.SessionID = sessionID
	payload.PID = reporter.ExplicitPID(input)
	if payload.PID == 0 {
		payload.PID = reporter.StableProcessID("antigravity-hook:" + sessionID)
	}
	payload.ParentPID = reporter.ExplicitParentPID(input)
	if payload.ParentPID == 0 {
		payload.ParentPID = os.Getppid()
	}
	payload.ParentSessionID = reporter.ParentSessionID(input)
	project := reporter.AntigravityWorkspace(input["workspace"])
	if project == "" {
		project = reporter.AntigravityWorkspace(input["cwd"])
	}
	payload.Project = project
	if model, ok := reporter.ModelName(input["model"]); ok {
		payload.Model = model
	}
	payload.State = session.State(state)
	payload.Reason = reason
	payload.StartedAt = reporter.Timestamp(firstPresent(input, "started_at", "startedAt", "session_start_time"), now)
	payload.UpdatedAt = now
	if tool := reporter.ObjectField(input, "toolCall"); tool != nil {
		payload.Tool = reporter.StringField(tool, "name")
	}
	payload.Terminal = "antigravity"
	return payload.Finish(now)
}

func firstPresent(input map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := input[key]; ok && value != nil {
			return value
		}
	}
	return nil
}
