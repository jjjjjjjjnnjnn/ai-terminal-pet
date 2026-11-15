package connect

// roots is the platform-shaped corner of this package: where the user profile
// and the per-user data directory are. It is one file so that E3 (the macOS
// port) has exactly one place to answer differently, and so that a unit test
// can point the whole package at a temp directory without touching the machine
// it runs on.
//
// Nothing else here is platform-specific: every check reads files and compares
// paths, which is why four of the inspectors were able to be written with no
// build tag at all.

import (
	"os"
	"path/filepath"
)

// osExecutable is os.Executable behind a name, so the seam is visible next to
// its siblings.
func osExecutable() (string, error) {
	return os.Executable()
}

// osUserHome is os.UserHomeDir behind the same name.
func osUserHome() (string, error) {
	return os.UserHomeDir()
}

// joinPath is filepath.Join behind a name, kept for symmetry with the two
// above: the three call sites in this package read as "the platform decides
// this", not as "a path happens to be joined".
func joinPath(elements ...string) string {
	return filepath.Join(elements...)
}
