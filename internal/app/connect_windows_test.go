package app

// The two things a test can hold this presentation to without a machine:
// --connect prints a block that names one real file per terminal, and it does
// not depend on what the developer who runs `go test` happens to have
// configured.
//
// The real-machine case is covered by the opt-in live test in this package's
// connect_windows_live_test.go (PET_CONNECT_PROBE=1), because a machine without
// OpenCode installed must still be able to run the rest of the suite.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ai-terminal-pet/internal/connect"
)

// decodeJSON is encoding/json behind the name the test reads with.
func decodeJSON(value string, target any) error {
	return json.Unmarshal([]byte(value), target)
}

// connectFixture lays out a profile plus install directory the way a working
// machine looks, and returns the Roots a test can point connectText at.
func connectFixture(t *testing.T) connect.Roots {
	t.Helper()
	root := t.TempDir()
	roots := connect.Roots{
		Home:         filepath.Join(root, "profile"),
		LocalAppData: filepath.Join(root, "local"),
		Executable:   filepath.Join(root, "repo", "bin", "AI-Terminal-Pet.exe"),
	}
	install := filepath.Join(roots.LocalAppData, installDirectoryName)
	reporterDir := filepath.Join(root, "repo", "reporters")
	for _, directory := range []string{roots.Home, install, filepath.Dir(roots.Executable), reporterDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{connect.HookBinaryName, connect.StatuslineBinaryName} {
		if err := os.WriteFile(filepath.Join(install, name), []byte("stub"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "repo", "reporters", "opencode.js"), []byte("module.exports = {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return roots
}

func TestConnectPrintsEveryTerminalInOrder(t *testing.T) {
	roots := connectFixture(t)
	text, err := connectText(roots, "")
	if err != nil {
		t.Fatalf("connectText: %v", err)
	}
	// Order is part of the contract: a report that reshuffles between runs is
	// unreadable in a bug thread.
	lastIndex := -1
	for _, terminal := range connect.TerminalNames() {
		index := strings.Index(text, "== "+terminal+" ==")
		if index < 0 {
			t.Fatalf("output has no %s block:\n%s", terminal, text)
		}
		if index < lastIndex {
			t.Fatalf("%s came out of order:\n%s", terminal, text)
		}
		lastIndex = index
	}
	if !strings.Contains(text, "written nothing") {
		t.Fatalf("the footer has to say the command wrote nothing:\n%s", text)
	}
}

func TestConnectUnknownTerminalIsAnError(t *testing.T) {
	roots := connectFixture(t)
	text, err := connectText(roots, "codex")
	if err == nil {
		t.Fatalf("an unknown terminal was accepted and printed:\n%s", text)
	}
	if !strings.Contains(err.Error(), "opencode") {
		t.Fatalf("the error has to list the terminals that do exist: %v", err)
	}
}

func TestConnectTextNamesARealFileForEveryTerminal(t *testing.T) {
	roots := connectFixture(t)
	for _, terminal := range connect.TerminalNames() {
		text, err := connectText(roots, terminal)
		if err != nil {
			t.Fatalf("connectText(%q): %v", terminal, err)
		}
		installRoot := filepath.Join(roots.LocalAppData, installDirectoryName)
		switch terminal {
		case connect.TerminalOpenCode:
			// The package is the form a normal user applies; the checkout URL
			// is the development form and only appears when one is reachable.
			if !strings.Contains(text, connect.OpenCodePackageName) {
				t.Fatalf("opencode block does not name the package:\n%s", text)
			}
			if !strings.Contains(text, "opencode plugin add") {
				t.Fatalf("opencode block does not name the install command:\n%s", text)
			}
			if !strings.Contains(text, "file:///") || !strings.Contains(text, "reporters/opencode.js") {
				t.Fatalf("opencode block lost the checkout form the fixture provides:\n%s", text)
			}
		case connect.TerminalClaude:
			if !strings.Contains(text, connect.HookBinaryName) {
				t.Fatalf("claude block should prefer the installed hook:\n%s", text)
			}
			// Every event line has to carry the same command, or the block is
			// a partial wiring that looks complete.
			count := strings.Count(text, connect.HookBinaryName)
			if count != len(connect.ClaudeHookEvents()) {
				t.Fatalf("claude block has %d command lines for %d events:\n%s",
					count, len(connect.ClaudeHookEvents()), text)
			}
		case connect.TerminalAntigravity:
			if !strings.Contains(text, connect.StatuslineBinaryName) {
				t.Fatalf("antigravity block does not name the statusline:\n%s", text)
			}
		}
		_ = installRoot
	}
}

func TestConnectJSONQuotedPathsAreValidJSON(t *testing.T) {
	roots := connectFixture(t)
	text, err := connectText(roots, connect.TerminalClaude)
	if err != nil {
		t.Fatal(err)
	}
	// Find the first command line and parse the whole block with the real
	// decoder: this is the one assertion that catches an unescaped Windows
	// path, which is how the first version produced
	// "command": ""C:\Users\..."".
	// The snippet is a fragment meant to be merged into a file that already has
	// braces, so the test wraps it exactly the way a user's file would.
	body := text[strings.Index(text, "\"hooks\""):]
	body = body[:strings.LastIndex(body, "}")+1]
	var decoded map[string]any
	if err := decodeJSON("{"+body+"}", &decoded); err != nil {
		t.Fatalf("the printed block is not valid JSON: %v\n%s", err, body)
	}
	hooks, ok := decoded["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("decoded block has no hooks object: %v", decoded)
	}
	if len(hooks) == 0 {
		t.Fatal("decoded hooks object is empty")
	}
	for event, rawGroups := range hooks {
		groups, ok := rawGroups.([]any)
		if !ok || len(groups) == 0 {
			t.Fatalf("%s is not a non-empty array", event)
		}
		group, ok := groups[0].(map[string]any)
		if !ok {
			t.Fatalf("%s[0] is not an object", event)
		}
		if _, ok := group["matcher"]; !ok {
			t.Fatalf("%s[0] has no matcher", event)
		}
	}
}
