package home

import "os"

// osUserHomeDir is os.UserHomeDir behind a name, so the tests' seam is visible
// next to its siblings in home.go.
func osUserHomeDir() (string, error) {
	return os.UserHomeDir()
}
