//go:build windows

package app

// SettingsWindowAvailable reports whether this build has a settings window to
// open. `--settings` asks it before choosing between the window and the terminal
// editor (internal/settingscli), because starting a panel and silently doing
// nothing is the shape of bug this project keeps removing.
func SettingsWindowAvailable() bool { return true }
