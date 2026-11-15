package pipeserver

// ROADMAP §3 L3 says: quantify the connection churn before touching it.
// The claim to check is "about 80k short-lived goroutines in ten hours", which
// was inferred from goroutine numbers in hang.log. Two things are measurable
// without adding a counter to the server (a counter in production code would
// move the overlay's vendored source digest and force a release):
//
//   - the cost of ONE delivery round trip, which is what a reporter pays per
//     host event (below);
//   - the event rate on this machine, measured as snapshot writes per minute and
//     quoted in docs/ROADMAP.md §3.
//
// Multiply the two and the churn stops being a mystery. This file is *_test.go
// on purpose: test sources are not in the vendored digest, so recording the
// measurement costs no version bump.
//
//	go test ./internal/pipeserver -run XXX -bench Delivery -benchtime 300x

import (
	"io"
	"net"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// BenchmarkDeliveryRoundTrip measures one complete one-line exchange over an
// in-memory pipe pair: the reporter's write plus the server's read and `OK`.
// It is the in-process floor - the real path adds a dial - but it is the part
// that decides how many connections per second the design can absorb.
func BenchmarkDeliveryRoundTrip(b *testing.B) {
	payload := deliveryBenchPayload()
	serverConnection, clientConnection := net.Pipe()
	server := &Server{
		active: make(map[io.Closer]struct{}),
		update: func(session.Session) {},
	}
	if !server.track(serverConnection, true) {
		b.Fatal("failed to register the bench connection")
	}
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		exchangeDone := make(chan struct{})
		go func() {
			defer close(exchangeDone)
			if _, err := clientConnection.Write(payload); err != nil {
				return
			}
			// The server writes the full "OK\n": a reader that stops at two
			// bytes leaves that write blocked on net.Pipe's unbuffered
			// rendezvous, which is a deadlock, not a slow benchmark - it
			// cost one debug cycle here and is worth the comment.
			acknowledgement := make([]byte, 3)
			io.ReadFull(clientConnection, acknowledgement)
		}()
		// track/read/Done is the production triple: read releases its own
		// registration in the deferred call.
		server.read(serverConnection)
		if !server.track(serverConnection, true) {
			b.Fatal("server stopped accepting after Close")
		}
		<-exchangeDone
	}
	b.StopTimer()
	_ = clientConnection.Close()
}

func deliveryBenchPayload() []byte {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return []byte(`{"schemaVersion":"ai-terminal-pet.v1","source":"bench","sessionId":"bench-1",` +
		`"parentSessionId":null,"pid":5000,"parentPid":94756,"project":"C:\\work","title":"benchmark",` +
		`"model":"provider/model","state":"working","reason":"step_streamed","startedAt":"` + now +
		`","updatedAt":"` + now + `","tool":"shell","terminal":"bench"}` + "\n")
}

// TestDeliveryRoundTripThroughTheRealRead is the same exchange asserted rather
// than measured: it proves the harness above exercises the production read path
// (parse, update, `OK`) and not a stand-in.
func TestDeliveryRoundTripThroughTheRealRead(t *testing.T) {
	updated := make(chan session.Session, 1)
	serverConnection, clientConnection := net.Pipe()
	defer clientConnection.Close()
	server := &Server{
		active: make(map[io.Closer]struct{}),
		update: func(value session.Session) { updated <- value },
	}
	if !server.track(serverConnection, true) {
		t.Fatal("failed to register the test connection")
	}
	go server.read(serverConnection)
	if _, err := clientConnection.Write(deliveryBenchPayload()); err != nil {
		t.Fatalf("write: %v", err)
	}
	acknowledgement := make([]byte, 2)
	if _, err := io.ReadFull(clientConnection, acknowledgement); err != nil {
		t.Fatalf("acknowledgement: %v", err)
	}
	if string(acknowledgement) != "OK" {
		t.Fatalf("acknowledgement = %q, want OK", acknowledgement)
	}
	select {
	case value := <-updated:
		if value.SessionID != "bench-1" {
			t.Fatalf("session id = %q", value.SessionID)
		}
	case <-time.After(time.Second):
		t.Fatal("the payload never reached update")
	}
}
