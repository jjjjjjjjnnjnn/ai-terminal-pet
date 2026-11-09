package ui

import (
	"encoding/binary"
	"os"
	"runtime"
	"testing"
	"time"
)

// The tone must be a real WAV image: PlaySoundW(SND_MEMORY) reads a complete
// RIFF/WAVE file from memory, and a malformed header silently plays nothing.
func TestBuildAlertWAVHeader(t *testing.T) {
	data := buildAlertWAV(beepExclamation, 85)
	if len(data) <= 44 {
		t.Fatalf("wave is only %d bytes", len(data))
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" ||
		string(data[12:16]) != "fmt " || string(data[36:40]) != "data" {
		t.Fatal("not a RIFF/WAVE PCM image")
	}
	if got := binary.LittleEndian.Uint32(data[4:8]); got != uint32(len(data)-8) {
		t.Fatalf("RIFF size %d, want %d", got, len(data)-8)
	}
	if got := binary.LittleEndian.Uint32(data[40:44]); got != uint32(len(data)-44) {
		t.Fatalf("data size %d, want %d", got, len(data)-44)
	}
	if got := binary.LittleEndian.Uint16(data[22:24]); got != 1 {
		t.Fatalf("channels = %d, want 1", got)
	}
	if got := binary.LittleEndian.Uint32(data[24:28]); got != soundSampleRate {
		t.Fatalf("sample rate = %d", got)
	}
	if got := binary.LittleEndian.Uint16(data[34:36]); got != 16 {
		t.Fatalf("bits per sample = %d", got)
	}
}

func peakAmplitude(data []byte) int {
	peak := 0
	for i := 44; i+1 < len(data); i += 2 {
		sample := int(int16(binary.LittleEndian.Uint16(data[i : i+2])))
		if sample < 0 {
			sample = -sample
		}
		if sample > peak {
			peak = sample
		}
	}
	return peak
}

// "The alert is too quiet" is fixed by making loudness a setting rather than
// relying on the host's default-beep scheme. It must scale monotonically,
// clamp out-of-range values, and never overflow int16.
func TestBuildAlertWAVVolumeScales(t *testing.T) {
	quiet := peakAmplitude(buildAlertWAV(beepExclamation, 25))
	normal := peakAmplitude(buildAlertWAV(beepExclamation, 85))
	loud := peakAmplitude(buildAlertWAV(beepExclamation, 100))
	if !(quiet < normal && normal < loud) {
		t.Fatalf("amplitude did not scale: 25=%d 85=%d 100=%d", quiet, normal, loud)
	}
	if loud > 32767 {
		t.Fatalf("clipping: peak %d exceeds int16 range", loud)
	}
	if got := peakAmplitude(buildAlertWAV(beepExclamation, 0)); got != 0 {
		t.Fatalf("volume 0 must be silence, got peak %d", got)
	}
	if a, b := peakAmplitude(buildAlertWAV(beepExclamation, 100)),
		peakAmplitude(buildAlertWAV(beepExclamation, 500)); a != b {
		t.Fatal("volume above 100 was not clamped")
	}
}

// The three tones have to be distinguishable without looking at the panel.
func TestBuildAlertWAVKindsDiffer(t *testing.T) {
	waiting := buildAlertWAV(beepExclamation, 85)
	errorTone := buildAlertWAV(beepError, 85)
	finished := buildAlertWAV(beepFinished, 85)
	if len(waiting) == len(finished) && string(waiting) == string(finished) {
		t.Fatal("waiting and finished tones are identical")
	}
	if len(errorTone) == len(waiting) && string(errorTone) == string(waiting) {
		t.Fatal("waiting and error tones are identical")
	}
	// The finished tone is deliberately the shortest and softest cue.
	if len(finished) >= len(waiting) {
		t.Fatalf("finished tone (%d bytes) should be shorter than an urgent one (%d)", len(finished), len(waiting))
	}
}

func TestSetSoundVolumeClamps(t *testing.T) {
	original := int(soundVolume.Load())
	defer setSoundVolume(original)
	setSoundVolume(500)
	if got := int(soundVolume.Load()); got != 100 {
		t.Fatalf("500 became %d, want 100", got)
	}
	setSoundVolume(-10)
	if got := int(soundVolume.Load()); got != 0 {
		t.Fatalf("-10 became %d, want 0", got)
	}
}

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

// Every tone must actually sound different. A scheme table that silently falls
// back to the same waveform would make the settings window's four options a
// lie, and the failure would be audible-only - i.e. invisible to every other
// test.
func TestSoundSchemesProduceDistinctWaveforms(t *testing.T) {
	seen := make(map[string]string, len(SoundSchemes))
	for _, scheme := range SoundSchemes {
		wave := buildAlertWAVScheme(beepExclamation, 85, scheme)
		if len(wave) <= 44 {
			t.Fatalf("scheme %q produced no samples", scheme)
		}
		for other, existing := range seen {
			if string(existing) == string(wave) {
				t.Fatalf("schemes %q and %q render identical audio", scheme, other)
			}
		}
		seen[scheme] = string(wave)
	}
}

func TestSetSoundSchemeRejectsUnknownValues(t *testing.T) {
	previous := currentSoundScheme()
	defer setSoundScheme(previous)
	setSoundScheme(soundSchemeBell)
	if got := currentSoundScheme(); got != soundSchemeBell {
		t.Fatalf("setSoundScheme(bell) = %q", got)
	}
	setSoundScheme("no-such-tone")
	if got := currentSoundScheme(); got != defaultScheme {
		t.Fatalf("an unknown scheme became %q, want the default", got)
	}
}
