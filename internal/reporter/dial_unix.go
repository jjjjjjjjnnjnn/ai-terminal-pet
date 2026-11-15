//go:build !windows

package reporter

// The unix transport for one delivery.
//
// Two platform facts decide this file, and both differ from Windows:
//
//  1. Which failures are worth retrying. On Windows the expensive one was
//     ERROR_ACCESS_DENIED (mandatory integrity levels, handover 29.4). Here the
//     equivalent is EACCES on a socket owned by another user - a misconfiguration
//     that no retry can fix - while "the file is not there" and "nothing is
//     listening on it" are both transient, and one launch fixes them.
//  2. There is no DETACHED_PROCESS. A child that must outlive its parent gets its
//     own process group instead, so a Ctrl-C in the terminal that ran the reporter
//     does not reach the overlay.

import (
	"errors"
	"io/fs"
	"net"
	"syscall"
	"time"
)

// dial opens one connection to the per-user socket.
func dial(path string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", path, timeout)
}

// retryableDialError reports whether another attempt could plausibly succeed.
//
// The rule is narrower than the Windows one and that is deliberate: a permission
// error means no future attempt can work, so the retry budget goes back to the
// host instead of being spent in microseconds. Everything else is treated as
// transient, because a reporter must prefer one wasted attempt to a lost report.
//
// The comparison is against syscall errors rather than message text: matching
// strings would make the retry policy depend on the language the OS reports
// failures in. Measured on this platform: a socket with no listener reports
// ENOENT through *net.OpError/*os.SyscallError, and a socket owned by another user
// reports EACCES through the same two wrappers.
func retryableDialError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, fs.ErrPermission) || errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) {
		return false
	}
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOENT) ||
		errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET)
}

// detachProcess marks a child as "do not wait for me".
//
// Setpgid is the POSIX equivalent of the Windows creation flag: the overlay lands
// in its own process group, and the parent drops the handle as soon as Start
// returns (the caller already passes nil for all three stdio descriptors).
func detachProcess() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
