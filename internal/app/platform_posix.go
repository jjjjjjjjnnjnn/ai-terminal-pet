//go:build !windows

package app

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/platform"
)

var errAlreadyRunning = errors.New("AI Terminal Pet is already running")

func newInstance(name string) (*lockInstance, error) {
	directory := home.Directory()
	if directory == "" {
		return nil, errors.New("this platform has no data directory")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, name+".lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := flockExclusive(file); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, errAlreadyRunning
		}
		return nil, err
	}
	return &lockInstance{file: file}, nil
}

type lockInstance struct {
	file *os.File
}

func (i *lockInstance) Close() {
	if i == nil || i.file == nil {
		return
	}
	_ = flockUnlock(i.file)
	_ = i.file.Close()
	i.file = nil
}

func newPresence() *platform.Presence {
	return platform.NewPresence(platform.DefaultPresenceTTL)
}

func hideConsole() {}
