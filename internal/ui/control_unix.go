//go:build linux || darwin

package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func controlSocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.Getenv("TMPDIR")
	}
	if dir == "" {
		dir = os.TempDir()
	}
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return filepath.Join(dir, "ai-terminal-pet-ctl.sock")
	}
	norm := strings.ToLower(strings.TrimRight(homeDir, `\/`))
	sum := sha256.Sum256([]byte(norm))
	name := "ai-terminal-pet-ctl-" + hex.EncodeToString(sum[:4]) + ".sock"
	path := filepath.Join(dir, name)
	// Keep within BSD/macOS sockaddr_un.sun_path 104-byte limit (103 chars + NUL).
	if len(path) > 103 {
		short := filepath.Join("/tmp", "petctl-"+hex.EncodeToString(sum[:4])+".sock")
		if len(short) <= 103 {
			return short
		}
	}
	return path
}

func postControlMessage(action ControlAction) bool {
	path := controlSocketPath()
	conn, err := net.DialTimeout("unix", path, 250*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(250 * time.Millisecond))
	msg := []byte{byte('0' + int(action)), '\n'}
	if _, err := conn.Write(msg); err != nil {
		return false
	}
	reply := make([]byte, 8)
	n, err := conn.Read(reply)
	if err != nil && err != io.EOF {
		return false
	}
	return strings.TrimSpace(string(reply[:n])) == "OK"
}

func startControlListener(dispatch func(ControlAction)) (io.Closer, error) {
	path := controlSocketPath()
	if _, err := os.Stat(path); err == nil {
		if probe, dErr := net.DialTimeout("unix", path, 100*time.Millisecond); dErr == nil {
			_ = probe.Close()
		} else {
			_ = os.Remove(path)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, aErr := listener.Accept()
			if aErr != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(250 * time.Millisecond))
				buf := make([]byte, 16)
				n, _ := c.Read(buf)
				if n > 0 && buf[0] >= '0' && buf[0] <= '4' {
					action := ControlAction(buf[0] - '0')
					if _, known := controlActionIsKnown(action); known {
						dispatch(action)
						_, _ = c.Write([]byte("OK\n"))
						return
					}
				}
				_, _ = c.Write([]byte("ERR\n"))
			}(conn)
		}
	}()
	return &controlListenerCloser{listener: listener, path: path}, nil
}

type controlListenerCloser struct {
	listener net.Listener
	path     string
}

func (c *controlListenerCloser) Close() error {
	err := c.listener.Close()
	_ = os.Remove(c.path)
	return err
}
