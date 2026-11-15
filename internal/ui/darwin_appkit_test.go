//go:build darwin

package ui

import (
	"os"
	"testing"
	"time"

	"ai-terminal-pet/internal/store"
	"github.com/ebitengine/purego/objc"
)

func TestDarwinCocoaClassesAndHelpersHeadlessSafe(t *testing.T) {
	if !ensureCocoaLoaded() {
		t.Fatal("ensureCocoaLoaded() = false on darwin")
	}
	if err := ensureDarwinAppClasses(); err != nil {
		t.Fatalf("ensureDarwinAppClasses() error = %v", err)
	}
	s := nsStringFromGo("AI Terminal Pet · 测试")
	if s == 0 {
		t.Fatal("nsStringFromGo returned 0")
	}
	s.Send(objc.RegisterName("release"))
}

func TestDarwinControlSocketRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	t.Setenv("HOME", tmp)

	got := make(chan ControlAction, 1)
	closer, err := startControlListener(func(action ControlAction) {
		got <- action
	})
	if err != nil {
		t.Fatalf("startControlListener() error = %v", err)
	}
	defer closer.Close()

	if !postControlMessage(ControlToggle) {
		t.Fatal("postControlMessage(ControlToggle) = false")
	}
	select {
	case action := <-got:
		if action != ControlToggle {
			t.Fatalf("received action = %v, want ControlToggle", action)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Unix control socket dispatch")
	}
}

func TestDarwinCocoaWindowSmokeProbe(t *testing.T) {
	if os.Getenv("PET_COCOA_PROBE") == "" {
		t.Skip("set PET_COCOA_PROBE=1 on a desktop macOS session to run live NSWindow smoke")
	}
	values := store.New()
	app := New(values, Options{})
	go func() {
		time.Sleep(400 * time.Millisecond)
		app.Close()
	}()
	if err := app.Run(); err != nil {
		t.Fatalf("App.Run() error = %v", err)
	}
}
