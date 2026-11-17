//go:build darwin

package settingscli

import (
	"os"

	"golang.org/x/sys/unix"
)

// IsTerminal asks the kernel whether this file descriptor is a terminal.
//
// macOS spells the same ioctl TIOCGETA rather than Linux's TCGETS; both are what
// isatty() uses underneath. os.ModeCharDevice would accept /dev/null, which is
// how `--settings < /dev/null` would end up in an editor with no input.
func IsTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	_, err := unix.IoctlGetTermios(int(file.Fd()), unix.TIOCGETA)
	return err == nil
}
