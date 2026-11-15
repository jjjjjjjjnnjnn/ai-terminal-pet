//go:build !windows

package reporter

// ConsoleAnchorTitle has no POSIX equivalent worth guessing at: a Unix session is
// identified by the terminal window that owns its pty, which the reporter never
// sees, and internal/platform ranks X11 windows by process ancestry instead.
// Returning "" means "no hint", which the caller already treats as the default.
func ConsoleAnchorTitle() string { return "" }
