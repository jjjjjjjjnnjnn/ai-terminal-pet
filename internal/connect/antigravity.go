package connect

// The Antigravity check: one statusline command, plus the question that took a
// whole round to answer.
//
// Handover §53.5 established the shape of this host: its statusline is a
// *level* interface -- every render feeds one "this is my current state" sample
// and there is no edge, so a finished or terminated agent arrives as `idle`,
// never as `closed`. Only a hook that carries `terminationReason` produces the
// closed edge. On this machine (measured in §53.5) the hook plugin ships
// BeforeAgent / AfterAgent / BeforeTool / Notification and no termination
// event, so its sessions legitimately end at idle.
//
// That is a design fact, not a defect, so it is reported as a warn with the
// reason spelled out: the next person to ask "why does agy stay in the panel
// after I close it" gets the answer from --doctor instead of from a handover
// section. What would be a defect is a statusline path that is gone.

import (
	"os"
	"path/filepath"
)

// antigravitySettingsPath is ~/.gemini/antigravity-cli/settings.json.
func antigravitySettingsPath(roots Roots) string {
	if roots.Home == "" {
		return ""
	}
	return filepath.Join(roots.Home, ".gemini", "antigravity-cli", "settings.json")
}

// antigravityPluginDir is where Antigravity keeps hook plugins:
// ~/.gemini/config/plugins/<name>/hooks.json.
func antigravityPluginDir(roots Roots) string {
	if roots.Home == "" {
		return ""
	}
	return filepath.Join(roots.Home, ".gemini", "config", "plugins")
}

// InspectAntigravity reports whether the statusline command still exists and
// whether anything in the hook set can produce a closed edge.
func InspectAntigravity(roots Roots) Finding {
	finding := Finding{Terminal: TerminalAntigravity}
	if roots.Home == "" {
		return finding.warn("cannot be checked (%s is not set)", homeVariable)
	}
	configPath := antigravitySettingsPath(roots)
	loaded := loadConfig(configPath)
	switch loaded.state {
	case configMissing:
		return finding.warn("not installed (no %s)", abbreviate(roots, configPath))
	case configUnreadable:
		return finding.warn("cannot read %s: %s", abbreviate(roots, configPath), loaded.reason)
	case configMalformed:
		return finding.warn("%s is %s", abbreviate(roots, configPath), loaded.reason)
	}

	command := antigravityStatusLine(loaded.object)
	termination := antigravityHasTerminationHook(roots)
	hasLevel := command != "" && refersToReporter(command)
	hasEdge := termination

	// A dangling path is a defect whatever else is wired: the statusline runs
	// on every render, and a missing binary makes each one fail silently.
	if hasLevel {
		target := commandPath(roots, command)
		if !exists(target) {
			return finding.fail("statusline -> %s (missing)", abbreviate(roots, target))
		}
	}

	switch {
	case hasEdge:
		// An edge signal outranks the level one: with a termination hook the
		// session really does close, so the level reporter being absent is
		// only a loss of turn-by-turn detail, not of the end signal.
		if hasLevel {
			return finding.ok("termination hook + statusline -> %s", abbreviate(roots, commandPath(roots, command)))
		}
		return finding.warn("termination hook only (no statusline reporter, so turn state is missing)")
	case hasLevel:
		return finding.warn("statusline only -> %s; sessions end at idle, never closed, because no hook sends terminationReason",
			abbreviate(roots, commandPath(roots, command)))
	}
	return finding.warn("not connected (no statusline reporter and no termination hook; see `ai-terminal-pet --connect=antigravity`)")
}

// antigravityStatusLine digs statusLine.command out of settings.json. The block
// may also be a plain string in some builds, so both are accepted.
func antigravityStatusLine(object map[string]any) string {
	value, ok := object["statusLine"]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]any:
		if command, ok := typed["command"].(string); ok {
			return command
		}
	}
	return ""
}

// antigravityHasTerminationHook scans every hook plugin under
// ~/.gemini/config/plugins/*/hooks.json for an event that can carry a
// termination reason.
//
// It looks for the event *name*, not for our binary: on this machine the
// termination-capable host events live in someone else's notifier plugin, and
// what matters for the "stays at idle" question is whether the host emits the
// event at all. A plugin we cannot parse is skipped rather than guessed at.
func antigravityHasTerminationHook(roots Roots) bool {
	directory := antigravityPluginDir(roots)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		loaded := loadConfig(filepath.Join(directory, entry.Name(), "hooks.json"))
		if loaded.state != configLoaded {
			continue
		}
		if objectMentionsTermination(loaded.object) {
			return true
		}
	}
	return false
}

// objectMentionsTermination walks a decoded JSON document for a key that
// mentions termination. Depth is unbounded but a hook file is shallow; the
// walk is a map/list recursive descent with no type assertions that can panic.
func objectMentionsTermination(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if mentionsTermination(key) {
				return true
			}
			if objectMentionsTermination(child) {
				return true
			}
		}
	case string:
		return mentionsTermination(typed)
	case []any:
		for _, child := range typed {
			if objectMentionsTermination(child) {
				return true
			}
		}
	}
	return false
}

func mentionsTermination(value string) bool {
	lower := toLowerASCII(value)
	for index := 0; index+11 <= len(lower); index++ {
		if lower[index:index+11] == "termination" {
			return true
		}
	}
	return false
}

// toLowerASCII avoids strings.ToLower's Unicode tables for a hot, tiny call;
// the keys we search for are ASCII.
func toLowerASCII(value string) string {
	lowered := make([]byte, len(value))
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		lowered[index] = character
	}
	return string(lowered)
}
