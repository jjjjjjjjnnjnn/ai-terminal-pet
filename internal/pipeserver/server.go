package pipeserver

// The pipe server's protocol half, with no platform code in it.
//
// Everything here was already transport-neutral: read one JSON line with a
// deadline, parse it, hand it to the store, write exactly "OK" or "ERR", and cap
// the fan-out at 32 connections. The only line that ever touched Win32 was
// winio.ListenPipe, and handover §55.4 measured the rest of the file as 187 lines
// of "与传输无关". This is that file, and newListener supplies the one platform
// verb each OS needs.
//
// The name derivation is the cross-language contract and MUST NOT move with the
// transport: first 4 bytes of sha256(lower-cased home, trailing separators
// trimmed), rendered as 8 hex chars. The container changes per platform (a
// named pipe on Windows, a unix socket elsewhere), the token does not -
// reporters/protocol.js computes the same 8 characters, and
// TestPipePathForUserMatchesReporterAlgorithm pins both sides.

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
)

const (
	// maxConnections bounds the fan-out. Reporters are short-lived and dial
	// one at a time, so 32 is already an order of magnitude above the
	// realistically concurrent set; the cap exists so a misbehaving client
	// cannot make the overlay allocate without limit.
	maxConnections = 32
	// connectionReadTimeout is per connection. A reporter that connects and
	// says nothing must not hold a slot: the overlay is a UI process and its
	// goroutine budget is not a client's to spend.
	connectionReadTimeout = 5 * time.Second
)

var errPayloadTooLarge = errors.New("pipe payload too large")

// Server accepts reporter connections and forwards parsed payloads to update.
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

// New binds the per-user endpoint and returns a server that is not yet serving.
//
// Preemption note (Windows): several listeners may share one pipe name; full
// exclusivity would need FILE_FLAG_FIRST_PIPE_INSTANCE, which go-winio does not
// expose. The mitigation is ordering - app.Run acquires the single-instance mutex
// BEFORE New is called. On unix the socket file itself is the lock, which is
// stronger, but the ordering is kept identical so both platforms behave the same
// way when a second instance races.
func New(path string, update func(session.Session)) (*Server, error) {
	if path == "" {
		return nil, errors.New("pipe path is empty")
	}
	if update == nil {
		return nil, errors.New("pipe update callback is empty")
	}
	listener, err := newListener(path)
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
	// Half-close contract: the client sends one JSON line then half-closes, and
	// the reply MUST be written before the FIN arrives. Any non-exact-OK reply
	// surfaces as EBADREPLY on the client, which retries until its deadline.
	defer func() {
		s.track(connection, false)
		_ = connection.Close()
		s.handlers.Done()
	}()
	line, err := readPayload(connection, connectionReadTimeout)
	if err != nil || len(line) == 0 {
		// An immediate EOF is not an error: --doctor connects and closes at
		// once, and treating that as a payload would make the overlay show a
		// phantom session.
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

// PipePathForUser returns the per-user endpoint so sessions of different users
// never share one. The token is the cross-language contract; the container is
// whatever pipeBasePath returns on this platform.
func PipePathForUser() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return pipeBasePath()
	}
	return userEndpoint(pipeBasePath(), pipeScopeToken(home))
}

// pipeScopeToken is the 8 hex characters both implementations agree on.
func pipeScopeToken(home string) string {
	normalized := strings.ToLower(strings.TrimRight(home, `\/`))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:4])
}
