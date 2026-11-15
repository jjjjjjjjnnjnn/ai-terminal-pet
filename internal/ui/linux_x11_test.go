//go:build linux

package ui

import (
	"os"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"github.com/jezek/xgb/xproto"
)

func TestKeySymToPanelKey(t *testing.T) {
	cases := []struct {
		sym  uint32
		want panelKey
	}{
		{0xff1b, panelKeyEscape},
		{0xff52, panelKeyUp},
		{0xff97, panelKeyUp},
		{0xff54, panelKeyDown},
		{0xff99, panelKeyDown},
		{0xff50, panelKeyHome},
		{0xff57, panelKeyEnd},
		{0xff0d, panelKeyEnter},
		{0xff8d, panelKeyEnter},
		{'a', panelKeyNone},
	}
	for _, tc := range cases {
		if got := keySymToPanelKey(xproto.Keysym(tc.sym)); got != tc.want {
			t.Fatalf("keySymToPanelKey(0x%x) = %v, want %v", tc.sym, got, tc.want)
		}
	}
}

func TestPutImageChunkRowsRespectsBudget(t *testing.T) {
	rows := putImageChunkRows(200, 1024)
	if rows != 1 {
		t.Fatalf("rows for 1024B budget width 200 = %d, want 1", rows)
	}
	rowsLarge := putImageChunkRows(200, 262140)
	if rowsLarge*200*4 > 262140-24 {
		t.Fatalf("chunk byte count %d exceeds budget", rowsLarge*200*4)
	}
	if rowsLarge < 300 {
		t.Fatalf("expected >300 rows per chunk for width 200, got %d", rowsLarge)
	}
}

func TestDisplaysFromMonitorsAppliesSingleWorkarea(t *testing.T) {
	work := displayRect{Left: 0, Top: 32, Right: 1920, Bottom: 1040}
	displays := displaysFromMonitors(1920, 1080, []x11Monitor{{
		Left:    0,
		Top:     0,
		Width:   1920,
		Height:  1080,
		Primary: false,
	}}, true, work)
	if len(displays) != 1 {
		t.Fatalf("expected 1 display, got %d", len(displays))
	}
	if displays[0].Top != 32 || displays[0].Bottom != 1040 {
		t.Fatalf("expected single monitor clipped to _NET_WORKAREA, got %+v", displays[0])
	}
}

func TestParseWorkareaCardinal(t *testing.T) {
	raw := encodeCardinals(10, 28, 1900, 1020)
	rect, ok := parseWorkareaCardinal(raw)
	if !ok {
		t.Fatal("expected valid workarea")
	}
	if rect.Left != 10 || rect.Top != 28 || rect.Right != 1910 || rect.Bottom != 1048 {
		t.Fatalf("unexpected rect %+v", rect)
	}
}

func TestLinuxCanvasRendersNonEmptyPanelPixels(t *testing.T) {
	c, err := newCanvas(expandedWidth, expandedPanelHeight(2))
	if err != nil {
		t.Fatalf("newCanvas failed: %v", err)
	}
	defer c.close()
	c.clear()
	now := time.Now()
	drawVerticalPanel(c, panelState{
		Items: []session.Session{
			{Source: "claude-hook", SessionID: "s1", PID: 101, Project: "/tmp/demo", State: session.StateWaitingPermission, UpdatedAt: now},
			{Source: "opencode", SessionID: "s2", PID: 102, Project: "/tmp/app", State: session.StateWorking, UpdatedAt: now},
		},
		Now:      now,
		Expanded: true,
	})
	nonZero := 0
	for i := 3; i < len(c.pixels); i += 4 {
		if c.pixels[i] > 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatal("expected Linux canvas to render non-zero alpha pixels")
	}
}

// TestX11LiveWindowProbe connects to DISPLAY when PET_X11_PROBE=1, maps the
// real overlay window, paints sessions onto it, and exits via ControlQuit.
func TestX11LiveWindowProbe(t *testing.T) {
	if os.Getenv("PET_X11_PROBE") == "" {
		t.Skip("set PET_X11_PROBE=1 to probe a live X11 display")
	}
	st := store.New()
	now := time.Now()
	st.UpsertPipe(session.Session{
		Source:    "opencode",
		SessionID: "probe-1",
		PID:       os.Getpid(),
		Project:   "/tmp/x11-probe",
		Title:     "X11 Probe",
		State:     session.StateWorking,
		UpdatedAt: now,
	}, now)
	app := New(st, Options{Settings: DefaultSettings()})
	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()
	time.Sleep(350 * time.Millisecond)
	app.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("App.Run failed on live X11 display: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for X11 App.Run to exit")
	}
}
