//go:build linux

package ui

// The Linux playback path is a subprocess, so these tests never need a sound
// card: they put a fake player on PATH or call startLinuxPlayer directly. What
// is pinned here is our side of the contract - which binary wins, what argv it
// gets, and that the WAV bytes reach its stdin - because "the speakers made a
// noise" is the one part no CI machine can answer.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// resetLinuxSoundProbe clears the cached PATH probe so a test can choose the
// player it wants. Production resolves it once per process on purpose.
func resetLinuxSoundProbe() {
	linuxPlayerMu.Lock()
	linuxPlayerResolved = false
	linuxPlayerPath = ""
	linuxPlayerArgs = nil
	linuxPlayerMu.Unlock()
}

// fakeLook answers LookPath from a set of names that exist.
func fakeLook(installed map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := installed[name]; ok {
			return path, nil
		}
		return "", os.ErrNotExist
	}
}

func TestProbeLinuxPlayerPrefersPipeWire(t *testing.T) {
	path, args := probeLinuxPlayer(fakeLook(map[string]string{
		"pw-play": "/usr/bin/pw-play",
		"paplay":  "/usr/bin/paplay",
		"aplay":   "/usr/bin/aplay",
	}))
	if path != "/usr/bin/pw-play" {
		t.Fatalf("player = %q, want the PipeWire one", path)
	}
	if len(args) != 1 || args[0] != "-" {
		t.Fatalf("pw-play args = %v, want [\"-\"] (stdin)", args)
	}
}

func TestProbeLinuxPlayerFallsBackInDocumentedOrder(t *testing.T) {
	cases := []struct {
		name      string
		installed map[string]string
		wantPath  string
		wantArgs  []string
	}{
		{
			name:      "pulseaudio only",
			installed: map[string]string{"paplay": "/usr/bin/paplay"},
			wantPath:  "/usr/bin/paplay",
			wantArgs:  nil,
		},
		{
			name:      "alsa only",
			installed: map[string]string{"aplay": "/usr/bin/aplay"},
			wantPath:  "/usr/bin/aplay",
			wantArgs:  []string{"-q", "-"},
		},
		{
			name:      "pipewire and alsa",
			installed: map[string]string{"pw-play": "/usr/bin/pw-play", "aplay": "/usr/bin/aplay"},
			wantPath:  "/usr/bin/pw-play",
			wantArgs:  []string{"-"},
		},
		{
			name:      "nothing",
			installed: map[string]string{},
			wantPath:  "",
			wantArgs:  nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, args := probeLinuxPlayer(fakeLook(tc.installed))
			if path != tc.wantPath {
				t.Fatalf("player = %q, want %q", path, tc.wantPath)
			}
			if len(args) != len(tc.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tc.wantArgs)
			}
			for index := range args {
				if args[index] != tc.wantArgs[index] {
					t.Fatalf("args = %v, want %v", args, tc.wantArgs)
				}
			}
		})
	}
}

// writeFakePlayer installs an executable shell script that copies its stdin to
// $PET_FAKE_OUT and records its argv in $PET_FAKE_ARGV.
func writeFakePlayer(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > \"$PET_FAKE_ARGV\"\n" +
		"cat > \"$PET_FAKE_OUT\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake player: %v", err)
	}
	return path
}

func TestStartLinuxPlayerFeedsTheSynthesisedWAVToStdin(t *testing.T) {
	player := writeFakePlayer(t, "paplay")
	out := filepath.Join(t.TempDir(), "played.wav")
	t.Setenv("PET_FAKE_OUT", out)
	t.Setenv("PET_FAKE_ARGV", filepath.Join(t.TempDir(), "argv.txt"))

	buffer := buildAlertWAVScheme(beepFinished, 60, soundSchemeChime)
	if len(buffer) <= 44 {
		t.Fatalf("the synthesiser produced no audio (%d bytes)", len(buffer))
	}
	command, err := startLinuxPlayer(player, nil, buffer)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("player exited with %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read what the player received: %v", err)
	}
	if !bytes.Equal(got, buffer) {
		t.Fatalf("player received %d bytes, want the %d synthesised ones", len(got), len(buffer))
	}
}

func TestBeepIsSilentWithoutAPlayer(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	resetLinuxSoundProbe()
	t.Cleanup(resetLinuxSoundProbe)

	previousVolume := int(soundVolume.Load())
	previousScheme := currentSoundScheme()
	t.Cleanup(func() {
		setSoundVolume(previousVolume)
		setSoundScheme(previousScheme)
	})
	setSoundVolume(80)

	// A missing player must be silence, not a panic and not a stall.
	beep(beepExclamation)
	if path, _ := linuxPlayer(); path != "" {
		t.Fatalf("probe found %q on an empty PATH", path)
	}
	if surface := SoundSurface(); len(surface) < 4 || surface[:4] != "none" {
		t.Fatalf("SoundSurface() = %q, want it to start with none", surface)
	}
}

func TestBeepAtZeroVolumeStartsNothing(t *testing.T) {
	player := writeFakePlayer(t, "pw-play")
	out := filepath.Join(t.TempDir(), "played.wav")
	t.Setenv("PATH", filepath.Dir(player))
	t.Setenv("PET_FAKE_OUT", out)
	t.Setenv("PET_FAKE_ARGV", filepath.Join(t.TempDir(), "argv.txt"))
	resetLinuxSoundProbe()
	t.Cleanup(resetLinuxSoundProbe)

	previousVolume := int(soundVolume.Load())
	t.Cleanup(func() { setSoundVolume(previousVolume) })
	setSoundVolume(0)

	beep(beepFinished)
	// The player would create the file within milliseconds; anything slower
	// than this guard means it never ran at all.
	time.Sleep(200 * time.Millisecond)
	if _, err := os.Stat(out); err == nil {
		t.Fatalf("volume 0 still started a player")
	}
}
