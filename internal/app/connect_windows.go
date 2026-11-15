package app

// --connect prints the paste-ready wiring for each host. It prints and nothing
// else.
//
// The split with --doctor is deliberate and is the whole reason both exist:
//
//	--doctor    reads the host configurations and says what is wired and what
//	            is broken (internal/connect.Inspect)
//	--connect   prints the block to paste, plus the file it goes into
//	            (internal/connect.Snippets)
//
// One reader, two outputs, so a snippet can never describe a form the check
// would then call missing. What --connect must never do is write: editing
// another product's configuration means guessing its schema, and this project
// will not guess (AGENTS rule 1 is the same rule, stated about the moment a
// reporter runs).

import (
	"fmt"
	"strings"

	"ai-terminal-pet/internal/connect"
)

// ConnectText returns the text --connect prints. An unknown terminal is an
// error the caller turns into a usage error, not an empty answer: silence
// would read as "you have no terminals to wire".
func ConnectText(terminal string) (string, error) {
	return connectText(connect.DefaultRoots(), terminal)
}

// connectText is the seam the tests use: the machine's real profile is replaced
// by a temp directory, so a test can assert the printed text without reading
// anything the developer happens to have configured.
func connectText(roots connect.Roots, terminal string) (string, error) {
	if terminal == "" {
		return connect.Render(connect.Snippets(roots)) + connectFooter(), nil
	}
	snippet, ok := connect.BuildOne(roots, terminal)
	if !ok {
		return "", fmt.Errorf("unknown terminal %q (want %s)", terminal, strings.Join(connect.TerminalNames(), ", "))
	}
	return connect.Render([]connect.Snippet{snippet}) + connectFooter(), nil
}

// connectFooter closes the loop back to --doctor: the snippet says what to
// paste, the check says whether it took. Without it, a paste that silently
// failed is still a silent failure -- which is the defect this round exists to
// remove.
func connectFooter() string {
	return "\nnext: paste the block above into the file named on the first line, then run\n" +
		"      ai-terminal-pet --doctor and look at the reporters line.\n" +
		"this command has written nothing to disk.\n"
}
