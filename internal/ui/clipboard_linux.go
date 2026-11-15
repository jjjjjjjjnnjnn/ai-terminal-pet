//go:build linux

package ui

import (
	"sync"

	"github.com/jezek/xgb/xproto"
)

var (
	x11ClipboardMu    sync.Mutex
	x11ClipboardText  string
	x11ClipboardOwner *App
)

func copyToClipboard(value string) bool {
	if value == "" {
		return false
	}
	x11ClipboardMu.Lock()
	owner := x11ClipboardOwner
	if owner == nil || owner.conn == nil || owner.atomClipboard == 0 {
		x11ClipboardMu.Unlock()
		return false
	}
	x11ClipboardText = value
	conn := owner.conn
	win := owner.win
	atom := owner.atomClipboard
	x11ClipboardMu.Unlock()

	if err := xproto.SetSelectionOwnerChecked(conn, win, atom, xproto.TimeCurrentTime).Check(); err != nil {
		return false
	}
	return true
}

func currentX11ClipboardText() string {
	x11ClipboardMu.Lock()
	defer x11ClipboardMu.Unlock()
	return x11ClipboardText
}
