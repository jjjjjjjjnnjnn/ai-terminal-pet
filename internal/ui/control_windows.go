//go:build windows

package ui

// The Windows delivery: find the window by class name and post the message.
//
// PostMessage is asynchronous and safe across threads, so a missing or busy window
// can never block the caller - which matters because the caller is usually a
// short-lived CLI process the user is waiting on.

import (
	"syscall"

	win "github.com/lxn/win"
)

func postControlMessage(action ControlAction) bool {
	name, err := syscall.UTF16FromString(windowClassName)
	if err != nil {
		return false
	}
	handle := win.FindWindow(&name[0], nil)
	if handle == 0 {
		return false
	}
	message := controlMessage(action)
	if message == 0 {
		return false
	}
	return win.PostMessage(handle, message, 0, 0) != 0
}

// controlMessage maps an action onto the window message the procedure handles.
func controlMessage(action ControlAction) uint32 {
	switch action {
	case ControlActivate:
		return wmActivate
	case ControlToggle:
		return wmToggle
	case ControlQuit:
		return wmClose
	case ControlReset:
		return wmReset
	case ControlSettings:
		return wmSettings
	}
	return 0
}
