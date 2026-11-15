package i18n

// The Windows locale lookup: GetUserDefaultLocaleName out of kernel32.
//
// Kept in its own file so the package's pure half (NormalizeSetting, Resolve,
// the two dictionaries) compiles on every platform. This is the first step of
// handover §55.6: the darwin build used to stop here with
// "undefined: syscall.NewLazyDLL" even though nothing about the dictionaries is
// platform-specific.

import (
	"syscall"
	"unsafe"
)

var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLocaleName = modKernel32.NewProc("GetUserDefaultLocaleName")
)

// systemLocale reads the user default locale via the Win32 API. LOCALE_NAME_MAX_LENGTH
// is 85; anything longer is truncated, which is still enough to detect the
// language prefix we care about.
func systemLocale() (string, error) {
	const maxLen = 85
	buf := make([]uint16, maxLen)
	ret, _, err := procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(maxLen),
	)
	if ret == 0 {
		if err != nil && err != syscall.Errno(0) {
			return "", err
		}
		return "", syscall.EINVAL
	}
	return syscall.UTF16ToString(buf), nil
}
