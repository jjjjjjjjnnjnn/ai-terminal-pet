package ui

// Durable record of a window-procedure fault.
//
// The overlay is started by a reporter with stdio ignored
// (reporters/protocol.js: spawn(..., { stdio: "ignore" })). A runtime panic's
// stderr output - the panic value and the whole traceback - therefore goes
// nowhere, and what remains on the machine is a panel that stopped responding
// plus a healthy-looking install: --doctor has no error.log to read, and the
// only Windows evidence is an AppHangB1 event with no dump (handover §79).
//
// error.log is where the command line already records faults it cannot show
// (cmd/ai-terminal-pet/main.go writeError), so a callback fault belongs there
// too. The write is deliberately best-effort like exit.log: a machine whose log
// cannot be written is still worth keeping alive.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ai-terminal-pet/internal/home"
)

// faultLogMaxBytes matches writeError's rotation point. One fault plus a
// truncated traceback is a few kilobytes, so this keeps many of them.
const faultLogMaxBytes = 1 << 20

// writeFault appends one timestamped record to error.log. It never returns an
// error: it runs from a deferred recover in the window procedure, where a
// second failure must not become a second panic.
func writeFault(kind, detail string) {
	path := home.Join("error.log")
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	if info, err := os.Stat(path); err == nil && info.Size() > faultLogMaxBytes {
		_ = os.Remove(path)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "%s %s: %s\n", time.Now().Format(time.RFC3339), kind, detail)
}
