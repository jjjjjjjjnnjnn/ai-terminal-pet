package connect

// The tests in this file are the acceptance for the whole package: they build
// three host configurations in a temp directory and drive every branch of every
// inspector, so none of them depends on what is installed on the machine that
// runs them.
//
// Three are worth reading before changing the code:
//
//   - TestPathToFileURLReproducesTheWorkingWrapper pins the exact string this
//     machine's OpenCode wrapper carried (handover §53.1). If that assertion
//     fails, the --connect output stops being pasteable and the OpenCode check
//     stops resolving, and both happen for the same reason.
//   - TestInspectOpenCodeFollowsTheSecondHop is the regression for handover
//     §52.3: the JSON entry and the directory can both be perfect while the
//     file the wrapper imports is gone.
//   - TestSnippetsWriteNothing is the rule --connect exists under: it prints,
//     it never edits another product's configuration.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// home builds a Roots over a throwaway profile with an install directory that
// contains the artifacts every snippet prefers, plus a repository checkout
// (bin/ + reporters/) the way a developer runs --connect the first time.
func home(t *testing.T) Roots {
	t.Helper()
	root := t.TempDir()
	roots := Roots{
		Home:         filepath.Join(root, "profile"),
		LocalAppData: filepath.Join(root, "local"),
		Executable:   filepath.Join(root, "repo", "bin", "AI-Terminal-Pet.exe"),
	}
	if err := os.MkdirAll(roots.Home, 0o700); err != nil {
		t.Fatal(err)
	}
	install := filepath.Join(roots.LocalAppData, installDirectoryName)
	if err := os.MkdirAll(install, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{hookBinaryName, statuslineBinaryName, "AI-Terminal-Pet.exe", "protocol.js"} {
		writeFile(t, filepath.Join(install, name), "stub "+name)
	}
	writeFile(t, roots.Executable, "stub overlay")
	writeFile(t, filepath.Join(root, "repo", "reporters", "opencode.js"), "module.exports = {}\n")
	return roots
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestInspectOpenCodeFollowsTheSecondHop(t *testing.T) {
	roots := home(t)
	reporter := filepath.Join(roots.Home, "repo", "reporters", "opencode.js")
	writeFile(t, reporter, "module.exports = {}\n")
	plugins := filepath.Join(roots.Home, ".config", "opencode", "plugins", "terminal-pet")
	if err := os.MkdirAll(plugins, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(plugins, "index.ts"),
		"import { plugin } from \""+mustURL(t, reporter)+"\"\n\nexport default plugin\n")
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"),
		`{"plugins": ["`+mustURL(t, plugins)+`"]}`)

	finding := InspectOpenCode(roots)
	if finding.Status != StatusOK {
		t.Fatalf("status = %s, want ok: %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "index.ts") || !stringContains(finding.Detail, "opencode.js") {
		t.Fatalf("detail has to name both hops: %s", finding.Detail)
	}

	// The hop nobody can see from opencode.json: delete the repository file
	// and the wrapper still looks perfect.
	if err := os.Remove(reporter); err != nil {
		t.Fatal(err)
	}
	finding = InspectOpenCode(roots)
	if finding.Status != StatusFail {
		t.Fatalf("a wrapper pointing at a deleted file must fail, got %s: %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "missing") {
		t.Fatalf("failure has to say what is missing: %s", finding.Detail)
	}
}

func TestInspectOpenCodeDirectFileEntry(t *testing.T) {
	roots := home(t)
	reporter := filepath.Join(roots.Home, "repo", "reporters", "opencode.js")
	writeFile(t, reporter, "module.exports = {}\n")
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"),
		`{"plugins": ["`+mustURL(t, reporter)+`"]}`)
	finding := InspectOpenCode(roots)
	if finding.Status != StatusOK {
		t.Fatalf("a direct file entry must be ok, got %s: %s", finding.Status, finding.Detail)
	}
}

func TestInspectOpenCodeReportsWhatItCannotSee(t *testing.T) {
	cases := []struct {
		name       string
		config     string
		wantStatus Status
		wantText   string
	}{
		{"no config", "", StatusWarn, "not installed"},
		{"no plugins key", `{"model": "x"}`, StatusWarn, "no plugins"},
		{"empty list", `{"plugins": []}`, StatusWarn, "empty"},
		{"foreign entry", `{"plugins": ["file:///C:/other/plugin.js"]}`, StatusWarn, "none of them ours"},
		{"wrong type", `{"plugins": {"a": 1}}`, StatusWarn, "not a list of strings"},
		{"comments", "{\n  // OpenCode allows a comment here\n  \"plugins\": []\n}", StatusWarn, "line 2"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			roots := home(t)
			if testCase.config != "" {
				writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"), testCase.config)
			}
			finding := InspectOpenCode(roots)
			if finding.Status != testCase.wantStatus {
				t.Fatalf("status = %s, want %s: %s", finding.Status, testCase.wantStatus, finding.Detail)
			}
			if !strings.Contains(finding.Detail, testCase.wantText) {
				t.Fatalf("detail %q does not contain %q", finding.Detail, testCase.wantText)
			}
		})
	}

	// A malformed file must read as "cannot inspect", never as "not
	// connected": claiming the wiring is absent sends the user to redo work
	// that is already done.
	roots := home(t)
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"), "{\n  // comment\n}")
	finding := InspectOpenCode(roots)
	if strings.Contains(finding.Detail, "not connected") {
		t.Fatalf("a parse failure was reported as a missing connection: %s", finding.Detail)
	}
}

func TestInspectOpenCodeDirectoryWithoutEntrypointFails(t *testing.T) {
	roots := home(t)
	plugins := filepath.Join(roots.Home, ".config", "opencode", "plugins", "terminal-pet")
	if err := os.MkdirAll(plugins, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"),
		`{"plugins": ["`+mustURL(t, plugins)+`"]}`)
	finding := InspectOpenCode(roots)
	if finding.Status != StatusFail {
		t.Fatalf("a plugin directory with no entrypoint must fail, got %s: %s", finding.Status, finding.Detail)
	}
}

func TestInspectClaudeCoverageAndBridges(t *testing.T) {
	roots := home(t)
	hook := filepath.Join(roots.LocalAppData, installDirectoryName, hookBinaryName)
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"), claudeSettings(8, `"`+hook+`" --mode=claude`))

	finding := InspectClaude(roots)
	if finding.Status != StatusOK {
		t.Fatalf("status = %s, want ok: %s", finding.Status, finding.Detail)
	}
	for _, event := range claudeHookEvents() {
		if !strings.Contains(finding.Detail, event) {
			t.Fatalf("detail does not report %s: %s", event, finding.Detail)
		}
	}

	// No SessionEnd: the host never sends the closed edge (handover §53.5).
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"), claudeSettings(7, `"`+hook+`" --mode=claude`))
	finding = InspectClaude(roots)
	if finding.Status != StatusWarn || !strings.Contains(finding.Detail, "SessionEnd") {
		t.Fatalf("a wiring without SessionEnd must warn and say why, got %s: %s", finding.Status, finding.Detail)
	}

	// The binary the command names is gone.
	if err := os.Remove(hook); err != nil {
		t.Fatal(err)
	}
	finding = InspectClaude(roots)
	if finding.Status != StatusFail || !strings.Contains(finding.Detail, "missing") {
		t.Fatalf("a dangling binary must fail, got %s: %s", finding.Status, finding.Detail)
	}
}

func TestInspectClaudeFollowsTheJSBridge(t *testing.T) {
	roots := home(t)
	// The installed wrapper on this machine loads pet-ipc.js out of the
	// install directory: a command that exists can still reach nothing
	// (handover §53.5).
	ipc := filepath.Join(roots.LocalAppData, installDirectoryName, "pet-ipc.js")
	writeFile(t, ipc, "module.exports = { reportStatus() {} }\n")
	wrapper := filepath.Join(roots.Home, ".claude", "hooks", "terminal-pet-hook.js")
	writeFile(t, wrapper, "const PET_IPC = '"+strings.ReplaceAll(ipc, `\`, `\\`)+"';\nconst ipc = require(PET_IPC)\n")
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"),
		claudeSettings(8, "node "+filepath.ToSlash(wrapper)))

	finding := InspectClaude(roots)
	if finding.Status != StatusOK {
		t.Fatalf("status = %s, want ok: %s", finding.Status, finding.Detail)
	}

	if err := os.Remove(ipc); err != nil {
		t.Fatal(err)
	}
	finding = InspectClaude(roots)
	if finding.Status != StatusFail || !strings.Contains(finding.Detail, "bridge") {
		t.Fatalf("a dead bridge must fail and name it, got %s: %s", finding.Status, finding.Detail)
	}
}

func TestInspectClaudeReportsMissingConfiguration(t *testing.T) {
	roots := home(t)
	if finding := InspectClaude(roots); finding.Status != StatusWarn || !strings.Contains(finding.Detail, "not installed") {
		t.Fatalf("got %s: %s", finding.Status, finding.Detail)
	}
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"), `{"env": {}}`)
	if finding := InspectClaude(roots); finding.Status != StatusWarn || !strings.Contains(finding.Detail, "not connected") {
		t.Fatalf("got %s: %s", finding.Status, finding.Detail)
	}
}

func TestInspectAntigravityLevelVersusEdge(t *testing.T) {
	roots := home(t)
	statusline := filepath.Join(roots.LocalAppData, installDirectoryName, statuslineBinaryName)
	writeFile(t, filepath.Join(roots.Home, ".gemini", "antigravity-cli", "settings.json"),
		`{"statusLine": {"command": "`+filepath.ToSlash(statusline)+`"}}`)

	finding := InspectAntigravity(roots)
	if finding.Status != StatusWarn {
		t.Fatalf("a statusline-only wiring must warn, got %s: %s", finding.Status, finding.Detail)
	}
	if !strings.Contains(finding.Detail, "idle") || !strings.Contains(finding.Detail, "terminationReason") {
		t.Fatalf("the warn has to explain the level/edge fact in one line: %s", finding.Detail)
	}

	// Add a hook plugin that can carry a termination reason: the edge exists
	// now, and the terminal is fully wired.
	writeFile(t, filepath.Join(roots.Home, ".gemini", "config", "plugins", "pet", "hooks.json"),
		`{"pet": {"Termination": [{"command": "pet-hook.exe --mode=antigravity"}]}}`)
	finding = InspectAntigravity(roots)
	if finding.Status != StatusOK {
		t.Fatalf("with an edge signal the wiring is complete, got %s: %s", finding.Status, finding.Detail)
	}

	if err := os.Remove(statusline); err != nil {
		t.Fatal(err)
	}
	finding = InspectAntigravity(roots)
	if finding.Status != StatusFail || !strings.Contains(finding.Detail, "missing") {
		t.Fatalf("a dangling statusline must fail, got %s: %s", finding.Status, finding.Detail)
	}
}

func TestInspectAntigravityNotInstalled(t *testing.T) {
	roots := home(t)
	if finding := InspectAntigravity(roots); finding.Status != StatusWarn || !strings.Contains(finding.Detail, "not installed") {
		t.Fatalf("got %s: %s", finding.Status, finding.Detail)
	}
}

func TestFindingsNeverEchoConfigurationContent(t *testing.T) {
	// A host configuration carries credentials. Every detail line is written
	// to be pasted into a public issue, so none of them may quote a value.
	const sentinel = "sk-live-DO-NOT-PRINT"
	roots := home(t)
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"),
		`{"apiKey": "`+sentinel+`", "plugins": ["file:///C:/other/plugin.js"]}`)
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"),
		`{"env": {"ANTHROPIC_AUTH_TOKEN": "`+sentinel+`"}, "hooks": {}}`)
	writeFile(t, filepath.Join(roots.Home, ".gemini", "antigravity-cli", "settings.json"),
		`{"apiKey": "`+sentinel+`", "statusLine": {"command": "/nowhere/pet-statusline.exe"}}`)

	report := Summary(Inspect(roots))
	if strings.Contains(report, sentinel) {
		t.Fatalf("a report line leaked configuration content: %s", report)
	}
	if strings.Contains(report, "\n") {
		t.Fatalf("the summary must stay on one line: %q", report)
	}
}

func TestWorstIsTheMostBrokenFinding(t *testing.T) {
	ok := Finding{Terminal: "a", Status: StatusOK}
	warn := Finding{Terminal: "b", Status: StatusWarn}
	fail := Finding{Terminal: "c", Status: StatusFail}
	if got := Worst([]Finding{ok, warn}); got != StatusWarn {
		t.Fatalf("Worst(ok,warn) = %s", got)
	}
	if got := Worst([]Finding{ok, fail, warn}); got != StatusFail {
		t.Fatalf("Worst(ok,fail,warn) = %s", got)
	}
	if got := Worst(nil); got != StatusOK {
		t.Fatalf("Worst(nil) = %s", got)
	}
	// An unknown status must never be reported as healthy.
	if got := (Status("weird")).String(); got != "fail" {
		t.Fatalf("unknown status stringified as %q", got)
	}
}

func TestSnippetsCarryResolvableTargets(t *testing.T) {
	roots := home(t)
	snippets := Snippets(roots)
	if len(snippets) != len(TerminalNames()) {
		t.Fatalf("got %d snippets, want %d", len(snippets), len(TerminalNames()))
	}

	openCode := snippetNamed(t, snippets, TerminalOpenCode)
	if !strings.Contains(openCode.Text, "file:///") || !strings.Contains(openCode.Text, "reporters/opencode.js") {
		t.Fatalf("the opencode snippet has to be the direct file form: %s", openCode.Text)
	}

	claude := snippetNamed(t, snippets, TerminalClaude)
	for _, event := range claudeHookEvents() {
		if !strings.Contains(claude.Text, event) {
			t.Fatalf("the claude snippet omits %s", event)
		}
	}
	if !strings.Contains(claude.Text, hookBinaryName) {
		t.Fatalf("the claude snippet should prefer the installed native hook: %s", claude.Text)
	}

	antigravity := snippetNamed(t, snippets, TerminalAntigravity)
	if !strings.Contains(antigravity.Text, "statusLine") || !strings.Contains(antigravity.Text, statuslineBinaryName) {
		t.Fatalf("the antigravity snippet has to wire the statusline: %s", antigravity.Text)
	}
	if !strings.Contains(antigravity.Note, "terminationReason") {
		t.Fatalf("the note has to state the level/edge fact: %s", antigravity.Note)
	}

	// Every block must be pasteable: the check is "it balances once wrapped
	// in braces", not "we reimplemented a parser".
	for _, snippet := range snippets {
		if strings.HasPrefix(snippet.Text, ";") {
			continue
		}
		if !jsonLooksBalanced("{" + snippet.Text + "}") {
			t.Fatalf("%s snippet is not balanced: %s", snippet.Terminal, snippet.Text)
		}
	}
}

func TestSnippetsSayWhatIsMissingInsteadOfPrintingADeadCommand(t *testing.T) {
	root := t.TempDir()
	roots := Roots{Home: filepath.Join(root, "profile"), LocalAppData: filepath.Join(root, "local")}
	if err := os.MkdirAll(roots.Home, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, snippet := range Snippets(roots) {
		if len(snippet.Missing) == 0 {
			t.Fatalf("%s reports nothing missing on a machine with no install", snippet.Terminal)
		}
		if strings.Contains(snippet.Text, hookPath(roots)) || strings.Contains(snippet.Text, statuslinePath(roots)) {
			t.Fatalf("%s printed a command for a file that does not exist: %s", snippet.Terminal, snippet.Text)
		}
	}
}

func TestOpenCodeSnippetRecognisesAnExistingWiring(t *testing.T) {
	// From an installed copy the repository is unreachable, so the honest
	// answer to --connect=opencode is "already wired" when the host config
	// resolves - not "nothing to paste", which reads as "you have no option".
	root := t.TempDir()
	roots := Roots{Home: filepath.Join(root, "profile"), LocalAppData: filepath.Join(root, "local")}
	reporter := filepath.Join(root, "repo", "reporters", "opencode.js")
	writeFile(t, reporter, "module.exports = {}\n")
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"),
		`{"plugins": ["`+mustURL(t, reporter)+`"]}`)

	snippet := snippetFor(roots, TerminalOpenCode)
	if len(snippet.Missing) != 0 {
		t.Fatalf("an already-wired terminal must not report anything missing: %v", snippet.Missing)
	}
	if !strings.Contains(snippet.Text, "already wired") || !strings.Contains(snippet.Text, "opencode.js") {
		t.Fatalf("the snippet has to say it is already wired and name the target: %s", snippet.Text)
	}
}

func TestSnippetsWriteNothing(t *testing.T) {
	roots := home(t)
	// Give every host a configuration worth clobbering.
	writeFile(t, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"), `{"plugins": []}`)
	writeFile(t, filepath.Join(roots.Home, ".claude", "settings.json"), `{"hooks": {}}`)
	writeFile(t, filepath.Join(roots.Home, ".gemini", "antigravity-cli", "settings.json"), `{"statusLine": {"command": "x"}}`)

	before := snapshotTree(t, ReadOnlyRoots(roots))
	for _, name := range TerminalNames() {
		snippet, ok := BuildOne(roots, name)
		if !ok {
			t.Fatalf("BuildOne(%q) said the terminal is unknown", name)
		}
		if snippet.Terminal != name {
			t.Fatalf("BuildOne(%q) returned %q", name, snippet.Terminal)
		}
	}
	_ = Render(Snippets(roots))
	after := snapshotTree(t, ReadOnlyRoots(roots))

	if len(before) != len(after) {
		t.Fatalf("file set changed while --connect was printing: %d -> %d entries", len(before), len(after))
	}
	for path, state := range before {
		if after[path] != state {
			t.Fatalf("%s changed while --connect was printing", path)
		}
	}
	if _, ok := BuildOne(roots, "codex"); ok {
		t.Fatal("an unknown terminal was accepted")
	}
}

func TestRenderIsStableAndAbbreviates(t *testing.T) {
	roots := home(t)
	text := Render(Snippets(roots))
	if !strings.Contains(text, "== "+TerminalOpenCode+" ==") {
		t.Fatalf("render does not label the first terminal: %s", text)
	}
	if strings.Contains(text, roots.Home) {
		t.Fatalf("render leaked the full profile path instead of ~: %s", text)
	}
	if again := Render(Snippets(roots)); again != text {
		// Map iteration anywhere in the render would show up here.
		t.Fatal("render is not deterministic")
	}
}

func TestImportedTargetsReadsBothDirections(t *testing.T) {
	source := "import { plugin } from \"file:///C:/a.js\"\n" +
		"export { other } from 'file:///C:/b.js'\n" +
		"const transform = 1\n" +
		"// from \"file:///C:/c.js\"\n" +
		"/* from \"file:///C:/d.js\" */\n" +
		"import(\"file:///C:/e.js\")\n"
	// The two paths in prose are comments, and a comment that names the very
	// file we are checking for is exactly how this machine's wrapper reads:
	// if those were taken as imports, deleting the real target would still
	// look healthy. The dynamic import is not a re-export either, so it is not
	// a target.
	want := []string{"file:///C:/a.js", "file:///C:/b.js"}
	if got := importedTargets(source); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("importedTargets = %v, want %v", got, want)
	}
	// And the thing that matters: the real file's own wrapper (this
	// repository's comment style) yields exactly one target.
	realistic := "// All of the reporter's logic lives in the repository:\n" +
		"//   reporters/opencode.js\n" +
		"import { plugin } from \"file:///C:/x/reporters/opencode.js\"\n\n" +
		"export default plugin\n"
	if got := importedTargets(realistic); len(got) != 1 || !strings.HasSuffix(got[0], "/reporters/opencode.js") {
		t.Fatalf("a realistic wrapper yielded %v", got)
	}
}

// helpers

func mustURL(t *testing.T, path string) string {
	t.Helper()
	value, err := PathToFileURL(path)
	if err != nil {
		t.Fatalf("PathToFileURL(%q): %v", path, err)
	}
	return value
}

// claudeSettings builds a settings.json with the first count events wired to one
// command.
func claudeSettings(count int, command string) string {
	events := claudeHookEvents()
	var builder strings.Builder
	builder.WriteString("{\"hooks\": {")
	for index := 0; index < count && index < len(events); index++ {
		if index > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("\"" + events[index] + "\": [{\"matcher\": \"*\", \"hooks\": [{\"type\": \"command\", \"command\": " + jsonString(command) + "}]}]")
	}
	builder.WriteString("}}")
	return builder.String()
}

func jsonString(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func snippetNamed(t *testing.T, snippets []Snippet, terminal string) Snippet {
	t.Helper()
	for _, snippet := range snippets {
		if snippet.Terminal == terminal {
			return snippet
		}
	}
	t.Fatalf("no snippet for %s", terminal)
	return Snippet{}
}

func hookPath(roots Roots) string {
	return installedHookBinary(roots)
}

func statuslinePath(roots Roots) string {
	return installedStatusline(roots)
}

func TestInstallDirResolvesDataDirAndXDG(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "custom-data")
	roots := Roots{
		Home:    filepath.Join(root, "profile"),
		DataDir: dataDir,
	}
	if got, want := installedHookBinary(roots), filepath.Join(dataDir, hookBinaryName); got != want {
		t.Fatalf("installedHookBinary(DataDir) = %q, want %q", got, want)
	}
	if got, want := installedStatusline(roots), filepath.Join(dataDir, statuslineBinaryName); got != want {
		t.Fatalf("installedStatusline(DataDir) = %q, want %q", got, want)
	}
	if empty := installedHookBinary(Roots{}); empty != "" {
		t.Fatalf("installedHookBinary(empty) = %q, want empty", empty)
	}
}

// stringContains is strings.Contains behind a name, used where the needle is a
// short filename that a reader would otherwise have to count characters to see.
func stringContains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// snapshotTree records every file under the given roots with its size and
// modification time, so two snapshots catch a modified file and a new one in the
// same assertion.
func snapshotTree(t *testing.T, directories []string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	for _, directory := range directories {
		for _, path := range listFiles(t, directory) {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			tree[path] = strconv.FormatInt(info.Size(), 10) + "|" + info.ModTime().UTC().String()
		}
	}
	return tree
}

func listFiles(t *testing.T, directory string) []string {
	t.Helper()
	paths := make([]string, 0, 8)
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return paths
}

// jsonLooksBalanced is a paste-safety check: braces and brackets balance, no
// string is left open, and no trailing comma sits before a close. It is
// deliberately not a parser (this package must not grow one for a diagnostic),
// but it catches the three ways a generated snippet breaks.
func jsonLooksBalanced(value string) bool {
	depth := 0
	inString := false
	for index := 0; index < len(value); index++ {
		character := value[index]
		if inString {
			if character == '\\' {
				index++
				continue
			}
			if character == '"' {
				inString = false
			}
			continue
		}
		switch character {
		case '"':
			inString = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth < 0 {
				return false
			}
		case ',':
			rest := strings.TrimLeft(value[index+1:], " \t\r\n")
			if rest != "" && (rest[0] == '}' || rest[0] == ']') {
				return false
			}
		}
	}
	return depth == 0 && !inString
}
