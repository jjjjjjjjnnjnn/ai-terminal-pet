// Package config persists the two small JSON files the overlay keeps under
// %LOCALAPPDATA%\AI-Terminal-Pet.
//
// The split is deliberate:
//
//   - config.json is the user's file. It is read at startup and rewritten
//     only when the user flips a tray toggle, so hand-editing it while the
//     app is stopped is safe.
//   - state.json belongs to the app: window position, zoom and expanded flag.
//     It is rewritten whenever any of those change.
//
// Keeping them apart means the app remembering where it was dragged can never
// clobber a setting the user edited by hand.
//
// Every read path is fail-safe: a missing, unreadable or corrupt file yields
// defaults rather than an error, because a broken settings file must never
// stop the overlay from starting.
package config

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MinOpacity keeps the panel usable. Below roughly 12% alpha the pet is
	// effectively invisible while still swallowing clicks, which reads as a
	// hung app rather than a hidden one.
	MinOpacity = 40
	MaxOpacity = 255

	// Zoom bounds mirror the wheel handler in internal/ui.
	MinZoom = 0.65
	MaxZoom = 1.6

	// FileName and StateFileName are the two file names inside the app
	// directory. Both live next to the executable's data, not the source.
	FileName      = "config.json"
	StateFileName = "state.json"
)

// Config is the user-editable settings file.
type Config struct {
	// Notifications enables tray balloon alerts when a session enters
	// waiting_permission or error.
	Notifications bool `json:"notifications"`
	// AlwaysOnTop keeps the overlay above other windows.
	AlwaysOnTop bool `json:"alwaysOnTop"`
	// Opacity is the layered window's constant alpha, MinOpacity..MaxOpacity.
	Opacity int `json:"opacity"`
}

// Default returns the settings used when config.json is absent or unreadable.
func Default() Config {
	return Config{Notifications: true, AlwaysOnTop: true, Opacity: 235}
}

// Normalize clamps every field into a usable range, substituting the default
// for anything out of bounds. It never fails, so callers can chain it.
func (c Config) Normalize() Config {
	if c.Opacity < MinOpacity || c.Opacity > MaxOpacity {
		c.Opacity = Default().Opacity
	}
	return c
}

// DecodeConfig parses config.json bytes. Absent fields keep their defaults and
// any parse failure returns Default(), so a truncated write cannot brick the
// app. Exported for tests.
func DecodeConfig(data []byte) Config {
	if len(data) == 0 {
		return Default()
	}
	parsed := Default()
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Default()
	}
	return parsed.Normalize()
}

// Load reads config.json, falling back to Default on any error.
func Load(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default()
	}
	return DecodeConfig(data)
}

// Save writes config.json atomically. The file is created 0600 and replaced by
// rename, so a crash mid-write leaves the previous settings intact.
func Save(path string, value Config) error {
	value = value.Normalize()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, append(data, '\n'))
}

// State is the app-owned window placement remembered between launches.
type State struct {
	// Placed is false until the window has actually been positioned once.
	// Without it, a state file of all zeroes would be indistinguishable from
	// "remembered at the top-left corner".
	Placed   bool    `json:"placed"`
	X        int32   `json:"x"`
	Y        int32   `json:"y"`
	Zoom     float64 `json:"zoom"`
	Expanded bool    `json:"expanded"`
}

// DefaultState is the placement used when nothing has been remembered yet:
// unplaced (so the app falls back to its work-area anchor) at zoom 1.
func DefaultState() State {
	return State{Zoom: 1}
}

// Normalize repairs out-of-range values. A NaN or zero zoom would otherwise
// scale the canvas to zero pixels and the overlay would never draw.
func (s State) Normalize() State {
	if math.IsNaN(s.Zoom) || math.IsInf(s.Zoom, 0) || s.Zoom < MinZoom || s.Zoom > MaxZoom {
		s.Zoom = DefaultState().Zoom
	}
	return s
}

// DecodeState parses state.json bytes, returning DefaultState on any error.
// Normalizing matters on this path too: the zero value has zoom 0, which
// would scale the canvas to nothing, so the error path must not hand that
// back. Exported for tests.
func DecodeState(data []byte) State {
	if len(data) == 0 {
		return DefaultState()
	}
	var parsed State
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DefaultState()
	}
	return parsed.Normalize()
}

// LoadState reads state.json, falling back to DefaultState on any error.
func LoadState(path string) State {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultState()
	}
	return DecodeState(data)
}

// SaveState writes state.json atomically.
func SaveState(path string, value State) error {
	value = value.Normalize()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, append(data, '\n'))
}

// TempPrefix marks the temporary files WriteFileAtomic creates. It is
// deliberately distinctive so a sweep can recognise them without touching
// anything a user put in the directory.
const TempPrefix = ".pet-"

// WriteFileAtomic writes data to path via a temporary file in the same
// directory followed by a rename. Same-directory is required: rename is only
// atomic within a volume.
func WriteFileAtomic(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, TempPrefix+"*.tmp")
	if err != nil {
		return err
	}
	name := temporary.Name()
	_ = temporary.Chmod(0600)
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		_ = os.Remove(name)
		return err
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}

// SweepTemporaryFiles removes atomic-write temporaries left in directory by an
// interrupted save and returns how many it deleted.
//
// WriteFileAtomic cleans up after itself on every error path, but a hard kill
// (taskkill /F, power loss) between CreateTemp and Rename cannot: the process
// simply stops existing. Nothing else revisits this directory, so without a
// sweep those orphans accumulate one per interrupted save.
//
// Only files older than minAge are considered, so a sweep can never race a
// save that another process still has in flight.
func SweepTemporaryFiles(directory string, minAge time.Duration) int {
	if directory == "" {
		return 0
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0
	}
	cutoff := time.Now().Add(-minAge)
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, TempPrefix) || !strings.HasSuffix(name, ".tmp") {
			continue
		}
		info, err := entry.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(directory, name)); err == nil {
			removed++
		}
	}
	return removed
}
