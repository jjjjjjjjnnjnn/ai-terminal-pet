package platform

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	processQueryLimitedInformation = 0x1000
	stillActive                    = 259
	errorAlreadyExists             = 183
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	openProcess        = kernel32.NewProc("OpenProcess")
	getExitCodeProcess = kernel32.NewProc("GetExitCodeProcess")
	closeHandle        = kernel32.NewProc("CloseHandle")
	createMutex        = kernel32.NewProc("CreateMutexW")
	freeConsole        = kernel32.NewProc("FreeConsole")
)

var ErrAlreadyRunning = errors.New("AI Terminal Pet is already running")

type Instance struct {
	handle syscall.Handle
}

func ProcessAlive(pid int) bool {
	if pid <= 0 || uint64(pid) > uint64(0xffffffff) {
		return false
	}
	handle, _, callErr := openProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if handle == 0 {
		return errors.Is(callErr, syscall.ERROR_ACCESS_DENIED)
	}
	defer closeHandle.Call(handle)
	var code uint32
	result, _, _ := getExitCodeProcess.Call(handle, uintptr(unsafe.Pointer(&code)))
	return result != 0 && code == stillActive
}

func AcquireInstance(name string) (*Instance, error) {
	if name == "" {
		return nil, errors.New("instance name is empty")
	}
	text, err := syscall.UTF16FromString(name)
	if err != nil {
		return nil, err
	}
	handle, _, callErr := createMutex.Call(0, 0, uintptr(unsafe.Pointer(&text[0])))
	if handle == 0 {
		return nil, callErr
	}
	if errors.Is(callErr, syscall.Errno(errorAlreadyExists)) {
		closeHandle.Call(handle)
		return nil, ErrAlreadyRunning
	}
	return &Instance{handle: syscall.Handle(handle)}, nil
}

func (i *Instance) Close() {
	if i == nil || i.handle == 0 {
		return
	}
	closeHandle.Call(uintptr(i.handle))
	i.handle = 0
}

func HideConsole() {
	freeConsole.Call()
}
