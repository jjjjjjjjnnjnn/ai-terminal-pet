//go:build !windows

package app

// The POSIX half of the overlay's lifecycle, as far as it exists.
//
// The single-instance handle and the console hide are the easy parts and they are
// here. The presence cache is deliberately only a pid probe: its SessionAlive rule
// is the accumulated result of handover §21, §43, §51 and §52, and every one of
// those defects came from answering "is this session still real" with less evidence
// than Windows had. A platform file may answer that question with fewer signals, but
// it must say so in its own words (§55.4's "if it has to degrade, write it in Known
// limits and --doctor") rather than inheriting a rule whose evidence it lacks.
//
// That is also what ErrNoWindowLayer stands in front of: with no window layer there
// is no overlay to run, so this file provides the instance handle, a conservative
// presence cache and the no-ops, and stops there.

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/session"
)

// errAlreadyRunning mirrors platform.ErrAlreadyRunning's contract: a second instance
// is not an error, it is a reason to poke the running one.
var errAlreadyRunning = errors.New("AI Terminal Pet is already running")

// newInstance takes a flock on a file in the project's data directory.
//
// A lock file is weaker than a Windows mutex, and the reason it still works is
// flock: the lock is tied to the open descriptor, so a process that is killed hard
// releases it. "Does the file exist" would have the opposite failure - a stale file
// from a crashed run would block the next start forever.
func newInstance(name string) (*lockInstance, error) {
	directory := home.Directory()
	if directory == "" {
		return nil, errors.New("this platform has no data directory")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, name+".lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := flockExclusive(file); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, errAlreadyRunning
		}
		return nil, err
	}
	return &lockInstance{file: file}, nil
}

// lockInstance holds the flock until Close.
type lockInstance struct {
	file *os.File
}

func (i *lockInstance) Close() {
	if i == nil || i.file == nil {
		return
	}
	_ = flockUnlock(i.file)
	_ = i.file.Close()
	i.file = nil
}

// noPresence is the cache a platform without window evidence gets.
//
// ProcessAlive is a signal-0 probe, which is exact for "this pid exists". SessionAlive
// is the same probe, and that is the honest limit: it cannot see that a terminal tab
// was closed while its host process lives on. A session therefore lingers until the
// absence grace expires, which docs/PLATFORMS.md states rather than hides.
type noPresence struct{}

func (p *noPresence) ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 is the POSIX "does this process exist" probe. EPERM means it exists
	// but belongs to another user, which is still alive.
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func (p *noPresence) SessionAlive(value session.Session) bool {
	return p.ProcessAlive(value.PID)
}

// newPresence builds the cache the scanner uses.
func newPresence() *noPresence {
	return &noPresence{}
}

// hideConsole is a no-op: a POSIX build has no console to detach from.
func hideConsole() {}

// uiAvailable reports whether this build has a window layer. It does not yet: see
// internal/ui/canvas_posix.go for what is missing, and docs/PLATFORMS.md for the
// table a reader can check.
func uiAvailable() bool { return false }
