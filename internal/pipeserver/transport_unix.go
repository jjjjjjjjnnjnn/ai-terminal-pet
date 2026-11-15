//go:build !windows

package pipeserver

// The unix transport: one unix-domain socket per user.
//
// The socket path is built in a directory that is already private to this user,
// so the DACL that Windows had to spell out is supplied by the filesystem:
//
//	linux  $XDG_RUNTIME_DIR/ai-terminal-pet-v1-<8hex>.sock  (/run/user/1000/...)
//	darwin $TMPDIR/ai-terminal-pet-v1-<8hex>.sock
//
// Two constraints decided the exact shape (handover §55.4):
//
//   - **The path must stay under 104 characters** (the sockaddr_un limit, and
//     macOS counts bytes, not runes). The per-user token is 8 hex chars and no
//     absolute home path goes in, which is what keeps this short even when the
//     profile lives somewhere deep.
//   - **The token must not move into the path**. The Go and Node implementations
//     hash the same home directory; only the container changes. Putting the
//     profile path in the socket name would make two machines with the same user
//     name collide, and would silently disagree with reporters/protocol.js.
//
// Listen removes a stale socket file first. That is safe because the socket
// exists only while an overlay holds it, and the single-instance mutex is
// acquired before this is called: a file still sitting there belongs to a process
// that was killed hard, not to a live one.

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// socketDirectory is where this platform puts per-user transient files.
func socketDirectory() string {
	if directory := os.Getenv("XDG_RUNTIME_DIR"); directory != "" {
		return directory
	}
	// macOS has no XDG; TMPDIR is per-user (and the sandboxed per-app variant
	// still resolves within the 104-character budget).
	if directory := os.Getenv("TMPDIR"); directory != "" {
		return directory
	}
	return os.TempDir()
}

// pipeBasePath is the full prefix a socket name is appended to.
func pipeBasePath() string {
	return filepath.Join(socketDirectory(), "ai-terminal-pet-v1")
}

func pipeSeparator() string { return "-" }

// socketSuffix completes a per-user socket path.
func socketPath(base, token string) string {
	if token == "" {
		return base + ".sock"
	}
	return base + "-" + token + ".sock"
}

// userEndpoint puts the token into the socket file name.
func userEndpoint(base, token string) string {
	return socketPath(base, token)
}

// newListener binds (or replaces) the per-user socket.
func newListener(path string) (net.Listener, error) {
	if path == "" {
		return nil, errors.New("socket path is empty")
	}
	// A stale file is a leftover, not a second instance: refuse only if
	// something is actually listening on it.
	if _, err := os.Stat(path); err == nil {
		probe, dialErr := net.Dial("unix", path)
		if dialErr == nil {
			_ = probe.Close()
			return nil, fmt.Errorf("%s is already in use", path)
		}
		if removeErr := os.Remove(path); removeErr != nil {
			return nil, fmt.Errorf("remove stale socket %s: %w", path, removeErr)
		}
	}
	// The 104-byte limit is a property of the ABI, so it is checked here rather
	// than discovered as a listen error on a user's machine.
	if len(path) > 104 {
		return nil, fmt.Errorf("socket path %q is %d bytes; the limit is 104", path, len(path))
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	return net.Listen("unix", path)
}
