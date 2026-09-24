// Command pet-statusline is a native (no Node.js) replacement for
// reporters/antigravity-statusline.js.
//
// Background: on this machine every new process pays ~1.5s creation tax
// (active AV hooks), bare `node` cold start costs ~4.3s and any `cmd /c`
// hop costs ~7.7s. The Antigravity statusline runner executes the configured
// command via shell with a tight budget, so the Node script can never
// finish in time (exit 1, empty stderr = runner-side timeout kill).
// This shim starts in ~1.5s and finishes in milliseconds.
//
// Contract (mirrors reporters/protocol.js, fail-fast, always exit 0):
//   - read stdin (<=1MiB, 2s), parse Antigravity status JSON loosely
//   - map agent_state/tool_confirmation_pending to state/reason
//   - write atomic runtime snapshot %LOCALAPPDATA%\AI-Terminal-Pet\runtime\
//   - dial per-user pipe, send one JSON line, expect exact "OK"
//   - ENOENT + state != closed -> launch AI-Terminal-Pet.exe once, retry once
//   - print one compact line to stdout (the visible statusline), never stderr
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/session"
	winio "github.com/Microsoft/go-winio"
)

const (
	maxStdinBytes   = 1 << 20
	stdinTimeout    = 2 * time.Second
	pipeDialTimeout = 150 * time.Millisecond
	pipeIOTimeout   = 150 * time.Millisecond
)

var (
	invalidFileChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
	trimFileEdge     = regexp.MustCompile(`^[. -]+|[. -]+$`)
	tokenChars       = regexp.MustCompile(`[^A-Za-z0-9_.:@/-]+`)
	trimTokenEdge    = regexp.MustCompile(`^[. -]+|[. -]+$`)

	redactRules = []struct {
		re   *regexp.Regexp
		repl string
	}{
		{regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`), "Bearer [REDACTED]"},
		{regexp.MustCompile(`\b(?:sk|pk|rk|ak)-[A-Za-z0-9_-]{8,}\b`), "[REDACTED]"},
		{regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{8,}\b`), "[REDACTED]"},
		{regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), "[REDACTED]"},
		{regexp.MustCompile(`(?s)-----BEGIN [^-]+-----.+?-----END [^-]+-----`), "[REDACTED PRIVATE KEY]"},
		{regexp.MustCompile(`(?i)\b([a-z][a-z0-9+.-]*:\/\/)([^\/\s:@]+):([^\/\s@]+)@`), "$1[REDACTED]@"},
		{regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`), "[REDACTED]"},
		{regexp.MustCompile(`(?i)\b((?:api[_-]?key|access[_-]?token|auth(?:orization)?|password|secret)\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)`), "$1[REDACTED]"},
		{regexp.MustCompile(`(?i)(--?(?:api-key|token|password|secret)\s+)([^\s,;]+)`), "$1[REDACTED]"},
	}
)

// cleanText mirrors protocol.js text(): redact, strip controls, collapse
// whitespace, enforce rune limit. Empty result returns "".
func cleanText(value string, limit int) string {
	for _, rule := range redactRules {
		value = rule.re.ReplaceAllString(value, rule.repl)
	}
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
	value = strings.TrimSpace(strings.Join(strings.Fields(value), " "))
	if value == "" {
		return ""
	}
	if utf8.RuneCountInString(value) > limit {
		runes := []rune(value)
		value = strings.TrimSpace(string(runes[:limit]))
	}
	return value
}

func cleanToken(value string, limit int) string {
	clean := cleanText(value, limit)
	if clean == "" {
		return ""
	}
	safe := trimTokenEdge.ReplaceAllString(tokenChars.ReplaceAllString(clean, "-"), "")
	if len(safe) > limit {
		safe = safe[:limit]
	}
	return safe
}

func boundedID(value string, limit int) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "unknown"
	}
	clean := cleanText(raw, limit)
	if clean == "" {
		return "unknown"
	}
	if len(raw) <= limit {
		return clean
	}
	sum := sha256.Sum256([]byte(raw))
	stem := cleanText(raw, limit-9)
	if stem == "" {
		stem = "unknown"
	}
	return stem + "-" + hex.EncodeToString(sum[:4])
}

// stableProcessID ports protocol.js stableProcessId (FNV-1a over UTF-16
// code units) so hook/statusline PIDs stay identical across languages.
func stableProcessID(value string) int {
	const (
		offset = 0x811c9dc5
		prime  = 0x01000193
	)
	hash := uint32(offset)
	for _, unit := range utf16.Encode([]rune(value)) {
		hash ^= uint32(unit)
		hash *= prime
	}
	return int(hash%0x7fffffff) + 1
}

func stringField(input map[string]any, keys ...string) string {
	for _, key := range keys {
		switch value := input[key].(type) {
		case string:
			if strings.TrimSpace(value) != "" {
				return value
			}
		case map[string]any:
			if nested := stringField(value, "type", "state", "status", "name", "cwd", "id", "model"); nested != "" {
				return nested
			}
		}
	}
	return ""
}

func objectField(input map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if value, ok := input[key].(map[string]any); ok {
			return value
		}
	}
	return nil
}

// statusFor mirrors antigravity-statusline.js statusFor().
func statusFor(input map[string]any) (session.State, string) {
	if pending, _ := input["tool_confirmation_pending"].(bool); pending {
		return session.StateWaitingPermission, "tool_confirmation_pending"
	}
	raw := input["agent_state"]
	var text string
	switch value := raw.(type) {
	case string:
		text = value
	case map[string]any:
		text = stringField(value, "type", "state", "status", "name")
	case nil:
		text = ""
	default:
		text = fmt.Sprint(value)
	}
	text = strings.ToLower(text)
	match := func(pattern string) bool {
		matched, _ := regexp.MatchString(pattern, text)
		return matched
	}
	switch {
	case match(`(closed|terminated|finished|complete|done|idle)`):
		return session.StateIdle, "agent_state"
	case match(`(error|failed|failure)`):
		return session.StateError, "agent_state"
	case match(`(permission|confirm|waiting)`):
		return session.StateWaitingPermission, "agent_state"
	case match(`(tool)`):
		return session.StateToolUse, "agent_state"
	case match(`(working|executing|busy|running|processing)`):
		return session.StateWorking, "agent_state"
	case match(`(thinking|reasoning|generat|stream)`):
		return session.StateThinking, "agent_state"
	default:
		return session.StateIdle, "agent_state"
	}
}

func modelName(input map[string]any) string {
	if value, ok := input["model"].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	if nested := objectField(input, "model"); nested != nil {
		provider := stringField(nested, "provider", "providerID", "providerId")
		model := stringField(nested, "model", "modelID", "modelId", "id", "name")
		parts := []string{}
		if strings.TrimSpace(provider) != "" {
			parts = append(parts, strings.TrimSpace(provider))
		}
		if strings.TrimSpace(model) != "" {
			parts = append(parts, strings.TrimSpace(model))
		}
		if len(parts) > 0 {
			return strings.Join(parts, "/")
		}
	}
	return ""
}

func workspacePath(input map[string]any) string {
	if value, ok := input["workspace"]; ok {
		switch workspace := value.(type) {
		case string:
			if strings.TrimSpace(workspace) != "" {
				return workspace
			}
		case map[string]any:
			if nested := stringField(workspace, "cwd", "current_dir", "project_dir", "path"); nested != "" {
				return nested
			}
		}
	}
	return stringField(input, "cwd", "current_dir", "project_dir", "path")
}

func toolName(input map[string]any) string {
	if value, ok := input["tool_name"].(string); ok && value != "" {
		return value
	}
	if call := objectField(input, "toolCall", "tool_call"); call != nil {
		return stringField(call, "name")
	}
	return ""
}

func timestampField(input map[string]any, fallback time.Time, keys ...string) time.Time {
	for _, key := range keys {
		var parsed time.Time
		switch value := input[key].(type) {
		case string:
			if value == "" {
				continue
			}
			parsedTime, err := time.Parse(time.RFC3339, value)
			if err != nil {
				continue
			}
			parsed = parsedTime
		case float64:
			// Epoch seconds or millis.
			if value < 1e11 {
				parsed = time.Unix(int64(value), 0)
			} else {
				parsed = time.Unix(int64(value)/1000, 0)
			}
		default:
			continue
		}
		if !parsed.IsZero() && !parsed.After(fallback.Add(5*time.Minute)) {
			return parsed.UTC()
		}
	}
	return fallback
}

func filenamePart(value, fallback string) string {
	input := value
	if strings.TrimSpace(input) == "" {
		input = fallback
	}
	safe := invalidFileChars.ReplaceAllString(input, "-")
	safe = trimFileEdge.ReplaceAllString(safe, "")
	if len(safe) > 48 {
		safe = safe[:48]
	}
	if safe != "" && safe == input && len(input) <= 48 {
		return safe
	}
	stem := safe
	if len(stem) > 39 {
		stem = stem[:39]
	}
	if stem == "" {
		stem = fallback
	}
	sum := sha256.Sum256([]byte(input))
	return stem + "-" + hex.EncodeToString(sum[:4])
}

func readStdin() []byte {
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes+1))
		done <- result{data, err}
	}()
	select {
	case res := <-done:
		if res.err != nil || len(res.data) > maxStdinBytes {
			return nil
		}
		return res.data
	case <-time.After(stdinTimeout):
		return nil
	}
}

func executablePath() string {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return ""
	}
	candidate := filepath.Join(root, "AI-Terminal-Pet", "AI-Terminal-Pet.exe")
	info, err := os.Stat(candidate)
	if err != nil || info.IsDir() {
		return ""
	}
	return candidate
}

func launchApplication() {
	executable := executablePath()
	if executable == "" {
		return
	}
	cmd := exec.Command(executable)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		// DETACHED_PROCESS (0x00000008); literal avoids a new dependency
		// for a single constant.
		CreationFlags: 0x00000008,
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	_ = cmd.Start()
}

func sendOnce(pipePath string, line []byte) bool {
	timeout := pipeDialTimeout
	conn, err := winio.DialPipe(pipePath, &timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(pipeIOTimeout))
	if _, err := conn.Write(append(append([]byte{}, line...), '\n')); err != nil {
		return false
	}
	reply := make([]byte, 32)
	count, err := conn.Read(reply)
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}
	return string(bytes.TrimSpace(reply[:count])) == "OK"
}

func run() int {
	now := time.Now().UTC()
	input := map[string]any{}
	if raw := readStdin(); len(bytes.TrimSpace(raw)) > 0 {
		var parsed any
		if err := json.Unmarshal(bytes.TrimSpace(raw), &parsed); err == nil {
			if obj, ok := parsed.(map[string]any); ok {
				input = obj
			}
		}
	}
	state, reason := statusFor(input)
	sessionID := stringField(input, "conversation_id", "session_id", "sessionId")
	if sessionID == "" {
		sessionID = "global"
	}
	pid := 0
	if rawPid, ok := input["pid"].(float64); ok && rawPid > 0 && rawPid <= 0xffffffff {
		pid = int(rawPid)
	} else {
		pid = stableProcessID("antigravity-statusline:" + sessionID)
	}
	parentPid := os.Getppid()
	if rawParent, ok := input["parentPid"].(float64); ok && rawParent >= 0 && rawParent <= 0xffffffff {
		parentPid = int(rawParent)
	}
	startedAt := timestampField(input, now, "started_at", "startedAt", "session_start_time")

	item := session.Session{
		SchemaVersion: session.SchemaVersion,
		Source:        "antigravity-statusline",
		SessionID:     boundedID(sessionID, 200),
		PID:           pid,
		ParentPID:     parentPid,
		Project:       cleanText(workspacePath(input), 1024),
		Model:         cleanText(modelName(input), 200),
		State:         state,
		Reason:        cleanToken(reason, 120),
		StartedAt:     startedAt,
		UpdatedAt:     now,
		Tool:          cleanText(toolName(input), 120),
		Terminal:      "antigravity",
	}
	if parent := stringField(input, "parent_session_id", "parentSessionId"); parent != "" {
		item.ParentSessionID = boundedID(parent, 200)
	}
	line, err := json.Marshal(item)
	if err != nil {
		fmt.Println("pet idle")
		return 0
	}
	// Self-check: only emit payloads the server accepts.
	if _, err := session.ParsePayload(append(line, '\n'), now); err != nil {
		fmt.Println("pet idle")
		return 0
	}

	root := os.Getenv("LOCALAPPDATA")
	if root != "" {
		directory := filepath.Join(root, "AI-Terminal-Pet", "runtime")
		name := filenamePart(item.Source, "unknown") + "-" +
			fmt.Sprint(item.PID) + "-" +
			filenamePart(item.SessionID, "unknown") + ".json"
		target := filepath.Join(directory, name)
		tmp, err := os.CreateTemp(directory, ".pet-*.tmp")
		if err == nil {
			tmpName := tmp.Name()
			_ = tmp.Chmod(0600)
			if _, err := tmp.Write(append(line, '\n')); err == nil {
				_ = tmp.Close()
				_ = os.Rename(tmpName, target)
			} else {
				_ = tmp.Close()
				_ = os.Remove(tmpName)
			}
		}
	}

	pipePath := pipeserver.PipePathForUser()
	launched := false
	if !sendOnce(pipePath, line) && state != session.StateClosed && !launched {
		launched = true
		launchApplication()
		sendOnce(pipePath, line)
	}

	fmt.Println("pet " + string(state))
	return 0
}

func main() {
	os.Exit(run())
}
