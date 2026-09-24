package pipeserver

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ai-terminal-pet/internal/session"
	winio "github.com/Microsoft/go-winio"
)

const (
	PipePath              = `\\.\pipe\ai-terminal-pet-v1`
	maxConnections        = 32
	connectionReadTimeout = 5 * time.Second
)

// PipePathForUser returns a per-user pipe name so sessions of different
// Windows users never share a pipe instance. The suffix is the first 4 bytes
// (8 hex chars) of SHA-256 over the lower-cased home directory, mirrored by
// reporters/protocol.js pipeScopeToken (cross-language golden test locked).
//
// Preemption note: Windows allows several listeners on the same pipe name.
// Full exclusivity would need CreateNamedPipe with FILE_FLAG_FIRST_PIPE_INSTANCE,
// which go-winio does not expose. Current mitigation is ordering: app.Run
// acquires the Local\AI-Terminal-Pet-v1-SingleInstance mutex BEFORE New() is
// called, so at most one server per logon session reaches ListenPipe.
func PipePathForUser() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return PipePath
	}
	normalized := strings.ToLower(strings.TrimRight(home, `\/`))
	sum := sha256.Sum256([]byte(normalized))
	return PipePath + "-" + hex.EncodeToString(sum[:4])
}

var errPayloadTooLarge = errors.New("pipe payload too large")

type Server struct {
	listener  net.Listener
	update    func(session.Session)
	mutex     sync.Mutex
	active    map[io.Closer]struct{}
	handlers  sync.WaitGroup
	closed    atomic.Bool
	closeOnce sync.Once
	closeDone chan struct{}
}

func New(path string, update func(session.Session)) (*Server, error) {
	if path == "" {
		return nil, errors.New("pipe path is empty")
	}
	if update == nil {
		return nil, errors.New("pipe update callback is empty")
	}
	listener, err := winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)",
		MessageMode:        false,
		InputBufferSize:    64 * 1024,
		OutputBufferSize:   4 * 1024,
	})
	if err != nil {
		return nil, err
	}
	return &Server{
		listener:  listener,
		update:    update,
		active:    make(map[io.Closer]struct{}),
		closeDone: make(chan struct{}),
	}, nil
}

func (s *Server) Serve(ctx context.Context) error {
	if s.closed.Load() {
		<-s.closeDone
		return nil
	}
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			s.Close()
		case <-stop:
		}
	}()
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() || ctx.Err() != nil {
				<-s.closeDone
				return nil
			}
			return err
		}
		if !s.track(connection, true) {
			_ = connection.Close()
			if s.closed.Load() {
				<-s.closeDone
				return nil
			}
			continue
		}
		go s.read(connection)
	}
}

func (s *Server) Close() {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		_ = s.listener.Close()
		s.mutex.Lock()
		connections := make([]io.Closer, 0, len(s.active))
		for connection := range s.active {
			connections = append(connections, connection)
		}
		s.mutex.Unlock()
		for _, connection := range connections {
			_ = connection.Close()
		}
		s.handlers.Wait()
		close(s.closeDone)
	})
}

func (s *Server) read(connection net.Conn) {
	// Half-close contract (Windows named pipes have weak socket.end semantics):
	// the client sends one JSON line then half-closes; the server MUST write
	// OK/ERR before the FIN arrives. Any non-exact-OK reply surfaces as
	// EBADREPLY on the client, which retries until its deadline. Verified.
	defer func() {
		s.track(connection, false)
		_ = connection.Close()
		s.handlers.Done()
	}()
	line, err := readPayload(connection, connectionReadTimeout)
	if err != nil || len(line) == 0 {
		return
	}
	value, err := session.ParsePayload(line, time.Now())
	if err != nil {
		_, _ = connection.Write([]byte("ERR\n"))
		return
	}
	s.update(value)
	_, _ = connection.Write([]byte("OK\n"))
}

func readPayload(connection net.Conn, timeout time.Duration) ([]byte, error) {
	if err := connection.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	line, err := bufio.NewReaderSize(connection, session.MaxPayloadBytes+1).ReadSlice('\n')
	if len(line) > session.MaxPayloadBytes {
		return nil, errPayloadTooLarge
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return line, nil
}

func (s *Server) track(connection io.Closer, add bool) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if add {
		if s.closed.Load() || len(s.active) >= maxConnections {
			return false
		}
		s.active[connection] = struct{}{}
		s.handlers.Add(1)
		return true
	}
	delete(s.active, connection)
	return true
}
