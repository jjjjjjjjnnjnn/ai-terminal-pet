//go:build petdebug

package ui

// The diagnostic build. One line per decision in exit.log beside config.json,
// plus one line per settings-window open/hide, so "the window went away" and
// "the process went away" can be told apart without asking the user what they
// remember seeing.
//
// PET_IDLE_GRACE_SEC shortens the idle window. It exists only here on purpose:
// a shipping build that honoured it would let an environment variable make the
// overlay quit before the ten minutes the documentation promises.
//
// PET_EXIT_TRACE=0 turns the log off even in this build, so a sandbox can run
// the debug binary without writing anything.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	exitTraceMaxBytes = 64 << 10
	exitGraceEnvVar   = "PET_IDLE_GRACE_SEC"
	exitTraceEnvVar   = "PET_EXIT_TRACE"
)

func init() {
	if seconds, err := strconv.Atoi(os.Getenv(exitGraceEnvVar)); err == nil && seconds > 0 && seconds <= 600 {
		idleGrace = time.Duration(seconds) * time.Second
	}
}

func traceExit(reason string) {
	appendExitTrace("exit", reason)
}

func traceSettings(reason string) {
	appendExitTrace("settings", reason)
}

// appendExitTrace is deliberately best-effort: a repro machine whose log
// cannot be written is still worth observing, and a failed write must never
// become an error on the overlay's exit path.
func appendExitTrace(kind, reason string) {
	if os.Getenv(exitTraceEnvVar) == "0" {
		return
	}
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return
	}
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return
	}
	target := filepath.Join(directory, "exit.log")
	if info, err := os.Stat(target); err == nil && info.Size() > exitTraceMaxBytes {
		_ = os.Remove(target)
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "%s %s %s pid=%d\n", time.Now().Format(time.RFC3339Nano), kind, reason, os.Getpid())
}
