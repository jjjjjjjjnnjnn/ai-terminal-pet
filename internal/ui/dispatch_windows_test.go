//go:build windows

package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A nested dispatch must not be allowed to fill the thread's stack, and the
// entry that is refused has to leave a record. The 2026-11-17 hang (handover
// §79) was a window procedure nested deeply enough to exhaust the 2 MB the PE
// header reserves; this is the bound that turns that into one log line.
func TestDispatchDepthIsBoundedAndTheRefusalIsLogged(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("XDG_DATA_HOME", root)

	app := &App{}
	for depth := 1; depth <= maxDispatchDepth; depth++ {
		if !app.beginDispatch() {
			t.Fatalf("depth %d of %d was refused", depth, maxDispatchDepth)
		}
	}
	if app.beginDispatch() {
		t.Fatalf("depth %d was allowed; the stack bound does nothing", maxDispatchDepth+1)
	}
	if app.beginDispatch() {
		t.Fatal("the refusal was not remembered between messages")
	}
	// Unwind what was entered so a later test cannot observe the depth.
	for app.dispatchDepth > 0 {
		app.endDispatch()
	}
	if app.dispatchDepth != 0 {
		t.Fatalf("depth after unwinding = %d, want 0", app.dispatchDepth)
	}

	path := filepath.Join(root, "AI-Terminal-Pet", "error.log")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no error.log after a refused dispatch: %v", err)
	}
	text := string(raw)
	if !strings.Contains(text, "dispatch-overflow") {
		t.Fatalf("error.log does not name the refusal: %q", text)
	}
	if got := strings.Count(text, "dispatch-overflow"); got != 1 {
		t.Fatalf("the refusal was logged %d times, want 1", got)
	}
}

// The watchdog's clock is refreshed by every entry, including a nested one, so
// a modal menu loop stays "alive". The guard must not change that, and must not
// let endDispatch drive the depth negative.
func TestNestedDispatchStillRefreshesTheWatchdogClock(t *testing.T) {
	var app App
	if !app.beginDispatch() {
		t.Fatal("the first dispatch was refused")
	}
	first := app.lastDispatchTicks.Load()
	time.Sleep(5 * time.Millisecond)
	if !app.beginDispatch() {
		t.Fatal("the second dispatch was refused")
	}
	if got := app.lastDispatchTicks.Load(); got == first && got != 0 {
		t.Fatal("a nested entry did not refresh lastDispatchTicks")
	}
	app.endDispatch()
	app.endDispatch()
	app.endDispatch() // one too many: must be a no-op
	if app.dispatchDepth != 0 {
		t.Fatalf("depth = %d after extra ends, want 0", app.dispatchDepth)
	}
}

// A panic inside a window procedure is contained and written where a bug
// report can find it. Before this, the runtime printed the panic to an ignored
// stderr and then wedged, so the machine only showed a hung window.
func TestACallbackPanicIsContainedAndRecorded(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("XDG_DATA_HOME", root)

	app := &App{}
	result := app.guardedDispatch(0, wmRefresh, 0, 0, "overlay", func() uintptr {
		panic("synthetic fault")
	})
	if result != 0 {
		t.Fatalf("a panicking handler returned %#x, want the neutral zero", result)
	}
	if app.dispatchDepth != 0 {
		t.Fatalf("depth after a recovered panic = %d, want 0", app.dispatchDepth)
	}

	path := filepath.Join(root, "AI-Terminal-Pet", "error.log")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no error.log after a callback panic: %v", err)
	}
	text := string(raw)
	for _, want := range []string{"panic:", "overlay", "message=WM_APP+1(refresh)", "synthetic fault"} {
		if !strings.Contains(text, want) {
			t.Fatalf("error.log is missing %q: %q", want, text)
		}
	}
}
