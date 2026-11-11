package reporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"ai-terminal-pet/internal/config"
)

// Auto-approve handshake, hook side. The overlay is a read-only panel: it can
// never click "allow" inside a terminal, so the protocol runs the other way
// round. The hook files a request and blocks for the veto window; silence
// means approval, and the overlay's countdown only has to write a deny file to
// stop it.
//
//	%LOCALAPPDATA%\AI-Terminal-Pet\approvals\pending-<source>-<session>-<pid>-<ms>.json
//	                         ...<that same path>.deny   <- written by the overlay
//
// This is the Go twin of the block documented in reporters/protocol.js and
// implemented against internal/ui/approve_windows.go. Both sides key on the
// JSON body, never on the file name, so the two stem sanitizers only have to
// stay similar, not identical.

const (
	// ApprovalPollInterval is how often the veto file is checked. 100 ms keeps
	// a tap responsive without touching the disk in a spin.
	ApprovalPollInterval = 100 * time.Millisecond
	// approvalPendingPrefix marks hook-owned request files.
	approvalPendingPrefix = "pending-"
	// DefaultApproveTimeoutSec mirrors config.DefaultAutoApproveTimeoutSec.
	DefaultApproveTimeoutSec = 5
)

// dangerousPatterns never auto-approve. A match forces the manual path: Claude
// falls through to its own prompt, Antigravity denies. Erring broad is
// deliberate - a missed pattern approves a destructive command.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+[^;|&]*-[a-z]*r`),          // rm -r / rm -rf
	regexp.MustCompile(`(?i)\brd\s+\/s`),                      // rd /s
	regexp.MustCompile(`(?i)\bmv\b[^;|&]*\/dev\/null`),        // mv ... /dev/null
	regexp.MustCompile(`(?i)\bmkfs\b`),                        // make filesystem
	regexp.MustCompile(`(?i)\bdd\b[^;|&]*\bof\s*=\s*\/dev\/`), // dd of=/dev/...
	regexp.MustCompile(`:\(\)\s*\{`),                          // fork bomb
	regexp.MustCompile(`(?i)\bgit\s+push\b[^;|&]*--force`),    // git push --force
	regexp.MustCompile(`(?i)\bformat\s+[a-zA-Z]:`),            // format C:
	regexp.MustCompile(`(?i)\bdel\b[^;|&]*\\windows`),         // del ... \Windows
	regexp.MustCompile(`(?i)\bshutdown\s+\/s`),                // shutdown /s
	regexp.MustCompile(`(?i)remove-item\b[^;|&]*system32`),    // Remove-Item ... System32
}

// IsDangerousCommand matches the pattern list against "tool name + tool
// input", the same text the JavaScript reporters hand to their blacklist.
func IsDangerousCommand(parts ...string) bool {
	text := ""
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		if text != "" {
			text += "\n"
		}
		text += part
	}
	if text == "" {
		return false
	}
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// AutoApproveConfig is the subset of config.json a hook cares about. Any read
// failure means "off": a hook that cannot read the user's intent must not
// approve anything.
type AutoApproveConfig struct {
	Enabled    bool
	TimeoutSec int
}

// LoadAutoApproveConfig reads config.json. Absent, unreadable or malformed all
// return the disabled fallback, which is also config.Default's answer.
func LoadAutoApproveConfig() AutoApproveConfig {
	fallback := AutoApproveConfig{Enabled: false, TimeoutSec: DefaultApproveTimeoutSec}
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return fallback
	}
	data, err := os.ReadFile(filepath.Join(root, "AI-Terminal-Pet", config.FileName))
	if err != nil {
		return fallback
	}
	var parsed struct {
		AutoApprove           bool `json:"autoApprove"`
		AutoApproveTimeoutSec int  `json:"autoApproveTimeoutSec"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fallback
	}
	timeout := parsed.AutoApproveTimeoutSec
	if timeout < 1 || timeout > 60 {
		timeout = DefaultApproveTimeoutSec
	}
	return AutoApproveConfig{Enabled: parsed.AutoApprove, TimeoutSec: timeout}
}

// ApprovalMode is the hook's decision for one permission gate.
type ApprovalMode string

const (
	// ApprovalOff means auto-approve is disabled: keep the host's own prompt.
	ApprovalOff ApprovalMode = "off"
	// ApprovalManual means the command looked destructive: alarm the panel but
	// never decide for the user.
	ApprovalManual ApprovalMode = "manual"
	// ApprovalAuto means the hook may approve if the user does not veto.
	ApprovalAuto ApprovalMode = "auto"
)

// ApprovalPlan decides which of the three paths a gate takes.
func ApprovalPlan(toolText string) (ApprovalMode, time.Duration) {
	settings := LoadAutoApproveConfig()
	if !settings.Enabled {
		return ApprovalOff, 0
	}
	if IsDangerousCommand(toolText) {
		return ApprovalManual, 0
	}
	return ApprovalAuto, time.Duration(settings.TimeoutSec) * time.Second
}

// ApprovalsDirectory is where hooks and the overlay exchange files.
func ApprovalsDirectory() string {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return ""
	}
	return filepath.Join(root, "AI-Terminal-Pet", "approvals")
}

// SanitizeApprovalStem mirrors protocol.js sanitizeStem and
// internal/ui sanitizeApprovalStem: [A-Za-z0-9_-] only, 64 characters.
func SanitizeApprovalStem(value string) string {
	out := make([]byte, 0, 64)
	for index := 0; index < len(value) && len(out) < 64; index++ {
		char := value[index]
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' {
			out = append(out, char)
		} else {
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "session"
	}
	return string(out)
}

// ApprovalRequest identifies one veto window.
type ApprovalRequest struct {
	Source    string
	SessionID string
	PID       int
	Tool      string
}

// WriteApprovalPending files the request atomically and returns its path, or
// "" when it cannot be filed. The overlay matches on the JSON body, so the
// name is only for humans and for the .deny sibling.
func WriteApprovalPending(request ApprovalRequest, now time.Time) string {
	directory := ApprovalsDirectory()
	if directory == "" {
		return ""
	}
	source := request.Source
	if source == "" {
		source = "hook"
	}
	sessionID := request.SessionID
	if sessionID == "" {
		sessionID = "global"
	}
	name := approvalPendingPrefix + SanitizeApprovalStem(source) + "-" +
		SanitizeApprovalStem(sessionID) + "-" +
		strconv.Itoa(request.PID) + "-" +
		strconv.FormatInt(now.UnixMilli(), 10) + ".json"
	target := filepath.Join(directory, name)
	body, err := json.Marshal(struct {
		Source    string  `json:"source"`
		SessionID string  `json:"sessionId"`
		PID       int     `json:"pid"`
		Tool      *string `json:"tool"`
		CreatedAt string  `json:"createdAt"`
	}{
		Source:    source,
		SessionID: sessionID,
		PID:       request.PID,
		Tool:      optionalString(request.Tool),
		CreatedAt: now.UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return ""
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return ""
	}
	if err := config.WriteFileAtomic(target, append(body, '\n')); err != nil {
		return ""
	}
	return target
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	copy := value
	return &copy
}

// PollApprovalDeny waits up to timeout for the overlay's veto file. True means
// the user tapped deny.
func PollApprovalDeny(pendingPath string, timeout time.Duration) bool {
	if pendingPath == "" {
		return false
	}
	if timeout < 0 {
		timeout = 0
	}
	veto := pendingPath + ".deny"
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(veto); err == nil {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		wait := ApprovalPollInterval
		if remaining := time.Until(deadline); remaining < wait {
			wait = remaining
		}
		if wait > 0 {
			time.Sleep(wait)
		}
	}
}

// RemoveApprovalPending deletes the request file. Leaving it behind would make
// the overlay keep showing a countdown for a hook that is already gone.
func RemoveApprovalPending(pendingPath string) {
	if pendingPath == "" {
		return
	}
	_ = os.Remove(pendingPath)
}

// WaitForApprovalDecision files a request and blocks for the veto window.
// "allow" is silence, "deny" is a tap on the panel. Fail-closed: when the
// request cannot be filed at all, the answer is deny, because an approval that
// cannot be vetoed is worse than a prompt.
func WaitForApprovalDecision(request ApprovalRequest, timeout time.Duration) string {
	pending := WriteApprovalPending(request, time.Now())
	if pending == "" {
		return "deny"
	}
	defer RemoveApprovalPending(pending)
	if PollApprovalDeny(pending, timeout) {
		return "deny"
	}
	return "allow"
}
