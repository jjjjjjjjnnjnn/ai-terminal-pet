//go:build linux

package settingscli

import (
	"os"

	"golang.org/x/sys/unix"
)

// IsTerminal asks the kernel whether this file descriptor is a terminal.
//
// os.ModeCharDevice is not enough: /dev/null is a character device too, so a
// `--settings < /dev/null` would be treated as "a person is watching" and drop
// into an editor that immediately hits EOF. TCGETS is the same test every
// isatty() implementation uses.
func IsTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	_, err := unix.IoctlGetTermios(int(file.Fd()), unix.TCGETS)
	return err == nil
}
