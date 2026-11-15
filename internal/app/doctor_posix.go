//go:build !windows

package app

import (
	"net"
	"os/exec"
	"time"
)

func probePipe(path string, timeout time.Duration) error {
	connection, err := net.DialTimeout("unix", path, timeout)
	if err != nil {
		return err
	}
	return connection.Close()
}

func hideChildConsole(_ *exec.Cmd) {}
