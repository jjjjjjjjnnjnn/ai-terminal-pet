//go:build windows

package ui

import (
	"os"
	"runtime"
	"testing"
	"time"
)

// TestPlaySoundLive is an opt-in desktop probe: it really plays one tone so
// we can confirm winmm accepts a WAV image built in Go memory.
//
//	PET_SOUND_PROBE=1 go test ./internal/ui -run TestPlaySoundLive -v
func TestPlaySoundLive(t *testing.T) {
	if os.Getenv("PET_SOUND_PROBE") == "" {
		t.Skip("set PET_SOUND_PROBE=1 to play a tone on the desktop")
	}
	buffer := buildAlertWAV(beepExclamation, 85)
	if len(buffer) <= 44 {
		t.Fatal("no waveform generated")
	}
	if !playSoundW(&buffer[0], sndMemory|sndAsync|sndNoDefault) {
		t.Fatal("PlaySoundW rejected the synthesised WAV")
	}
	time.Sleep(1200 * time.Millisecond) // let the async playback finish
	runtime.KeepAlive(buffer)
	t.Log("played a 0.9 s two-note chime at 85% amplitude")
}
