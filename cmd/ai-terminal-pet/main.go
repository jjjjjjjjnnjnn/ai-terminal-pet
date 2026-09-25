// Command ai-terminal-pet is the always-on-top overlay that aggregates AI
// terminal session states.
//
// It runs two ways:
//
//   - With no arguments, as the overlay itself. It is normally started on
//     demand by a reporter (reporters/*.js, cmd/pet-statusline), and removes
//     itself once no session is left.
//   - With a control flag, as a short-lived client that talks to a running
//     overlay and exits. These exist so the pet can be driven from a script,
//     a shell alias or a terminal status line without knowing anything about
//     named pipes or window handles.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"ai-terminal-pet/internal/app"
	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/ui"
)

const usage = `ai-terminal-pet - floating status overlay for AI coding terminals

usage:
  ai-terminal-pet                 start the overlay (no-op if already running)
  ai-terminal-pet --show          reveal the running overlay
  ai-terminal-pet --toggle        switch between compact and expanded
  ai-terminal-pet --reset         re-anchor to the screen edge at 100% zoom
  ai-terminal-pet --quit          stop the running overlay
  ai-terminal-pet --version       print the version
  ai-terminal-pet --help          print this message

The overlay also exits on its own once no terminal session is reporting.
`

// errUsage marks a bad invocation. A typo at a shell prompt deserves a message
// on stderr and exit code 2, not a silent append to error.log - the caller is
// a human who is watching, and error.log is for things that went wrong on
// their own.
var errUsage = errors.New("bad usage")

func main() {
	debug.SetGCPercent(50)
	os.Exit(entry(os.Args[1:]))
}

func entry(args []string) int {
	err := run(args)
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errUsage):
		return 2
	default:
		writeError(err)
		return 1
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return app.Run()
	}
	switch args[0] {
	case "--help", "-h", "help":
		fmt.Print(usage)
		return nil
	case "--version", "-v", "version":
		fmt.Printf("ai-terminal-pet %s\n", app.Version)
		return nil
	case "--show":
		// Fall through to a normal start when nothing is running, so the
		// command is safe to bind to a hotkey.
		if ui.Control(ui.ControlActivate) {
			return nil
		}
		return app.Run()
	case "--toggle":
		if ui.Control(ui.ControlToggle) {
			return nil
		}
		return app.Run()
	case "--reset":
		if err := clearState(); err != nil {
			return err
		}
		if ui.Control(ui.ControlReset) {
			return nil
		}
		return app.Run()
	case "--quit":
		ui.Control(ui.ControlQuit)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "ai-terminal-pet: unknown argument %q\n\n%s", args[0], usage)
		return errUsage
	}
}

// clearState removes the remembered window placement. The running instance
// persists a fresh anchor straight after wmReset, so deleting first only
// matters for the "nothing was running" path.
func clearState() error {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return errors.New("LOCALAPPDATA is not set")
	}
	path := filepath.Join(root, "AI-Terminal-Pet", config.StateFileName)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func writeError(value error) {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" || value == nil {
		return
	}
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return
	}
	path := filepath.Join(directory, "error.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	if info, statErr := file.Stat(); statErr == nil && info.Size() > 1<<20 {
		_ = file.Close()
		_ = os.Remove(path)
		file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return
		}
		defer file.Close()
	}
	_, _ = file.WriteString(time.Now().Format(time.RFC3339) + " " + value.Error() + "\n")
}
