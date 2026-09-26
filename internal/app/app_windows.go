package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ai-terminal-pet/internal/autostart"
	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/platform"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"ai-terminal-pet/internal/ui"
)

// Version is reported by `ai-terminal-pet --version`.
const Version = "0.5.0"

func Run() error {
	// Single-instance gate doubles as pipe-preemption mitigation: only the
	// mutex holder proceeds to pipeserver.New/ListenPipe (see PipePathForUser).
	// A second instance does not just exit silently - it pokes the live
	// overlay so a user double-clicking the executable sees it appear
	// instead of nothing happening.
	instance, err := platform.AcquireInstance(`Local\AI-Terminal-Pet-v1-SingleInstance`)
	if errors.Is(err, platform.ErrAlreadyRunning) {
		ui.Control(ui.ControlActivate)
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
	directory := filepath.Join(localAppData, "AI-Terminal-Pet")
	// Reclaim atomic-write temporaries orphaned by a hard kill. Two minutes is
	// already orders of magnitude longer than the write it guards, and the
	// overlay is single-instance, so nothing else can hold one in flight.
	// install.ps1 force-stops the overlay before every upgrade, so a longer
	// window would leave a visible orphan behind after each reinstall.
	config.SweepTemporaryFiles(directory, 2*time.Minute)
	runtimeDirectory := filepath.Join(directory, "runtime")
	scanner := store.SnapshotScanner{
		Directory: runtimeDirectory,
		TTL:       store.DefaultSnapshotTTL,
		Now:       time.Now,
		Alive:     platform.ProcessAlive,
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	configPath := filepath.Join(directory, config.FileName)
	statePath := filepath.Join(directory, config.StateFileName)
	settings := config.Load(configPath)
	placement := config.LoadState(statePath)
	// Saves are coalesced onto a background goroutine: a placement write
	// happens right after a drag, and on machines with AV write hooks an
	// inline write could stall the UI thread for hundreds of milliseconds.
	saver := newPlacementSaver(statePath)
	defer saver.flush()

	autostartManager := autostart.New()
	executable, executableErr := os.Executable()
	autostartEnabled := false
	if executableErr == nil {
		// Repair a stale path first. The install path is the only thing in
		// the Run value, so a moved or reinstalled binary would fail silently
		// at logon - the worst possible way to fail. Sync never enables a
		// disabled entry, so this cannot turn autostart on by itself.
		if _, err := autostartManager.Sync(executable); err != nil {
			// Autostart is a convenience, not a prerequisite: a registry that
			// refuses to cooperate must not stop the overlay from running.
			_ = err
		}
		if status, err := autostartManager.Status(); err == nil {
			autostartEnabled = status.Enabled
		}
	}

	values := store.New()
	window := ui.New(values, ui.Options{
		Opacity: uint8(settings.Opacity),
		Settings: ui.Settings{
			Notifications: settings.Notifications,
			Sound:         settings.Sound,
			AlwaysOnTop:   settings.AlwaysOnTop,
			EdgeRetract:   settings.EdgeRetract,
		},
		AutoStart: autostartEnabled,
		Initial: ui.Placement{
			Placed:   placement.Placed,
			X:        placement.X,
			Y:        placement.Y,
			Zoom:     placement.Zoom,
			Expanded: placement.Expanded,
		},
		OnPlacementChanged: func(value ui.Placement) {
			saver.save(config.State{
				Placed:   value.Placed,
				X:        value.X,
				Y:        value.Y,
				Zoom:     value.Zoom,
				Expanded: value.Expanded,
			})
		},
		OnSettingsChanged: func(value ui.Settings) {
			// Reload first so a hand-edited field the tray does not expose
			// (opacity) survives the round trip.
			current := config.Load(configPath)
			current.Notifications = value.Notifications
			current.Sound = value.Sound
			current.AlwaysOnTop = value.AlwaysOnTop
			current.EdgeRetract = value.EdgeRetract
			_ = config.Save(configPath, current)
		},
		OnAutoStartChanged: func(enabled bool) bool {
			if executableErr != nil {
				return false
			}
			if enabled {
				return autostartManager.Enable(executable) == nil
			}
			return autostartManager.Disable() == nil
		},
	})
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

// placementSaver writes state.json from a single background goroutine,
// coalescing bursts down to the newest value.
//
// The UI calls back on drag release, zoom steps and expand toggles. Writing
// inline would run a file write on the UI thread; on this class of machine
// (AV hooks on every file operation) that can block for hundreds of
// milliseconds, which is exactly the kind of UI-thread stall that surfaces as
// AppHang. A single writer also keeps two overlapping saves from racing.
type placementSaver struct {
	path   string
	wake   chan struct{}
	mutex  sync.Mutex
	latest config.State
	dirty  bool
}

func newPlacementSaver(path string) *placementSaver {
	saver := &placementSaver{path: path, wake: make(chan struct{}, 1)}
	go saver.loop()
	return saver
}

func (s *placementSaver) save(value config.State) {
	s.mutex.Lock()
	s.latest = value
	s.dirty = true
	s.mutex.Unlock()
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *placementSaver) loop() {
	for range s.wake {
		for {
			s.mutex.Lock()
			if !s.dirty {
				s.mutex.Unlock()
				break
			}
			value := s.latest
			s.dirty = false
			s.mutex.Unlock()
			_ = config.SaveState(s.path, value)
		}
	}
}

// flush writes any pending placement synchronously. The process is about to
// exit, so a write still queued on the background goroutine might never land.
func (s *placementSaver) flush() {
	s.mutex.Lock()
	value, pending := s.latest, s.dirty
	s.dirty = false
	s.mutex.Unlock()
	if pending {
		_ = config.SaveState(s.path, value)
	}
}

// Report is the machine-readable view printed by `--json`.
type Report struct {
	SchemaVersion string            `json:"schemaVersion"`
	State         session.State     `json:"state"`
	Count         int               `json:"count"`
	GeneratedAt   time.Time         `json:"generatedAt"`
	Sessions      []session.Session `json:"sessions"`
}

// Snapshot reads the on-disk reporter snapshots and returns the same aggregate
// the overlay would show.
//
// It deliberately does not talk to the running overlay. Reading runtime/ works
// whether or not the overlay is alive, needs no IPC, and cannot be blocked by
// the single-instance mutex - which matters because the main use of --json is
// to feed a status line or a script, and those must never hang on a GUI
// process. The trade is freshness: the view is only as current as the
// snapshots, and entries older than store.DefaultSnapshotTTL count as gone,
// exactly as they do in the overlay.
func Snapshot() (Report, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return Report{}, errors.New("LOCALAPPDATA is not set")
	}
	scanner := store.SnapshotScanner{
		Directory: filepath.Join(localAppData, "AI-Terminal-Pet", "runtime"),
		TTL:       store.DefaultSnapshotTTL,
		Now:       time.Now,
		Alive:     platform.ProcessAlive,
	}
	values := store.New()
	now := time.Now()
	for _, value := range scanner.Scan() {
		values.UpsertSnapshot(value, now)
	}
	values.Cleanup(now, platform.ProcessAlive, store.DefaultSnapshotTTL)
	items := values.List()
	if items == nil {
		items = []session.Session{}
	}
	return Report{
		SchemaVersion: session.SchemaVersion,
		State:         session.Aggregate(items),
		Count:         len(items),
		GeneratedAt:   now.UTC(),
		Sessions:      items,
	}, nil
}

// EncodeReport renders a report as JSON. pretty switches to indented output for
// humans; the default single line is what scripts and status lines want.
func EncodeReport(report Report, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(report, "", "  ")
	}
	return json.Marshal(report)
}
