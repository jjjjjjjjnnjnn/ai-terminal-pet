package connect

// The Claude Code check: settings.json -> hooks.<Event>[].hooks[].command.
//
// Two shapes of command appear in the wild and both are accepted:
//
//	"%LOCALAPPDATA%\AI-Terminal-Pet\pet-hook.exe" --mode=claude   (native)
//	node C:/Users/<you>/.claude/hooks/terminal-pet-hook.js Stop   (JS path)
//	node <repo>/reporters/claude-hook.js Stop                     (repo path)
//
// The JS path is not necessarily the end of the chain: this machine's installed
// wrapper (handover §53.5) is a reporter that requires a bridge file out of the
// install directory, so its command can be valid while the wiring behind it is
// dead. That second hop is checked too, because it is invisible to the host:
// Claude Code starts the wrapper, the wrapper swallows the failure, and the
// panel just stays empty.

import (
	"path/filepath"
	"sort"
	"strings"
)

// claudeSettingsPath is ~/.claude/settings.json.
func claudeSettingsPath(roots Roots) string {
	if roots.Home == "" {
		return ""
	}
	return filepath.Join(roots.Home, ".claude", "settings.json")
}

// InspectClaude reports whether Claude Code is wired to a reporter that still
// exists, and how many events it covers.
func InspectClaude(roots Roots) Finding {
	finding := Finding{Terminal: TerminalClaude}
	if roots.Home == "" {
		return finding.warn("cannot be checked (%s is not set)", homeVariable)
	}
	configPath := claudeSettingsPath(roots)
	loaded := loadConfig(configPath)
	switch loaded.state {
	case configMissing:
		return finding.warn("not installed (no %s)", abbreviate(roots, configPath))
	case configUnreadable:
		return finding.warn("cannot read %s: %s", abbreviate(roots, configPath), loaded.reason)
	case configMalformed:
		return finding.warn("%s is %s", abbreviate(roots, configPath), loaded.reason)
	}

	entries := claudeHookCommands(loaded.object)
	if len(entries) == 0 {
		return finding.warn("not connected (%s has no hook that reaches a reporter; run `ai-terminal-pet --connect=claude`)",
			abbreviate(roots, configPath))
	}

	events := make([]string, 0, len(entries))
	seen := map[string]bool{}
	for _, entry := range entries {
		if !seen[entry.event] {
			seen[entry.event] = true
			events = append(events, entry.event)
		}
	}
	// claudeHookCommands walks a map, so its order is randomized per run. A
	// report line that reshuffles is unreadable in a bug thread and would make
	// --doctor's output non-diffable, so the events are put back into the
	// documented order and then alphabetically for anything unknown.
	sortEvents(events)

	// The first entry decides the verdict; the rest only widen the count. A
	// machine with two reporters wired (native plus the JS fallback) is not a
	// finding worth a second line, and reporting both would bury the events.
	command := entries[0].command
	target, kind := claudeCommandTarget(roots, command)
	if target == "" {
		return finding.warn("%d event(s) wired, but %q names no file to run", len(entries), truncate(command, 60))
	}
	if !exists(target) {
		return finding.fail("%d event(s) -> %s (%s missing)", len(entries), abbreviate(roots, target), kind)
	}
	// Second hop: a JS reporter that pulls its bridge out of the install
	// directory. Only a path this project owns is checked -- a reporter that
	// requires a package name has nothing here to verify.
	if bridge := bridgeReference(target); bridge != "" && !exists(bridge) {
		return finding.fail("%d event(s) -> %s, whose bridge %s is missing", len(entries), abbreviate(roots, target), abbreviate(roots, bridge))
	}

	detail := strings.Join(events, "/")
	if !seen["SessionEnd"] {
		// Handover §53.5: the closed signal is the host's SessionEnd hook. A
		// wiring without it leaves every session sitting at idle forever,
		// which reads as "the pet is stuck" rather than as a missing hook.
		return finding.warn("%d event(s) -> %s, but no SessionEnd, so a finished session stays idle (%s)",
			len(entries), abbreviate(roots, target), detail)
	}
	return finding.ok("%d event(s) -> %s (%s)", len(entries), abbreviate(roots, target), detail)
}

// hookEntry is one command found in settings.json, with the event it belongs
// to. The event name is kept because the set is part of the verdict.
type hookEntry struct {
	event   string
	command string
}

// sortEvents orders hook event names by the documented set first and
// alphabetically after it. Hand-rolled rather than sort.Strings because the
// documented order is the reading order: SessionStart ... SessionEnd tells the
// story of a session, and alphabetical does not.
func sortEvents(events []string) {
	order := make(map[string]int, 8)
	for index, name := range claudeHookEvents() {
		order[name] = index
	}
	sort.SliceStable(events, func(left, right int) bool {
		leftRank, leftKnown := order[events[left]]
		rightRank, rightKnown := order[events[right]]
		switch {
		case leftKnown && rightKnown:
			return leftRank < rightRank
		case leftKnown != rightKnown:
			return leftKnown
		default:
			return events[left] < events[right]
		}
	})
}

// claudeHookCommands walks hooks.<Event>[].hooks[].command and returns only
// the commands that reach us. Claude Code's own shape allows several matcher
// groups per event and several hooks per group; flattening is correct because
// the check is "is anything wired at all".
func claudeHookCommands(object map[string]any) []hookEntry {
	rawHooks, ok := object["hooks"].(map[string]any)
	if !ok {
		return nil
	}
	entries := make([]hookEntry, 0, 8)
	for event, rawGroups := range rawHooks {
		groups, ok := rawGroups.([]any)
		if !ok {
			continue
		}
		for _, rawGroup := range groups {
			group, ok := rawGroup.(map[string]any)
			if !ok {
				continue
			}
			commands, ok := group["hooks"].([]any)
			if !ok {
				continue
			}
			for _, rawCommand := range commands {
				commandGroup, ok := rawCommand.(map[string]any)
				if !ok {
					continue
				}
				command, ok := commandGroup["command"].(string)
				if !ok || !refersToReporter(command) {
					continue
				}
				entries = append(entries, hookEntry{event: event, command: command})
			}
		}
	}
	return entries
}

// claudeCommandTarget resolves the file a hook command runs, and names the kind
// so a report explains itself without the reader opening settings.json.
func claudeCommandTarget(roots Roots, command string) (string, string) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return "", ""
	}
	first := commandPath(roots, trimmed)
	// "node <script> <Event>": the file we care about is the script, not node.
	if baseName(first) == "node" || baseName(first) == "node.exe" {
		fields := splitCommand(trimmed)
		if len(fields) < 2 {
			return "", ""
		}
		return expandUser(roots, fields[1]), "script"
	}
	if strings.HasSuffix(strings.ToLower(baseName(first)), ".js") {
		return first, "script"
	}
	return first, "binary"
}

// splitCommand splits on whitespace but keeps a quoted first argument whole,
// because every install path on Windows is one space away from needing quotes.
func splitCommand(command string) []string {
	if command == "" {
		return nil
	}
	fields := make([]string, 0, 4)
	rest := strings.TrimSpace(command)
	for rest != "" {
		if rest[0] == '"' {
			end := strings.Index(rest[1:], `"`)
			if end < 0 {
				fields = append(fields, rest)
				break
			}
			fields = append(fields, rest[:end+2])
			rest = strings.TrimSpace(rest[end+2:])
			continue
		}
		next := strings.IndexAny(rest, " \t")
		if next < 0 {
			fields = append(fields, rest)
			break
		}
		fields = append(fields, rest[:next])
		rest = strings.TrimSpace(rest[next:])
	}
	return fields
}

// bridgeReference looks inside a JS reporter for a bridge file it loads from
// this project's install directory. It returns the first such path, already
// resolved, or "" when the script carries none.
//
// The scan is deliberately narrow: only string literals that look like an
// absolute path AND name something this project installs. A reporter that
// requires "node:fs" or a package by name must not be read as a path here.
func bridgeReference(script string) string {
	raw, err := readSmallFile(script)
	if err != nil {
		return ""
	}
	for _, literal := range stringLiterals(raw) {
		if !looksLikeAbsolutePath(literal) || !refersToBridge(literal) {
			continue
		}
		return unescapeWindowsPath(literal)
	}
	return ""
}

// refersToBridge is narrower than refersToReporter: it only matches the files
// a JS reporter loads out of the install directory, so a sibling reporter that
// happens to be named claude-hook.js is not treated as a bridge.
func refersToBridge(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"pet-ipc", "protocol.js", "pet-statusline-shim.js"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// looksLikeAbsolutePath rejects a bare package name: only something with a
// drive letter or a leading slash can be stat()ed from here.
func looksLikeAbsolutePath(value string) bool {
	if len(value) >= 2 && value[1] == ':' {
		return true
	}
	return strings.HasPrefix(value, `\\`) || strings.HasPrefix(value, "/")
}

// unescapeWindowsPath turns the escaped form a JavaScript literal carries
// ("C:\\Users\\x") into a real path. Only the escape a path actually needs is
// collapsed: a blanket unescape would turn a literal "\t" inside a folder name
// into a tab and hand back a path that does not exist.
func unescapeWindowsPath(value string) string {
	return strings.ReplaceAll(value, `\\`, `\`)
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
