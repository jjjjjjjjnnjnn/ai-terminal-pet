package reporter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/session"
	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

const (
	// PipeDialTimeout mirrors protocol.js connectTimeoutMs' default: the
	// overlay is local and either listening or not, so there is nothing to
	// wait for beyond a scheduler hiccup.
	PipeDialTimeout = 150 * time.Millisecond
	// PipeIOTimeout bounds the write and the acknowledgement read.
	PipeIOTimeout = 150 * time.Millisecond
	// DefaultRetryBudget is what the reporters pass as retryMs.
	DefaultRetryBudget = 750 * time.Millisecond
	// DefaultConnectTimeout caps one attempt inside the retry loop.
	DefaultConnectTimeout = 150 * time.Millisecond
)

// SendOptions tunes one delivery. The zero value is the behaviour every
// reporter script has: retry for DefaultRetryBudget, launch the overlay once
// when the pipe is missing, never launch for a "closed" report.
type SendOptions struct {
	// Retry is the total budget, not a per-attempt timeout.
	Retry time.Duration
	// ConnectTimeout caps a single dial.
	ConnectTimeout time.Duration
	// Launch allows starting AI-Terminal-Pet.exe when the pipe is absent.
	Launch bool
	// PipePath overrides the per-user pipe (tests).
	PipePath string
}

// SendOnce dials the pipe, writes one JSON line and requires the exact "OK"
// acknowledgement. Any other outcome - including a partial read - counts as
// failure, mirroring sendOnce in protocol.js: only "OK" means the store took
// the report.
func SendOnce(pipePath string, line []byte) bool {
	ok, _ := sendOnce(pipePath, line)
	return ok
}

// sendOnce is SendOnce plus the dial error, so Send can tell "the overlay is
// not listening yet" (worth a launch and a retry) from "this process will never
// be allowed in" (retrying only burns the reporter's budget).
func sendOnce(pipePath string, line []byte) (bool, error) {
	timeout := PipeDialTimeout
	conn, err := winio.DialPipe(pipePath, &timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(PipeIOTimeout))
	if _, err := conn.Write(append(append([]byte{}, line...), '\n')); err != nil {
		return false, nil
	}
	reply := make([]byte, 32)
	count, err := conn.Read(reply)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, nil
	}
	if string(bytes.TrimSpace(reply[:count])) != "OK" {
		return false, nil
	}
	return true, nil
}

// retryableDialError reports whether another attempt could plausibly succeed.
//
// The distinction is not cosmetic. A hook runs once per host event and the host
// waits for it, so when delivery is impossible the retry budget is pure added
// latency. "No listener yet" is transient: one launch and one retry can fix it.
// "Access is denied" is not: it means this token will never open that pipe -
// typically a process at a lower mandatory integrity level than the one that
// created the pipe, which happens when the executable itself lives in a
// low-integrity-labelled directory. Every remaining attempt would fail in
// microseconds and the reporter would still pay the whole budget.
func retryableDialError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, windows.ERROR_PIPE_BUSY) {
		return true
	}
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ERROR_FILE_NOT_FOUND)
}

// Send writes the runtime snapshot and then delivers the payload, retrying
// until the budget is spent. It returns true only when the server
// acknowledged the line.
//
// The order matters and is the same as sendStatus in protocol.js: the
// snapshot lands first, so even an overlay that has to be launched reads this
// session from disk on its next scan. That is what keeps a report filed during
// startup from being lost.
func Send(p Payload, options SendOptions) bool {
	if options.Retry <= 0 {
		options.Retry = DefaultRetryBudget
	}
	if options.ConnectTimeout <= 0 {
		options.ConnectTimeout = DefaultConnectTimeout
	}
	line, err := p.MarshalJSON()
	if err != nil {
		return false
	}
	if err := p.Validate(time.Now()); err != nil {
		// The server would reject this payload, so filing a snapshot would
		// only leave a file the scanner keeps re-reading and discarding.
		return false
	}
	WriteSnapshot(p)

	pipePath := options.PipePath
	if pipePath == "" {
		pipePath = pipeserver.PipePathForUser()
	}
	launchAllowed := options.Launch && p.State != session.StateClosed
	launchAttempted := false
	deadline := time.Now().Add(options.Retry)
	for {
		ok, dialErr := sendOnce(pipePath, line)
		if ok {
			return true
		}
		if !launchAttempted && launchAllowed && retryableDialError(dialErr) {
			launchAttempted = true
			LaunchApplication()
		}
		if dialErr != nil && !retryableDialError(dialErr) {
			// Denied (or otherwise permanent). Retrying cannot change the
			// answer, and the host is waiting on this process.
			return false
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false
		}
		wait := DefaultConnectTimeout / 2
		if remaining < wait {
			wait = remaining
		}
		time.Sleep(wait)
	}
}

// ApplicationPath is the installed overlay this user should launch, or "" when
// there is nothing to launch. AI_TERMINAL_PET_EXE is honoured only when it
// resolves inside the install directory, mirroring protocol.js: a reporter
// that would run an arbitrary path from the environment is a way to turn a
// config typo into code execution.
func ApplicationPath() string {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return ""
	}
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if override := os.Getenv("AI_TERMINAL_PET_EXE"); override != "" {
		if os.Getenv("AI_TERMINAL_PET_ALLOW_UNSAFE_EXE") == "1" {
			return override
		}
		absolute, err := filepath.Abs(override)
		if err != nil {
			return ""
		}
		relative, err := filepath.Rel(directory, absolute)
		if err != nil || relative == ".." || len(relative) > 2 && relative[:3] == ".."+string(filepath.Separator) {
			return ""
		}
		return absolute
	}
	return filepath.Join(directory, "AI-Terminal-Pet.exe")
}

// LaunchApplication starts the overlay detached, once, and ignores every
// error: a reporter must never fail its host because the pet could not start.
func LaunchApplication() bool {
	executable := ApplicationPath()
	if executable == "" {
		return false
	}
	info, err := os.Stat(executable)
	if err != nil || info.IsDir() {
		return false
	}
	cmd := exec.Command(executable)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		// DETACHED_PROCESS (0x00000008); the literal avoids a dependency
		// for a single constant.
		CreationFlags: 0x00000008,
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	return cmd.Start() == nil
}

// DecodeStdin parses the hook payload the host writes to stdin. A malformed or
// oversized body yields an empty map rather than an error, because every
// reporter path treats "no payload" as "report idle" and must not exit
// non-zero into the host CLI.
func DecodeStdin(raw []byte) map[string]any {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return map[string]any{}
	}
	var parsed any
	if err := json.Unmarshal(trimmed, &parsed); err != nil {
		return map[string]any{}
	}
	if object, ok := parsed.(map[string]any); ok {
		return object
	}
	return map[string]any{}
}
