package autostart

// The platform-neutral half of the run-at-logon entry: the Manager, its
// semantics, and the backend interface the two platform files implement.
//
// The registry is reached through an interface rather than directly, and that is
// not only for tests: a unit test that silently turned autostart on for whoever
// ran `go test` would be a genuinely bad citizen, and registry writes are machine
// state, not a temporary directory. The same interface now lets one implementation
// point at HKCU\...\Run and another at ~/Library/LaunchAgents (handover §55.6).

import (
	"errors"
	"strings"
)

// ValueName is the entry this app owns. It is exported so the CLI and the docs can
// refer to the same string, and it is the same on every platform: on Windows it is
// the Run value's name, on macOS the LaunchAgent label's last element, on Linux the
// systemd unit's base name.
const ValueName = "AITerminalPet"

// Command is the command line stored in the entry.
//
// Windows parses the value as a command line, so a path containing spaces must be
// quoted or the shell would try to execute "C:\Program". Quoting is unconditional
// rather than conditional on whitespace: it is always correct and it means a user
// whose install path has no spaces today cannot be surprised after moving it.
func Command(executable string) string {
	return `"` + executable + `"`
}

// Status is the observed state of the autostart entry.
type Status struct {
	// Enabled reports whether the entry exists.
	Enabled bool
	// Target is the executable the entry currently points at, unquoted. It is
	// empty when disabled.
	Target string
}

// backend is the slice of native behaviour this package needs.
type backend interface {
	read(name string) (value string, found bool, err error)
	write(name, value string) error
	remove(name string) error
}

// Manager reads and writes the autostart entry.
type Manager struct {
	backend backend
}

// New returns a Manager backed by the running platform's own mechanism.
func New() *Manager {
	return &Manager{backend: newBackend()}
}

// newWith builds a Manager over an arbitrary backend. Test seam.
func newWith(value backend) *Manager {
	return &Manager{backend: value}
}

// Status reports whether autostart is on and what it points at.
func (m *Manager) Status() (Status, error) {
	value, found, err := m.backend.read(ValueName)
	if err != nil {
		return Status{}, err
	}
	if !found {
		return Status{}, nil
	}
	return Status{Enabled: true, Target: unquote(value)}, nil
}

// Enable points the autostart entry at executable, replacing any previous value.
func (m *Manager) Enable(executable string) error {
	if strings.TrimSpace(executable) == "" {
		return errors.New("autostart executable is empty")
	}
	return m.backend.write(ValueName, Command(executable))
}

// Disable removes the autostart entry. Disabling when already disabled is not an
// error: the caller's intent is "make sure it is off".
func (m *Manager) Disable() error {
	return m.backend.remove(ValueName)
}

// Sync re-points an enabled entry at executable and reports whether it had to
// change anything.
//
// The install path is the only thing in the entry, so moving or reinstalling the
// app leaves it pointing at a binary that no longer exists - which fails silently
// at logon, the worst possible way to fail. Repairing it on every start costs one
// read.
//
// A disabled entry is left disabled: Sync never turns autostart on behind the
// user's back.
func (m *Manager) Sync(executable string) (bool, error) {
	status, err := m.Status()
	if err != nil || !status.Enabled {
		return false, err
	}
	if status.Target == executable {
		return false, nil
	}
	if err := m.Enable(executable); err != nil {
		return false, err
	}
	return true, nil
}

// unquote strips the surrounding quotes Command adds, tolerating a hand-edited
// value that has none.
func unquote(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
		return trimmed[1 : len(trimmed)-1]
	}
	return trimmed
}
