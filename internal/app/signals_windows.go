//go:build windows

package app

import (
	"context"
	"os"
	"os/signal"
)

func watchShutdownSignals(ctx context.Context, shutdown func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer signal.Stop(ch)
	select {
	case <-ctx.Done():
	case <-ch:
		shutdown()
	}
}
