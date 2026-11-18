package connect

// The second shape of a Claude Code wiring: a plugin.
//
// A user who installs `packages/claude-ai-terminal-pet` does not write anything
// into ~/.claude/settings.json. The host registers the plugin's
// hooks/hooks.json from ~/.claude/plugins/cache/... itself, and keeps the
// bookkeeping in ~/.claude/plugins/installed_plugins.json. Reading only the
// settings file would report that machine as "not connected" - a false warning
// about work the user just did, which is exactly what --doctor exists to avoid.
//
// The check stays narrow on purpose: it recognizes the plugin by id
// (ai-terminal-pet@<market>), expands the host's ${CLAUDE_PLUGIN_ROOT}, and
// then reuses the same command parser as the settings shape. Nothing here
// executes a hook or trusts a version number.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// claudePluginName is the plugin id this project publishes. A marketplace may
// rename its container; the plugin name is ours.
const claudePluginName = "ai-terminal-pet"

// installedPluginsFile is the version-2 shape Claude Code writes:
//
//	{"version":2,"plugins":{"ai-terminal-pet@market":[{"scope":"user",
//	  "installPath":"…/plugins/cache/market/ai-terminal-pet/1.0.22", …}]}}
type installedPluginsFile struct {
	Version int                                `json:"version"`
	Plugins map[string][]installedPluginRecord `json:"plugins"`
}

type installedPluginRecord struct {
	Scope       string `json:"scope"`
	InstallPath string `json:"installPath"`
	Version     string `json:"version"`
}

func claudePluginsPath(roots Roots) string {
	if roots.Home == "" {
		return ""
	}
	return filepath.Join(roots.Home, ".claude", "plugins", "installed_plugins.json")
}

// inspectClaudePlugin returns a finding and whether a plugin install was found
// at all. found=false means "no plugin to comment on" and the caller keeps its
// original "not connected" warning.
func inspectClaudePlugin(roots Roots, settings map[string]any) (Finding, bool) {
	finding := Finding{Terminal: TerminalClaude}
	pluginsPath := claudePluginsPath(roots)
	if pluginsPath == "" {
		return finding, false
	}
	raw, err := os.ReadFile(pluginsPath)
	if err != nil {
		return finding, false
	}
	var installed installedPluginsFile
	if err := json.Unmarshal(raw, &installed); err != nil {
		return finding.warn("%s is not valid JSON: %v", abbreviate(roots, pluginsPath), err), true
	}

	ids := make([]string, 0, len(installed.Plugins))
	for id := range installed.Plugins {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if !strings.HasPrefix(id, claudePluginName+"@") {
			continue
		}
		for _, record := range installed.Plugins[id] {
			if record.InstallPath == "" {
				continue
			}
			hooksPath := filepath.Join(record.InstallPath, "hooks", "hooks.json")
			hooks := loadConfig(hooksPath)
			if hooks.state != configLoaded {
				continue
			}
			entries := expandPluginRoot(claudeHookCommands(hooks.object), record.InstallPath)
			if len(entries) == 0 {
				continue
			}

			events := make([]string, 0, len(entries))
			seen := map[string]bool{}
			for _, entry := range entries {
				if !seen[entry.event] {
					seen[entry.event] = true
					events = append(events, entry.event)
				}
			}
			sortEvents(events)

			target, _ := claudeCommandTarget(roots, entries[0].command)
			if target == "" || !exists(target) {
				return finding.fail("plugin %s %s is installed, but %s is missing", id, record.Version, abbreviate(roots, target)), true
			}
			if !claudePluginEnabled(settings, id, record.Scope) {
				return finding.warn("plugin %s %s is installed but disabled (claude plugin enable %s)", id, record.Version, id), true
			}
			detail := strings.Join(events, "/")
			if !seen["SessionEnd"] {
				return finding.warn("%d event(s) -> %s (plugin %s %s), but no SessionEnd, so a finished session stays idle (%s)",
					len(entries), abbreviate(roots, target), id, record.Version, detail), true
			}
			return finding.ok("%d event(s) -> %s (plugin %s %s, %s)",
				len(entries), abbreviate(roots, target), id, record.Version, detail), true
		}
	}
	return finding, false
}

// expandPluginRoot substitutes the host variable a plugin command carries.
// Both spellings occur in host documentation and in the wild.
func expandPluginRoot(entries []hookEntry, root string) []hookEntry {
	expanded := make([]hookEntry, len(entries))
	for index, entry := range entries {
		command := strings.ReplaceAll(entry.command, "${CLAUDE_PLUGIN_ROOT}", root)
		command = strings.ReplaceAll(command, "$CLAUDE_PLUGIN_ROOT", root)
		expanded[index] = hookEntry{event: entry.event, command: command}
	}
	return expanded
}

// claudePluginEnabled reads enabledPlugins from the user settings. A missing
// map means an older host that did not track the flag; a missing id inside an
// existing map is how the CLI writes a disabled plugin. Project- and
// local-scoped installs keep their flag in the project settings, which this
// check does not read, so they are assumed enabled rather than misreported.
func claudePluginEnabled(settings map[string]any, id, scope string) bool {
	enabled, ok := settings["enabledPlugins"].(map[string]any)
	if !ok {
		return true
	}
	value, present := enabled[id]
	if !present {
		return scope != "user"
	}
	flag, ok := value.(bool)
	if !ok {
		return true
	}
	return flag
}
