//go:build windows

package reporter

// The console title, as a jump hint.
//
// Why this exists (measured 2026-11-15, handover §61):
//
// The statusline reporter runs inside a shell the host spawns for one render, so
// `os.Getppid()` names a process that is already gone by the time the panel reads
// the snapshot - a 40-sample wall-clock trace of the live snapshot showed
// parentPid cycling through 32152, 31332, 97196, 17492 ... every two to three
// seconds, each one already DEAD on the next read. internal/platform's resolver
// then has no ancestor chain to rank windows with, every candidate loses its
// depth bonus, and the click ends in "could not find the terminal".
//
// The console title is the missing piece, and it is exact: Windows Terminal
// renders one window per tab whose title is the tab title, and the console a
// process inherits is that tab's console. Verified on this machine against two
// live agy instances - the one in the tab titled "命令提示符 - agy" reported
// exactly "命令提示符 - agy", the one in "Windows PowerShell" reported exactly
// "Windows PowerShell".
//
// So the reporter sends the console title in the payload's optional `title`
// field. No protocol change: `title` is already "human readable title", and the
// resolver already scores an exact title match highest. The pid that produced it
// may be dead; the string is what survives.
//
// Best effort and silent: this package's charter is "never break the host CLI",
// and a jump hint is not worth an error, a detached console, or a moved stdout.

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	consoleKernel32       = syscall.NewLazyDLL("kernel32.dll")
	getConsoleTitleWProc  = consoleKernel32.NewProc("GetConsoleTitleW")
	getConsoleProcList    = consoleKernel32.NewProc("GetConsoleProcessList")
	attachConsoleProc     = consoleKernel32.NewProc("AttachConsole")
	freeConsoleProc       = consoleKernel32.NewProc("FreeConsole")
	consoleTitleMaxLength = 512
)

// ConsoleAnchorTitle returns the title of the console this reporter is running
// under, or "" when there is none worth sending.
//
// Two attempts, cheapest first:
//
//  1. our own console. The reporter is a child of the shell running inside the
//     terminal, so this is normally the host's console and needs no syscall
//     beyond GetConsoleTitleW.
//  2. the parent's console, by AttachConsole. This covers a host that gives its
//     reporter a fresh (untitled) console: the parent still sits on the tab's
//     console. The attach is bracketed by a restore, because the statusline line
//     this command exists to print has to survive it.
func ConsoleAnchorTitle() string {
	if title := sanitizeConsoleTitle(currentConsoleTitle()); title != "" {
		return title
	}
	parent := os.Getppid()
	if parent <= 0 {
		return ""
	}
	root := ownConsoleRoot()
	freeConsoleProc.Call()
	title := ""
	if ok, _, _ := attachConsoleProc.Call(uintptr(parent)); ok != 0 {
		title = currentConsoleTitle()
		freeConsoleProc.Call()
	}
	if root > 0 {
		attachConsoleProc.Call(uintptr(root))
	}
	return sanitizeConsoleTitle(title)
}

// ownConsoleRoot picks the process that created the console this reporter was
// started on: the lowest pid in GetConsoleProcessList that is not us. That is
// what a later AttachConsole restores output through.
func ownConsoleRoot() int {
	var owners [16]uint32
	count, _, _ := getConsoleProcList.Call(uintptr(unsafe.Pointer(&owners[0])), uintptr(len(owners)))
	self := os.Getpid()
	root := 0
	for index := 0; index < int(count) && index < len(owners); index++ {
		current := int(owners[index])
		if current <= 0 || current == self {
			continue
		}
		if root == 0 || current < root {
			root = current
		}
	}
	return root
}

func currentConsoleTitle() string {
	buffer := make([]uint16, consoleTitleMaxLength)
	length, _, _ := getConsoleTitleWProc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if length == 0 || length >= uintptr(len(buffer)) {
		return ""
	}
	return syscall.UTF16ToString(buffer[:length])
}
