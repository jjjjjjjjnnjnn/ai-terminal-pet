//go:build !windows && !linux && !darwin

package settingscli

import "os"

// IsTerminal is false on the platforms that have no window layer and no tested
// terminal binding: `--settings` prints the table there rather than promising an
// editor nobody has run. See docs/PLATFORMS.md.
func IsTerminal(file *os.File) bool { return false }
