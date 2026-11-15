//go:build !windows && !linux && !darwin

package ui

func copyToClipboard(_ string) bool {
	return false
}
