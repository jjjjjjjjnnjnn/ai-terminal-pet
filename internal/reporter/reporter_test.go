package reporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// Node produced these values with stableProcessId from reporters/protocol.js:
//
//	node -e "const {stableProcessId}=require('./reporters/protocol.js');console.log(stableProcessId('claude-hook:abc'))"
//
// Pinning them is the whole point of the FNV-1a-over-UTF-16 detail: an ASCII
// seed hashes identically either way, a CJK one does not.
func TestStableProcessIDMatchesReporter(t *testing.T) {
	cases := []struct {
		seed string
		want int
	}{
		{"claude-hook:global", 482609443},
		{"claude-hook:abc", 967448088},
		{"antigravity-hook:\u4f1a\u8bdd-1", 143134961},
	}
	for _, testCase := range cases {
		if got := StableProcessID(testCase.seed); got != testCase.want {
			t.Errorf("StableProcessID(%q) = %d, want %d", testCase.seed, got, testCase.want)
		}
	}
}

func TestCleanTextAndToken(t *testing.T) {
	if got := CleanText("  hello \t world  ", 200); got != "hello world" {
		t.Fatalf("CleanText collapsed incorrectly: %q", got)
	}
	if got := CleanText("token sk-abcdefgh12345678 here", 200); got != "token [REDACTED] here" {
		t.Fatalf("secret was not redacted: %q", got)
	}
	if got := CleanText("abcdefghij", 4); got != "abcd" {
		t.Fatalf("CleanText did not cut to the rune limit: %q", got)
	}
	if got := CleanText("\x00\x01  ", 10); got != "" {
		t.Fatalf("control-only input should collapse to empty, got %q", got)
	}
	// token() keeps [A-Za-z0-9_.:@/-] and maps every other run to one hyphen,
	// which is why the drive colon survives as a leading "C:-" stem.
	if got := CleanToken(`C:\work\my proj`, 200); got != "C:-work-my-proj" {
		t.Fatalf("CleanToken kept characters the protocol forbids: %q", got)
	}
}

func TestBoundedIDKeepsShortValuesAndSuffixesLongOnes(t *testing.T) {
	if got := BoundedID("short-id", 200); got != "short-id" {
		t.Fatalf("short id should pass through, got %q", got)
	}
	long := ""
	for index := 0; index < 40; index++ {
		long += "a"
	}
	got := BoundedID(long, 10)
	if len(got) > 10 {
		t.Fatalf("bounded id is longer than the limit: %q", got)
	}
	// Two ids that share a prefix must not collide.
	if BoundedID(long, 10) == BoundedID(long+"b", 10) {
		t.Fatal("colliding bounded ids: the digest suffix did not separate them")
	}
	if blank := BoundedID("   ", 200); blank != "unknown" {
		t.Fatalf(`blank id should become "unknown", got %q`, blank)
	}
}

func TestNormalizeTimestampsClampsFutureAndInverts(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	future := now.Add(2 * time.Hour)
	started, updated := NormalizeTimestamps(future, future, now)
	if updated.After(now.Add(4 * time.Minute)) {
		t.Fatalf("future timestamps were not clamped: %v", updated)
	}
	if started.After(updated) {
		t.Fatalf("startedAt must never follow updatedAt: %v > %v", started, updated)
	}
	started, updated = NormalizeTimestamps(now, now.Add(-time.Hour), now)
	if updated.Before(started) {
		t.Fatalf("inverted pair was not repaired: %v < %v", updated, started)
	}
}

func TestFilenamePartUsesDigestForUnsafeNames(t *testing.T) {
	if got := FilenamePart("simple-name", "unknown"); got != "simple-name" {
		t.Fatalf("safe name should pass through, got %q", got)
	}
	unsafe := `C:\work\proj`
	got := FilenamePart(unsafe, "unknown")
	if got == unsafe || len(got) > 48 {
		t.Fatalf("unsafe name was not sanitized: %q", got)
	}
	if FilenamePart(unsafe, "unknown") == FilenamePart(`C:\work\other`, "unknown") {
		t.Fatal("two different unsafe names collided")
	}
}

func TestPayloadFinishProducesServerAcceptedJSON(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "s-1"
	payload.Project = `C:\work`
	payload.State = session.StateToolUse
	payload.Tool = "Bash"
	payload.Terminal = "claude"
	payload.PID = 1234
	payload.ParentPID = 1200
	payload = payload.Finish(now)

	if err := payload.Validate(now); err != nil {
		t.Fatalf("the server would reject this payload: %v", err)
	}
	line, err := payload.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(line, &decoded); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	for _, key := range []string{
		"schemaVersion", "source", "sessionId", "parentSessionId", "pid", "parentPid",
		"project", "title", "model", "state", "reason", "startedAt", "updatedAt", "tool", "terminal",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("payload is missing %q", key)
		}
	}
	if decoded["state"] != "tool_use" {
		t.Errorf("state should serialize as a bare string, got %v", decoded["state"])
	}
}

func TestPayloadFinishRepairsOutOfRangeFields(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	payload := NewPayload(now)
	payload.SessionID = "   "
	payload.PID = 0
	payload.ParentPID = -5
	payload.State = session.State("nonsense")
	payload = payload.Finish(now)
	if payload.SessionID != "unknown" {
		t.Fatalf("blank session id was not repaired: %q", payload.SessionID)
	}
	if payload.PID <= 0 {
		t.Fatalf("zero pid was not replaced by a synthetic one: %d", payload.PID)
	}
	if payload.ParentPID != 0 {
		t.Fatalf("negative parent pid should clamp to 0, got %d", payload.ParentPID)
	}
	if payload.State != session.StateIdle {
		t.Fatalf("invalid state should fall back to idle, got %q", payload.State)
	}
}

func TestWriteSnapshotRoundTrips(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	directory := filepath.Join(root, "AI-Terminal-Pet", "runtime")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "snap"
	payload.PID = 42
	payload.State = session.StateWorking
	payload = payload.Finish(now)

	if !WriteSnapshot(payload) {
		t.Fatal("WriteSnapshot reported failure")
	}
	path := SnapshotPath(payload)
	if path == "" {
		t.Fatal("SnapshotPath returned nothing")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one snapshot, found %d", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(directory, entries[0].Name()))
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if _, err := session.ParsePayload(data, now); err != nil {
		t.Fatalf("snapshot is not a payload the server accepts: %v", err)
	}
	_ = path
}

// A fresh install has no runtime directory until the overlay has run once, and
// this is exactly the situation a snapshot exists for: the overlay is not
// running, so the only way the report survives is the file. Refusing to create
// the directory would drop that first report silently. sendStatus in
// protocol.js calls mkdirSync({recursive:true}) for the same reason.
func TestWriteSnapshotCreatesAMissingRuntimeDirectory(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "fresh-install"
	payload.PID = 43
	payload.State = session.StateWorking
	payload = payload.Finish(now)

	if _, err := os.Stat(filepath.Join(root, "AI-Terminal-Pet", "runtime")); !os.IsNotExist(err) {
		t.Fatalf("the test expected no runtime directory to exist yet: %v", err)
	}
	if !WriteSnapshot(payload) {
		t.Fatal("WriteSnapshot must create the runtime directory instead of giving up")
	}
	data, err := os.ReadFile(SnapshotPath(payload))
	if err != nil {
		t.Fatalf("snapshot not readable: %v", err)
	}
	if _, err := session.ParsePayload(data, now); err != nil {
		t.Fatalf("snapshot is not a payload the server accepts: %v", err)
	}
}

// A machine without the overlay hits this path on every single host event, so
// it has to be quiet, bounded and non-fatal. The pipe name is overridden
// because the real user pipe may well have a live listener.
func TestSendWithoutListenerFailsQuietly(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	now := time.Now().UTC()
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "no-pipe"
	payload.PID = 7
	payload = payload.Finish(now)
	options := SendOptions{
		Retry:    20 * time.Millisecond,
		Launch:   false,
		PipePath: `\\.\pipe\ai-terminal-pet-test-does-not-exist`,
	}
	if Send(payload, options) {
		t.Fatal("Send reported success without a listener")
	}
}

func TestApprovalPlanModes(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	if mode, timeout := ApprovalPlan("Bash: ls"); mode != ApprovalOff || timeout != 0 {
		t.Fatalf("a missing config.json must mean auto-approve off, got %q/%v", mode, timeout)
	}

	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := []byte(`{"autoApprove":true,"autoApproveTimeoutSec":2}`)
	if err := os.WriteFile(filepath.Join(directory, "config.json"), body, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if mode, timeout := ApprovalPlan("Bash: ls"); mode != ApprovalAuto || timeout != 2*time.Second {
		t.Fatalf("plain command should be auto, got %q/%v", mode, timeout)
	}
	if mode, _ := ApprovalPlan("Bash: rm -rf /tmp/x"); mode != ApprovalManual {
		t.Fatalf("a destructive command must never be auto-approved, got %q", mode)
	}
	if mode, _ := ApprovalPlan("Bash: git push --force origin main"); mode != ApprovalManual {
		t.Fatalf("force push must never be auto-approved, got %q", mode)
	}
	if mode, _ := ApprovalPlan("Bash: Remove-Item C:\\Windows\\System32 -Recurse -Force"); mode != ApprovalManual {
		t.Fatalf("System32 removal must never be auto-approved, got %q", mode)
	}
}

func TestApprovalPendingFileAndVeto(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	directory := filepath.Join(root, "AI-Terminal-Pet", "approvals")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	request := ApprovalRequest{Source: "claude-hook", SessionID: "s1", PID: 11, Tool: "Bash"}
	pending := WriteApprovalPending(request, time.Now())
	if pending == "" {
		t.Fatal("WriteApprovalPending returned nothing")
	}
	data, err := os.ReadFile(pending)
	if err != nil {
		t.Fatalf("pending file was not created: %v", err)
	}
	var body struct {
		Source    string `json:"source"`
		SessionID string `json:"sessionId"`
		PID       int    `json:"pid"`
		Tool      string `json:"tool"`
		CreatedAt string `json:"createdAt"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("pending body is not JSON: %v", err)
	}
	if body.Source != "claude-hook" || body.SessionID != "s1" || body.PID != 11 || body.Tool != "Bash" {
		t.Fatalf("pending body lost the identity the overlay matches on: %+v", body)
	}
	if !strings.HasPrefix(filepath.Base(pending), "pending-") {
		t.Fatalf("pending file does not use the prefix the overlay scans for: %s", pending)
	}
	if _, err := time.Parse(time.RFC3339Nano, body.CreatedAt); err != nil {
		t.Fatalf("createdAt is not an RFC3339 timestamp: %q", body.CreatedAt)
	}
	// The overlay only ever writes the sibling .deny file.
	if PollApprovalDeny(pending, 50*time.Millisecond) {
		t.Fatal("PollApprovalDeny reported a veto that was never written")
	}
	if err := os.WriteFile(pending+".deny", []byte(`{"decision":"deny"}`), 0600); err != nil {
		t.Fatalf("write veto: %v", err)
	}
	if !PollApprovalDeny(pending, 50*time.Millisecond) {
		t.Fatal("PollApprovalDeny missed the veto file")
	}
	RemoveApprovalPending(pending)
	if _, err := os.Stat(pending); !os.IsNotExist(err) {
		t.Fatal("RemoveApprovalPending left the request behind")
	}
}

// A veto that arrives while the hook is waiting has to end the wait, which is
// the whole contract between the overlay and this binary.
func TestWaitForApprovalDecisionObservesLateVeto(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	directory := filepath.Join(root, "AI-Terminal-Pet", "approvals")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	request := ApprovalRequest{Source: "claude-hook", SessionID: "late", PID: 12, Tool: "Bash"}
	answer := make(chan string, 1)
	go func() { answer <- WaitForApprovalDecision(request, 2*time.Second) }()

	deadline := time.Now().Add(time.Second)
	for {
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatalf("read dir: %v", err)
		}
		if len(entries) > 0 {
			veto := filepath.Join(directory, entries[0].Name()+".deny")
			if err := os.WriteFile(veto, []byte(`{"decision":"deny"}`), 0600); err != nil {
				t.Fatalf("write veto: %v", err)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the hook never filed a pending request")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if decision := <-answer; decision != "deny" {
		t.Fatalf("a late veto must produce deny, got %q", decision)
	}
}

func TestWaitForApprovalDecisionFailsClosedWithoutADirectory(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")
	request := ApprovalRequest{Source: "claude-hook", SessionID: "s1", PID: 11}
	if decision := WaitForApprovalDecision(request, 50*time.Millisecond); decision != "deny" {
		t.Fatalf("an unfilable request must answer deny, got %q", decision)
	}
}

func TestHookStatusMappingMatchesReporterScripts(t *testing.T) {
	claude := []struct {
		payload map[string]any
		state   string
		reason  string
	}{
		{map[string]any{"hook_event_name": "SessionEnd"}, "closed", "session_end"},
		{map[string]any{"hook_event_name": "PreToolUse"}, "tool_use", "tool"},
		{map[string]any{"hook_event_name": "PostToolUse"}, "working", "tool_result"},
		{map[string]any{"hook_event_name": "UserPromptSubmit"}, "thinking", "prompt_submitted"},
		{map[string]any{"hook_event_name": "Stop"}, "idle", "idle"},
		{map[string]any{"hook_event_name": "PermissionRequest"}, "waiting_permission", "permission"},
		{map[string]any{"error": "boom"}, "error", "error"},
		{map[string]any{}, "idle", "hook"},
	}
	for _, testCase := range claude {
		state, reason := ClaudeStatus(testCase.payload)
		if state != testCase.state || reason != testCase.reason {
			t.Errorf("ClaudeStatus(%v) = %q/%q, want %q/%q", testCase.payload, state, reason, testCase.state, testCase.reason)
		}
	}

	antigravity := []struct {
		payload map[string]any
		state   string
		reason  string
	}{
		{map[string]any{"termination_reason": "done"}, "closed", "termination_reason"},
		{map[string]any{"fullyIdle": true}, "idle", "fully_idle"},
		{map[string]any{"toolCall": map[string]any{"name": "run"}}, "tool_use", "tool_call"},
		{map[string]any{}, "idle", "hook"},
	}
	for _, testCase := range antigravity {
		state, reason := AntigravityStatus(testCase.payload)
		if state != testCase.state || reason != testCase.reason {
			t.Errorf("AntigravityStatus(%v) = %q/%q, want %q/%q", testCase.payload, state, reason, testCase.state, testCase.reason)
		}
	}
}

func TestAntigravityPermissionGateDetection(t *testing.T) {
	gates := []map[string]any{
		{"tool_confirmation_pending": true},
		{"permission_required": true},
		{"toolCall": map[string]any{"requiresApproval": true}},
		{"event": "waiting_for_approval"},
	}
	for _, payload := range gates {
		if !AntigravityIsPermissionGate(payload) {
			t.Errorf("%v should be a permission gate", payload)
		}
	}
	passive := []map[string]any{
		{},
		{"toolCall": map[string]any{"name": "read_file"}},
		{"fullyIdle": true},
	}
	for _, payload := range passive {
		if AntigravityIsPermissionGate(payload) {
			t.Errorf("%v must not block for a veto window", payload)
		}
	}
}
