package home

// The resolver in the platform-neutral form (this is the file that says out loud
// what each platform's convention is, and why the Windows one must not move).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DirectoryName is the folder this project owns. It is the same on every
// platform except the Linux XDG case, where the convention is lower-case; that
// difference is deliberate, because a stray capitalised directory in
// ~/.local/share is not what any other Linux application does.
const (
	WindowsDirectoryName = "AI-Terminal-Pet"
	XDGDirectoryName     = "ai-terminal-pet"
)

// Directory returns this project's data directory on the running platform.
//
// An empty result means the platform gave us nothing to work with (a Windows
// process with no LOCALAPPDATA and no profile, or no resolvable home directory
// anywhere); every caller treats that as "no disk state" and becomes a no-op
// rather than guessing a relative path.
func Directory() string {
	homeDir, _ := os.UserHomeDir()
	return DirectoryFor(runtime.GOOS, os.Getenv("LOCALAPPDATA"), os.Getenv("XDG_DATA_HOME"), homeDir)
}

// DirectoryFor is the pure resolver the tests drive: the same rules, with the
// platform and the environment supplied instead of read.
//
// The Windows answer is deliberately what it always was. Every existing install
// keeps config.json, state.json, runtime/, approvals/ and hang.log in
// %LOCALAPPDATA%\AI-Terminal-Pet, and the packaged DSH plugin copies its overlay
// into that directory on first use - so moving it would orphan the state of every
// machine that upgrades. macOS and Linux are new, and take the conventions their
// users expect.
func DirectoryFor(goos, localAppData, xdgDataHome, homeDir string) string {
	switch goos {
	case "windows":
		return windowsDirectory(localAppData, homeDir)
	case "darwin":
		return darwinDirectory(homeDir)
	default:
		return linuxDirectory(xdgDataHome, homeDir)
	}
}

// windowsDirectory is %LOCALAPPDATA%\AI-Terminal-Pet.
//
// A missing LOCALAPPDATA yields an empty directory - no state, not a guessed path.
// That is the behaviour this project has always had, and two reporter tests pin
// it: an approval request that cannot be filed must answer deny, and a payload that
// cannot be filed must not be sent. Falling back to the profile (which an earlier
// draft of this file did) would have made those two cases unreachable on Windows.
func windowsDirectory(localAppData, homeDir string) string {
	if localAppData == "" {
		return ""
	}
	return joinNative(localAppData, WindowsDirectoryName)
}

// darwinDirectory is ~/Library/Application Support/AI-Terminal-Pet.
func darwinDirectory(homeDir string) string {
	if homeDir == "" {
		return ""
	}
	return joinNative(homeDir, "Library", "Application Support", WindowsDirectoryName)
}

// linuxDirectory is $XDG_DATA_HOME/ai-terminal-pet, or ~/.local/share/ai-terminal-pet.
//
// XDG_DATA_HOME is honoured only when it is absolute: the specification says a
// relative value must be ignored, and a relative value would put the overlay's
// state next to whatever directory it happened to start in.
//
// The absolute test is spelled by hand rather than with filepath.IsAbs, because
// IsAbs answers for the *running* platform: filepath.IsAbs("/x") is false on
// Windows, which would make this function answer differently depending on which
// machine compiled it. The rule here is XDG's, so it is written as XDG's rule.
func linuxDirectory(xdgDataHome, homeDir string) string {
	if xdgDataHome != "" && isRootedPath(xdgDataHome) {
		return posixJoin(xdgDataHome, XDGDirectoryName)
	}
	if homeDir == "" {
		return ""
	}
	return posixJoin(homeDir, ".local", "share", XDGDirectoryName)
}

// Join returns a path inside the data directory. An empty directory yields "",
// which is how every caller learns there is no disk state to touch.
func Join(elements ...string) string {
	base := Directory()
	if base == "" {
		return ""
	}
	return joinNative(append([]string{base}, elements...)...)
}

// DirectoryNameFor reports the folder name a platform uses, for the tests that pin
// the case difference.
func DirectoryNameFor(goos string) string {
	if WantsLowercaseName(goos) {
		return XDGDirectoryName
	}
	return WindowsDirectoryName
}

// WantsLowercaseName reports the one convention where the folder is lower-case,
// so a test can say so out loud instead of hard-coding it twice.
func WantsLowercaseName(goos string) bool {
	return goos != "windows" && goos != "darwin"
}

// joinNative joins with the host's separator: it is only used for the running
// platform's own answer, never to describe another platform.
func joinNative(elements ...string) string {
	return filepath.Join(elements...)
}

// isRootedPath reports whether a value starts with a separator or a drive letter,
// independent of the host's own path rules.
func isRootedPath(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '/' || value[0] == '\\' {
		return true
	}
	return len(value) >= 2 && value[1] == ':'
}

// posixJoin joins with forward slashes, for the platforms whose paths are not
// the host's paths.
func posixJoin(elements ...string) string {
	result := ""
	for _, element := range elements {
		if element == "" {
			continue
		}
		if result == "" {
			result = strings.TrimRight(element, "/\\")
			continue
		}
		result += "/" + strings.Trim(element, "/\\")
	}
	return result
}
