//go:build windows

package reporter

// The Windows transport for one delivery: a named pipe, plus the two facts Send
// needs to decide whether another attempt is worth the host's time.
//
// Both halves are in this file because they are about the same thing - what a
// failed dial means on Windows. "Access is denied" is permanent and specific to
// this platform (a process at a lower integrity level cannot open a pipe whose
// DACL grants only the owner and SYSTEM, handover §29.4); a missing pipe is
// transient, because one launch and one retry can fix it. Unix has no integrity-label
// equivalent, so its file answers these questions differently - see dial_unix.go.

import (
	"errors"
	"io/fs"
	"net"
	"syscall"
	"time"

	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

// dial opens one connection to the per-user pipe.
func dial(path string, timeout time.Duration) (net.Conn, error) {
	deadline := timeout
	return winio.DialPipe(path, &deadline)
}

// retryableDialError reports whether another attempt could plausibly succeed.
//
// The distinction is not cosmetic. A hook runs once per host event and the host
// waits for it, so when delivery is impossible the retry budget is pure added
// latency. "No listener yet" is transient: one launch and one retry can fix it.
// "Access is denied" is not: it means this token will never open that pipe, and
// every remaining attempt would fail in microseconds while the reporter still paid
// the whole budget (measured: 829-862 ms down to 89-119 ms, handover §29.3).
func retryableDialError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, windows.ERROR_PIPE_BUSY) {
		return true
	}
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ERROR_FILE_NOT_FOUND)
}

// detachProcess marks a child as "do not wait for me": the overlay is launched
// detached so a reporter can exit while the pet is still starting.
func detachProcess() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
		// DETACHED_PROCESS (0x00000008); the literal avoids a dependency
		// for a single constant.
		CreationFlags: 0x00000008,
	}
}
