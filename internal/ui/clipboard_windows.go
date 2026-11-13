package ui

// Putting a string on the Windows clipboard, for the one gesture that needs it
// (a tap on a session row copies that session's working directory).
//
// The order of the three steps is the whole design:
//
//  1. allocate the global memory block and fill it;
//  2. only then open the clipboard;
//  3. empty it and hand the block over.
//
// The intuitive order (open, empty, allocate) has a failure mode that is worse
// than doing nothing: if the allocation fails after EmptyClipboard, the user is
// left with an EMPTY clipboard, and "the copy failed" becomes indistinguishable
// from "the click did nothing" - while their previous clipboard content is
// gone. Allocating first means the worst case is "nothing changed".
//
// Every primitive is already declared by lxn/win (kernel32 + user32), so this
// file adds no dependency and no cgo.

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

// copyToClipboard places value on the clipboard as CF_UNICODETEXT. It reports
// whether the data is actually there: the caller shows its receipt only on
// true, because a receipt over an empty clipboard is a lie.
//
// Called from the UI thread only (it runs inside a click handler, and
// OpenClipboard makes the calling process the clipboard owner).
func copyToClipboard(value string) bool {
	if value == "" {
		return false
	}
	// UTF16FromString appends the NUL terminator CF_UNICODETEXT expects, and
	// rejects a value containing an embedded NUL.
	encoded, err := syscall.UTF16FromString(value)
	if err != nil || len(encoded) == 0 {
		return false
	}
	block := win.GlobalAlloc(win.GMEM_MOVEABLE, uintptr(len(encoded))*unsafe.Sizeof(uint16(0)))
	if block == 0 {
		return false
	}
	pointer := win.GlobalLock(block)
	if pointer == nil {
		win.GlobalFree(block)
		return false
	}
	copy(unsafe.Slice((*uint16)(pointer), len(encoded)), encoded)
	win.GlobalUnlock(block)

	if !win.OpenClipboard(0) {
		win.GlobalFree(block)
		return false
	}
	// CloseClipboard is required even on the failure path below: leaving the
	// clipboard open blocks every other process on the desktop until this
	// window handles another click.
	defer win.CloseClipboard()
	win.EmptyClipboard()
	// The system takes ownership of the block on success, so it must NOT be
	// freed afterwards; on failure ownership stayed with us.
	if win.SetClipboardData(win.CF_UNICODETEXT, win.HANDLE(block)) == 0 {
		win.GlobalFree(block)
		return false
	}
	return true
}
