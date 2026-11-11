package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunRefusesWithoutAPayload is the guard that matters most: a plain
// `go build ./cmd/pet-setup` must produce a binary that refuses to install,
// because the alternative is an installer that writes a half-empty tree into
// somebody's %LOCALAPPDATA% and reports success.
func TestRunRefusesWithoutAPayload(t *testing.T) {
	// The committed payload tree has no .staged marker, so the embedded FS in
	// this test binary is exactly the "plain checkout" case.
	defer silenceOutput(t)()

	for _, args := range [][]string{{}, {"--add-to-path"}} {
		if code := run(args, os.Stdout, os.Stderr); code != exitNotBuilt {
			t.Fatalf("args %v: exit %d, want %d (no payload)", args, code, exitNotBuilt)
		}
	}
}

func TestVersionAndHelpDoNotNeedAPayload(t *testing.T) {
	defer silenceOutput(t)()

	if code := run([]string{"--version"}, os.Stdout, os.Stderr); code != exitOK {
		t.Fatalf("--version: exit %d", code)
	}
	if code := run([]string{"--help"}, os.Stdout, os.Stderr); code != exitOK {
		t.Fatalf("--help: exit %d", code)
	}
	// -h and -v are the forms people type first.
	if code := run([]string{"-h"}, os.Stdout, os.Stderr); code != exitOK {
		t.Fatalf("-h: exit %d", code)
	}
	if code := run([]string{"-v"}, os.Stdout, os.Stderr); code != exitOK {
		t.Fatalf("-v: exit %d", code)
	}
}

func TestUnknownOptionIsUsageNotPowerShell(t *testing.T) {
	defer silenceOutput(t)()

	code := run([]string{"--bogus"}, os.Stdout, os.Stderr)
	if code != exitUsage {
		t.Fatalf("exit %d, want %d", code, exitUsage)
	}
}

// TestParseArgsMapsOntoInstallerSwitches pins the forwarding table. Each of
// these strings is passed verbatim to scripts\install.ps1, so a typo here
// becomes a PowerShell parameter-binding error in front of the user.
func TestParseArgsMapsOntoInstallerSwitches(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{nil, nil},
		{[]string{"--install"}, nil},
		{[]string{"--verify"}, []string{"-VerifyOnly"}},
		{[]string{"--verify-only"}, []string{"-VerifyOnly"}},
		{[]string{"--uninstall"}, []string{"-Uninstall"}},
		{[]string{"--add-to-path"}, []string{"-AddToPath"}},
		{[]string{"--ADD-TO-PATH"}, []string{"-AddToPath"}},
	}
	for _, tc := range cases {
		got, err := parseArgs(tc.args)
		if err != nil {
			t.Fatalf("args %v: %v", tc.args, err)
		}
		if got.mode != modeInstall {
			t.Fatalf("args %v: mode %v, want install", tc.args, got.mode)
		}
		if strings.Join(got.passthrough, "|") != strings.Join(tc.want, "|") {
			t.Fatalf("args %v: passthrough %v, want %v", tc.args, got.passthrough, tc.want)
		}
	}
}

// TestParseArgsDoesNotForwardUninstallForce: deletion still asks, even from a
// setup exe. -Force is a deliberate spelling, not a side effect of --uninstall.
func TestParseArgsDoesNotForwardUninstallForce(t *testing.T) {
	got, err := parseArgs([]string{"--uninstall"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.passthrough) != 1 || got.passthrough[0] != "-Uninstall" {
		t.Fatalf("passthrough %v, want exactly -Uninstall", got.passthrough)
	}
}

func TestParseSumsAcceptsTheSelfTestFormat(t *testing.T) {
	data := []byte(strings.Join([]string{
		"# written by scripts/package.ps1",
		"HASH-ONE  bin/AI-Terminal-Pet.exe",
		"HASH-TWO  scripts/install.ps1",
		"",
	}, "\n"))
	// Real hashes are 64 hex chars; this uses placeholders on purpose to keep
	// the parser's two independent failures separate in this test.
	if _, err := parseSums(data); err == nil {
		t.Fatal("a 7-character hash must be rejected")
	}
}

func TestParseSumsRejectsMalformedLines(t *testing.T) {
	good := strings.Repeat("a", sha256.Size*2)
	cases := []struct {
		name string
		body string
	}{
		{"no double space", good + " bin/x.exe"},
		{"non-hex hash", strings.Repeat("z", sha256.Size*2) + "  bin/x.exe"},
		{"empty", ""},
		{"no path", good + "  "},
	}
	for _, tc := range cases {
		if _, err := parseSums([]byte(tc.body)); err == nil {
			t.Fatalf("%s: expected an error", tc.name)
		}
	}
}

func TestParseSumsKeepsPathsWithSingleSpaces(t *testing.T) {
	hash := strings.Repeat("A", sha256.Size*2)
	lines, err := parseSums([]byte(hash + "  docs/assets/my file.png\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d line(s)", len(lines))
	}
	if lines[0].Path != "docs/assets/my file.png" {
		t.Fatalf("path %q lost its content", lines[0].Path)
	}
	// The hash is normalised to lowercase so a summary written by a tool that
	// uppercases still compares equal to crypto/sha256's output.
	if lines[0].Hash != strings.ToLower(hash) {
		t.Fatalf("hash %q was not lowercased", lines[0].Hash)
	}
}

// TestVerifyPayloadCatchesEachFailure: three separate failures, because "the
// payload is wrong" has to name which file and which way.
func TestVerifyPayloadCatchesEachFailure(t *testing.T) {
	body := []byte("hello")
	hash := sha256.Sum256(body)
	sums := []byte(hex.EncodeToString(hash[:]) + "  a.txt\n")

	t.Run("complete", func(t *testing.T) {
		files := map[string][]byte{"a.txt": body, sumsName: sums}
		if err := verifyPayload(files); err != nil {
			t.Fatalf("a matching payload was rejected: %v", err)
		}
	})

	t.Run("wrong bytes", func(t *testing.T) {
		files := map[string][]byte{"a.txt": []byte("HELLO"), sumsName: sums}
		err := verifyPayload(files)
		if err == nil || !strings.Contains(err.Error(), "hash differs: a.txt") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		files := map[string][]byte{sumsName: sums}
		err := verifyPayload(files)
		if err == nil || !strings.Contains(err.Error(), "missing: a.txt") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("extra file", func(t *testing.T) {
		files := map[string][]byte{"a.txt": body, "smuggled.dll": []byte("x"), sumsName: sums}
		err := verifyPayload(files)
		if err == nil || !strings.Contains(err.Error(), "not listed in") {
			t.Fatalf("got %v", err)
		}
	})
}

// TestExtractPayloadRefusesToEscape mirrors the ".." guard. A payload coming
// from a future zip rather than embed.FS would not get that guarantee for free.
func TestExtractPayloadRefusesToEscape(t *testing.T) {
	stage := t.TempDir()
	err := extractPayload(map[string][]byte{"../escape.txt": []byte("x")}, stage)
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("got %v", err)
	}
}

func TestExtractPayloadWritesNestedPaths(t *testing.T) {
	stage := t.TempDir()
	files := map[string][]byte{
		"bin/AI-Terminal-Pet.exe":       []byte("exe"),
		"scripts/install.ps1":           []byte("ps1"),
		"integrations/x/locale/en.json": []byte("{}"),
	}
	if err := extractPayload(files, stage); err != nil {
		t.Fatal(err)
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(stage, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s: content differs", name)
		}
	}
}

// silenceOutput keeps the report lines out of `go test` output and restores
// the process streams afterwards. The program writes to *os.File, so the only
// way to capture it is to point the file descriptors at /dev/null equivalents.
func silenceOutput(t *testing.T) func() {
	t.Helper()
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	originalOut, originalErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = null, null
	return func() {
		os.Stdout, os.Stderr = originalOut, originalErr
		null.Close()
	}
}
