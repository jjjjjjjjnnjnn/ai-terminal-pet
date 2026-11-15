//go:build !windows

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func watchShutdownSignals(ctx context.Context, shutdown func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(ch)
	select {
	case <-ctx.Done():
	case <-ch:
		shutdown()
	}
}
