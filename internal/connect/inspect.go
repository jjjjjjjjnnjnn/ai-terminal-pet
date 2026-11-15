package connect

// The three inspectors.
//
// Each one answers the same question in a different file format, and the
// formats are not interchangeable (handover §53.1 / §53.5):
//
//	OpenCode     opencode.json -> plugins[] -> may name a *directory* whose
//	             index.ts re-exports the real reporter through a static
//	             file:// import. Two hops, and the second one is invisible
//	             in the JSON that names it.
//	Claude Code  settings.json -> hooks.<Event>[].hooks[].command -> either a
//	             native "<exe> --mode=claude" or "node <script> <Event>".
//	Antigravity  antigravity-cli/settings.json -> statusLine.command, plus
//	             its hook plugin under ~/.gemini/config/plugins/*/hooks.json.
//
// Three rules hold across all of them:
//
//  1. A missing terminal is a warn, never a fail. "This machine has no
//     Antigravity" and "Antigravity is wired to a file that no longer exists"
//     are different machines; only the second one is broken. This is also
//     load-bearing for tests/fresh-machine.mjs S5, which asserts that a
//     standalone install on a bare sandbox reports zero fails -- the sandbox
//     has no host configuration at all.
//  2. A configuration file we cannot parse is reported as exactly that
//     (warn), not as "not connected": claiming the wiring is absent when we
//     failed to read it would send the user to re-do work that is already
//     done.
//  3. Detail never echoes file content. opencode.json on a real machine
//     carries API keys; the line carries paths, counts and event names only.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// configState separates the three ways a configuration file can be unusable
// from the one way it can be usable. The distinction is what rule 2 needs.
type configState int

const (
	configLoaded configState = iota
	configMissing
	configUnreadable
	configMalformed
)

// configLimit keeps one pathological file from turning a diagnostic into a
// memory event. Host configurations are kilobytes; 4 MiB is generous.
const configLimit = 4 << 20

// config is one loaded configuration file.
type config struct {
	state  configState
	reason string
	object map[string]any
}

// loadConfig reads and parses one JSON file. Encoding is not guessed: the
// three hosts all write UTF-8, and a file that is not plain JSON (comments,
// trailing commas) is reported as malformed with the line it broke on, which
// is the most useful thing a diagnostic can say about it.
func loadConfig(path string) config {
	if path == "" {
		return config{state: configMissing, reason: "no path"}
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config{state: configMissing, reason: path}
		}
		return config{state: configUnreadable, reason: err.Error()}
	}
	if info.IsDir() {
		return config{state: configUnreadable, reason: path + " is a directory"}
	}
	if info.Size() > configLimit {
		return config{state: configUnreadable, reason: fmt.Sprintf("%s is %d bytes (limit %d)", path, info.Size(), configLimit)}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return config{state: configUnreadable, reason: err.Error()}
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return config{state: configMalformed, reason: malformedReason(raw, err)}
	}
	if object == nil {
		return config{state: configMalformed, reason: "the document is JSON null"}
	}
	return config{state: configLoaded, object: object}
}

// malformedReason turns a decoder error into "line N: reason", which is what
// makes a JSON-with-comments file actionable instead of mysterious.
func malformedReason(raw []byte, err error) string {
	var syntax *json.SyntaxError
	if !asSyntaxError(err, &syntax) {
		return err.Error()
	}
	line := 1
	for index := 0; index < int(syntax.Offset) && index < len(raw); index++ {
		if raw[index] == '\n' {
			line++
		}
	}
	return fmt.Sprintf("not plain JSON (line %d: %s)", line, syntax.Error())
}

func asSyntaxError(err error, target **json.SyntaxError) bool {
	if syntax, ok := err.(*json.SyntaxError); ok {
		*target = syntax
		return true
	}
	return false
}

// abbreviate is how a path appears in a report line: the machine's own
// directories become ~ and %LOCALAPPDATA%, everything else is shown as its
// last two elements. A bug report stays readable without leaking a full
// profile layout, and two machines' reports are comparable at a glance.
func abbreviate(roots Roots, path string) string {
	if path == "" {
		return "(no path)"
	}
	slashed := strings.ReplaceAll(path, `\`, "/")
	// LocalAppData is tried first because on Windows it sits *inside* the
	// profile: matching ~ first would label every install artifact
	// "~/AppData/Local/..." and hide the one abbreviation users recognise.
	for _, pair := range []struct {
		root  string
		label string
	}{{roots.LocalAppData, "%LOCALAPPDATA%"}, {roots.Home, "~"}} {
		if pair.root == "" {
			continue
		}
		rootSlashed := strings.ReplaceAll(strings.TrimSuffix(pair.root, `\`), `\`, "/")
		if rootSlashed == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(slashed), strings.ToLower(rootSlashed)) {
			rest := slashed[len(rootSlashed):]
			rest = strings.TrimPrefix(rest, "/")
			if rest == "" {
				return pair.label
			}
			return pair.label + "/" + rest
		}
	}
	parts := strings.Split(slashed, "/")
	if len(parts) <= 2 {
		return slashed
	}
	return ".../" + strings.Join(parts[len(parts)-2:], "/")
}

// expandUser resolves "~", "%LOCALAPPDATA%", "$env:LOCALAPPDATA%" and
// "$LOCALAPPDATA" against the Roots a test supplies. Windows hosts write
// paths in all three styles, sometimes in the same file.
func expandUser(roots Roots, value string) string {
	trimmed := unquote(strings.TrimSpace(value))
	for _, name := range []string{"LOCALAPPDATA", "USERPROFILE", "APPDATA", "HOME"} {
		var env string
		switch name {
		case "LOCALAPPDATA":
			env = roots.LocalAppData
		case "USERPROFILE", "HOME":
			env = roots.Home
		}
		if env == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		for _, form := range []string{"%" + strings.ToLower(name) + "%", "$env:" + strings.ToLower(name), "$" + strings.ToLower(name)} {
			if index := strings.Index(lower, form); index >= 0 {
				original := trimmed[index : index+len(form)]
				trimmed = strings.ReplaceAll(trimmed, original, env)
				lower = strings.ToLower(trimmed)
			}
		}
	}
	if strings.HasPrefix(trimmed, "~") {
		rest := strings.TrimPrefix(trimmed, "~")
		rest = strings.TrimLeft(rest, "/\\")
		if rest == "" {
			return roots.Home
		}
		return joinPath(roots.Home, rest)
	}
	return trimmed
}

// commandPath is the first token of a shell command, unquoted and expanded.
// "node C:/.../x.js Stop" -> "node", but for a bare path it is the path
// itself, which is the only case the caller cares about.
func commandPath(roots Roots, command string) string {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return ""
	}
	var first string
	if trimmed[0] == '"' {
		if end := strings.Index(trimmed[1:], `"`); end >= 0 {
			first = trimmed[0 : end+2]
		} else {
			first = trimmed
		}
	} else {
		first = strings.Fields(trimmed)[0]
	}
	return expandUser(roots, first)
}

// exists reports whether a resolved reference is still on disk.
func exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// refersToReporter is the "is this entry about us?" test. It matches the
// names this project actually writes into a host configuration -- the repo
// checkout, the installed directory, and the reporter file names -- and
// nothing else, so an unrelated plugin that happens to live in the same
// directory is not misreported as a broken pet.
func refersToReporter(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"ai-terminal-pet", "terminal-pet", "pet-hook", "pet-statusline",
		"claude-hook.js", "antigravity-hook.js", "opencode.js",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// readSmallFile reads a file only when it is small enough to be a reporter
// script or a wrapper. Stated as a size gate rather than a guess: a 4 MiB
// PowerShell profile that happens to be named in a command must not be read
// into memory just to look for a bridge path.
func readSmallFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() || info.Size() > 256<<10 {
		return "", fmt.Errorf("%s is not a script (%d bytes)", path, info.Size())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// token is one lexical element of a JS/TS source. Only two kinds matter here:
// a word (which is how an `from` keyword is recognised) and a string literal
// (which is what a wrapper's target is). Comments and punctuation are dropped,
// which is precisely the behaviour both callers want and the reason this is one
// scanner instead of two loops with their own comment heuristics.
type token struct {
	kind  tokenKind
	value string
}

type tokenKind uint8

const (
	tokenWord tokenKind = iota
	tokenString
)

// scanSource walks a JavaScript/TypeScript source once, tracking the five states
// that matter (normal, the three string delimiters, line comment, block
// comment) and emitting words and string contents.
//
// Why not a regex: this machine's own wrapper quotes `reporters/opencode.js`
// three times in prose before the import that matters, so a pattern that does
// not understand comments sends --doctor after a file the wrapper never loads.
// Why not a parser: this package is linked into the overlay binary, and a
// JavaScript parser would be a dependency bought for a diagnostic.
func scanSource(source string) []token {
	tokens := make([]token, 0, 32)
	index := 0
	for index < len(source) {
		character := source[index]
		switch {
		case character == '/' && index+1 < len(source) && source[index+1] == '/':
			if end := strings.IndexByte(source[index:], '\n'); end >= 0 {
				index += end + 1
			} else {
				index = len(source)
			}
		case character == '/' && index+1 < len(source) && source[index+1] == '*':
			end := strings.Index(source[index+2:], "*/")
			if end < 0 {
				index = len(source)
			} else {
				index += end + 4
			}
		case character == '"' || character == '\'' || character == '`':
			value, next := scanString(source, index)
			tokens = append(tokens, token{kind: tokenString, value: value})
			index = next
		case isIdentifierByte(character):
			start := index
			for index < len(source) && isIdentifierByte(source[index]) {
				index++
			}
			tokens = append(tokens, token{kind: tokenWord, value: source[start:index]})
		default:
			index++
		}
	}
	return tokens
}

// scanString reads one quoted string starting at the opening quote and returns
// its content and the index just past the closing quote. An unterminated string
// takes the rest of the source, which is how a broken wrapper reads anyway.
func scanString(source string, start int) (string, int) {
	quote := source[start]
	index := start + 1
	var builder strings.Builder
	for index < len(source) {
		character := source[index]
		if character == '\\' && index+1 < len(source) {
			builder.WriteByte(source[index+1])
			index += 2
			continue
		}
		if character == quote {
			return builder.String(), index + 1
		}
		builder.WriteByte(character)
		index++
	}
	return builder.String(), index
}

// stringLiterals returns every string literal's content, in source order.
func stringLiterals(source string) []string {
	literals := make([]string, 0, 8)
	for _, item := range scanSource(source) {
		if item.kind == tokenString {
			literals = append(literals, item.value)
		}
	}
	return literals
}
