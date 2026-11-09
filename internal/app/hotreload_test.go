package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchConfigNotifiesOnlyOnChange(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.json")
	if err := os.WriteFile(path, []byte(`{"language":"en"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := make(chan struct{}, 8)
	go watchConfigWithInterval(ctx, path, func() {
		select {
		case changes <- struct{}{}:
		default:
		}
	}, 10*time.Millisecond)

	// An untouched file must not wake the UI over and over.
	select {
	case <-changes:
		t.Fatal("unchanged config triggered a reload")
	case <-time.After(60 * time.Millisecond):
	}

	// Force a distinct mtime; some filesystems have coarse timestamps.
	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte(`{"language":"zh-CN"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-changes:
	case <-time.After(time.Second):
		t.Fatal("edited config was not hot-reloaded")
	}

	// Repeated ticks on the now-stable file must not re-notify.
	drained := false
	for {
		select {
		case <-changes:
			drained = true
		default:
			if drained {
				t.Fatal("a single edit triggered more than one reload")
			}
			return
		}
	}
}

func TestWatchConfigSurvivesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.json")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	notified := make(chan struct{}, 1)
	watchConfigWithInterval(ctx, path, func() { notified <- struct{}{} }, 10*time.Millisecond)
	select {
	case <-notified:
		t.Fatal("a missing config file must not trigger a reload")
	default:
	}
}
