//go:build linux

package ui

// Linux has no audio API this project can call in-process: ALSA needs libasound
// (cgo), PulseAudio and PipeWire need a protocol client, and docs/PLATFORMS.md
// records why neither is on the table. So the overlay keeps the WAV the
// synthesiser in sound.go already builds and hands it to the first system
// player it finds on PATH - PipeWire, then PulseAudio, then plain ALSA.
//
// That is the direction recorded before the port (round 61): no libc, no
// link-time audio dependency, and a machine with no player stays silent instead
// of failing. Nothing here runs per frame - beep is called from the alert plan
// and the handful of confirmation points, both event-rate.

import (
	"bytes"
	"os/exec"
	"sync"
)

// linuxPlayers is the probe order, and each entry also carries the arguments
// that make that player read a WAV file from stdin:
//
//   - pw-play reads stdin when the file is "-";
//   - paplay reads stdin when no file is given at all;
//   - aplay reads stdin from "-", and -q keeps it off stderr.
//
// The order is preference, not capability: PipeWire is the default on current
// desktops, PulseAudio is what the older ones ship, and aplay is the fallback
// that needs no session server.
var linuxPlayers = []struct {
	name string
	args []string
}{
	{name: "pw-play", args: []string{"-"}},
	{name: "paplay", args: nil},
	{name: "aplay", args: []string{"-q", "-"}},
}

// probeLinuxPlayer returns the first player that exists on PATH. look is
// exec.LookPath in production and a stub in tests, so the choice is testable
// without installing any of the three.
func probeLinuxPlayer(look func(string) (string, error)) (string, []string) {
	for _, candidate := range linuxPlayers {
		path, err := look(candidate.name)
		if err == nil && path != "" {
			return path, candidate.args
		}
	}
	return "", nil
}

var (
	// linuxPlayerMu guards the cached probe. A PATH walk per alert would put a
	// filesystem scan between the decision and the sound, and the answer cannot
	// change while the process runs.
	linuxPlayerMu       sync.Mutex
	linuxPlayerResolved bool
	linuxPlayerPath     string
	linuxPlayerArgs     []string

	// linuxSoundMu guards the previous player, which is stopped before the next
	// tone starts. Windows does the same by handing PlaySoundW a second buffer
	// and macOS by stopping the previous NSSound, so "one tone at a time" is the
	// behaviour all three platforms share.
	linuxSoundMu   sync.Mutex
	linuxSoundLast *exec.Cmd
)

func linuxPlayer() (string, []string) {
	linuxPlayerMu.Lock()
	defer linuxPlayerMu.Unlock()
	if !linuxPlayerResolved {
		linuxPlayerPath, linuxPlayerArgs = probeLinuxPlayer(exec.LookPath)
		linuxPlayerResolved = true
	}
	return linuxPlayerPath, linuxPlayerArgs
}

// SoundSurface names what the `sound` setting will actually reach on this
// machine, for --doctor. No player is reported as "none" rather than as a
// failure: silence is the documented behaviour when nothing is installed.
func SoundSurface() string {
	path, _ := linuxPlayer()
	if path == "" {
		return "none (no pw-play, paplay or aplay in PATH: alerts are silent; install pipewire-utils, pulseaudio-utils or alsa-utils)"
	}
	return "external player " + path
}

// beep synthesises the alert WAV and plays it through the probed player.
func beep(kind uint32) {
	volume := int(soundVolume.Load())
	if volume <= 0 {
		return
	}
	buffer := buildAlertWAV(kind, volume)
	if len(buffer) <= 44 {
		return
	}
	path, args := linuxPlayer()
	if path == "" {
		return
	}
	command, err := startLinuxPlayer(path, args, buffer)
	if err != nil {
		return
	}
	linuxSoundMu.Lock()
	previous := linuxSoundLast
	linuxSoundLast = command
	linuxSoundMu.Unlock()
	if previous != nil && previous.Process != nil {
		_ = previous.Process.Kill()
	}
	// Wait is what reaps the child: a player left unreaped would sit in the
	// process table for as long as the overlay runs.
	go func() { _ = command.Wait() }()
}

// startLinuxPlayer is split out so a test can wait for one play and inspect the
// bytes that reached the player's stdin without racing the UI thread.
func startLinuxPlayer(path string, args []string, buffer []byte) (*exec.Cmd, error) {
	command := exec.Command(path, args...)
	command.Stdin = bytes.NewReader(buffer)
	if err := command.Start(); err != nil {
		return nil, err
	}
	return command, nil
}
