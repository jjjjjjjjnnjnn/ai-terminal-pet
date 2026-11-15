//go:build linux

package platform

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/screensaver"
	"github.com/jezek/xgb/xproto"
)

var silenceXGBOnce sync.Once

func openX11Conn() (*xgb.Conn, xproto.Window, bool) {
	if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
		return nil, 0, false
	}
	silenceXGBOnce.Do(func() {
		xgb.Logger = log.New(io.Discard, "", 0)
	})
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, 0, false
	}
	setup := xproto.Setup(conn)
	if setup == nil || len(setup.Roots) == 0 {
		conn.Close()
		return nil, 0, false
	}
	return conn, setup.DefaultScreen(conn).Root, true
}

func internAtom(conn *xgb.Conn, name string) (xproto.Atom, bool) {
	reply, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil || reply == nil || reply.Atom == 0 {
		return 0, false
	}
	return reply.Atom, true
}

func getPropertyBytes(conn *xgb.Conn, win xproto.Window, prop, typ xproto.Atom) ([]byte, bool) {
	reply, err := xproto.GetProperty(conn, false, win, prop, typ, 0, 1024).Reply()
	if err != nil || reply == nil || reply.ValueLen == 0 {
		return nil, false
	}
	return reply.Value, true
}

func getPropertyUint32List(conn *xgb.Conn, win xproto.Window, prop, typ xproto.Atom) []uint32 {
	raw, ok := getPropertyBytes(conn, win, prop, typ)
	if !ok || len(raw) < 4 {
		return nil
	}
	count := len(raw) / 4
	out := make([]uint32, count)
	for i := 0; i < count; i++ {
		out[i] = xgb.Get32(raw[i*4:])
	}
	return out
}

func getWindowTitle(conn *xgb.Conn, win xproto.Window, netWMName, utf8String xproto.Atom) string {
	if netWMName != 0 && utf8String != 0 {
		if raw, ok := getPropertyBytes(conn, win, netWMName, utf8String); ok {
			if title := strings.TrimSpace(string(bytes.TrimRight(raw, "\x00"))); title != "" {
				return title
			}
		}
	}
	if raw, ok := getPropertyBytes(conn, win, xproto.AtomWmName, xproto.AtomAny); ok {
		return strings.TrimSpace(string(bytes.TrimRight(raw, "\x00")))
	}
	return ""
}

func getWindowClass(conn *xgb.Conn, win xproto.Window) string {
	raw, ok := getPropertyBytes(conn, win, xproto.AtomWmClass, xproto.AtomString)
	if !ok {
		return ""
	}
	parts := strings.Split(strings.TrimRight(string(raw), "\x00"), "\x00")
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		return parts[1]
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func snapshotProcesses() map[int]focusProcess {
	result := make(map[int]focusProcess, 128)
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, convErr := strconv.Atoi(entry.Name())
		if convErr != nil || pid <= 0 {
			continue
		}
		statPath := filepath.Join("/proc", entry.Name(), "stat")
		raw, readErr := os.ReadFile(statPath)
		if readErr != nil {
			continue
		}
		openParen := bytes.IndexByte(raw, '(')
		closeParen := bytes.LastIndexByte(raw, ')')
		if openParen < 0 || closeParen <= openParen || closeParen+2 >= len(raw) {
			continue
		}
		comm := string(raw[openParen+1 : closeParen])
		fields := strings.Fields(string(raw[closeParen+2:]))
		if len(fields) < 2 {
			continue
		}
		ppid, ppidErr := strconv.Atoi(fields[1])
		if ppidErr != nil {
			ppid = 0
		}
		name := comm
		if target, linkErr := os.Readlink(filepath.Join("/proc", entry.Name(), "exe")); linkErr == nil && target != "" {
			if base := filepath.Base(target); base != "" {
				name = base
			}
		}
		result[pid] = focusProcess{parent: ppid, name: name}
	}
	return result
}

func visibleWindows(processes map[int]focusProcess) []focusWindow {
	conn, root, ok := openX11Conn()
	if !ok {
		return nil
	}
	defer conn.Close()

	clientListAtom, ok := internAtom(conn, "_NET_CLIENT_LIST")
	if !ok {
		return nil
	}
	netWMName, _ := internAtom(conn, "_NET_WM_NAME")
	utf8String, _ := internAtom(conn, "UTF8_STRING")
	netWMPid, _ := internAtom(conn, "_NET_WM_PID")

	ids := getPropertyUint32List(conn, root, clientListAtom, xproto.AtomWindow)
	if len(ids) == 0 {
		return nil
	}
	out := make([]focusWindow, 0, len(ids))
	for _, id := range ids {
		win := xproto.Window(id)
		pid := 0
		if netWMPid != 0 {
			if pids := getPropertyUint32List(conn, win, netWMPid, xproto.AtomCardinal); len(pids) > 0 {
				pid = int(pids[0])
			}
		}
		class := getWindowClass(conn, win)
		title := getWindowTitle(conn, win, netWMName, utf8String)
		exe := processes[pid].name
		out = append(out, focusWindow{
			handle: uintptr(id),
			pid:    pid,
			class:  class,
			title:  title,
			exe:    exe,
		})
	}
	return out
}

// SampleForeground queries the active X11 window and optional MIT-SCREEN-SAVER idle duration.
func SampleForeground() ForegroundSample {
	conn, root, ok := openX11Conn()
	if !ok {
		return ForegroundSample{}
	}
	defer conn.Close()

	var sample ForegroundSample
	if activeAtom, has := internAtom(conn, "_NET_ACTIVE_WINDOW"); has {
		if active := getPropertyUint32List(conn, root, activeAtom, xproto.AtomWindow); len(active) > 0 && active[0] != 0 {
			sample.Handle = uintptr(active[0])
			sample.Known = true
		}
	}
	if err := screensaver.Init(conn); err == nil {
		if reply, qErr := screensaver.QueryInfo(conn, xproto.Drawable(root)).Reply(); qErr == nil && reply != nil {
			sample.Idle = time.Duration(reply.MsSinceUserInput) * time.Millisecond
			sample.IdleKnown = true
		}
	}
	return sample
}

// ActivateTarget asks the EWMH window manager to bring handle to the front.
func ActivateTarget(handle uintptr) bool {
	if handle == 0 {
		return false
	}
	conn, root, ok := openX11Conn()
	if !ok {
		return false
	}
	defer conn.Close()

	win := xproto.Window(handle)
	if activeAtom, has := internAtom(conn, "_NET_ACTIVE_WINDOW"); has {
		event := xproto.ClientMessageEvent{
			Format: 32,
			Window: win,
			Type:   activeAtom,
			Data: xproto.ClientMessageDataUnionData32New([]uint32{
				2, // source indication: pager / application switcher
				xproto.TimeCurrentTime,
				0,
				0,
				0,
			}),
		}
		mask := uint32(xproto.EventMaskSubstructureRedirect | xproto.EventMaskSubstructureNotify)
		_ = xproto.SendEventChecked(conn, false, root, mask, string(event.Bytes())).Check()
	}
	_ = xproto.ConfigureWindow(conn, win, xproto.ConfigWindowStackMode, []uint32{xproto.StackModeAbove})
	_ = xproto.SetInputFocus(conn, xproto.InputFocusPointerRoot, win, xproto.TimeCurrentTime)
	conn.Sync()
	return true
}
