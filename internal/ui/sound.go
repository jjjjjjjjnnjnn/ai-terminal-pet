package ui

import (
	"encoding/binary"
	"math"
	"sync"
	"sync/atomic"
)

// The ui package deliberately does not import internal/config (the app layer
// owns the file paths), so the scheme names are mirrored here. A test in the
// app package pins the two lists together.
const (
	soundSampleRate = 44100

	// MessageBeep / alert tone kinds shared across platforms.
	beepError       uint32 = 0x00000010 // MB_ICONHAND
	beepExclamation uint32 = 0x00000030 // MB_ICONEXCLAMATION
	beepFinished    uint32 = 0x00000040
)

const (
	waveSine = iota
	waveTriangle
	waveSquare

	envelopeFade
	envelopeDecay
)

type tone struct {
	frequency float64
	lengthMs  int
	gapMs     int
}

type schemeSpec struct {
	gain     float64
	envelope int
	notes    map[uint32][]tone
}

// schemeSpecs is a pure description of each timbre: no files, no registry,
// no external assets. Switching a scheme only changes which pattern is fed to
// the synthesiser, so auditioning in the settings window is instant.
func schemeSpecs() map[string]schemeSpec {
	return map[string]schemeSpec{
		soundSchemeChime: {
			gain: 1, envelope: envelopeFade,
			notes: map[uint32][]tone{
				beepExclamation: {{880, 130, 55}, {1175, 190, 190}, {880, 130, 55}, {1175, 190, 0}},
				beepError:       {{440, 190, 70}, {330, 260, 120}, {440, 190, 70}, {330, 260, 0}},
				beepFinished:    {{659, 95, 35}, {880, 95, 35}, {1175, 190, 0}},
			},
		},
		soundSchemeBell: {
			gain: 0.92, envelope: envelopeDecay,
			notes: map[uint32][]tone{
				beepExclamation: {{1046, 230, 90}, {1568, 300, 0}},
				beepError:       {{784, 220, 80}, {587, 320, 0}},
				beepFinished:    {{1046, 130, 45}, {1318, 130, 45}, {1568, 230, 0}},
			},
		},
		soundSchemeSoft: {
			gain: 0.7, envelope: envelopeFade,
			notes: map[uint32][]tone{
				beepExclamation: {{523, 200, 90}, {659, 260, 0}},
				beepError:       {{392, 220, 90}, {330, 280, 0}},
				beepFinished:    {{523, 150, 50}, {659, 220, 0}},
			},
		},
		soundSchemePulse: {
			gain: 0.55, envelope: envelopeFade,
			notes: map[uint32][]tone{
				beepExclamation: {{440, 80, 45}, {880, 80, 45}, {440, 80, 180}, {880, 110, 0}},
				beepError:       {{330, 100, 55}, {220, 100, 55}, {330, 150, 0}},
				beepFinished:    {{660, 70, 35}, {990, 110, 0}},
			},
		},
	}
}

func waveFor(scheme string) int {
	switch scheme {
	case soundSchemeSoft:
		return waveTriangle
	case soundSchemePulse:
		return waveSquare
	default:
		return waveSine
	}
}

var (
	soundVolume atomic.Int32 // 0..100
	soundScheme atomic.Value // string

	soundMutex   sync.Mutex
	soundBuffers = map[uint32][]byte{} // keeps async PCM buffers alive
)

func init() {
	soundScheme.Store(defaultScheme)
}

func currentSoundScheme() string {
	value, _ := soundScheme.Load().(string)
	return validSoundScheme(value)
}

// toneKind maps an alert decision's timbre name onto the synthesised kind constant.
func toneKind(tone string) uint32 {
	switch tone {
	case toneError:
		return beepError
	case toneFinished:
		return beepFinished
	default:
		return beepExclamation
	}
}

// setSoundVolume / setSoundScheme are called on the UI thread at startup,
// whenever config.json is hot-reloaded, and on every settings-window change.
func setSoundVolume(volume int) {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	soundVolume.Store(int32(volume))
}

func setSoundScheme(value string) {
	soundScheme.Store(validSoundScheme(value))
}

// buildAlertWAV returns a complete mono 16-bit PCM WAV file in memory using
// the currently selected scheme.
func buildAlertWAV(kind uint32, volume int) []byte {
	return buildAlertWAVScheme(kind, volume, currentSoundScheme())
}

// buildAlertWAVScheme is the pure synthesiser. volume is 0..100 and scales the
// peak amplitude linearly; a short fade in and a longer fade out (or an
// exponential decay for the bell) remove the click at each note boundary.
func buildAlertWAVScheme(kind uint32, volume int, scheme string) []byte {
	scheme = validSoundScheme(scheme)
	spec := schemeSpecs()[scheme]
	pattern, ok := spec.notes[kind]
	if !ok {
		pattern = spec.notes[beepExclamation]
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	totalSamples := 0
	for _, note := range pattern {
		totalSamples += (note.lengthMs + note.gapMs) * soundSampleRate / 1000
	}
	if totalSamples <= 0 {
		return nil
	}

	data := make([]byte, 44+totalSamples*2)
	copy(data[0:4], "RIFF")
	binary.LittleEndian.PutUint32(data[4:8], uint32(36+totalSamples*2))
	copy(data[8:12], "WAVE")
	copy(data[12:16], "fmt ")
	binary.LittleEndian.PutUint32(data[16:20], 16) // PCM header size
	binary.LittleEndian.PutUint16(data[20:22], 1)  // PCM
	binary.LittleEndian.PutUint16(data[22:24], 1)  // mono
	binary.LittleEndian.PutUint32(data[24:28], soundSampleRate)
	binary.LittleEndian.PutUint32(data[28:32], soundSampleRate*2) // byte rate
	binary.LittleEndian.PutUint16(data[32:34], 2)                 // block align
	binary.LittleEndian.PutUint16(data[34:36], 16)                // bits per sample
	copy(data[36:40], "data")
	binary.LittleEndian.PutUint32(data[40:44], uint32(totalSamples*2))

	peak := 32767 * 0.9 * spec.gain * float64(volume) / 100
	wave := waveFor(scheme)
	offset := 44
	for _, note := range pattern {
		length := note.lengthMs * soundSampleRate / 1000
		fadeIn := length / 12
		fadeOut := length / 4
		if fadeIn < 1 {
			fadeIn = 1
		}
		for i := 0; i < length; i++ {
			envelope := 1.0
			if spec.envelope == envelopeDecay {
				envelope = math.Exp(-3.4 * float64(i) / float64(length))
				if i < fadeIn {
					envelope *= float64(i) / float64(fadeIn)
				}
			} else {
				if i < fadeIn {
					envelope = float64(i) / float64(fadeIn)
				} else if i > length-fadeOut {
					envelope = float64(length-i) / float64(fadeOut)
				}
			}
			if envelope < 0 {
				envelope = 0
			}
			phase := 2 * math.Pi * note.frequency * float64(i) / soundSampleRate
			var shape float64
			switch wave {
			case waveTriangle:
				shape = 2 / math.Pi * math.Asin(math.Sin(phase))
			case waveSquare:
				// A soft-clipped square keeps the "digital" character without
				// the harsh aliasing of a raw sign() step.
				shape = math.Tanh(3*math.Sin(phase)) / math.Tanh(3)
			default:
				shape = math.Sin(phase)
			}
			sample := int16(peak * envelope * shape)
			binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(sample))
			offset += 2
		}
		// The gap is silence; the backing slice is already zeroed.
		offset += note.gapMs * soundSampleRate / 1000 * 2
	}
	return data[:offset]
}
