package app

// The platform-neutral half of the overlay: the scanner loop, the config watcher,
// the two savers, the config-to-UI mapping and the --json side.
//
// RunWith (run.go) is the only place that knows the whole lifecycle; everything here
// is a piece of it with no OS in it, which is why internal/app can now be compiled
// for two more platforms. Splitting them is handover §55.6 step 7.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/session"
	"ai-terminal-pet/internal/store"
	"ai-terminal-pet/internal/ui"
)

// scanInterval is deliberately slow. Presence caches the OS view for 10s, reporters
// push state changes immediately over the pipe, and the panel already skips
// identical frames. Five seconds is enough to notice a terminal that started
// without a reporter event, without waking the process twice a second forever.
const scanInterval = 5 * time.Second

// scanIntervalForTest is the same value the scanner goroutine uses, exposed so a
// test can assert the cadence instead of waiting five seconds to observe it.
func scanIntervalForTest() time.Duration { return scanInterval }

// configWatchInterval is intentionally slow. Hand edits to config.json are rare and
// a 30s delay is imperceptible, while a filesystem watcher would add a handle and a
// thread to a process whose whole job is to stay lightweight.
const configWatchInterval = 30 * time.Second

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
	// reflect.DeepEqual rather than ==: Config carries a slice now (the
	// per-session mute list), and this comparison is "is this reload the echo
	// of our own write", which has to mean field-for-field equality.
	return s.hasWritten && reflect.DeepEqual(s.written, value)
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
	// The three-way completion policy is the live field; NotifyOnFinish above
	// stays as the legacy two-way view, and Normalize keeps them consistent
	// when the file is read back.
	target.DonePolicy = config.ValidDonePolicy(value.DonePolicy)
	target.ForegroundAware = value.ForegroundAware
	if value.IdleSeconds >= 0 {
		target.IdleMinutes = value.IdleSeconds / 60
	}
	target.QuietCurrent = value.QuietCurrent
	// The mute list is written back the way the ui package holds it, then
	// normalised on read; storing it here means a mute survives the overlay
	// exiting when no session is left (which is its normal end of life).
	target.MutedSessions = append([]string(nil), value.MutedSessions...)
	target.QuietSubagents = value.QuietSubagents
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
	target.SettingsTextSize = config.ValidTextSize(value.SettingsTextSize)
	target.SettingsKeyboard = value.SettingsKeyboard
	target.SettingsTooltips = value.SettingsTooltips
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
	// Alerts is the last alert decision the overlay recorded, or null when
	// there is no receipt. It is how a script (or a bug report) answers "was
	// it silent because it decided to be, or because nothing ran": the record
	// carries one reason token per considered session.
	Alerts *AlertReport `json:"alerts"`
}

// AlertReport is the subset of the overlay's receipt that belongs in the
// machine-readable report. It mirrors internal/ui's record rather than
// importing it (the app layer does not reach into a window package), and a
// test pins the two field names together.
type AlertReport struct {
	At       time.Time            `json:"at"`
	Presence string               `json:"presence"`
	Present  bool                 `json:"present"`
	Balloon  string               `json:"balloon,omitempty"`
	Tone     string               `json:"tone,omitempty"`
	Skipped  string               `json:"skipped,omitempty"`
	Sessions []AlertReportSession `json:"sessions,omitempty"`
}

type AlertReportSession struct {
	Identity string `json:"identity"`
	State    string `json:"state"`
	Kind     string `json:"kind"`
	Decision string `json:"decision"`
	Unread   bool   `json:"unread,omitempty"`
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
	directory := home.Directory()
	if directory == "" {
		return Report{}, ErrNoDataDirectory
	}
	presence := newPresence()
	scanner := store.SnapshotScanner{
		Directory: filepath.Join(directory, "runtime"),
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
	// The alert receipt is read here rather than fetched from the live window:
	// --json must never depend on a GUI process, and this file is written by
	// the overlay for exactly this purpose. It lives beside config.json and
	// NOT in runtime/, because every *.json in runtime/ is a reporter snapshot
	// and the store scanner deletes the ones it cannot parse (see
	// internal/store/runtime_contract_test.go).
	alerts := readAlertReport(directory)
	return Report{
		SchemaVersion: session.SchemaVersion,
		State:         session.Aggregate(items),
		Count:         len(items),
		GeneratedAt:   now.UTC(),
		Sessions:      items,
		Alerts:        alerts,
	}, nil
}

// readAlertReport reads alerts.json from the app directory. A missing or
// unreadable file is null, not a failure: an overlay that has not run since the
// upgrade has no receipt, and that is not a fault in the reporter a script is
// asking about.
func readAlertReport(directory string) *AlertReport {
	path := filepath.Join(directory, "alerts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var record AlertReport
	if err := json.Unmarshal(data, &record); err != nil {
		return nil
	}
	return &record
}

// EncodeReport renders a report as JSON. pretty switches to indented output for
// humans; the default single line is what scripts and status lines want.
func EncodeReport(report Report, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(report, "", "  ")
	}
	return json.Marshal(report)
}
