//go:build windows

package settingscli

import (
	"os"

	"golang.org/x/sys/windows"
)

// IsTerminal asks Windows whether this handle is a console.
//
// os.ModeCharDevice is not enough: NUL is a character device, so
// `--settings < NUL` would look like a person watching and drop into an editor
// that immediately hits EOF. GetConsoleMode fails for anything redirected.
func IsTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	var mode uint32
	return windows.GetConsoleMode(windows.Handle(file.Fd()), &mode) == nil
}
