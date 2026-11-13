package ui

// Clipboard tests.
//
// The real round trip needs a desktop session: on a machine with no clipboard
// window station (headless CI) OpenClipboard fails, which the code correctly
// reports as "not copied". So the round trip is an opt-in probe, like the
// project's other desktop probes, while the rules that must hold everywhere
// (empty input, embedded NUL) are plain unit tests.

import (
	"os"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/lxn/win"
)

// Nothing to copy must report failure rather than clear the clipboard: the
// receipt the caller shows is a claim that the path is on the clipboard.
func TestCopyToClipboardRejectsNothingToCopy(t *testing.T) {
	if copyToClipboard("") {
		t.Fatal("an empty value must not be reported as copied")
	}
	// An embedded NUL cannot round-trip through CF_UNICODETEXT, which is
	// NUL-terminated: UTF16FromString refuses it and so must we.
	if copyToClipboard("C:\\a\x00b") {
		t.Fatal("a value with an embedded NUL must not be reported as copied")
	}
}

// TestClipboardProbe is the opt-in desktop probe:
//
//	$env:PET_CLIPBOARD_PROBE = "1"; go test ./internal/ui -run TestClipboardProbe
//
// It is what proves the feaature on a real machine: it copies a path, reads it
// back through the clipboard API, and checks the round trip. CI skips it
// because the clipboard belongs to a window station a headless runner may not
// have - documented in the file header rather than silently passing.
func TestClipboardProbe(t *testing.T) {
	if os.Getenv("PET_CLIPBOARD_PROBE") == "" {
		t.Skip("set PET_CLIPBOARD_PROBE=1 to exercise the real clipboard")
	}
	value := `C:\Users\probe\Desktop\ai-terminal-pet`
	if !copyToClipboard(value) {
		t.Fatal("copyToClipboard reported failure on a desktop session")
	}
	got, ok := readClipboard(t)
	if !ok {
		t.Fatal("the clipboard has no CF_UNICODETEXT after a successful copy")
	}
	if got != value {
		t.Fatalf("clipboard round trip = %q, want %q", got, value)
	}
	// A second copy must replace the first, not append: the gesture is a
	// copy, and a clipboard holding "pathpath" is the classic symptom of
	// forgetting EmptyClipboard.
	if !copyToClipboard(value + "2") {
		t.Fatal("the second copy failed")
	}
	if got, _ := readClipboard(t); got != value+"2" {
		t.Fatalf("the second copy left %q on the clipboard", got)
	}
}

// readClipboard is the test-side readback. It fails the test rather than
// returning an error because every caller treats "cannot read" as a defect in
// the probe itself.
func readClipboard(t *testing.T) (string, bool) {
	t.Helper()
	if !win.OpenClipboard(0) {
		return "", false
	}
	defer win.CloseClipboard()
	handle := win.GetClipboardData(win.CF_UNICODETEXT)
	if handle == 0 {
		return "", false
	}
	pointer := win.GlobalLock(win.HGLOBAL(handle))
	if pointer == nil {
		return "", false
	}
	defer win.GlobalUnlock(win.HGLOBAL(handle))
	// CF_UNICODETEXT is NUL-terminated; a copy of a real path is far shorter
	// than this bound, and the slice stops at the terminator anyway.
	units := unsafe.Slice((*uint16)(pointer), 4096)
	text := syscall.UTF16ToString(units)
	if text == "" {
		return "", false
	}
	return text, strings.Contains(text, "\\") || text == "C:"
}
