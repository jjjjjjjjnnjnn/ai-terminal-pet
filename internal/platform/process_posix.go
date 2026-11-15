//go:build !windows

package platform

import (
	"errors"
	"syscall"
)

// ErrAlreadyRunning reports an existing single-instance owner.
var ErrAlreadyRunning = errors.New("AI Terminal Pet is already running")

// Instance represents a platform single-instance handle (on POSIX internal/app
// owns the flock handle directly; this stub satisfies platform references).
type Instance struct{}

// ProcessAlive checks whether pid is alive via signal 0 probing.
func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// AcquireInstance is unsupported here on POSIX because internal/app uses flock
// on a directory-scoped file directly.
func AcquireInstance(name string) (*Instance, error) {
	return &Instance{}, nil
}

// Close releases the instance handle.
func (i *Instance) Close() {}

// HideConsole is a no-op on POSIX.
func HideConsole() {}
