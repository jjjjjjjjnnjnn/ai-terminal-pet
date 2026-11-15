//go:build !windows

package connect

// The portable assertions, run on the POSIX spelling too.
//
// Two things are checked here that connect_windows_test.go cannot check:
//
//   - the same package works with a POSIX path, so --connect and the reporters
//     check are usable on linux (handover §55.6 step 3), and
//   - a backslash is a *character* in a POSIX file name, not a separator. Turning
//     it into a separator here would be the "looks the same on both" trap §55.7
//     listed, and it would make a directory named "a\b" unreachable.

import (
	"strings"
	"testing"
)

func TestPathToFileURLUsesThePlatformSpelling(t *testing.T) {
	local := "/home/pet/repo/reporters/opencode.js"
	got, err := PathToFileURL(local)
	if err != nil {
		t.Fatalf("PathToFileURL: %v", err)
	}
	// No drive letter is invented, and the three-slash form is still produced:
	// that part is the URL's grammar, not the platform's.
	if got != "file:///home/pet/repo/reporters/opencode.js" {
		t.Fatalf("got %q", got)
	}
	back, err := FileURLToPath(got)
	if err != nil || back != local {
		t.Fatalf("round trip gave %q (%v), want %q", back, err, local)
	}
}

func TestFileURLToPathKeepsBackslashesInPOSIXNames(t *testing.T) {
	// A POSIX file name may contain a backslash. It must survive both directions,
	// because connect's whole job is to name a file that exists.
	local := `/home/pet/repo/a\b/opencode.js`
	url, err := PathToFileURL(local)
	if err != nil {
		t.Fatalf("PathToFileURL: %v", err)
	}
	// The backslash is percent-encoded as data, not rewritten as a separator.
	if !strings.Contains(url, "%5C") {
		t.Fatalf("backslash was not encoded as data: %s", url)
	}
	back, err := FileURLToPath(url)
	if err != nil || back != local {
		t.Fatalf("round trip gave %q (%v), want %q", back, err, local)
	}
}

func TestPathToFileURLRefusesRelativePOSIXPaths(t *testing.T) {
	for _, input := range []string{"reporters/opencode.js", "./x.js"} {
		if _, err := PathToFileURL(input); err == nil {
			t.Fatalf("%q was accepted; a host resolves a file URL from its own directory", input)
		}
	}
}
