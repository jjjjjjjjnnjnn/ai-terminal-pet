package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-terminal-pet/internal/reporter"
)

// captured is one payload the command tried to send, with the answer faked so
// a test never needs a pipe, the 750 ms retry budget or a real runtime dir.
type captured struct {
	payload reporter.Payload
}

// useSender replaces the process-wide sender for one test and restores it.
func useSender(t *testing.T) *[]captured {
	t.Helper()
	sent := &[]captured{}
	previous := sender
	sender = func(payload reporter.Payload) bool {
		*sent = append(*sent, captured{payload: payload})
		return true
	}
	t.Cleanup(func() { sender = previous })
	return sent
}

func runHook(t *testing.T, args []string, stdin string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := run(args, strings.NewReader(stdin), &out)
	return out.String(), code
}

func TestClaudeToolUseReportsAndDefers(t *testing.T) {
	sent := useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	out, code := runHook(t, []string{"--mode=claude"},
		`{"hook_event_name":"PreToolUse","session_id":"abc","cwd":"C:\\work","tool_name":"Bash","tool_input":{"command":"ls"}}`)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Fatalf("claude-hook defers with an empty object, got %q", out)
	}
	if len(*sent) != 1 {
		t.Fatalf("expected one report, got %d", len(*sent))
	}
	payload := (*sent)[0].payload
	if payload.Source != "claude-hook" || payload.State != "tool_use" || payload.Reason != "tool" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Project != `C:\work` || payload.Tool != "Bash" || payload.Terminal != "claude" {
		t.Fatalf("payload lost fields: %+v", payload)
	}
	if payload.PID == 0 {
		t.Fatal("pid should be synthesized from the session id when the host omits it")
	}
}

func TestClaudeStopEventIsIdle(t *testing.T) {
	sent := useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	out, _ := runHook(t, []string{"--mode=claude"}, `{"hook_event_name":"Stop","session_id":"s2"}`)
	if strings.TrimSpace(out) != "{}" {
		t.Fatalf("stop must defer with {}, got %q", out)
	}
	if (*sent)[0].payload.State != "idle" {
		t.Fatalf("stop should map to idle, got %q", (*sent)[0].payload.State)
	}
}

func TestClaudeSessionEndClosesTheSession(t *testing.T) {
	sent := useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	runHook(t, []string{"--mode=claude"}, `{"hook_event_name":"SessionEnd","session_id":"s3"}`)
	if (*sent)[0].payload.State != "closed" {
		t.Fatalf("SessionEnd should close the session, got %q", (*sent)[0].payload.State)
	}
}

func TestAntigravityGateWithoutAutoApproveDefersToTheHost(t *testing.T) {
	sent := useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	out, _ := runHook(t, []string{"--mode=antigravity"},
		`{"conversation_id":"c1","tool_confirmation_pending":true,"toolCall":{"name":"run_command"}}`)
	if strings.TrimSpace(out) != `{"decision":"allow"}` {
		t.Fatalf("antigravity defers with an explicit allow, got %q", out)
	}
	if len(*sent) != 1 || (*sent)[0].payload.State != "waiting_permission" {
		t.Fatalf("a gate must still be reported as waiting_permission, got %+v", *sent)
	}
}

func TestDangerousCommandNeverAutoApproves(t *testing.T) {
	sent := useSender(t)
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	writeConfig(t, root, `{"autoApprove":true,"autoApproveTimeoutSec":1}`)

	// Claude: fall through to its own prompt, having alarmed the panel.
	out, _ := runHook(t, []string{"--mode=claude"},
		`{"hook_event_name":"PermissionRequest","session_id":"c1","tool_name":"Bash","tool_input":{"command":"rm -rf /tmp/x"}}`)
	if strings.TrimSpace(out) != "{}" {
		t.Fatalf("a dangerous claude command must defer, got %q", out)
	}
	if (*sent)[0].payload.Reason != "permission_dangerous" {
		t.Fatalf("the panel should be told why it alarmed, got %q", (*sent)[0].payload.Reason)
	}

	// Antigravity: this hook IS the decider, so the only safe answer is deny.
	out, _ = runHook(t, []string{"--mode=antigravity"},
		`{"conversation_id":"a1","tool_confirmation_pending":true,"toolCall":{"name":"run_command","input":{"command":"git push --force origin main"}}}`)
	if strings.TrimSpace(out) != `{"decision":"deny"}` {
		t.Fatalf("a dangerous antigravity command must be denied, got %q", out)
	}
}

func TestAutoApproveApprovesWhenNothingVetoes(t *testing.T) {
	sent := useSender(t)
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	writeConfig(t, root, `{"autoApprove":true,"autoApproveTimeoutSec":1}`)
	if err := os.MkdirAll(filepath.Join(root, "AI-Terminal-Pet", "approvals"), 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	started := time.Now()
	out, code := runHook(t, []string{"--mode=claude"},
		`{"hook_event_name":"PermissionRequest","session_id":"auto-1","tool_name":"Read"}`)
	elapsed := time.Since(started)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if strings.TrimSpace(out) != `{"decision":"allow"}` {
		t.Fatalf("silence must approve, got %q", out)
	}
	if elapsed < 900*time.Millisecond {
		t.Fatalf("the veto window was skipped (waited %v, expected about 1s)", elapsed)
	}
	// The report is filed before the wait, so the overlay can count down.
	if len(*sent) != 1 || (*sent)[0].payload.State != "waiting_permission" {
		t.Fatalf("the wait must report the pending session first, got %+v", *sent)
	}
}

func TestAutoApproveHonoursAVetoFromTheOverlay(t *testing.T) {
	sent := useSender(t)
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	writeConfig(t, root, `{"autoApprove":true,"autoApproveTimeoutSec":5}`)
	approvals := filepath.Join(root, "AI-Terminal-Pet", "approvals")
	if err := os.MkdirAll(approvals, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Stand in for the overlay's countdown: wait for the request, veto it.
	// Matching on the pending-*.json name is not cosmetic: WriteApprovalPending
	// publishes atomically, so a ".pet-*.tmp" appears in this directory first
	// and vetoing that name would leave the real request unvetoed.
	done := make(chan struct{})
	go func() {
		defer close(done)
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			entries, err := os.ReadDir(approvals)
			if err == nil {
				for _, entry := range entries {
					name := entry.Name()
					if !strings.HasPrefix(name, "pending-") || !strings.HasSuffix(name, ".json") {
						continue
					}
					_ = os.WriteFile(filepath.Join(approvals, name+".deny"), []byte(`{"decision":"deny"}`), 0600)
					return
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	out, _ := runHook(t, []string{"--mode=claude"},
		`{"hook_event_name":"PermissionRequest","session_id":"veto-1","tool_name":"Read"}`)
	<-done
	if strings.TrimSpace(out) != `{"decision":"deny"}` {
		t.Fatalf("a veto must turn into deny, got %q", out)
	}
	if len(*sent) == 0 {
		t.Fatal("the pending session should have been reported")
	}
}

func TestUnknownModeStillExitsZeroAndAnswers(t *testing.T) {
	useSender(t)
	out, code := runHook(t, []string{"--mode=nonsense"}, "")
	if code != 0 {
		t.Fatalf("a reporter must never exit non-zero, got %d", code)
	}
	if strings.TrimSpace(out) != `{"decision":"allow"}` {
		t.Fatalf("an unusable invocation should still answer allow, got %q", out)
	}
}

func TestEmptyAndMalformedStdinStillReport(t *testing.T) {
	sent := useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	for _, stdin := range []string{"", "   ", "not json", `[1,2,3]`, `{"hook_event_name":`} {
		out, code := runHook(t, []string{"--mode=claude"}, stdin)
		if code != 0 {
			t.Fatalf("stdin %q produced exit code %d", stdin, code)
		}
		if strings.TrimSpace(out) != "{}" {
			t.Fatalf("stdin %q produced %q", stdin, out)
		}
	}
	if len(*sent) != 5 {
		t.Fatalf("every invocation should file one idle report, got %d", len(*sent))
	}
	for _, item := range *sent {
		if item.payload.State != "idle" {
			t.Fatalf("an unparsable event should report idle, got %q", item.payload.State)
		}
	}
}

func TestOversizedStdinIsRejected(t *testing.T) {
	sent := useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	huge := strings.Repeat("a", reporter.MaxStdinBytes+1024)
	out, code := runHook(t, []string{"--mode=claude"}, huge)
	if code != 0 || strings.TrimSpace(out) != "{}" {
		t.Fatalf("oversized stdin should degrade to idle: code=%d out=%q", code, out)
	}
	if len(*sent) != 1 || (*sent)[0].payload.State != "idle" {
		t.Fatalf("oversized stdin should report idle, got %+v", *sent)
	}
}

func TestHelpPrintsUsageAndExitsZero(t *testing.T) {
	useSender(t)
	out, code := runHook(t, []string{"--help"}, "")
	if code != 0 {
		t.Fatalf("--help must exit 0, got %d", code)
	}
	if !strings.Contains(out, "pet-hook") || !strings.Contains(out, "--mode=claude") {
		t.Fatalf("usage text is missing: %q", out)
	}
}

func TestModeFlagForms(t *testing.T) {
	cases := []struct {
		args []string
		want string
		ok   bool
	}{
		{[]string{"--mode=claude"}, "claude", true},
		{[]string{"--mode=ANTIGRAVITY"}, "antigravity", true},
		{[]string{"--mode= antigravity "}, "antigravity", true},
		{[]string{"--claude"}, "claude", true},
		{[]string{"--antigravity"}, "antigravity", true},
		{nil, "claude", true},
		{[]string{"--mode=other"}, "unknown", false},
		{[]string{"--bogus"}, "claude", false},
	}
	for _, testCase := range cases {
		got, ok := parseMode(testCase.args)
		if ok != testCase.ok {
			t.Errorf("parseMode(%v) ok = %v, want %v", testCase.args, ok, testCase.ok)
		}
		if testCase.ok && got != testCase.want {
			t.Errorf("parseMode(%v) = %q, want %q", testCase.args, got, testCase.want)
		}
	}
}

func TestPayloadsAreValidForTheServer(t *testing.T) {
	useSender(t)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	runHook(t, []string{"--mode=claude"},
		`{"hook_event_name":"PreToolUse","session_id":"valid","started_at":"2026-11-11T09:00:00Z","model":{"provider":"anthropic","model":"opus"}}`)
	runHook(t, []string{"--mode=antigravity"},
		`{"conversation_id":"valid-2","workspace":{"cwd":"C:\\w"},"toolCall":{"name":"read"}}`)

	// run() built them; rebuild both through the same helpers and ask the
	// server's own parser. This is the check that keeps a drift between the
	// binary and internal/session from shipping silently.
	for _, stdin := range []struct {
		mode  string
		input string
	}{
		{"claude", `{"hook_event_name":"PreToolUse","session_id":"valid","started_at":"2026-11-11T09:00:00Z","model":{"provider":"anthropic","model":"opus"}}`},
		{"antigravity", `{"conversation_id":"valid-2","workspace":{"cwd":"C:\\w"},"toolCall":{"name":"read"}}`},
	} {
		payload := buildPayload(t, stdin.mode, stdin.input)
		if err := payload.Validate(time.Now()); err != nil {
			t.Fatalf("%s payload would be rejected by the server: %v", stdin.mode, err)
		}
		line, err := payload.MarshalJSON()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(line, &decoded); err != nil {
			t.Fatalf("%s payload is not JSON: %v", stdin.mode, err)
		}
	}
}

// buildPayload drives the same mapping helpers run() uses, so a test can
// inspect the payload that would have been sent.
func buildPayload(t *testing.T, mode, stdin string) reporter.Payload {
	t.Helper()
	input := reporter.DecodeStdin([]byte(stdin))
	now := time.Date(2026, 11, 11, 9, 30, 0, 0, time.UTC)
	if mode == "claude" {
		state, reason := reporter.ClaudeStatus(input)
		return claudePayload(input, state, reason, now)
	}
	state, reason := reporter.AntigravityStatus(input)
	return antigravityPayload(input, state, reason, now)
}

func writeConfig(t *testing.T, root, body string) {
	t.Helper()
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "config.json"), []byte(body), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

var _ io.Reader = (*strings.Reader)(nil)
