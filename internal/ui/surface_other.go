//go:build !windows && !linux && !darwin

package ui

// Platforms without a native window layer still build this package: App is the
// placeholder in app_posix.go, and internal/app's doctor is compiled from the
// same untagged code on every OS. Both surface questions therefore need an
// answer here rather than a build error.

// SoundSurface names what the `sound` setting would reach.
func SoundSurface() string { return "none (no window layer on this platform)" }

// NotificationSurface names what `notifications` would reach.
func NotificationSurface() string { return "none (no window layer on this platform)" }

// notificationSurfaceAvailable answers for the alert receipt.
func notificationSurfaceAvailable() bool { return false }
