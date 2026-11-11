// Command pet-statusline is a native (no Node.js) replacement for
// reporters/antigravity-statusline.js.
//
// Background: on this machine every new process pays ~1.5s creation tax
// (active AV hooks), bare `node` cold start costs ~4.3s and any `cmd /c`
// hop costs ~7.7s. The Antigravity statusline runner executes the configured
// command via shell with a tight budget, so the Node script can never
// finish in time (exit 1, empty stderr = runner-side timeout kill).
// This shim finishes in milliseconds.
//
// Contract (mirrors reporters/antigravity-statusline.js, fail-fast, always
// exit 0):
//   - read stdin (<=1MiB, 2s), parse Antigravity status JSON loosely
//   - map agent_state/tool_confirmation_pending to state/reason
//   - write atomic runtime snapshot %LOCALAPPDATA%\AI-Terminal-Pet\runtime\
//   - dial per-user pipe, send one JSON line, expect exact "OK"
//   - ENOENT + state != closed -> launch AI-Terminal-Pet.exe once, retry once
//   - print one compact line to stdout (the visible statusline), never stderr
//
// Everything except the state mapping itself lives in internal/reporter, which
// is shared with cmd/pet-hook. That is deliberate: the redaction rules, the id
// bounds, the synthetic PID hash and the timestamp clamps are wire-protocol
// rules, and every copy of them is a chance to drift from reporters/protocol.js
// and from the other native binary.
//
// This command keeps its own single-shot delivery instead of reporter.Send on
// purpose. The statusline runs on every refresh inside a runner with a tight
// time budget, so it must not pay Send's 750 ms retry budget when the overlay
// is simply not running.
package main

import (
	"fmt"
	"os"
	"time"

	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/reporter"
	"ai-terminal-pet/internal/session"
)

func run() int {
	now := time.Now().UTC()
	input := reporter.DecodeStdin(reporter.ReadStdin(os.Stdin))

	state, reason := reporter.StatuslineStatus(input)
	payload := statuslinePayload(input, state, reason, now)
	line, err := payload.MarshalJSON()
	if err != nil {
		fmt.Println("pet idle")
		return 0
	}
	// Self-check: only emit payloads the server accepts.
	if _, err := session.ParsePayload(append(line, '\n'), now); err != nil {
		fmt.Println("pet idle")
		return 0
	}

	_ = reporter.WriteSnapshot(payload)

	pipePath := pipeserver.PipePathForUser()
	if !reporter.SendOnce(pipePath, line) && payload.State != session.StateClosed {
		_ = reporter.LaunchApplication()
		_ = reporter.SendOnce(pipePath, line)
	}

	fmt.Println("pet " + string(payload.State))
	return 0
}

// statuslinePayload builds the payload for one refresh. Every field is set
// raw: reporter.Payload.Finish applies the protocol's cleaning, bounding and
// timestamp clamps, which is exactly the same code cmd/pet-hook runs.
func statuslinePayload(input map[string]any, state session.State, reason string, now time.Time) reporter.Payload {
	sessionID := reporter.AntigravitySessionID(input)
	payload := reporter.NewPayload(now)
	payload.Source = "antigravity-statusline"
	payload.SessionID = sessionID
	payload.PID = reporter.ExplicitPID(input)
	if payload.PID == 0 {
		payload.PID = reporter.StableProcessID("antigravity-statusline:" + sessionID)
	}
	payload.ParentPID = reporter.ExplicitParentPID(input)
	if payload.ParentPID == 0 {
		payload.ParentPID = os.Getppid()
	}
	payload.ParentSessionID = reporter.ParentSessionID(input)
	payload.Project = reporter.StatuslineWorkspace(input)
	payload.Model, _ = reporter.ModelName(input["model"])
	payload.State = state
	payload.Reason = reason
	payload.StartedAt = statuslineTimestamp(input, now, "started_at", "startedAt", "session_start_time")
	payload.UpdatedAt = now
	payload.Tool = reporter.StatuslineToolName(input)
	payload.Terminal = "antigravity"
	return payload.Finish(now)
}

// statuslineTimestamp differs from reporter.Timestamp in one way that is worth
// keeping: the statusline simply ignores a start time more than five minutes in
// the future instead of clamping it, because a host with a skewed clock would
// otherwise show a session that has been running for hours. A non-positive
// epoch is ignored too (reporter.Timestamp's rule), which also means a 1969
// value can never reach the server as startedAt.
func statuslineTimestamp(input map[string]any, fallback time.Time, keys ...string) time.Time {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		parsed := reporter.Timestamp(value, time.Time{})
		if parsed.IsZero() {
			continue
		}
		if parsed.After(fallback.Add(5 * time.Minute)) {
			continue
		}
		return parsed
	}
	return fallback
}

func main() {
	os.Exit(run())
}
