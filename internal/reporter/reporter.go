// Package reporter holds the host-side half of the ai-terminal-pet wire
// protocol: everything a native reporter binary needs to turn a host CLI
// event into one session payload, and to deliver it the same way
// reporters/protocol.js does.
//
// It exists because two commands need exactly the same rules. cmd/pet-hook
// (Claude Code and Antigravity hook events) uses it today; cmd/pet-statusline
// (Antigravity's statusline) still carries its own copy of these helpers and
// should move here the next time it is touched. reporters/protocol.js is the
// reference implementation, and a divergence between any two of them would file
// two different sessions for what the user sees as one.
//
// Nothing in this package touches Win32 windows; it deals only with stdin,
// JSON, the runtime snapshot directory and the named pipe.
package reporter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	// MaxStdinBytes mirrors protocol.js MAX_STDIN_BYTES. A payload over this
	// size is rejected outright rather than truncated: a truncated JSON line
	// is worse than no report at all.
	MaxStdinBytes = 1 << 20
	// StdinTimeout mirrors protocol.js STDIN_TIMEOUT_MS.
	StdinTimeout = 2 * time.Second
)

var (
	invalidFileChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
	trimFileEdge     = regexp.MustCompile(`^[. -]+|[. -]+$`)
	tokenChars       = regexp.MustCompile(`[^A-Za-z0-9_.:@/-]+`)
	trimTokenEdge    = regexp.MustCompile(`^[. -]+|[. -]+$`)

	// redactRules is the Go twin of redact() in reporters/protocol.js, in the
	// same order. Order matters: a Bearer token inside a URL userinfo would
	// otherwise be rewritten twice.
	redactRules = []struct {
		re   *regexp.Regexp
		repl string
	}{
		{regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`), "Bearer [REDACTED]"},
		{regexp.MustCompile(`\b(?:sk|pk|rk|ak)-[A-Za-z0-9_-]{8,}\b`), "[REDACTED]"},
		{regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{8,}\b`), "[REDACTED]"},
		{regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), "[REDACTED]"},
		{regexp.MustCompile(`(?s)-----BEGIN [^-]+-----.+?-----END [^-]+-----`), "[REDACTED PRIVATE KEY]"},
		{regexp.MustCompile(`(?i)\b([a-z][a-z0-9+.-]*:\/\/)([^\/\s:@]+):([^\/\s@]+)@`), "${1}[REDACTED]@"},
		{regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`), "[REDACTED]"},
		{regexp.MustCompile(`(?i)\b((?:api[_-]?key|access[_-]?token|auth(?:orization)?|password|secret)\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)`), "${1}[REDACTED]"},
		{regexp.MustCompile(`(?i)(--?(?:api-key|token|password|secret)\s+)([^\s,;]+)`), "${1}[REDACTED]"},
	}
)

// CleanText mirrors protocol.js text(): redact, strip control characters,
// collapse whitespace, then cut to limit runes. An input that is empty after
// all that returns "" and means "the field is absent".
func CleanText(value string, limit int) string {
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

// CleanToken mirrors protocol.js token(): text() plus a restricted alphabet,
// because the value is used as a machine-readable reason, not prose.
func CleanToken(value string, limit int) string {
	clean := CleanText(value, limit)
	if clean == "" {
		return ""
	}
	safe := trimTokenEdge.ReplaceAllString(tokenChars.ReplaceAllString(clean, "-"), "")
	if len(safe) > limit {
		safe = safe[:limit]
	}
	return safe
}

// BoundedID mirrors protocol.js boundedId(): a value that fits is kept
// verbatim (after cleaning); a longer one is cut and suffixed with the first
// 4 bytes of its sha256, so two long ids that share a prefix stay distinct.
func BoundedID(value string, limit int) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "unknown"
	}
	clean := CleanText(raw, limit)
	if clean == "" {
		return "unknown"
	}
	if len(raw) <= limit {
		return clean
	}
	sum := sha256.Sum256([]byte(raw))
	stem := CleanText(raw, limit-9)
	if stem == "" {
		stem = "unknown"
	}
	return stem + "-" + hex.EncodeToString(sum[:4])
}

// StableProcessID ports protocol.js stableProcessId: FNV-1a over UTF-16 code
// units, because JavaScript's charCodeAt indexes UTF-16 and a byte-wise hash
// would disagree on every non-ASCII session id. The mapping is pinned by
// TestStableProcessIDMatchesReporter, which carries values produced by Node.
func StableProcessID(value string) int {
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

// PositiveInt mirrors protocol.js positiveInteger(v, fallback): only a
// positive integer that fits in 32 bits is accepted. JSON numbers arrive as
// float64, strings are coerced the way Number() would for plain integers.
func PositiveInt(value any, fallback int) int {
	switch typed := value.(type) {
	case float64:
		if typed > 0 && typed <= 0xffffffff && typed == float64(int(typed)) {
			return int(typed)
		}
	case int:
		if typed > 0 && typed <= 0xffffffff {
			return typed
		}
	case string:
		if number, err := parsePositiveInt(typed); err == nil {
			return number
		}
	}
	return fallback
}

func parsePositiveInt(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("empty")
	}
	number := 0
	for _, digit := range trimmed {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("not an integer")
		}
		number = number*10 + int(digit-'0')
		if number > 0xffffffff {
			return 0, fmt.Errorf("out of range")
		}
	}
	if number <= 0 {
		return 0, fmt.Errorf("not positive")
	}
	return number, nil
}

// Truthy mirrors a JavaScript truthiness test for the subset of JSON values a
// hook can carry. An explicit false, 0, "" and null are falsy; everything else
// is truthy, which is what `if (input.error)` means in the reporter scripts.
func Truthy(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case bool:
		return typed
	case float64:
		return typed != 0
	case string:
		return typed != ""
	case map[string]any:
		return true
	case []any:
		return true
	default:
		return true
	}
}

// Timestamp mirrors protocol.js timestamp(): ISO strings, epoch seconds and
// epoch milliseconds are all accepted, and an unparseable value falls back
// instead of failing the whole report.
func Timestamp(value any, fallback time.Time) time.Time {
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return fallback
		}
		parsed, err := time.Parse(time.RFC3339Nano, typed)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, typed)
		}
		if err != nil {
			return fallback
		}
		return parsed.UTC()
	case float64:
		if typed <= 0 {
			return fallback
		}
		if typed < 1e11 {
			return time.Unix(int64(typed), 0).UTC()
		}
		return time.Unix(int64(typed)/1000, 0).UTC()
	default:
		return fallback
	}
}

// NormalizeTimestamps applies the two clamps protocol.js normalizePayload
// documents, in the same order:
//
//   - neither timestamp may sit more than four minutes in the future (the
//     server rejects five, and a clock-skewed reporter would otherwise lose
//     its whole session silently);
//   - updatedAt never precedes startedAt.
//
// It is idempotent, because callers normalize twice (snapshot, then send).
func NormalizeTimestamps(started, updated, now time.Time) (time.Time, time.Time) {
	if started.IsZero() {
		started = now
	}
	if updated.IsZero() {
		updated = now
	}
	ceiling := now.Add(4 * time.Minute)
	clamp := func(value time.Time) time.Time {
		if value.After(ceiling) {
			return ceiling
		}
		return value
	}
	started = clamp(started)
	updated = clamp(updated)
	if updated.Before(started) {
		started = updated
	}
	return started.UTC(), updated.UTC()
}

// FilenamePart mirrors protocol.js filenamePart(): a readable stem when the
// value is already safe, otherwise a truncated stem plus a sha256 prefix so
// two different unsanitary ids cannot collide on one snapshot file.
func FilenamePart(value, fallback string) string {
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

// StringField returns the first key whose value is a non-blank string, or the
// first string found one level down for object values. Hook payloads change
// shape between host versions, so most fields are read through this.
func StringField(input map[string]any, keys ...string) string {
	for _, key := range keys {
		switch value := input[key].(type) {
		case string:
			if strings.TrimSpace(value) != "" {
				return value
			}
		case map[string]any:
			if nested := StringField(value, "type", "state", "status", "name", "cwd", "id", "model"); nested != "" {
				return nested
			}
		}
	}
	return ""
}

// ScalarString renders a JSON scalar the way String(value) would in the
// reporter scripts, and returns "" for objects, arrays and null.
func ScalarString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%v", typed)
	case bool:
		return fmt.Sprintf("%t", typed)
	default:
		return ""
	}
}

// ObjectField returns the first key that holds a JSON object.
func ObjectField(input map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if value, ok := input[key].(map[string]any); ok {
			return value
		}
	}
	return nil
}

// CompactJSON renders a decoded JSON value the way JSON.stringify would for
// the purpose it is used for here: being scanned by the dangerous-command
// patterns. Key order is sorted, which is fine because every rule is a
// substring/regex match rather than a structural one.
func CompactJSON(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(encoded)
}
