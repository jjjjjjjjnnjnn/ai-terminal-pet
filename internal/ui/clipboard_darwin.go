//go:build darwin

package ui

import "github.com/ebitengine/purego/objc"

func copyToClipboard(value string) bool {
	if value == "" || !ensureCocoaLoaded() {
		return false
	}
	var copied bool
	withAutoreleasePool(func() {
		clsPB := objc.GetClass("NSPasteboard")
		if clsPB == 0 {
			return
		}
		pb := objc.ID(clsPB).Send(objc.RegisterName("generalPasteboard"))
		if pb == 0 {
			return
		}
		pb.Send(objc.RegisterName("clearContents"))
		nsValue := nsStringFromGo(value)
		nsType := nsStringFromGo("public.utf8-plain-text")
		if nsValue != 0 {
			defer nsValue.Send(objc.RegisterName("release"))
		}
		if nsType != 0 {
			defer nsType.Send(objc.RegisterName("release"))
		}
		if nsValue == 0 || nsType == 0 {
			return
		}
		copied = objc.Send[bool](pb, objc.RegisterName("setString:forType:"), nsValue, nsType)
	})
	return copied
}
