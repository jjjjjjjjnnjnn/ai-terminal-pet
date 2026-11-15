package reporter

// The platform-independent half of the console-title jump hint: what counts as a
// title worth sending. See console_windows.go for the measurement behind it.

import "strings"

// sanitizeConsoleTitle rejects values that describe a binary rather than a
// conversation. A console that never had its title set reports the shell's
// executable path ("C:\windows\system32\WindowsPowerShell\v1.0\powershell.EXE"),
// and sending that as a title would only add noise: no window title will ever
// contain it.
func sanitizeConsoleTitle(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if strings.ContainsAny(trimmed, `\/`) || strings.HasSuffix(lower, ".exe") {
		return ""
	}
	return trimmed
}
