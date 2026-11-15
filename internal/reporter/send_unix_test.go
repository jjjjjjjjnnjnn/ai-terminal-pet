//go:build !windows

package reporter

// The unix case table for the retry classification, plus the probe that proves
// Send spends no time on a refusal it cannot fix.
//
// The errors are a different set from the Windows ones and the rule is different
// too: on Windows a busy pipe is worth retrying and a denied pipe is not; here a
// socket that nobody listens on reports ENOENT/ECONNREFUSED (transient, one launch
// fixes it) while a socket this process may not open reports EACCES (permanent).
// Writing this out keeps the two policies from being "unified" later by somebody
// who only read one of them - the mistake handover §55.5 warns about.
//
// Measured on this platform: net.Dial to a unix socket wraps the syscall error in
// *os.SyscallError under *net.OpError, so errors.Is is what finds it.

import (
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRetryableDialErrorClassification(t *testing.T) {
	cases := []struct {
		name  string
		err   error
		retry bool
	}{
		{"no socket yet", &fs.PathError{Op: "dial", Path: "/run/user/1000/x.sock", Err: syscall.ENOENT}, true},
		{"nothing listening", &fs.PathError{Op: "dial", Path: "/run/user/1000/x.sock", Err: syscall.ECONNREFUSED}, true},
		{"owned by another user", &fs.PathError{Op: "dial", Path: "/run/user/1000/x.sock", Err: syscall.EACCES}, false},
		{"nil", nil, false},
	}
	for _, testCase := range cases {
		if got := retryableDialError(testCase.err); got != testCase.retry {
			t.Errorf("%s: retryableDialError = %v, want %v", testCase.name, got, testCase.retry)
		}
	}
}

// A socket this process may not open returns immediately rather than burning the
// host's time: the same rule as the Windows DACL probe, checked through the file
// permissions this platform actually has instead of a named pipe's security
// descriptor.
//
// The socket is real, and unlink-on-close is turned off on purpose: a path that
// simply does not exist reports ENOENT (transient, one launch fixes it), while a
// file with no permissions reports EACCES. Only the second one tests this rule.
func TestSendStopsImmediatelyOnAPermissionRefusal(t *testing.T) {
	dataDirectory(t)
	path := filepath.Join(t.TempDir(), "ai-terminal-pet-v1-denied.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Skipf("cannot create the probe socket: %v", err)
	}
	// *net.UnixListener owns the file and removes it on Close; taking the concrete
	// type is what lets the probe leave the file behind to chmod it.
	unix, ok := listener.(*net.UnixListener)
	if !ok {
		t.Skip("this platform's listener is not a unix listener")
	}
	unix.SetUnlinkOnClose(false)
	_ = unix.Close()
	if err := os.Chmod(path, 0o000); err != nil {
		t.Skipf("cannot drop the socket's permissions: %v", err)
	}
	defer os.Chmod(path, 0o600)

	now := time.Now().UTC()
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "denied"
	payload.PID = 21
	payload = payload.Finish(now)

	started := time.Now()
	delivered := Send(payload, SendOptions{
		Retry:          750 * time.Millisecond,
		ConnectTimeout: 150 * time.Millisecond,
		PipePath:       path,
	})
	elapsed := time.Since(started)
	if delivered {
		t.Fatal("Send claimed success against a socket it cannot open")
	}
	// The budget is 750 ms; a permanent refusal has to return in a fraction of it.
	if elapsed > 300*time.Millisecond {
		t.Fatalf("Send spent %v on a refusal it cannot fix", elapsed)
	}
}
