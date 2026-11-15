//go:build linux

package platform

import (
	"os"
	"testing"
)

func TestLinuxProcSnapshotContainsSelfAndParent(t *testing.T) {
	procs := snapshotProcesses()
	pid := os.Getpid()
	self, ok := procs[pid]
	if !ok {
		t.Fatalf("snapshotProcesses missing self pid %d", pid)
	}
	if self.name == "" {
		t.Fatalf("snapshotProcesses self pid %d has empty name", pid)
	}
	if ppid := os.Getppid(); ppid > 0 && self.parent != ppid {
		t.Fatalf("self parent = %d, os.Getppid() = %d", self.parent, ppid)
	}
	if !ProcessAlive(pid) {
		t.Fatalf("ProcessAlive(%d) = false for current process", pid)
	}
}
