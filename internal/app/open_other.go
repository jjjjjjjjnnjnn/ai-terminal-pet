//go:build !windows

package app

import "errors"

// openExternal has no implementation off Windows: the only surface that calls it
// is the Win32 settings window's footer. The terminal settings editor
// (internal/settingscli) prints the same two links instead, so a Linux or macOS
// user copies them rather than having a browser launched for them - which is
// also why this returns an error instead of pretending to succeed.
func openExternal(url string) error {
	return errors.New("opening a browser is only implemented on Windows; print the URL instead")
}
