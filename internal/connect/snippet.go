package connect

// What --connect prints.
//
// The rule this file exists to keep: it writes nothing. Every snippet is a
// block of text the user pastes themselves, into a file they own, after reading
// it. A tool that edits another product's configuration on the user's behalf
// has to guess that product's schema, and this project's whole reporter story
// is built on not guessing (AGENTS rule 1: never fail into the host CLI). The
// measured failure this replaces (handover §52.3) was a *stale* reference, and
// no snippet writer can see staleness -- which is why --doctor reads and
// --connect only prints.
//
// Both halves of what a user needs are here: the block to paste, and the path
// it goes into. A snippet that only carried the block would leave the same
// question the README's connect table leaves ("which file, exactly?").

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	pethome "ai-terminal-pet/internal/home"
)

// Snippet is one terminal's paste-ready answer.
type Snippet struct {
	// Terminal is the --connect value that produced it.
	Terminal string
	// Where is the configuration file the block belongs in, already resolved
	// and abbreviated the same way a --doctor line abbreviates it.
	Where string
	// Text is the exact block to paste. Always ASCII-configurable JSON or TS,
	// never a diff.
	Text string
	// Note carries what the block cannot: an ordering requirement, a caveat,
	// or the one reason a host will still look quiet after pasting.
	Note string
	// Missing lists the reporter artifacts this terminal needs that are not on
	// disk. Non-empty means "install first, then paste".
	Missing []string
}

// Snippets builds every terminal's answer, in TerminalNames() order.
func Snippets(roots Roots) []Snippet {
	snippets := make([]Snippet, 0, len(TerminalNames()))
	for _, name := range TerminalNames() {
		snippets = append(snippets, snippetFor(roots, name))
	}
	return snippets
}

// BuildOne builds one terminal's answer. ok is false for a name that is not one
// of ours, which the caller turns into a usage error rather than a silent
// empty output.
func BuildOne(roots Roots, terminal string) (Snippet, bool) {
	for _, name := range TerminalNames() {
		if strings.EqualFold(name, terminal) {
			return snippetFor(roots, name), true
		}
	}
	return Snippet{}, false
}

func snippetFor(roots Roots, terminal string) Snippet {
	switch terminal {
	case TerminalOpenCode:
		return openCodeSnippet(roots)
	case TerminalClaude:
		return claudeSnippet(roots)
	default:
		return antigravitySnippet(roots)
	}
}

// installedHookBinary and installedStatusline resolve the copies every reporter
// is meant to run from. They are preferred over a repository checkout because
// that is the set --doctor's install-files check verifies and the set the plugin
// package bootstraps.
func installDir(roots Roots) string {
	if roots.DataDir != "" {
		return roots.DataDir
	}
	if roots.LocalAppData != "" {
		return filepath.Join(roots.LocalAppData, installDirectoryName)
	}
	return pethome.DirectoryFor(runtime.GOOS, "", roots.XDGDataHome, roots.Home)
}

func installedHookBinary(roots Roots) string {
	dir := installDir(roots)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, hookBinaryName)
}

func installedStatusline(roots Roots) string {
	dir := installDir(roots)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, statuslineBinaryName)
}

// HookBinaryName and StatuslineBinaryName expose the two deployable names so a
// test fixture (or a future doctor check in another package) can name the same
// artifacts without a second literal that could drift.
const (
	HookBinaryName       = hookBinaryName
	StatuslineBinaryName = statuslineBinaryName
)

// repositoryRoot is where reporters/ lives, derived from the running executable
// only when that derivation can be verified: an exe under <repo>/bin/ means
// <repo>, and <repo> must actually contain the reporter file we are about to
// name. Anything else returns "" and the caller falls back to the install
// directory.
func repositoryRoot(roots Roots) string {
	if roots.ExecutableErr != nil || roots.Executable == "" {
		return ""
	}
	candidate := filepath.Dir(filepath.Dir(roots.Executable))
	if exists(filepath.Join(candidate, "reporters", "opencode.js")) {
		return candidate
	}
	return ""
}

// openCodeSnippet wires OpenCode through the repository checkout.
//
// One hop, not two: the README's documented form is a direct file URL, and the
// directory re-export on this machine exists only to keep the event table in one
// place (handover §53.1). Printing the direct form means the snippet has one
// target to go stale instead of two.
func openCodeSnippet(roots Roots) Snippet {
	snippet := Snippet{Terminal: TerminalOpenCode}
	snippet.Where = abbreviate(roots, filepath.Join(roots.Home, ".config", "opencode", "opencode.json"))
	root := repositoryRoot(roots)
	if root == "" {
		// Running from the installed copy, where the repository is not
		// reachable. If OpenCode is already wired, saying so is far more
		// useful than "nothing to paste": the user's next question is "is it
		// already done?", and this is the one command that can answer it.
		if finding := InspectOpenCode(roots); finding.Status == StatusOK {
			snippet.Text = "; already wired - " + finding.Detail
			snippet.Note = "Nothing to paste: OpenCode already loads a reporter. " +
				"Run `ai-terminal-pet --doctor` to see the reporters line, or paste this block " +
				"yourself if you would rather point it at a repository checkout."
			return snippet
		}
		snippet.Missing = []string{"reporters/opencode.js (a repository checkout that contains it)"}
		snippet.Text = "; nothing to paste yet - run this from the repository, or install first"
		snippet.Note = "OpenCode loads plugins from a file URL, so the snippet has to name a file that exists. " +
			"From an installed copy the repository is not reachable: clone the repository, then run " +
			"`ai-terminal-pet --connect=opencode` from its bin/ directory."
		return snippet
	}
	url, err := PathToFileURL(filepath.Join(root, "reporters", "opencode.js"))
	if err != nil {
		snippet.Missing = []string{"reporters/opencode.js (could not be encoded as a URL: " + err.Error() + ")"}
		snippet.Text = "; nothing to paste yet"
		return snippet
	}
	var builder strings.Builder
	builder.WriteString("\"plugins\": [\n")
	fmt.Fprintf(&builder, "  \"%s\"\n", url)
	builder.WriteString("]")
	snippet.Text = builder.String()
	snippet.Note = "Add this entry to the existing plugins array (keep the entries already there). " +
		"OpenCode resolves the entry to a directory's index.ts or to a .js file; this snippet is the direct file form."
	return snippet
}

// claudeSnippet wires Claude Code through the native hook.
//
// Native first: pet-hook.exe measures 51-73 ms steady state against the JS
// reporter's 93-157 ms on the machine these numbers were taken on (handover
// §29.6), and a hook runs on every event. The JS form stays the fallback for a
// machine where the installer has not run.
func claudeSnippet(roots Roots) Snippet {
	snippet := Snippet{Terminal: TerminalClaude}
	snippet.Where = abbreviate(roots, filepath.Join(roots.Home, ".claude", "settings.json"))

	binary := installedHookBinary(roots)
	command := ""
	if exists(binary) {
		command = `"` + binary + `" --mode=claude`
	} else if root := repositoryRoot(roots); root != "" && exists(filepath.Join(root, "reporters", "claude-hook.js")) {
		// The JS reporter takes the event name as its one argument, so each
		// hook entry needs its own line. A missing pet-hook.exe is a warning,
		// not a failure to print anything.
		command = "node " + filepath.Join(root, "reporters", "claude-hook.js")
		snippet.Missing = []string{abbreviate(roots, binary) + " (the installer deploys it; the node fallback is shown instead)"}
	} else {
		snippet.Missing = []string{abbreviate(roots, binary), "reporters/claude-hook.js (a repository checkout that contains it)"}
		snippet.Text = "; nothing to paste yet - run `ai-terminal-pet --doctor` to see what is missing"
		snippet.Note = "Every hook command needs a reporter that exists. Install with " + installHint + ", then re-run --connect=claude."
		return snippet
	}

	events := claudeHookEvents()
	var builder strings.Builder
	builder.WriteString("\"hooks\": {\n")
	for index, event := range events {
		// The event name is the reporter's one argument: pet-hook takes its
		// dialect from --mode and its event from stdin, while the JS fallback
		// reads the event as argv[2]. Both forms are covered by the command
		// built above.
		one := command
		if strings.HasSuffix(one, "claude-hook.js") {
			one += " " + event
		}
		fmt.Fprintf(&builder, "  %s: [%s]", jsonQuote(event), claudeHookGroup(one))
		if index < len(events)-1 {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("}")
	snippet.Text = builder.String()
	snippet.Note = "Merge this into the existing hooks object, keeping the entries already there " +
		"(this machine also runs a notifier on PreToolUse and Stop). SessionEnd is what produces the closed state; " +
		"without it a finished session stays at idle."
	return snippet
}

// claudeHookGroup renders one matcher group for one command. One line per event
// keeps the block short enough to compare by eye against the file it is pasted
// into, which matters because the realistic merge is "keep the notifier that is
// already there" (handover §53.5 records a machine that runs both).
func claudeHookGroup(command string) string {
	return `{"matcher": "*", "hooks": [{"type": "command", "command": ` + jsonQuote(command) + `}]}`
}

// jsonQuote escapes a Windows path or command into a JSON string. Backslashes
// have to be doubled, otherwise the pasted block is not JSON at all -- the first
// version of this snippet printed the raw path and produced
// "command": ""C:\Users\..."".
func jsonQuote(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// ClaudeHookEvents exposes the documented event set so a presentation test can
// assert that the printed block covers all of it without keeping a second copy
// of the list.
func ClaudeHookEvents() []string {
	return claudeHookEvents()
}

// claudeHookEvents is the set this repository documents and has measured
// (handover §53.5, README "Connect your terminal"): the eight events that carry
// a state transition, an end signal, or a permission prompt.
func claudeHookEvents() []string {
	return []string{
		"SessionStart", "UserPromptSubmit", "PreToolUse", "PostToolUse",
		"PermissionRequest", "Notification", "Stop", "SessionEnd",
	}
}

// antigravitySnippet wires Antigravity's level reporter and states the fact
// that a snippet cannot fix: without a termination hook the session ends at
// idle, never closed (handover §53.5).
func antigravitySnippet(roots Roots) Snippet {
	snippet := Snippet{Terminal: TerminalAntigravity}
	snippet.Where = abbreviate(roots, filepath.Join(roots.Home, ".gemini", "antigravity-cli", "settings.json"))

	binary := installedStatusline(roots)
	if !exists(binary) {
		snippet.Missing = []string{abbreviate(roots, binary)}
		snippet.Text = "; nothing to paste yet - run `ai-terminal-pet --doctor` to see what is missing"
		snippet.Note = "The statusline command has to name an executable that exists; install with " + installHint + " first."
		return snippet
	}
	snippet.Text = "\"statusLine\": {\n  \"command\": " + jsonQuote(binary) + "\n}"
	snippet.Note = "This is a level reporter: every render reports the current state, so a finished or terminated " +
		"agent arrives as idle. Only a hook carrying terminationReason produces closed. Paste " +
		"`" + hookBinaryName + " --mode=antigravity` into a plugin's hooks.json if you need sessions to close."
	return snippet
}

// Render turns the snippets into the text --connect prints. Plain text, one
// block per terminal, because the user's next action is copy-paste into another
// program -- not reading a report.
func Render(snippets []Snippet) string {
	var builder strings.Builder
	for index, snippet := range snippets {
		if index > 0 {
			builder.WriteString("\n")
		}
		fmt.Fprintf(&builder, "== %s ==\n", snippet.Terminal)
		fmt.Fprintf(&builder, "where: %s\n", snippet.Where)
		if len(snippet.Missing) > 0 {
			fmt.Fprintf(&builder, "missing: %s\n", strings.Join(snippet.Missing, "; "))
		}
		builder.WriteString(snippet.Text)
		builder.WriteString("\n")
		if snippet.Note != "" {
			for _, line := range wrapNote(snippet.Note, 78) {
				builder.WriteString("note: " + line + "\n")
			}
		}
	}
	return builder.String()
}

// wrapNote keeps a note readable in a console that is 80 columns wide. Words
// are never split; a note longer than the width simply takes more lines.
func wrapNote(note string, width int) []string {
	words := strings.Fields(note)
	lines := make([]string, 0, 3)
	current := ""
	for _, word := range words {
		switch {
		case current == "":
			current = word
		case len(current)+1+len(word) <= width:
			current += " " + word
		default:
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

// ReadOnlyRoots is what a test uses to prove --connect wrote nothing: the
// directories are hashed before and after and compared. The list lives here
// rather than in a test file so that the "prints only" rule has one
// implementation: a future terminal has to be added to it, in the same place
// its snippet is written.
func ReadOnlyRoots(roots Roots) []string {
	return []string{
		filepath.Join(roots.Home, ".config", "opencode"),
		filepath.Join(roots.Home, ".claude"),
		filepath.Join(roots.Home, ".gemini"),
	}
}
