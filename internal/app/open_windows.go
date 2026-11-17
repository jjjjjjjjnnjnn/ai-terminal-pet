//go:build windows

package app

import (
	"fmt"
	"syscall"
	"unsafe"
)

// openExternal asks Windows to open a URL in the user's default browser.
//
// ShellExecuteW is the documented way to do this: it is the same call Explorer
// makes on a double-click, so it honours the user's default browser rather than
// guessing at one. The call is made only from a click on the settings footer -
// nothing here runs on its own, and the URL is a constant.
//
// shell32 is one of the three DLLs AGENTS.md allows (user32, gdi32, shell32).
var (
	shell32DLL     = syscall.NewLazyDLL("shell32.dll")
	shellExecuteWI = shell32DLL.NewProc("ShellExecuteW")
)

func openExternal(url string) error {
	verb, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := syscall.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	const swShowNormal = 1
	// ShellExecuteW returns a value greater than 32 on success; the low values
	// are its own error codes, so a "> 32" check is the documented contract and
	// not a magic number.
	result, _, callErr := shellExecuteWI.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		0,
		0,
		swShowNormal,
	)
	if result <= 32 {
		if callErr != nil && callErr.Error() != "The operation completed successfully." {
			return fmt.Errorf("ShellExecuteW %q: %w (code %d)", url, callErr, result)
		}
		return fmt.Errorf("ShellExecuteW %q failed (code %d)", url, result)
	}
	return nil
}
