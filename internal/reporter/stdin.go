package reporter

import (
	"io"
	"time"
)

// ReadStdin reads one hook payload from stdin under the two limits the protocol
// documents: at most MaxStdinBytes, and at most StdinTimeout. A payload over the
// size cap is discarded rather than truncated - half a JSON object is worse
// than no report at all - and the timeout exists because a host that never
// closes stdin must not hang its own hook.
//
// The read happens on a goroutine so the timeout can actually fire: a blocking
// os.Stdin.Read cannot be interrupted, but the process may exit while it is
// still parked. Nothing here writes, so an abandoned goroutine is harmless.
func ReadStdin(stdin io.Reader) []byte {
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(stdin, MaxStdinBytes+1))
		done <- result{data, err}
	}()
	select {
	case res := <-done:
		if res.err != nil || len(res.data) > MaxStdinBytes {
			return nil
		}
		return res.data
	case <-time.After(StdinTimeout):
		return nil
	}
}
