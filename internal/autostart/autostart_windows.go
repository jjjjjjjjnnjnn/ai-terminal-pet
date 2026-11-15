//go:build windows

package autostart

// The Windows backend: HKCU\Software\Microsoft\Windows\CurrentVersion\Run.
//
// That location is deliberate: HKLM would need administrator rights and would
// apply to every account on the machine, neither of which is appropriate for a
// personal overlay. HKCU also means "disable" is a plain delete that always
// succeeds, with nothing left behind to clean up on uninstall.

import (
	"syscall"
	"unsafe"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

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

// newBackend returns the first non-nil backend among the platform
// implementations, in the order a machine is likely to have them. On Windows there
// is exactly one candidate, and that is the point: the entry must exist in the
// place the user's logon actually reads.
func newBackend() backend {
	return registryBackend{}
}

// registryBackend talks to the real registry.
type registryBackend struct{}

func (registryBackend) read(name string) (string, bool, error) {
	key, err := openRunKey(keyQueryValue)
	if err != nil {
		if err == errNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	defer regCloseKey.Call(uintptr(key))

	var valueType uint32
	var size uint32
	status, _, _ := regQueryValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if status == errorFileNotFound {
		return "", false, nil
	}
	if status != errorSuccess {
		return "", false, syscall.Errno(status)
	}
	if size == 0 || size > 64*1024 {
		return "", false, syscall.EINVAL
	}
	buffer := make([]uint16, size/2+1)
	status, _, _ = regQueryValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if status != errorSuccess {
		return "", false, syscall.Errno(status)
	}
	return syscall.UTF16ToString(buffer), true, nil
}

func (registryBackend) write(name, value string) error {
	key, err := openRunKey(keySetValue)
	if err != nil {
		return err
	}
	defer regCloseKey.Call(uintptr(key))

	encoded, encodeErr := syscall.UTF16FromString(value)
	if encodeErr != nil {
		return encodeErr
	}
	status, _, _ := regSetValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		0,
		regSZ,
		uintptr(unsafe.Pointer(&encoded[0])),
		uintptr(len(encoded)*2),
	)
	if status != errorSuccess {
		return syscall.Errno(status)
	}
	return nil
}

func (registryBackend) remove(name string) error {
	key, err := openRunKey(keySetValue)
	if err != nil {
		if err == errNotFound {
			return nil
		}
		return err
	}
	defer regCloseKey.Call(uintptr(key))

	status, _, _ := regDeleteValueW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
	)
	if status == errorFileNotFound {
		return nil
	}
	if status != errorSuccess {
		return syscall.Errno(status)
	}
	return nil
}

var errNotFound = syscall.Errno(errorFileNotFound)

func openRunKey(access uint32) (syscall.Handle, error) {
	var key syscall.Handle
	status, _, _ := regOpenKeyExW.Call(
		uintptr(hkeyCurrentUser),
		uintptr(unsafe.Pointer(utf16Ptr(runKeyPath))),
		0,
		uintptr(access),
		uintptr(unsafe.Pointer(&key)),
	)
	if status == errorFileNotFound {
		return 0, errNotFound
	}
	if status != errorSuccess {
		return 0, syscall.Errno(status)
	}
	return key, nil
}

// utf16Ptr is the one place a Go string becomes a UTF-16 buffer for the registry
// calls. syscall.UTF16PtrFromString rejects embedded NULs, which is the behaviour
// wanted here: a name containing one would address a different value.
func utf16Ptr(value string) *uint16 {
	pointer, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return nil
	}
	return pointer
}
