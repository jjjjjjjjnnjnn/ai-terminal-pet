//go:build windows

package pipeserver

// The Windows transport: a named pipe with a per-user DACL.
//
// Three facts decide this file, all measured in earlier rounds:
//
//  1. The DACL is D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY) - owner and SYSTEM only, and it
//     stays that way. A process at a lower mandatory integrity level cannot open
//     the pipe (that is handover §29.4, a fact about where the binary lives, not
//     a bug), and widening the ACL to make one machine's low-integrity directory
//     work would weaken every other user's boundary.
//  2. MessageMode is false: reporters write one JSON line, not a framed message.
//  3. The buffers are sized for a 1 MiB payload cap with a tiny reply, which is
//     what the protocol allows.

import (
	"net"

	winio "github.com/Microsoft/go-winio"
)

// pipeBasePath is "no per-user suffix": the container is the pipe namespace.
func pipeBasePath() string {
	return `\\.\pipe\ai-terminal-pet-v1`
}

// userEndpoint appends the per-user token to the pipe namespace. On Windows the
// token is part of the pipe *name*; on unix it is part of the socket file name,
// which is why this is a function rather than string concatenation at the call
// site.
func userEndpoint(base, token string) string {
	if token == "" {
		return base
	}
	return base + "-" + token
}

// newListener binds the pipe.
func newListener(path string) (net.Listener, error) {
	return winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)",
		MessageMode:        false,
		InputBufferSize:    64 * 1024,
		OutputBufferSize:   4 * 1024,
	})
}
