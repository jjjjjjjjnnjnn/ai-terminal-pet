//go:build darwin

package ui

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

var (
	darwinSoundMu sync.Mutex
	darwinSounds  = map[uint32]objc.ID{}
)

// beep synthesises an alert WAV and plays it asynchronously via Cocoa NSSound.
func beep(kind uint32) {
	volume := int(soundVolume.Load())
	if volume <= 0 || !ensureCocoaLoaded() {
		return
	}
	buffer := buildAlertWAV(kind, volume)
	if len(buffer) <= 44 {
		return
	}
	withAutoreleasePool(func() {
		clsData := objc.GetClass("NSData")
		clsSound := objc.GetClass("NSSound")
		if clsData == 0 || clsSound == 0 {
			return
		}
		data := objc.ID(clsData).Send(objc.RegisterName("alloc")).Send(
			objc.RegisterName("initWithBytes:length:"),
			unsafe.Pointer(&buffer[0]),
			uintptr(len(buffer)),
		)
		if data == 0 {
			return
		}
		defer data.Send(objc.RegisterName("release"))

		sound := objc.ID(clsSound).Send(objc.RegisterName("alloc")).Send(
			objc.RegisterName("initWithData:"),
			data,
		)
		if sound == 0 {
			return
		}
		_ = objc.Send[bool](sound, objc.RegisterName("play"))

		darwinSoundMu.Lock()
		prev := darwinSounds[kind]
		darwinSounds[kind] = sound
		darwinSoundMu.Unlock()
		if prev != 0 {
			_ = objc.Send[bool](prev, objc.RegisterName("stop"))
			prev.Send(objc.RegisterName("release"))
		}
	})
}

// SoundSurface names what the `sound` setting reaches on this machine, for
// --doctor. AppKit ships with the OS, so the answer cannot depend on the
// machine - only on whether the window layer loaded at all.
func SoundSurface() string {
	return "AppKit NSSound (built in)"
}
