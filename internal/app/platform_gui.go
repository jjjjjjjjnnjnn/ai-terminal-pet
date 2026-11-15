//go:build linux || darwin

package app

// uiAvailable reports whether this Unix build includes the pure-Go window layer.
func uiAvailable() bool { return true }
