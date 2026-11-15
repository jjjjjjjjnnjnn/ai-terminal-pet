//go:build linux

package app

// The flock primitives, in the two files their build tags require.
//
// flock is one syscall with different names on the two unix platforms this project
// targets (syscall.Flock exists on linux and darwin alike, but keeping the call
// behind a named helper is what makes the intent visible next to the reason for
// using it: see newInstance).

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
