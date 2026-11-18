package connect

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

const pluginID = "ai-terminal-pet@ai-terminal-pet-market"

func writeClaudeSettings(t *testing.T, roots Roots, body string) {
	t.Helper()
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"), body)
}

func writePluginHooks(t *testing.T, installPath string, withBinary bool) {
	t.Helper()
	events := map[string]any{}
	for _, event := range claudeHookEvents() {
		events[event] = []any{
			map[string]any{
				"matcher": "*",
				"hooks": []any{
					map[string]any{
						"type":    "command",
						"command": "\"${CLAUDE_PLUGIN_ROOT}/vendor/pet-hook.exe\" --mode=claude",
					},
				},
			},
		}
	}
	encoded, err := json.Marshal(map[string]any{"hooks": events})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(installPath, "hooks", "hooks.json"), string(encoded))
	if withBinary {
		writeFile(t, filepath.Join(installPath, "vendor", "pet-hook.exe"), "stub hook")
	}
}

func writeInstalledPlugins(t *testing.T, roots Roots, installPath, scope string) {
	t.Helper()
	record := map[string]any{
		"version": 2,
		"plugins": map[string]any{
			pluginID: []any{
				map[string]any{
					"scope":       scope,
					"installPath": installPath,
					"version":     "1.0.22",
				},
			},
		},
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, claudePluginsPath(roots), string(encoded))
}

func pluginInstallPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "plugins", "cache", "ai-terminal-pet-market", "ai-terminal-pet", "1.0.22")
}

func TestInspectClaudeFindsThePluginWhenSettingsHasNoHooks(t *testing.T) {
	roots := home(t)
	writeClaudeSettings(t, roots, `{"enabledPlugins":{"`+pluginID+`":true}}`)
	install := pluginInstallPath(t)
	writePluginHooks(t, install, true)
	writeInstalledPlugins(t, roots, install, "user")

	finding := InspectClaude(roots)
	if finding.Status != StatusOK {
		t.Fatalf("status = %s, detail = %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "plugin "+pluginID) {
		t.Fatalf("detail does not name the plugin: %s", finding.Detail)
	}
	if !strings.Contains(finding.Detail, "8 event(s)") {
		t.Fatalf("detail does not count the events: %s", finding.Detail)
	}
}

func TestInspectClaudeFailsWhenThePluginBinaryIsMissing(t *testing.T) {
	roots := home(t)
	writeClaudeSettings(t, roots, `{"enabledPlugins":{"`+pluginID+`":true}}`)
	install := pluginInstallPath(t)
	writePluginHooks(t, install, false)
	writeInstalledPlugins(t, roots, install, "user")

	finding := InspectClaude(roots)
	if finding.Status != StatusFail {
		t.Fatalf("status = %s, detail = %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "is missing") {
		t.Fatalf("detail should say what is missing: %s", finding.Detail)
	}
}

func TestInspectClaudeWarnsWhenThePluginIsDisabled(t *testing.T) {
	roots := home(t)
	writeClaudeSettings(t, roots, `{"enabledPlugins":{"`+pluginID+`":false}}`)
	install := pluginInstallPath(t)
	writePluginHooks(t, install, true)
	writeInstalledPlugins(t, roots, install, "user")

	finding := InspectClaude(roots)
	if finding.Status != StatusWarn {
		t.Fatalf("status = %s, detail = %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "disabled") {
		t.Fatalf("detail should say disabled: %s", finding.Detail)
	}
}

func TestInspectClaudeKeepsThePlainWarningWithoutAPlugin(t *testing.T) {
	roots := home(t)
	writeClaudeSettings(t, roots, `{}`)

	finding := InspectClaude(roots)
	if finding.Status != StatusWarn {
		t.Fatalf("status = %s, detail = %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "not connected") {
		t.Fatalf("detail should still say not connected: %s", finding.Detail)
	}
}

func TestClaudePluginEnabledDefaults(t *testing.T) {
	enabled := map[string]any{"enabledPlugins": map[string]any{pluginID: true}}
	disabled := map[string]any{"enabledPlugins": map[string]any{pluginID: false}}
	if !claudePluginEnabled(enabled, pluginID, "user") {
		t.Fatal("explicit true must be enabled")
	}
	if claudePluginEnabled(disabled, pluginID, "user") {
		t.Fatal("explicit false must be disabled")
	}
	if !claudePluginEnabled(map[string]any{}, pluginID, "user") {
		t.Fatal("an older host without enabledPlugins must be assumed enabled")
	}
	if claudePluginEnabled(map[string]any{"enabledPlugins": map[string]any{}}, pluginID, "user") {
		t.Fatal("a user-scoped plugin missing from an existing map is disabled")
	}
}
