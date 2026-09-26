// Package autostart manages the per-user "run at logon" entry.
//
// The entry is a value under HKCU\Software\Microsoft\Windows\CurrentVersion\Run.
// That location is deliberate: HKLM would need administrator rights and would
// apply to every account on the machine, neither of which is appropriate for a
// personal overlay. HKCU also means "disable" is a plain delete that always
// succeeds, with nothing left behind to clean up on uninstall.
//
// The registry is reached through the backend interface rather than directly,
// so the enable/disable/repair logic is testable. A unit test that silently
// turned on autostart for whoever ran `go test` would be a genuinely bad
// citizen, and registry writes are machine state, not a temporary directory.
package autostart

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

	// ValueName is the Run value this app owns. It is exported so the CLI and
	// the docs can refer to the same string.
	ValueName = "AITerminalPet"

	hkeyCurrentUser = 0x80000001

	keyQueryValue = 0x0001
	keySetValue   = 0x0002

	regSZ = 1

	errorSuccess      = 0
	errorFileNotFound = 2
)

var (
	advapi32         = syscall.NewLazyDLL("advapi32.dll")
	regOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	regQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	regSetValueExW   = advapi32.NewProc("RegSetValueExW")
	regDeleteValueW  = advapi32.NewProc("RegDeleteValueW")
	regCloseKey      = advapi32.NewProc("RegCloseKey")
)

// Command is the command line stored in the Run value.
//
// Windows parses the value as a command line, so a path containing spaces must
// be quoted or the shell would try to execute "C:\Program". Quoting is
// unconditional rather than conditional on whitespace: it is always correct and
// it means a user whose install path has no spaces today cannot be surprised
// after moving it.
func Command(executable string) string {
	return `"` + executable + `"`
}

// Status is the observed state of the autostart entry.
type Status struct {
	// Enabled reports whether the Run value exists.
	Enabled bool
	// Target is the executable the value currently points at, unquoted. It is
	// empty when disabled.
	Target string
}

// backend is the slice of registry behaviour this package needs.
type backend interface {
	read(name string) (value string, found bool, err error)
	write(name, value string) error
	remove(name string) error
}

// Manager reads and writes the autostart entry.
type Manager struct {
	backend backend
}

// New returns a Manager backed by the real per-user Run key.
func New() *Manager {
	return &Manager{backend: registryBackend{}}
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

// Enable points the autostart entry at executable, replacing any previous
// value.
func (m *Manager) Enable(executable string) error {
	if strings.TrimSpace(executable) == "" {
		return errors.New("autostart executable is empty")
	}
	return m.backend.write(ValueName, Command(executable))
}

// Disable removes the autostart entry. Disabling when already disabled is not
// an error: the caller's intent is "make sure it is off".
func (m *Manager) Disable() error {
	return m.backend.remove(ValueName)
}

// Sync re-points an enabled entry at executable and reports whether it had to
// change anything.
//
// The install path is the only thing in the entry, so moving or reinstalling
// the app leaves the Run value pointing at a binary that no longer exists -
// which fails silently at logon, the worst possible way to fail. Repairing it
// on every start costs one registry read.
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

// registryBackend talks to the real registry.
type registryBackend struct{}

func (registryBackend) read(name string) (string, bool, error) {
	key, err := openRunKey(keyQueryValue)
	if err != nil {
		return "", false, err
	}
	defer regCloseKey.Call(uintptr(key))

	valueName, err := syscall.UTF16FromString(name)
	if err != nil {
		return "", false, err
	}
	var kind uint32
	var size uint32
	result, _, _ := regQueryValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(&valueName[0])),
		0,
		uintptr(unsafe.Pointer(&kind)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if result == errorFileNotFound {
		return "", false, nil
	}
	if result != errorSuccess {
		return "", false, fmt.Errorf("RegQueryValueEx(%s): %w", name, syscall.Errno(result))
	}
	if kind != regSZ || size == 0 {
		// A REG_EXPAND_SZ or binary value under our name is not ours; treat it
		// as absent rather than trying to interpret it.
		return "", false, nil
	}
	buffer := make([]uint16, (size+1)/2)
	result, _, _ = regQueryValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(&valueName[0])),
		0,
		uintptr(unsafe.Pointer(&kind)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if result != errorSuccess {
		return "", false, fmt.Errorf("RegQueryValueEx(%s): %w", name, syscall.Errno(result))
	}
	return syscall.UTF16ToString(buffer), true, nil
}

func (registryBackend) write(name, value string) error {
	key, err := openRunKey(keySetValue)
	if err != nil {
		return err
	}
	defer regCloseKey.Call(uintptr(key))

	valueName, err := syscall.UTF16FromString(name)
	if err != nil {
		return err
	}
	encoded, err := syscall.UTF16FromString(value)
	if err != nil {
		return err
	}
	result, _, _ := regSetValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(&valueName[0])),
		0,
		regSZ,
		uintptr(unsafe.Pointer(&encoded[0])),
		uintptr(len(encoded)*2),
	)
	if result != errorSuccess {
		return fmt.Errorf("RegSetValueEx(%s): %w", name, syscall.Errno(result))
	}
	return nil
}

func (registryBackend) remove(name string) error {
	key, err := openRunKey(keySetValue)
	if err != nil {
		return err
	}
	defer regCloseKey.Call(uintptr(key))

	valueName, err := syscall.UTF16FromString(name)
	if err != nil {
		return err
	}
	result, _, _ := regDeleteValueW.Call(uintptr(key), uintptr(unsafe.Pointer(&valueName[0])))
	if result == errorFileNotFound || result == errorSuccess {
		return nil
	}
	return fmt.Errorf("RegDeleteValue(%s): %w", name, syscall.Errno(result))
}

// openRunKey opens the per-user Run key with the requested access. The key
// always exists on a normal Windows profile; a failure here means the profile
// is broken and the caller should surface it rather than silently pretend
// autostart is off.
func openRunKey(access uint32) (syscall.Handle, error) {
	subKey, err := syscall.UTF16FromString(runKeyPath)
	if err != nil {
		return 0, err
	}
	var key syscall.Handle
	result, _, _ := regOpenKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(&subKey[0])),
		0,
		uintptr(access),
		uintptr(unsafe.Pointer(&key)),
	)
	if result != errorSuccess {
		return 0, fmt.Errorf("RegOpenKeyEx(%s): %w", runKeyPath, syscall.Errno(result))
	}
	return key, nil
}
