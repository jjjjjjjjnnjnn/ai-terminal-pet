package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeConfigFallsBackToDefaults(t *testing.T) {
	defaults := Default()
	for _, tc := range []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"truncated json", `{"opacity":`},
		{"wrong type", `{"opacity":"bright"}`},
		{"json array", `[1,2,3]`},
		{"not json", "hello"},
	} {
		if got := DecodeConfig([]byte(tc.data)); got != defaults {
			t.Fatalf("%s: got %+v want %+v", tc.name, got, defaults)
		}
	}
}

func TestDecodeConfigKeepsAbsentFieldsAtDefault(t *testing.T) {
	// A partial file must not silently turn notifications off: absent keys
	// keep their default, only present keys override.
	got := DecodeConfig([]byte(`{"opacity": 200}`))
	if !got.Notifications || !got.AlwaysOnTop {
		t.Fatalf("absent booleans must keep defaults, got %+v", got)
	}
	if got.Opacity != 200 {
		t.Fatalf("opacity must be honoured, got %d", got.Opacity)
	}
}

func TestDecodeConfigHonoursExplicitFalse(t *testing.T) {
	got := DecodeConfig([]byte(`{"notifications": false, "alwaysOnTop": false}`))
	if got.Notifications || got.AlwaysOnTop {
		t.Fatalf("explicit false must stick, got %+v", got)
	}
}

func TestConfigNormalizeClampsOpacity(t *testing.T) {
	for _, opacity := range []int{-1, 0, MinOpacity - 1, MaxOpacity + 1, 100000} {
		if got := (Config{Opacity: opacity}).Normalize(); got.Opacity != Default().Opacity {
			t.Fatalf("opacity %d must fall back to default, got %d", opacity, got.Opacity)
		}
	}
	for _, opacity := range []int{MinOpacity, 128, MaxOpacity} {
		if got := (Config{Opacity: opacity}).Normalize(); got.Opacity != opacity {
			t.Fatalf("opacity %d must be preserved, got %d", opacity, got.Opacity)
		}
	}
}

func TestDecodeStateRejectsBadZoom(t *testing.T) {
	for _, raw := range []string{`{"zoom":0}`, `{"zoom":-3}`, `{"zoom":99}`, `{}`, `nonsense`} {
		if got := DecodeState([]byte(raw)); got.Zoom != 1 {
			t.Fatalf("%s: zoom must normalise to 1, got %v", raw, got.Zoom)
		}
	}
	got := DecodeState([]byte(`{"placed":true,"x":-1200,"y":40,"zoom":1.3,"expanded":true}`))
	if !got.Placed || got.X != -1200 || got.Y != 40 || got.Zoom != 1.3 || !got.Expanded {
		t.Fatalf("state must round-trip, got %+v", got)
	}
}

func TestRoundTripThroughDisk(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, FileName)
	statePath := filepath.Join(directory, StateFileName)

	if err := Save(configPath, Config{Notifications: false, AlwaysOnTop: true, Opacity: 120}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if got := Load(configPath); got.Notifications || !got.AlwaysOnTop || got.Opacity != 120 {
		t.Fatalf("config round-trip mismatch: %+v", got)
	}

	if err := SaveState(statePath, State{Placed: true, X: 10, Y: -20, Zoom: 0.8, Expanded: true}); err != nil {
		t.Fatalf("save state: %v", err)
	}
	if got := LoadState(statePath); !got.Placed || got.X != 10 || got.Y != -20 || got.Zoom != 0.8 || !got.Expanded {
		t.Fatalf("state round-trip mismatch: %+v", got)
	}
}

func TestLoadMissingFilesUseDefaults(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.json")
	if got := Load(missing); got != Default() {
		t.Fatalf("missing config must default, got %+v", got)
	}
	if got := LoadState(missing); got != DefaultState() {
		t.Fatalf("missing state must be a normalized zero, got %+v", got)
	}
}

func TestSaveLeavesNoTemporaryFiles(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, StateFileName)
	if err := SaveState(path, State{Placed: true, Zoom: 1}); err != nil {
		t.Fatalf("save: %v", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() != StateFileName {
			t.Fatalf("unexpected leftover file %q", entry.Name())
		}
	}
}
