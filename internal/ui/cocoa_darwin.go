//go:build darwin

package ui

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const nsUTF8StringEncoding = uintptr(4)

var (
	cocoaLoadOnce sync.Once
	cocoaLoaded   bool
)

func ensureCocoaLoaded() bool {
	cocoaLoadOnce.Do(func() {
		if _, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_GLOBAL|purego.RTLD_LAZY); err != nil {
			return
		}
		if objc.GetClass("NSApplication") == 0 || objc.GetClass("NSWindow") == 0 {
			return
		}
		cocoaLoaded = true
	})
	return cocoaLoaded
}

func withAutoreleasePool(fn func()) {
	if !ensureCocoaLoaded() {
		fn()
		return
	}
	clsPool := objc.GetClass("NSAutoreleasePool")
	if clsPool == 0 {
		fn()
		return
	}
	pool := objc.ID(clsPool).Send(objc.RegisterName("new"))
	if pool != 0 {
		defer pool.Send(objc.RegisterName("drain"))
	}
	fn()
}

// nsStringFromGo creates a retained NSString from a UTF-8 Go string. Caller owns
// the returned object and must call release when non-zero.
func nsStringFromGo(value string) objc.ID {
	clsStr := objc.GetClass("NSString")
	if clsStr == 0 {
		return 0
	}
	if len(value) == 0 {
		return objc.ID(clsStr).Send(objc.RegisterName("alloc")).Send(objc.RegisterName("init"))
	}
	raw := []byte(value)
	return objc.ID(clsStr).Send(objc.RegisterName("alloc")).Send(
		objc.RegisterName("initWithBytes:length:encoding:"),
		unsafe.Pointer(&raw[0]),
		uintptr(len(raw)),
		nsUTF8StringEncoding,
	)
}
