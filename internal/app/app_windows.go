package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"ai-terminal-pet/internal/ui"
)

func Run() error {
	// Single-instance gate doubles as pipe-preemption mitigation: only the
	// mutex holder proceeds to pipeserver.New/ListenPipe (see PipePathForUser).
	// A second instance exits silently with nil so double-click/launch storms
	// never surface an error dialog.
	instance, err := platform.AcquireInstance(`Local\AI-Terminal-Pet-v1-SingleInstance`)
	if errors.Is(err, platform.ErrAlreadyRunning) {
		return nil
	}
	if err != nil {
		return err
	}
	defer instance.Close()
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return errors.New("LOCALAPPDATA is not set")
	}
	platform.HideConsole()
	runtimeDirectory := filepath.Join(localAppData, "AI-Terminal-Pet", "runtime")
	scanner := store.SnapshotScanner{
		Directory: runtimeDirectory,
		TTL:       store.DefaultSnapshotTTL,
		Now:       time.Now,
		Alive:     platform.ProcessAlive,
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	values := store.New()
	window := ui.New(values)
	apply := func() bool {
		changed := false
		for _, value := range scanner.Scan() {
			changed = values.UpsertSnapshot(value, time.Now()) || changed
		}
		changed = values.Cleanup(time.Now(), platform.ProcessAlive, store.DefaultSnapshotTTL) || changed
		return changed
	}
	apply()
	server, err := pipeserver.New(pipeserver.PipePathForUser(), func(value session.Session) {
		changed := values.UpsertPipe(value, time.Now())
		if changed {
			window.Refresh()
			if value.State != session.StateClosed {
				window.Activate()
			}
		}
	})
	if err != nil {
		return fmt.Errorf("pipe server: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan struct{})
	serverErr := make(chan error, 1)
	go func() {
		defer close(serverDone)
		if serveErr := server.Serve(ctx); serveErr != nil {
			select {
			case serverErr <- serveErr:
			default:
			}
			window.Close()
		}
	}()
	defer func() {
		cancel()
		server.Close()
		<-serverDone
	}()
	go scan(ctx, scanner, apply, window.Refresh)
	if err := window.Run(); err != nil {
		return fmt.Errorf("window: %w", err)
	}
	select {
	case err := <-serverErr:
		return fmt.Errorf("pipe server: %w", err)
	default:
		return nil
	}
}

func scan(ctx context.Context, scanner store.SnapshotScanner, apply func() bool, refresh func()) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if apply() {
				refresh()
			}
		}
	}
}
