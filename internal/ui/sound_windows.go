//go:build windows

package ui

import "runtime"

// Win32 PlaySoundW / shell notification flags used on Windows.
const (
	sndAsync     = 0x0001
	sndNoDefault = 0x0002
	sndMemory    = 0x0004

	// niifNoSound stops the tray balloon from layering a second, usually
	// quieter system sound on top of the one we just played.
	niifNoSound = 0x00000010
	// niifInfo is NIIF_INFO; lxn/win only exports the warning/error flags.
	niifInfo = 0x00000001
)

// beep plays an alert tone via Winmm PlaySoundW. The PCM buffer is cached by
// kind: PlaySoundW with SND_ASYNC keeps reading it after this function returns,
// so the Go slice stays reachable for the duration of playback.
func beep(kind uint32) {
	volume := int(soundVolume.Load())
	if volume <= 0 {
		return
	}
	buffer := buildAlertWAV(kind, volume)
	if len(buffer) <= 44 {
		return
	}
	soundMutex.Lock()
	soundBuffers[kind] = buffer
	held := soundBuffers[kind]
	soundMutex.Unlock()
	if !playSoundW(&held[0], sndMemory|sndAsync|sndNoDefault) {
		// Very old / stripped systems may lack winmm. Fall back to the system
		// tone; beepFinished is not a valid MB_ICON* value.
		fallback := kind
		if kind == beepFinished {
			fallback = beepExclamation
		}
		messageBeepProc.Call(uintptr(fallback))
	}
	runtime.KeepAlive(held)
}

// SoundSurface names what the `sound` setting reaches on this machine, for
// --doctor. Winmm is a system DLL, so the answer cannot depend on the machine.
func SoundSurface() string {
	return "winmm PlaySoundW (built in)"
}
