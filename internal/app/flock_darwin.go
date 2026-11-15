//go:build darwin

package app

// The flock primitives on darwin.
//
// Same syscall as linux, kept as a separate file only because the build tag has to
// be: one file cannot be tagged "linux or darwin" without also dragging in every
// other unix. The body is identical on purpose - if the two ever need to differ,
// that difference becomes visible right here instead of inside a shared helper.

import (
	"os"
	"syscall"
)

func flockExclusive(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func flockUnlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
