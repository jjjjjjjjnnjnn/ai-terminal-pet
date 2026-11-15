package ui

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// The watchdog end to end, without a window and without sleeping a machine.
//
// watchHang only needs three things to run: a non-zero handle, the awake clock,
// and LOCALAPPDATA for the log. A real window is not one of them, which is what
// makes the case that matters - "a resume from standby must not be dumped" -
// testable at all. Suspend cannot be simulated; the pair of readings it produces
// can be.
func TestWatchdogDumpsARealStallAndNotAResume(t *testing.T) {
	cases := []struct {
		name      string
		quietFor  time.Duration
		wantDump  bool
		reasonWhy string
	}{
		{
			name:     "eight seconds of awake silence is a real stall",
			quietFor: 8 * time.Second,
			wantDump: true,
			// The UI thread has not dispatched for eight seconds of awake time.
			reasonWhy: "a block must still be reported",
		},
		{
			name:     "three milliseconds after a resume is not a stall",
			quietFor: 3 * time.Millisecond,
			wantDump: false,
			// This is what a resume produces: the previous dispatch was moments
			// before the machine slept, so the awake gap is milliseconds even
			// though the wall gap was hours.
			reasonWhy: "standby must not look like a freeze",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			t.Setenv("LOCALAPPDATA", directory)

			app := &App{}
			app.handle.Store(1) // watchHang's only precondition on a window
			// One unit of interrupt time is 100 ns.
			units := uint64(tc.quietFor / (100 * time.Nanosecond))
			app.lastDispatchTicks.Store(interruptTicks() - units)

			stop := make(chan struct{})
			done := make(chan struct{})
			go func() {
				defer close(done)
				app.watchHang(stop)
			}()
			// The watchdog ticks once a second; two ticks is enough to decide.
			time.Sleep(2500 * time.Millisecond)
			close(stop)
			<-done

			path := filepath.Join(directory, "AI-Terminal-Pet", "hang.log")
			_, statErr := os.Stat(path)
			if tc.wantDump && statErr != nil {
				t.Fatalf("%s: no hang.log was written (%v)", tc.reasonWhy, statErr)
			}
			if !tc.wantDump && statErr == nil {
				content, _ := os.ReadFile(path)
				t.Fatalf("%s: a dump was written: %s", tc.reasonWhy, firstLine(string(content)))
			}
			if tc.wantDump {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				header := firstLine(string(content))
				if !strings.Contains(header, "stall 1/3") {
					t.Fatalf("the header does not name the dump: %q", header)
				}
				// The header now states how long the UI thread was quiet *while
				// awake*, which is the number that makes a dump readable: before
				// the clock change it would have said an hour.
				if !strings.Contains(header, "awakeQuiet=") {
					t.Fatalf("the header no longer reports awakeQuiet: %q", header)
				}
				// The watchdog tick adds a second of slack, so 8 s reads as 8-9.
				if !strings.Contains(header, "awakeQuiet=8") && !strings.Contains(header, "awakeQuiet=9") {
					t.Fatalf("awakeQuiet is not the measured gap: %q", header)
				}
			}
		})
	}
}

func firstLine(text string) string {
	if index := strings.IndexByte(text, '\n'); index >= 0 {
		return text[:index]
	}
	return text
}

// A missing or zero reading must not produce a dump: on a host where the API
// answered nothing, the watchdog is silent rather than loud. This is the one
// failure mode that would be worse than the false positives it replaces.
func TestWatchdogStaysQuietWithoutAClock(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("LOCALAPPDATA", directory)

	app := &App{}
	app.handle.Store(1)
	app.lastDispatchTicks.Store(0)

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		app.watchHang(stop)
	}()
	time.Sleep(1200 * time.Millisecond)
	close(stop)
	<-done

	if _, err := os.Stat(filepath.Join(directory, "AI-Terminal-Pet", "hang.log")); err == nil {
		t.Fatal("a dump was written although the clock was never recorded")
	}
}

// lastDispatchTicks is written from the UI thread (both window procedures) and
// read from the watchdog goroutine, which is why it is atomic rather than a
// plain field. This is a compile-time reminder, not a race test: -race needs a
// C toolchain this machine does not have (docs/ROADMAP.md §6).
var _ = atomic.Uint64{}
