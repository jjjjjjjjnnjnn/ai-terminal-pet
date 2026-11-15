package home

// The case table, checked on every platform: the resolver is pure, so the same
// expectations run whether the machine is Windows, macOS or Linux.
//
// The Windows rows are the important ones. Every existing install has its state
// in %LOCALAPPDATA%\AI-Terminal-Pet, and the packaged plugin copies files there on
// first use - so a change to that row would orphan a config.json, a runtime/
// directory and a list of muted sessions on every machine that already upgraded.
// That is why the table asserts the Windows answer rather than assuming it.

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDirectoryForMatchesEachPlatformsConvention(t *testing.T) {
	cases := []struct {
		name        string
		goos        string
		localApp    string
		xdg         string
		homeDir     string
		wantSuffix  []string
		wantExactly string
	}{
		{
			name:        "windows uses LOCALAPPDATA and keeps the capitalised name",
			goos:        "windows",
			localApp:    `C:\Users\example\AppData\Local`,
			homeDir:     `C:\Users\example`,
			wantExactly: `C:\Users\example\AppData\Local\AI-Terminal-Pet`,
		},
		{
			name:        "windows falls back to the profile when LOCALAPPDATA is missing",
			goos:        "windows",
			homeDir:     `C:\Users\example`,
			wantExactly: `C:\Users\example\AppData\Local\AI-Terminal-Pet`,
		},
		{
			name:       "windows with nothing at all is empty, not relative",
			goos:       "windows",
			wantSuffix: []string{""},
		},
		{
			name:       "darwin uses Application Support",
			goos:       "darwin",
			homeDir:    "/Users/example",
			wantSuffix: []string{"Library", "Application Support", "AI-Terminal-Pet"},
		},
		{
			name:       "linux honours XDG_DATA_HOME and lower-cases the folder",
			goos:       "linux",
			xdg:        "/home/example/.xdg",
			homeDir:    "/home/example",
			wantSuffix: []string{".xdg", "ai-terminal-pet"},
		},
		{
			name:       "linux defaults to .local/share",
			goos:       "linux",
			homeDir:    "/home/example",
			wantSuffix: []string{".local", "share", "ai-terminal-pet"},
		},
		{
			name:       "linux ignores a relative XDG_DATA_HOME",
			goos:       "linux",
			xdg:        "relative/xdg",
			homeDir:    "/home/example",
			wantSuffix: []string{".local", "share", "ai-terminal-pet"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := DirectoryFor(testCase.goos, testCase.localApp, testCase.xdg, testCase.homeDir)
			if testCase.wantExactly != "" {
				// Windows paths are compared as strings: filepath.Join on a
				// linux test host would normalise them differently.
				if normalizeForCompare(got) != normalizeForCompare(testCase.wantExactly) {
					t.Fatalf("got %q, want %q", got, testCase.wantExactly)
				}
				return
			}
			if len(testCase.wantSuffix) == 1 && testCase.wantSuffix[0] == "" {
				if got != "" {
					t.Fatalf("got %q, want an empty directory", got)
				}
				return
			}
			if !suffixMatches(got, testCase.wantSuffix) {
				t.Fatalf("got %q, want a path ending in %v", got, testCase.wantSuffix)
			}
		})
	}
}

func TestDirectoryNameCaseFollowsTheConvention(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		name := DirectoryNameFor(goos)
		if WantsLowercaseName(goos) {
			if name != strings.ToLower(name) {
				t.Fatalf("%s should use a lower-case folder name, got %q", goos, name)
			}
			continue
		}
		if name != WindowsDirectoryName {
			t.Fatalf("%s should use %q, got %q", goos, WindowsDirectoryName, name)
		}
	}
}

func TestJoinBuildsInsideTheDirectory(t *testing.T) {
	if got := Join("runtime"); got != filepath.Join(Directory(), "runtime") {
		t.Fatalf("Join(runtime) = %q", got)
	}
	for _, element := range []string{"config.json", "state.json", "approvals", "alerts.json", "hang.log"} {
		path := Join(element)
		if path == "" || filepath.Base(path) != element {
			t.Fatalf("Join(%q) = %q", element, path)
		}
	}
}

// normalizeForCompare makes a Windows path comparable no matter which host runs
// the test: backslashes become slashes and trailing separators are dropped.
func normalizeForCompare(value string) string {
	return strings.TrimRight(strings.ReplaceAll(value, `\`, "/"), "/")
}

// suffixMatches compares a resolved path against the segments it must end with,
// after normalising separators.
func suffixMatches(value string, segments []string) bool {
	normalized := normalizeForCompare(value)
	wanted := normalizeForCompare(filepath.Join(segments...))
	return strings.HasSuffix(normalized, wanted)
}

// The one platform-dependent assertion: on the host this test runs on, Directory
// must agree with DirectoryFor for its own GOOS.
func TestDirectoryAgreesWithTheRunningPlatform(t *testing.T) {
	got := Directory()
	if got == "" {
		t.Fatalf("a test host with no home directory cannot run this suite (%s)", runtime.GOOS)
	}
	want := DirectoryFor(runtime.GOOS, "", "", mustHomeDir(t))
	if normalizeForCompare(got) == normalizeForCompare(want) {
		return
	}
	// LOCALAPPDATA (windows) or XDG_DATA_HOME (linux) being set legitimately
	// changes the answer; assert the shape instead of the exact string.
	if !suffixMatches(got, []string{DirectoryNameFor(runtime.GOOS)}) {
		t.Fatalf("Directory() = %q, which does not end in %q", got, DirectoryNameFor(runtime.GOOS))
	}
}

func mustHomeDir(t *testing.T) string {
	t.Helper()
	homeDir, err := osUserHomeDir()
	if err != nil || homeDir == "" {
		t.Fatalf("no home directory: %v", err)
	}
	return homeDir
}
