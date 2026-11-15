//go:build windows

package reporter

import (
	"syscall"
	"testing"
	"unsafe"
)

// TestConsoleAnchorTitleRoundTrip reads the console title back through the same
// path the statusline uses. It runs on a real console (Windows Terminal here,
// the null console in CI), so it is a probe of the syscall rather than of a
// fixture: the title is set, read, and restored, and the assertion is on the
// read.
func TestConsoleAnchorTitleRoundTrip(t *testing.T) {
	const probeTitle = "pet-console-anchor-probe"
	setConsoleTitle := consoleKernel32.NewProc("SetConsoleTitleW")
	original := currentConsoleTitle()

	set := func(value string) bool {
		text, err := syscall.UTF16FromString(value)
		if err != nil {
			t.Fatalf("UTF16FromString(%q): %v", value, err)
		}
		ok, _, _ := setConsoleTitle.Call(uintptr(unsafe.Pointer(&text[0])))
		return ok != 0
	}
	if !set(probeTitle) {
		t.Skip("this process has no console to retitle; the title path cannot be exercised here")
	}
	got := ConsoleAnchorTitle()
	set(original)

	if got != probeTitle {
		t.Fatalf("ConsoleAnchorTitle() = %q, want %q (the console title this process just set)", got, probeTitle)
	}
}

// TestConsoleAnchorTitleNeverReturnsABinaryPath pins the second half of the
// contract: whatever the console reports, the hint is either a conversation
// title or empty - never a path that could only ever be noise in the resolver.
func TestConsoleAnchorTitleNeverReturnsABinaryPath(t *testing.T) {
	if title := ConsoleAnchorTitle(); title != "" && sanitizeConsoleTitle(title) != title {
		t.Fatalf("ConsoleAnchorTitle() returned an unsanitised value %q", title)
	}
}
