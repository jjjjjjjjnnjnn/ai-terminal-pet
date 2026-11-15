//go:build windows

package app

import (
	"context"
	"os/exec"
	"syscall"
	"time"

	winio "github.com/Microsoft/go-winio"
)

func probePipe(path string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	connection, err := winio.DialPipeContext(ctx, path)
	if err != nil {
		return err
	}
	return connection.Close()
}

func hideChildConsole(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
