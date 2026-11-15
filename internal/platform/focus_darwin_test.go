//go:build darwin

package platform

import (
	"os"
	"testing"
)

func TestDarwinProcessSnapshotIncludesSelf(t *testing.T) {
	procs := snapshotProcesses()
	if len(procs) == 0 {
		t.Fatal("snapshotProcesses() returned empty map")
	}
	if _, ok := procs[os.Getpid()]; !ok {
		t.Fatalf("snapshotProcesses() missing current pid %d", os.Getpid())
	}
}

func TestDarwinFocusQueriesDoNotPanicHeadless(t *testing.T) {
	if ActivateTarget(0) {
		t.Fatal("ActivateTarget(0) = true, want false")
	}
	_ = visibleWindows(snapshotProcesses())
	_ = SampleForeground()
}
