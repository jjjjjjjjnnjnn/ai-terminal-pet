package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/ui"
)

// A settings change reaches config.json through a goroutine, never through the
// click handler: configSaver.save stages the value and wakes a writer, because
// on a machine with an AV write hook a synchronous write can stall the message
// pump long enough to make the window feel dead.
//
// That split has one consequence worth pinning here. When someone verifies
// "does clicking this control actually persist?" by starting the overlay,
// clicking, and then **killing** it, the file legitimately still holds the old
// value: a kill skips `defer settingsSaver.flush()`, which is the only place a
// staged-but-unwritten value is written for certain. 44.7 in docs/HANDOVER.md
// recorded exactly that as an unexplained bug until it was re-run with a
// graceful exit. These two tests state both halves of the contract so the next
// reading of "the file did not change" starts from the mechanism rather than
// from a suspicion about the click handler.

// TestConfigSaverPersistsWithoutAFlush covers the live path: save() alone must
// reach the disk, because a running overlay never calls flush() until it exits.
func TestConfigSaverPersistsWithoutAFlush(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	saver := newConfigSaver(path)
	defer saver.flush()

	// Values that differ from config.Default() in more than one field, so a
	// test that accidentally reads the defaults cannot pass.
	saver.save(ui.Settings{
		Sound:            false,
		SoundVolume:      42,
		AlwaysOnTop:      false,
		Theme:            config.ThemePaper,
		SettingsTextSize: config.TextSizeLarge,
	})

	deadline := time.Now().Add(5 * time.Second)
	for {
		saved := config.Load(path)
		if saved.SoundVolume == 42 && saved.Theme == config.ThemePaper &&
			saved.SettingsTextSize == config.TextSizeLarge && !saved.Sound {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("after 5s config.json still reads soundVolume=%d theme=%q textSize=%q sound=%v: "+
				"the writer goroutine never landed the value",
				saved.SoundVolume, saved.Theme, saved.SettingsTextSize, saved.Sound)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestConfigSaverFlushPersistsAStagedValue covers the exit path, deterministically:
// the saver is built without its goroutine, so the value can only appear because
// flush() wrote it. This is the guarantee a graceful exit relies on, and the one
// a kill takes away.
func TestConfigSaverFlushPersistsAStagedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	saver := &configSaver{path: path, wake: make(chan struct{}, 1)}

	saver.save(ui.Settings{SoundVolume: 17, Theme: config.ThemeContrast})

	if _, err := os.Stat(path); err == nil {
		t.Fatal("config.json exists before flush(): save() must stage, not write")
	}

	saver.flush()

	saved := config.Load(path)
	if saved.SoundVolume != 17 {
		t.Fatalf("soundVolume = %d after flush(), want 17", saved.SoundVolume)
	}
	if saved.Theme != config.ThemeContrast {
		t.Fatalf("theme = %q after flush(), want %q", saved.Theme, config.ThemeContrast)
	}
}

// TestConfigSaverFlushWithoutAPendingValueIsSilent stops flush() from creating a
// file for a process that changed nothing: the overlay exits on its own when the
// last session ends, and that exit must not fabricate preferences.
func TestConfigSaverFlushWithoutAPendingValueIsSilent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	saver := &configSaver{path: path, wake: make(chan struct{}, 1)}

	saver.flush()

	if _, err := os.Stat(path); err == nil {
		t.Fatal("flush() created config.json without a staged value")
	}
}
