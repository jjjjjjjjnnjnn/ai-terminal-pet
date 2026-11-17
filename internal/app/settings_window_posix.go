//go:build !windows

package app

// SettingsWindowAvailable is false everywhere without the Win32 settings
// window: internal/ui's settings layout is drawn by GDI, and docs/ROADMAP.md §5
// rules out a toolkit port (it would mean cgo). The terminal editor in
// internal/settingscli is the settings surface here instead, so `--settings`
// means "edit them" on every platform rather than "start a panel and print a
// hint" on two of them.
func SettingsWindowAvailable() bool { return false }
