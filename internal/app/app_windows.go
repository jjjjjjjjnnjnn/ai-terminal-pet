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
//
// 1.0.2 rather than 1.0.1: the look is now a setting. Three named palettes
// (slate stays the default, so nothing changes unless you ask), a switch that
// tints each row's source name by terminal, and "start hidden" for people who
// want the strip in the tray until a session needs them.
const Version = "1.0.2"

func Run() error {
	return RunWith(false)
}

// RunWith starts the overlay, optionally opening the settings window once the
// panel is up. `--settings` needs this: when nothing is running yet there is no
// window to post to, and silently starting a panel without the settings window
// the user asked for would make the flag a lie.
func RunWith(openSettings bool) error {
	// Single-instance gate doubles as pipe-preemption mitigation: only the
	// mutex holder proceeds to pipeserver.New/ListenPipe (see PipePathForUser).
	// A second instance does not just exit silently - it pokes the live
	// overlay so a user double-clicking the executable sees it appear
	// instead of nothing happening.
	instance, err := platform.AcquireInstance(`Local\AI-Terminal-Pet-v1-SingleInstance`)
	if errors.Is(err, platform.ErrAlreadyRunning) {
		if openSettings {
			ui.Control(ui.ControlSettings)
		} else {
			ui.Control(ui.ControlActivate)
		}
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
	// scripts/install.ps1 force-stops the overlay before every upgrade, so a longer
	// window would leave a visible orphan behind after each reinstall.
	config.SweepTemporaryFiles(directory, 2*time.Minute)
	runtimeDirectory := filepath.Join(directory, "runtime")
	// One cached OS view drives snapshot retention, liveness and cleanup for
	// the whole process. It replaces the old model where "no reporter push for
	// 15/30 seconds" was taken to mean the session had ended.
	presence := platform.NewPresence(platform.DefaultPresenceTTL)
	scanner := store.SnapshotScanner{
		Directory: runtimeDirectory,
		TTL:       store.DefaultSnapshotTTL,
		Now:       time.Now,
		Alive:     presence.ProcessAlive,
		KeepAlive: presence.SessionAlive,
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
	// config.json gets the same treatment. The settings window writes on every
	// click and once per slider drag, and an inline write there would put a
	// file write on the UI thread for a window that is already repainting.
	settingsSaver := newConfigSaver(configPath)
	defer settingsSaver.flush()

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
			Notifications:         settings.Notifications,
			Sound:                 settings.Sound,
			SoundVolume:           settings.SoundVolume,
			SoundScheme:           settings.SoundScheme,
			SoundOnWaiting:        settings.SoundOnWaiting,
			SoundOnError:          settings.SoundOnError,
			NotifyOnFinish:        settings.NotifyOnFinish,
			AlwaysOnTop:           settings.AlwaysOnTop,
			EdgeRetract:           settings.EdgeRetract,
			Language:              settings.Language,
			AutoApprove:           settings.AutoApprove,
			AutoApproveTimeoutSec: settings.AutoApproveTimeoutSec,
			Opacity:               settings.Opacity,
			Theme:                 settings.Theme,
			PerSourceColor:        settings.PerSourceColor,
			HiddenAtLaunch:        settings.HiddenAtLaunch,
		},
		AutoStart:             autostartEnabled,
		AutoApproveTimeoutSec: settings.AutoApproveTimeoutSec,
		OpenSettings:          openSettings,
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
			// Staged, not written: the read, the merge and the write all happen
			// on the saver's goroutine so a click can never block the pump.
			settingsSaver.save(value)
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
	// Snapshot files are the durable record of which conversations exist.
	// seenSnapshots remembers every identity that ever had one, so a file
	// disappearing is an explicit "this session ended" signal - reporters
	// remove their snapshot on close, and writing a `closed` payload has the
	// same effect. Without this, a pipe session whose file was deleted stayed
	// in memory forever, propped up by a long-lived parent pid.
	seenSnapshots := make(map[string]bool)
	apply := func() bool {
		now := time.Now()
		changed := false
		items, present := scanner.ScanWithKeys()
		for _, value := range items {
			changed = values.UpsertSnapshot(value, now) || changed
			seenSnapshots[store.Key(value)] = true
		}
		alive := snapshotAwareAlive(present, seenSnapshots, presence.SessionAlive)
		// Reconcile, not Cleanup: a live terminal keeps its session no matter
		// how long the reporter has been quiet. Reconcile only bumps the
		// revision when a session is really added or removed, so a quiet
		// desktop schedules no repaints at all.
		changed = values.Reconcile(now, alive, store.DefaultAbsenceGrace) || changed
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
	// Hot-reload the human-editable config without restarting the overlay.
	// A single Stat every 30s is cheaper than any Windows file watcher and
	// needs no extra handle or thread.
	go watchConfig(ctx, configPath, func() {
		current := config.Load(configPath)
		// A settings-window change lands on disk and bumps mtime; reloading it
		// straight back would echo the write. Only a change this process did
		// not make is worth applying.
		if settingsSaver.echo(current) {
			return
		}
		window.QueueSettings(ui.Settings{
			Notifications:         current.Notifications,
			Sound:                 current.Sound,
			SoundVolume:           current.SoundVolume,
			SoundScheme:           current.SoundScheme,
			SoundOnWaiting:        current.SoundOnWaiting,
			SoundOnError:          current.SoundOnError,
			NotifyOnFinish:        current.NotifyOnFinish,
			AlwaysOnTop:           current.AlwaysOnTop,
			EdgeRetract:           current.EdgeRetract,
			Language:              current.Language,
			AutoApprove:           current.AutoApprove,
			AutoApproveTimeoutSec: current.AutoApproveTimeoutSec,
			Opacity:               current.Opacity,
			Theme:                 current.Theme,
			PerSourceColor:        current.PerSourceColor,
			HiddenAtLaunch:        current.HiddenAtLaunch,
		})
	})
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

// snapshotAwareAlive combines the durable disk record with OS liveness.
//
// The disk snapshot IS the conversation's existence. A reporter that finishes
// writes `closed` or removes its file; until it does, the session is real even
// when the OS evidence is weak (a long-lived host pid whose tab closed, a
// stale window title, a quiet CLI). Deferring to Presence in that case made
// the overlay evict a session that `--json` still reported from the very same
// file, emptied HasAvailable(), and eventually self-destructed the window.
//
// A session that once had a snapshot but no longer does has been explicitly
// ended by its reporter; no amount of "the parent pid is still alive" should
// resurrect it. Sessions that never had a snapshot (pipe-only reporters) keep
// the ordinary OS liveness rules so they are not penalised.
func snapshotAwareAlive(present, seen map[string]bool, fallback func(session.Session) bool) func(session.Session) bool {
	return func(value session.Session) bool {
		if value.State == session.StateClosed {
			return false
		}
		identity := store.Key(value)
		if seen[identity] && !present[identity] {
			return false
		}
		if present[identity] {
			return true
		}
		return fallback(value)
	}
}

// scanInterval is deliberately slow. Presence caches the OS view for 10s,
// reporters push state changes immediately over the pipe, and the panel
// already skips identical frames. Five seconds is enough to notice a terminal
// that started without a reporter event, without waking the process twice a
// second forever.
const scanInterval = 5 * time.Second

func scan(ctx context.Context, scanner store.SnapshotScanner, apply func() bool, refresh func()) {
	ticker := time.NewTicker(scanInterval)
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

// configWatchInterval is intentionally slow. Hand edits to config.json are
// rare and a 30s delay is imperceptible, while a filesystem watcher would add
// a handle and a thread to a process whose whole job is to stay lightweight.
const configWatchInterval = 30 * time.Second

func watchConfig(ctx context.Context, path string, onChange func()) {
	watchConfigWithInterval(ctx, path, onChange, configWatchInterval)
}

func watchConfigWithInterval(ctx context.Context, path string, onChange func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var stamp time.Time
	if info, err := os.Stat(path); err == nil {
		stamp = info.ModTime()
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(path)
			if err != nil || !info.ModTime().After(stamp) {
				continue
			}
			stamp = info.ModTime()
			onChange()
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

// configSaver is the same coalescing writer for config.json.
//
// The settings window publishes on every toggle and once per slider drag. Those
// calls arrive on the UI thread, and on machines with an AV write hook a single
// config write can block for hundreds of milliseconds - long enough to stall the
// message pump and make the window feel dead. Both the read and the write are
// therefore moved off the UI thread: the saver merges the new preferences into
// whatever is currently on disk, which is also what keeps a hand-edited key this
// build does not know about from being dropped.
type configSaver struct {
	path   string
	wake   chan struct{}
	mutex  sync.Mutex
	latest ui.Settings
	dirty  bool
	// written is the exact value this process last persisted. The reload
	// watcher compares against it so a write here is not echoed straight back
	// into the UI, which would briefly revert the switch the user just flipped.
	written    config.Config
	hasWritten bool
}

func newConfigSaver(path string) *configSaver {
	saver := &configSaver{path: path, wake: make(chan struct{}, 1)}
	go saver.loop()
	return saver
}

func (s *configSaver) save(value ui.Settings) {
	s.mutex.Lock()
	s.latest = value
	s.dirty = true
	s.mutex.Unlock()
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *configSaver) loop() {
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
			s.write(value)
		}
	}
}

func (s *configSaver) write(value ui.Settings) {
	current := config.Load(s.path)
	applyUISettings(&current, value)
	if err := config.Save(s.path, current); err != nil {
		return
	}
	s.mutex.Lock()
	s.written = current
	s.hasWritten = true
	s.mutex.Unlock()
}

// echo reports whether value is exactly what this process last wrote.
func (s *configSaver) echo(value config.Config) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.hasWritten && s.written == value
}

func (s *configSaver) flush() {
	s.mutex.Lock()
	value, pending := s.latest, s.dirty
	s.dirty = false
	s.mutex.Unlock()
	if pending {
		s.write(value)
	}
}

// applyUISettings copies the preferences the settings window owns onto a config
// value. Fields the window cannot change are left alone.
func applyUISettings(target *config.Config, value ui.Settings) {
	target.Notifications = value.Notifications
	target.Sound = value.Sound
	target.SoundVolume = value.SoundVolume
	target.SoundScheme = value.SoundScheme
	target.SoundOnWaiting = value.SoundOnWaiting
	target.SoundOnError = value.SoundOnError
	target.NotifyOnFinish = value.NotifyOnFinish
	target.AlwaysOnTop = value.AlwaysOnTop
	target.EdgeRetract = value.EdgeRetract
	target.Language = value.Language
	target.AutoApprove = value.AutoApprove
	target.AutoApproveTimeoutSec = value.AutoApproveTimeoutSec
	// ValidTheme is the second line of defence: the ui package already
	// normalises an unknown name to slate, but config.json is a documented
	// hand-editable file, and a value written by hand must not read back as
	// a palette that does not exist.
	target.Theme = config.ValidTheme(value.Theme)
	target.PerSourceColor = value.PerSourceColor
	target.HiddenAtLaunch = value.HiddenAtLaunch
	if value.Opacity > 0 {
		target.Opacity = value.Opacity
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
// process.
//
// It uses the same existence model as the overlay: a snapshot with an old
// timestamp is still reported while its host terminal is observable, so a
// quiet session does not vanish from a script's view either.
func Snapshot() (Report, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return Report{}, errors.New("LOCALAPPDATA is not set")
	}
	presence := platform.NewPresence(platform.DefaultPresenceTTL)
	scanner := store.SnapshotScanner{
		Directory: filepath.Join(localAppData, "AI-Terminal-Pet", "runtime"),
		TTL:       store.DefaultSnapshotTTL,
		Now:       time.Now,
		Alive:     presence.ProcessAlive,
		KeepAlive: presence.SessionAlive,
	}
	values := store.New()
	now := time.Now()
	for _, value := range scanner.Scan() {
		values.UpsertSnapshot(value, now)
	}
	values.Reconcile(now, presence.SessionAlive, store.DefaultAbsenceGrace)
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
