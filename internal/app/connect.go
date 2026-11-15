package app

// --connect prints the paste-ready wiring for each host. It prints and nothing
// else.

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

func connectFooter() string {
	return "\nnext: paste the block above into the file named on the first line, then run\n" +
		"      ai-terminal-pet --doctor and look at the reporters line.\n" +
		"this command has written nothing to disk.\n"
}
