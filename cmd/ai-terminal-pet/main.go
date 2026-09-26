// Command ai-terminal-pet is the always-on-top overlay that aggregates AI
// terminal session states.
//
// It runs two ways:
//
//   - With no arguments, as the overlay itself. It is normally started on
//     demand by a reporter (reporters/*.js, cmd/pet-statusline), and removes
//     itself once no session is left.
//   - With a control flag, as a short-lived client that talks to a running
//     overlay, reads local state, or prints a snapshot, then exits. These exist
//     so the pet can be driven from a script, a shell alias or a terminal
//     status line without knowing anything about named pipes or window handles.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"ai-terminal-pet/internal/app"
	"ai-terminal-pet/internal/autostart"
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
  ai-terminal-pet --json          print the current sessions as JSON
  ai-terminal-pet --autostart     report the run-at-logon state
  ai-terminal-pet --autostart=on  run the overlay at logon
  ai-terminal-pet --autostart=off stop running the overlay at logon
  ai-terminal-pet --version       print the version
  ai-terminal-pet --help          print this message

--json prints one line; add --pretty for indented output. It reads the
reporter snapshots on disk, so it works whether or not the overlay is running.

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
	command := args[0]
	switch {
	case command == "--help" || command == "-h" || command == "help":
		fmt.Print(usage)
		return nil
	case command == "--version" || command == "-v" || command == "version":
		fmt.Printf("ai-terminal-pet %s\n", app.Version)
		return nil
	case command == "--json":
		return printSnapshot(args[1:])
	case command == "--autostart" || strings.HasPrefix(command, "--autostart="):
		return controlAutoStart(command)
	case command == "--show":
		// Fall through to a normal start when nothing is running, so the
		// command is safe to bind to a hotkey.
		if ui.Control(ui.ControlActivate) {
			return nil
		}
		return app.Run()
	case command == "--toggle":
		if ui.Control(ui.ControlToggle) {
			return nil
		}
		return app.Run()
	case command == "--reset":
		if err := clearState(); err != nil {
			return err
		}
		if ui.Control(ui.ControlReset) {
			return nil
		}
		return app.Run()
	case command == "--quit":
		ui.Control(ui.ControlQuit)
		return nil
	default:
		return usageError("unknown argument %q", command)
	}
}

// printSnapshot renders the on-disk session view. Extra arguments are rejected
// rather than ignored: silently accepting a typo like --pretry would leave the
// caller staring at compact output wondering why.
func printSnapshot(extra []string) error {
	pretty := false
	for _, value := range extra {
		switch value {
		case "--pretty":
			pretty = true
		default:
			return usageError("unknown argument %q for --json", value)
		}
	}
	report, err := app.Snapshot()
	if err != nil {
		return err
	}
	encoded, err := app.EncodeReport(report, pretty)
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

// controlAutoStart reports or changes the run-at-logon entry. The registry is
// the source of truth, so a bare --autostart always re-reads it rather than
// consulting any cached flag.
func controlAutoStart(command string) error {
	manager := autostart.New()
	action := ""
	if strings.HasPrefix(command, "--autostart=") {
		action = strings.TrimPrefix(command, "--autostart=")
	}
	switch action {
	case "":
		status, err := manager.Status()
		if err != nil {
			return err
		}
		if !status.Enabled {
			fmt.Println("autostart: off")
			return nil
		}
		fmt.Printf("autostart: on -> %s\n", status.Target)
		return nil
	case "on", "off":
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		if action == "on" {
			if err := manager.Enable(executable); err != nil {
				return err
			}
			fmt.Printf("autostart: on -> %s\n", executable)
			return nil
		}
		if err := manager.Disable(); err != nil {
			return err
		}
		fmt.Println("autostart: off")
		return nil
	default:
		return usageError("--autostart takes on or off, got %q", action)
	}
}

// usageError prints to stderr and returns the sentinel main maps to exit 2.
func usageError(format string, args ...any) error {
	fmt.Fprintf(os.Stderr, "ai-terminal-pet: "+format+"\n\n"+usage, args...)
	return errUsage
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
