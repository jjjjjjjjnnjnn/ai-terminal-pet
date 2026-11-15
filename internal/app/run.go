package app

// RunWith is the overlay's entry point, and it is deliberately platform-neutral:
// everything that differs per OS arrives through the small interfaces below.
//
// This file is handover §55.6 step 7's centrepiece. Before it, internal/app was
// one 654-line _windows.go whose first half was the whole lifecycle (single
// instance, presence, config, autostart, window, pipe server, watcher) and whose
// second half was a file dump. Almost none of that lifecycle is Windows-specific:
//
//	single instance   "either I am the first or somebody already is"
//	presence          "is this session's terminal still on the machine"
//	hide console      a no-op on POSIX
//	window            ui.Run is the one real platform boundary
//
// So the lifecycle lives here, once, and each platform file supplies:
//
//	newInstance(name)  the single-instance handle
//	newPresence(ttl)   the process/window cache
//	hideConsole()      no-op on POSIX
//	windowRun          through ui, which has its own per-platform entry
//
// The one thing that must NOT be unified is the presence rule itself: §51 and §52
// were both caused by "pid alive == session alive", and §55.4 says explicitly not
// to simplify SessionAlive on a platform without window evidence. A platform that
// can only prove a pid is alive therefore reports that, and its own file says so
// (see platform/presence_darwin.go once it exists) rather than pretending.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ai-terminal-pet/internal/autostart"
	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"ai-terminal-pet/internal/ui"
)

// Run starts the overlay.
func Run() error {
	return RunWith(false)
}

// RunWith starts the overlay, optionally opening the settings window once the
// panel is up. `--settings` needs this: when nothing is running yet there is no
// window to post to, and silently starting a panel without the settings window the
// user asked for would make the flag a lie.
func RunWith(openSettings bool) error {
	// Single-instance gate doubles as pipe-preemption mitigation: only the
	// mutex holder proceeds to pipeserver.New (see PipePathForUser).
	// A second instance does not just exit silently - it pokes the live overlay so
	// a user double-clicking the executable sees it appear instead of nothing
	// happening.
	instance, err := newInstance(instanceName)
	if errors.Is(err, errAlreadyRunning) {
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

	directory := home.Directory()
	if directory == "" {
		return errors.New("this platform has no data directory, so there is nothing to run")
	}
	hideConsole()
	// Reclaim atomic-write temporaries orphaned by a hard kill. Two minutes is
	// already orders of magnitude longer than the write it guards, and the overlay
	// is single-instance, so nothing else can hold one in flight. scripts/install.ps1
	// force-stops the overlay before every upgrade, so a longer window would leave a
	// visible orphan behind after each reinstall.
	config.SweepTemporaryFiles(directory, 2*time.Minute)
	runtimeDirectory := filepath.Join(directory, "runtime")
	// One cached OS view drives snapshot retention, liveness and cleanup for the
	// whole process. It replaces the old model where "no reporter push for 15/30
	// seconds" was taken to mean the session had ended.
	presence := newPresence()
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
	if !uiAvailable() {
		return ErrNoWindowLayer
	}

	configPath := filepath.Join(directory, config.FileName)
	statePath := filepath.Join(directory, config.StateFileName)
	settings := config.Load(configPath)
	placement := config.LoadState(statePath)
	// Saves are coalesced onto a background goroutine: a placement write happens
	// right after a drag, and on machines with AV write hooks an inline write could
	// stall the UI thread for hundreds of milliseconds.
	saver := newPlacementSaver(statePath)
	defer saver.flush()
	// config.json gets the same treatment. The settings window writes on every click
	// and once per slider drag, and an inline write there would put a file write on
	// the UI thread for a window that is already repainting.
	settingsSaver := newConfigSaver(configPath)
	defer settingsSaver.flush()

	autostartManager := autostart.New()
	executable, executableErr := os.Executable()
	autostartEnabled := false
	if executableErr == nil {
		// Repair a stale path first. The install path is the only thing in the entry,
		// so a moved or reinstalled binary would fail silently at logon - the worst
		// possible way to fail. Sync never enables a disabled entry, so this cannot
		// turn autostart on by itself.
		if _, err := autostartManager.Sync(executable); err != nil {
			// Autostart is a convenience, not a prerequisite: a backend that refuses
			// to cooperate must not stop the overlay from running.
			_ = err
		}
		if status, err := autostartManager.Status(); err == nil {
			autostartEnabled = status.Enabled
		}
	}

	values := store.New()
	window := ui.New(values, ui.Options{
		Opacity:               uint8(settings.Opacity),
		Settings:              uiSettings(settings),
		AutoStart:             autostartEnabled,
		AutoApproveTimeoutSec: settings.AutoApproveTimeoutSec,
		OpenSettings:          openSettings,
		// The attendance sampler reuses this exact cache: one process snapshot and
		// one window pass serve both liveness and "is the user looking at that
		// session", instead of enumerating twice.
		Presence: presence,
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
			// Staged, not written: the read, the merge and the write all happen on
			// the saver's goroutine so a click can never block the pump.
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
	// disappearing is an explicit "this session ended" signal - reporters remove
	// their snapshot on close, and writing a `closed` payload has the same effect.
	// Without this, a pipe session whose file was deleted stayed in memory forever,
	// propped up by a long-lived parent pid.
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
		// Reconcile, not Cleanup: a live terminal keeps its session no matter how
		// long the reporter has been quiet. Reconcile only bumps the revision when a
		// session is really added or removed, so a quiet desktop schedules no
		// repaints at all.
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
	// Hot-reload the human-editable config without restarting the overlay. A single
	// Stat every 30s is cheaper than any OS file watcher and needs no extra handle
	// or thread.
	go watchConfig(ctx, configPath, func() {
		current := config.Load(configPath)
		// A settings-window change lands on disk and bumps mtime; reloading it
		// straight back would echo the write. Only a change this process did not
		// make is worth applying.
		if settingsSaver.echo(current) {
			return
		}
		window.QueueSettings(uiSettings(current))
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

// ErrNoWindowLayer is what a platform without a window implementation returns from
// the overlay entry point.
//
// It exists because the alternative is worse than a missing feature. A build that
// started and drew nothing would look exactly like the silent failures this project
// spent round after round removing (§51: "the pet just does not appear"), and the
// panel would be a window nobody can see. Saying so - and pointing at the commands
// that do work - is the honest failure. The reporters, --json, --doctor, --connect,
// --autostart and the pipe server all work on every platform; only the window does
// not, and --doctor reports it as a warning.
var ErrNoWindowLayer = errors.New("this platform has no window layer yet: the reporters, --json, --doctor and --connect work; see docs/PLATFORMS.md")

// uiSettings maps the persisted config onto the UI's public settings struct. It is
// one function rather than two copies of nineteen fields (the startup path and the
// hot-reload path used to spell it twice, which is how a new setting reaches one
// and not the other).
func uiSettings(settings config.Config) ui.Settings {
	return ui.Settings{
		Notifications:         settings.Notifications,
		Sound:                 settings.Sound,
		SoundVolume:           settings.SoundVolume,
		SoundScheme:           settings.SoundScheme,
		SoundOnWaiting:        settings.SoundOnWaiting,
		SoundOnError:          settings.SoundOnError,
		NotifyOnFinish:        settings.NotifyOnFinish,
		DonePolicy:            settings.DonePolicy,
		ForegroundAware:       settings.ForegroundAware,
		IdleSeconds:           settings.IdleMinutes * 60,
		QuietCurrent:          settings.QuietCurrent,
		MutedSessions:         append([]string(nil), settings.MutedSessions...),
		QuietSubagents:        settings.QuietSubagents,
		AlwaysOnTop:           settings.AlwaysOnTop,
		EdgeRetract:           settings.EdgeRetract,
		Language:              settings.Language,
		AutoApprove:           settings.AutoApprove,
		AutoApproveTimeoutSec: settings.AutoApproveTimeoutSec,
		Opacity:               settings.Opacity,
		Theme:                 settings.Theme,
		PerSourceColor:        settings.PerSourceColor,
		HiddenAtLaunch:        settings.HiddenAtLaunch,
		SettingsTextSize:      settings.SettingsTextSize,
		SettingsKeyboard:      settings.SettingsKeyboard,
		SettingsTooltips:      settings.SettingsTooltips,
	}
}

// instanceName is the single-instance key. It is a name, not a path, because each
// platform names its own kernel object (a mutex on Windows, a lock file on POSIX).
const instanceName = "AI-Terminal-Pet-v1-SingleInstance"

// ErrNoDataDirectory is what a command that needs the overlay's data directory reports
// when the platform gives it none (a Windows process with no LOCALAPPDATA, a POSIX one
// with no HOME). It is a distinct error because the fix is environmental, not a bug.
var ErrNoDataDirectory = errors.New("this platform resolved no data directory for AI-Terminal-Pet")
