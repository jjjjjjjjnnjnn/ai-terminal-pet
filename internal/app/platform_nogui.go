//go:build !windows && !linux && !darwin

package app

// uiAvailable reports whether this POSIX platform includes a window layer.
func uiAvailable() bool { return false }
