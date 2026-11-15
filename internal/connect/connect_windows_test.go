//go:build windows

package connect

// The assertions that describe Windows path spelling.
//
// A Windows path and a POSIX path are spelled differently on purpose: that split
// is localpath_<goos>.go, and it is what lets the checks themselves run on every
// platform. These cases therefore belong to this platform; the portable ones stay
// in connect_test.go.
//
// Two of them are golden samples:
// TestPathToFileURLReproducesTheWorkingWrapper pins the exact string this machine's
// OpenCode wrapper carried (handover §53.1), which is the input the encoder has to
// match and the output --connect prints.

import (
	"strings"
	"testing"
)

func TestPathToFileURLReproducesTheWorkingWrapper(t *testing.T) {
	// The exact text this machine's index.ts carried when OpenCode loaded it.
	const want = "file:///C:/Users/example/Desktop/%E9%A1%B9%E7%9B%AE%E5%BC%80%E5%8F%91/ai-terminal-pet/reporters/opencode.js"
	local := `C:\Users\example\Desktop\项目开发\ai-terminal-pet\reporters\opencode.js`

	got, err := PathToFileURL(local)
	if err != nil {
		t.Fatalf("PathToFileURL: %v", err)
	}
	if got != want {
		t.Fatalf("encoding drifted from the wrapper OpenCode actually loaded\n got: %s\nwant: %s", got, want)
	}
	// And back: the check and the snippet have to agree, or --connect would
	// print a reference --doctor then calls missing.
	decoded, err := FileURLToPath(got)
	if err != nil {
		t.Fatalf("FileURLToPath: %v", err)
	}
	if strings.ReplaceAll(decoded, `\`, "/") != strings.ReplaceAll(local, `\`, "/") {
		t.Fatalf("round trip lost the path: %q", decoded)
	}
}

func TestPathToFileURLRejectsWhatItCannotExpress(t *testing.T) {
	for _, input := range []string{"", "   ", "reporters/opencode.js", "./relative/path.js"} {
		if _, err := PathToFileURL(input); err == nil {
			t.Fatalf("%q was accepted; a host cannot resolve a relative file URL", input)
		}
	}
	unc, err := PathToFileURL(`\\server\share\plugins\opencode.js`)
	if err != nil {
		t.Fatalf("UNC: %v", err)
	}
	if !strings.HasPrefix(unc, "file://server/share/") {
		t.Fatalf("UNC did not become a host form: %s", unc)
	}
}

func TestFileURLToPathAcceptsEveryShapeAHostWrites(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"three slashes", "file:///C:/Users/x/reporters/opencode.js", `C:\Users\x\reporters\opencode.js`},
		{"one slash", "file:C:/Users/x/reporters/opencode.js", `C:\Users\x\reporters\opencode.js`},
		{"quoted", `"C:/Users/x/reporters/opencode.js"`, `C:\Users\x\reporters\opencode.js`},
		{"bare path", `%LOCALAPPDATA%\AI-Terminal-Pet\pet-hook.exe`, `%LOCALAPPDATA%\AI-Terminal-Pet\pet-hook.exe`},
		{"percent encoded", "file:///C:/Users/x/%E9%A1%B9%E7%9B%AE/opencode.js", `C:\Users\x\项目\opencode.js`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := FileURLToPath(testCase.input)
			if err != nil {
				t.Fatalf("FileURLToPath(%q): %v", testCase.input, err)
			}
			if got != testCase.want {
				t.Fatalf("FileURLToPath(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
	if _, err := FileURLToPath("https://example.com/plugin.js"); err == nil {
		t.Fatal("a non-file URL was accepted")
	}
}

func TestAbbreviateShortensTheProfilesOwnDirectories(t *testing.T) {
	roots := Roots{Home: `C:\Users\example`, LocalAppData: `C:\Users\example\AppData\Local`}
	cases := []struct {
		input string
		want  string
	}{
		{`C:\Users\example\.claude\settings.json`, "~/.claude/settings.json"},
		{`C:\Users\example\AppData\Local\AI-Terminal-Pet\pet-hook.exe`, "%LOCALAPPDATA%/AI-Terminal-Pet/pet-hook.exe"},
		{`D:\elsewhere\deep\tree\a.json`, ".../tree/a.json"},
		{"", "(no path)"},
	}
	for _, testCase := range cases {
		if got := abbreviate(roots, testCase.input); got != testCase.want {
			t.Fatalf("abbreviate(%q) = %q, want %q", testCase.input, got, testCase.want)
		}
	}
}

func TestExpandUserResolvesEveryWindowsForm(t *testing.T) {
	roots := Roots{Home: `C:\Users\example`, LocalAppData: `C:\Users\example\AppData\Local`}
	cases := map[string]string{
		`%LOCALAPPDATA%\AI-Terminal-Pet\pet-hook.exe`:    `C:\Users\example\AppData\Local\AI-Terminal-Pet\pet-hook.exe`,
		`$env:LOCALAPPDATA\AI-Terminal-Pet\pet-hook.exe`: `C:\Users\example\AppData\Local\AI-Terminal-Pet\pet-hook.exe`,
		`$LOCALAPPDATA\AI-Terminal-Pet\pet-hook.exe`:     `C:\Users\example\AppData\Local\AI-Terminal-Pet\pet-hook.exe`,
		`~/.claude/hooks/terminal-pet-hook.js`:           `C:\Users\example\.claude\hooks\terminal-pet-hook.js`,
		`"C:\Users\example\.claude\settings.json"`:       `C:\Users\example\.claude\settings.json`,
	}
	for input, want := range cases {
		if got := expandUser(roots, input); got != want {
			t.Fatalf("expandUser(%q) = %q, want %q", input, got, want)
		}
	}
}
