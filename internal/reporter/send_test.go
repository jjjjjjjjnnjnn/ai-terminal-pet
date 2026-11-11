package reporter

import (
	"io/fs"
	"os"
	"testing"
	"time"

	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

// The retry budget is the reporter's entire cost when delivery is impossible,
// because the host waits for this process. Getting the classification wrong in
// either direction is a real bug: treating "denied" as retryable adds the whole
// budget to every single host event, and treating "no listener yet" as final
// throws away the report the launch was supposed to enable.
func TestRetryableDialErrorClassification(t *testing.T) {
	cases := []struct {
		name  string
		err   error
		retry bool
	}{
		{"no listener yet", &fs.PathError{Op: "open", Path: `\\.\pipe\x`, Err: windows.ERROR_FILE_NOT_FOUND}, true},
		{"every instance busy", &fs.PathError{Op: "open", Path: `\\.\pipe\x`, Err: windows.ERROR_PIPE_BUSY}, true},
		{"access denied", &fs.PathError{Op: "open", Path: `\\.\pipe\x`, Err: windows.ERROR_ACCESS_DENIED}, false},
		{"broken pipe", &fs.PathError{Op: "open", Path: `\\.\pipe\x`, Err: windows.ERROR_BROKEN_PIPE}, false},
		{"nil", nil, false},
	}
	for _, testCase := range cases {
		if got := retryableDialError(testCase.err); got != testCase.retry {
			t.Errorf("%s: retryableDialError = %v, want %v", testCase.name, got, testCase.retry)
		}
	}
}

// A pipe whose DACL grants only SYSTEM is the cheap local stand-in for the
// real-world case this guards: a reporter whose token can never open that
// pipe. Send must come back immediately instead of burning the retry budget.
func TestSendStopsImmediatelyWhenThePipeRefusesTheCaller(t *testing.T) {
	if testing.Short() {
		t.Skip("creates a real named pipe")
	}
	path := `\\.\pipe\ai-terminal-pet-test-denied-` + itoa(os.Getpid())
	listener, err := winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GRGW;;;SY)",
	})
	if err != nil {
		t.Skipf("cannot create the probe pipe: %v", err)
	}
	defer listener.Close()
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr == nil {
			connection.Close()
		}
	}()

	t.Setenv("LOCALAPPDATA", t.TempDir())
	now := time.Now().UTC()
	payload := NewPayload(now)
	payload.Source = "claude-hook"
	payload.SessionID = "denied"
	payload.PID = 21
	payload = payload.Finish(now)

	started := time.Now()
	ok := Send(payload, SendOptions{Retry: 750 * time.Millisecond, PipePath: path})
	elapsed := time.Since(started)

	if ok {
		t.Fatal("Send claimed success against a pipe that denies this caller")
	}
	if elapsed > 300*time.Millisecond {
		t.Fatalf("Send spent %v retrying a permanent denial; the host pays that on every event", elapsed)
	}
}
