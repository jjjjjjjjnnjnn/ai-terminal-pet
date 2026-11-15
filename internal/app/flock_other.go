//go:build !linux && !darwin && !windows

package app

// The fallback for the platforms nobody has ported yet (freebsd, openbsd, ...).
//
// An exclusive lock that always "succeeds" would be worse than none: two overlays
// would race for the same socket and one would silently lose. So the honest answer
// here is an error, and ErrNoWindowLayer has already stopped the caller before
// anything is written.

import (
	"errors"
	"os"
)

var errUnsupportedPlatform = errors.New("this platform has no single-instance lock yet")

func flockExclusive(file *os.File) error { return errUnsupportedPlatform }

func flockUnlock(file *os.File) error { return nil }
