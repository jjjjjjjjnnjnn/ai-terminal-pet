//go:build windows

package platform

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	th32csSnapProcess = 0x00000002
	swRestore         = 9
)

var (
	focusUser32 = syscall.NewLazyDLL("user32.dll")

	createToolhelp32SnapshotProc = kernel32.NewProc("CreateToolhelp32Snapshot")
	process32FirstProc           = kernel32.NewProc("Process32FirstW")
	process32NextProc            = kernel32.NewProc("Process32NextW")
	enumWindowsProc              = focusUser32.NewProc("EnumWindows")
	isWindowVisibleProc          = focusUser32.NewProc("IsWindowVisible")
	getWindowThreadProcessIdProc = focusUser32.NewProc("GetWindowThreadProcessId")
	getClassNameProc             = focusUser32.NewProc("GetClassNameW")
	getWindowTextProc            = focusUser32.NewProc("GetWindowTextW")
	isIconicProc                 = focusUser32.NewProc("IsIconic")
	showWindowProc               = focusUser32.NewProc("ShowWindow")
	setForegroundWindowProc      = focusUser32.NewProc("SetForegroundWindow")
	getForegroundWindowProc      = focusUser32.NewProc("GetForegroundWindow")
)

type processEntry32W struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

// ActivateTarget brings exactly one window forward on the UI thread.
func ActivateTarget(handle uintptr) bool {
	if handle == 0 {
		return false
	}
	if minimized, _, _ := isIconicProc.Call(handle); minimized != 0 {
		showWindowProc.Call(handle, swRestore)
	}
	if result, _, _ := setForegroundWindowProc.Call(handle); result != 0 {
		return true
	}
	now, _, _ := getForegroundWindowProc.Call()
	return now == handle
}

func snapshotProcesses() map[int]focusProcess {
	result := make(map[int]focusProcess)
	handle, _, _ := createToolhelp32SnapshotProc.Call(th32csSnapProcess, 0)
	if handle == 0 || handle == ^uintptr(0) {
		return result
	}
	defer closeHandle.Call(handle)
	var entry processEntry32W
	entry.Size = uint32(unsafe.Sizeof(entry))
	ok, _, _ := process32FirstProc.Call(handle, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		if entry.ProcessID != 0 {
			result[int(entry.ProcessID)] = focusProcess{
				parent: int(entry.ParentProcessID),
				name:   syscall.UTF16ToString(entry.ExeFile[:]),
			}
		}
		ok, _, _ = process32NextProc.Call(handle, uintptr(unsafe.Pointer(&entry)))
	}
	return result
}

var (
	focusEnumMu     sync.Mutex
	focusEnumTarget *[]focusWindow
	focusEnumProcs  map[int]focusProcess
)

var focusEnumCallback = syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
	if focusEnumTarget == nil {
		return 1
	}
	if visible, _, _ := isWindowVisibleProc.Call(hwnd); visible == 0 {
		return 1
	}
	var pid uint32
	getWindowThreadProcessIdProc.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	classBuf := make([]uint16, 256)
	getClassNameProc.Call(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), uintptr(len(classBuf)))
	titleBuf := make([]uint16, 512)
	getWindowTextProc.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), uintptr(len(titleBuf)))
	*focusEnumTarget = append(*focusEnumTarget, focusWindow{
		handle: hwnd,
		pid:    int(pid),
		class:  syscall.UTF16ToString(classBuf),
		title:  syscall.UTF16ToString(titleBuf),
		exe:    focusEnumProcs[int(pid)].name,
	})
	return 1
})

func visibleWindows(processes map[int]focusProcess) []focusWindow {
	focusEnumMu.Lock()
	defer focusEnumMu.Unlock()
	collected := make([]focusWindow, 0, 32)
	focusEnumTarget = &collected
	focusEnumProcs = processes
	enumWindowsProc.Call(uintptr(focusEnumCallback), 0)
	focusEnumTarget = nil
	focusEnumProcs = nil
	return collected
}
