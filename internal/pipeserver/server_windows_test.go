package pipeserver

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

func TestReadPayloadLimit(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{name: "maximum", size: session.MaxPayloadBytes},
		{name: "over maximum", size: session.MaxPayloadBytes + 1, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverConnection, clientConnection := net.Pipe()
			payload := payloadOfSize(t, test.size)
			writeDone := make(chan error, 1)
			go func() {
				_, err := clientConnection.Write(payload)
				if closeErr := clientConnection.Close(); err == nil {
					err = closeErr
				}
				writeDone <- err
			}()

			line, err := readPayload(serverConnection, time.Second)
			_ = serverConnection.Close()
			if writeErr := <-writeDone; writeErr != nil {
				t.Fatal(writeErr)
			}
			if test.wantErr {
				if !errors.Is(err, errPayloadTooLarge) {
					t.Fatalf("expected payload limit error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(line) != session.MaxPayloadBytes {
				t.Fatalf("read %d bytes", len(line))
			}
			if _, err := session.ParsePayload(line, time.Date(2026, 9, 24, 10, 2, 0, 0, time.UTC)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReadAcknowledgesValidPayload(t *testing.T) {
	serverConnection, clientConnection := net.Pipe()
	defer clientConnection.Close()
	updated := make(chan session.Session, 1)
	server := &Server{
		update: func(value session.Session) { updated <- value },
		active: make(map[io.Closer]struct{}),
	}
	if !server.track(serverConnection, true) {
		t.Fatal("failed to track connection")
	}
	go server.read(serverConnection)
	payload := payloadOfSize(t, 512)
	if _, err := clientConnection.Write(payload); err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(clientConnection)
	if err != nil {
		t.Fatal(err)
	}
	if string(response) != "OK\n" {
		t.Fatalf("unexpected response: %q", response)
	}
	select {
	case value := <-updated:
		if value.Source != "test" || value.State != session.StateIdle {
			t.Fatalf("unexpected update: %#v", value)
		}
	case <-time.After(time.Second):
		t.Fatal("update callback did not run")
	}
}

func TestReadRejectsInvalidPayload(t *testing.T) {
	serverConnection, clientConnection := net.Pipe()
	defer clientConnection.Close()
	server := &Server{
		update: func(session.Session) { t.Fatal("update must not run for invalid payload") },
		active: make(map[io.Closer]struct{}),
	}
	if !server.track(serverConnection, true) {
		t.Fatal("failed to track connection")
	}
	go server.read(serverConnection)
	if _, err := clientConnection.Write([]byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(clientConnection)
	if err != nil {
		t.Fatal(err)
	}
	if string(response) != "ERR\n" {
		t.Fatalf("unexpected response: %q", response)
	}
}

func TestPipePathForUserMatchesReporterAlgorithm(t *testing.T) {
	t.Setenv("USERPROFILE", `C:\Users\rongj`)
	want := PipePath + "-42ce7d78"
	if got := PipePathForUser(); got != want {
		t.Fatalf("pipe path mismatch: got %q want %q", got, want)
	}
}

func TestReadPayloadDeadline(t *testing.T) {
	serverConnection, clientConnection := net.Pipe()
	defer serverConnection.Close()
	defer clientConnection.Close()

	started := time.Now()
	_, err := readPayload(serverConnection, 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected read deadline")
	}
	var timeout net.Error
	if !errors.As(err, &timeout) || !timeout.Timeout() {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("read deadline took %s", elapsed)
	}
}

func TestConnectionLimit(t *testing.T) {
	server := &Server{active: make(map[io.Closer]struct{})}
	connections := make([]io.Closer, 0, maxConnections)
	defer func() {
		for _, connection := range connections {
			server.track(connection, false)
			server.handlers.Done()
		}
	}()

	for range maxConnections {
		connection := io.NopCloser(strings.NewReader(""))
		if !server.track(connection, true) {
			t.Fatal("connection below the limit was rejected")
		}
		connections = append(connections, connection)
	}
	if server.track(io.NopCloser(strings.NewReader("")), true) {
		t.Fatal("connection above the limit was accepted")
	}
}

func TestCloseWaitsForHandlers(t *testing.T) {
	listener := &blockingListener{closed: make(chan struct{})}
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseUpdate := func() {
		releaseOnce.Do(func() {
			close(release)
		})
	}
	server := &Server{
		listener: listener,
		update: func(session.Session) {
			close(started)
			<-release
		},
		active:    make(map[io.Closer]struct{}),
		closeDone: make(chan struct{}),
	}
	serverConnection, clientConnection := net.Pipe()
	defer server.Close()
	defer releaseUpdate()
	defer clientConnection.Close()

	if !server.track(serverConnection, true) {
		t.Fatal("failed to track connection")
	}
	go server.read(serverConnection)
	payload := payloadOfSize(t, 512)
	writeDone := make(chan error, 1)
	go func() {
		_, err := clientConnection.Write(payload)
		writeDone <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("update callback did not start")
	}
	closeReturned := make(chan struct{})
	go func() {
		server.Close()
		close(closeReturned)
	}()
	select {
	case <-closeReturned:
		t.Fatal("Close returned before the handler completed")
	case <-time.After(50 * time.Millisecond):
	}
	releaseUpdate()
	select {
	case <-closeReturned:
	case <-time.After(time.Second):
		t.Fatal("Close did not return after the handler completed")
	}
	if err := <-writeDone; err != nil {
		t.Fatal(err)
	}
	select {
	case <-server.closeDone:
	default:
		t.Fatal("shutdown completion was not recorded")
	}
}

func payloadOfSize(t *testing.T, size int) []byte {
	t.Helper()
	const prefix = `{"schemaVersion":"ai-terminal-pet.v1",`
	const suffix = "\"source\":\"test\",\"sessionId\":\"test\",\"pid\":1,\"state\":\"idle\",\"startedAt\":\"2026-09-24T10:00:00Z\",\"updatedAt\":\"2026-09-24T10:01:00Z\"}\n"
	padding := size - len(prefix) - len(suffix)
	if padding < 0 {
		t.Fatalf("payload size %d is too small", size)
	}
	payload := make([]byte, 0, size)
	payload = append(payload, prefix...)
	payload = append(payload, bytes.Repeat([]byte{' '}, padding)...)
	return append(payload, suffix...)
}

type blockingListener struct {
	closed chan struct{}
	once   sync.Once
}

func (l *blockingListener) Accept() (net.Conn, error) {
	<-l.closed
	return nil, net.ErrClosed
}

func (l *blockingListener) Close() error {
	l.once.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *blockingListener) Addr() net.Addr {
	return testAddress("pipe")
}

type testAddress string

func (a testAddress) Network() string {
	return string(a)
}

func (a testAddress) String() string {
	return string(a)
}
