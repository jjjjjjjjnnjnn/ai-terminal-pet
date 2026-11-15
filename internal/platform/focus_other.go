//go:build !windows && !linux && !darwin

package platform

func snapshotProcesses() map[int]focusProcess {
	return map[int]focusProcess{}
}

func visibleWindows(_ map[int]focusProcess) []focusWindow {
	return nil
}

// SampleForeground is unavailable on this platform.
func SampleForeground() ForegroundSample {
	return ForegroundSample{}
}

// ActivateTarget is unavailable on this platform.
func ActivateTarget(_ uintptr) bool {
	return false
}
