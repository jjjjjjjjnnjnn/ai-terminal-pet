// pet-setup is the single-file installer the release ships.
//
// It embeds the same payload the portable zip carries - three binaries,
// protocol.js, the Harness bundle, and install.ps1/install.cmd - unpacks it into
// a temporary directory, and then runs the installer that ships inside it.
//
// That indirection is the whole design. scripts/install.ps1 stays the only
// implementation of "what gets installed and how it is verified": the setup
// binary does not reimplement the artifact list, the hashing, the stop-the-
// process-retry loop, the HKCU Run entry or the PATH edit. A second
// implementation would drift from the first, and the drift would be invisible
// until a user's -VerifyOnly reported a mismatch the installer never wrote.
// The cost is that pet-setup needs PowerShell on the machine, which every
// Windows install has and which the portable zip already needs.
//
// What pet-setup adds on top of the zip is the one thing the zip cannot do:
// verify the payload it is about to install against the SHA256SUMS it carries,
// before anything is written. A downloaded zip can be missing a file or carry a
// truncated one and the user finds out later; a self-extracting exe finds out
// before it touches %LOCALAPPDATA%.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// buildVersion is stamped by scripts/package.ps1 through
// `-ldflags "-X main.buildVersion=<overlay version>"`. It stays "dev" for a
// plain `go build`, so a locally built setup exe never claims to be a release.
var buildVersion = "dev"

const (
	payloadRoot     = "payload"
	stagedMarker    = ".staged"
	keepAliveMarker = ".gitkeep"
	sumsName        = "SHA256SUMS.txt"
	installerEntry  = `scripts\install.ps1`

	exitOK       = 0
	exitFailed   = 1
	exitUsage    = 2
	exitNotBuilt = 3
)

const usageText = `pet-setup - install AI Terminal Pet

Usage:
  pet-setup                 install (or upgrade) into %LOCALAPPDATA%\AI-Terminal-Pet
  pet-setup --add-to-path   install, then add the install directory to HKCU Path
  pet-setup --verify        check the installed files against the payload hashes
  pet-setup --uninstall     stop the overlay and delete the install directory
  pet-setup --version       print the version of the payload this build carries
  pet-setup --help          print this text

Every option is forwarded to scripts\install.ps1, which does the actual work;
this program only unpacks its payload and verifies it first.
`

// summaryLine is one parsed line of SHA256SUMS.txt.
type summaryLine struct {
	Hash string
	Path string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main without the exit, so the tests can drive it.
func run(args []string, stdout, stderr *os.File) int {
	options, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "[fail] %v\n\n%s", err, usageText)
		return exitUsage
	}

	switch options.mode {
	case modeHelp:
		fmt.Fprint(stdout, usageText)
		return exitOK
	case modeVersion:
		fmt.Fprintf(stdout, "pet-setup %s\n", buildVersion)
		return exitOK
	}

	// A setup exe built without the packaging step would silently install a
	// placeholder file. Refusing is the only honest behaviour, and the message
	// says which command produces a usable build.
	if !payloadReady() {
		fmt.Fprintf(stderr, "[fail] pet-setup: this build carries no payload.\n")
		fmt.Fprintf(stderr, "       Build it with scripts\\package.ps1, which stages the files and then compiles this program.\n")
		return exitNotBuilt
	}

	payload, err := readPayload()
	if err != nil {
		fmt.Fprintf(stderr, "[fail] pet-setup: %v\n", err)
		return exitFailed
	}

	fmt.Fprintf(stdout, "pet-setup %s: %d payload file(s)\n", buildVersion, len(payload))

	// Verify before writing anything. This is the one check the portable zip
	// cannot perform, and it is cheap: 5 MB of SHA256 over files already in
	// memory because they are embedded.
	if err := verifyPayload(payload); err != nil {
		fmt.Fprintf(stderr, "[fail] pet-setup: the payload does not match its SHA256SUMS.txt: %v\n", err)
		fmt.Fprintf(stderr, "       Nothing was written. Re-download the installer; the file you have is damaged.\n")
		return exitFailed
	}
	fmt.Fprintln(stdout, "[ok] payload matches SHA256SUMS.txt")

	stage, err := os.MkdirTemp("", "ai-terminal-pet-setup-")
	if err != nil {
		fmt.Fprintf(stderr, "[fail] pet-setup: cannot create a temporary directory: %v\n", err)
		return exitFailed
	}
	defer os.RemoveAll(stage)

	if err := extractPayload(payload, stage); err != nil {
		fmt.Fprintf(stderr, "[fail] pet-setup: cannot unpack into %s: %v\n", stage, err)
		return exitFailed
	}

	installer := filepath.Join(stage, installerEntry)
	if _, err := os.Stat(installer); err != nil {
		fmt.Fprintf(stderr, "[fail] pet-setup: the payload has no %s: %v\n", installerEntry, err)
		return exitFailed
	}

	code, err := runInstaller(installer, options.passthrough, stdout, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "[fail] pet-setup: cannot start PowerShell: %v\n", err)
		fmt.Fprintf(stderr, "       The installer is a PowerShell script; Windows PowerShell is part of every Windows install.\n")
		return exitFailed
	}
	return code
}

type mode int

const (
	modeInstall mode = iota
	modeHelp
	modeVersion
)

type options struct {
	mode        mode
	passthrough []string
}

// parseArgs maps the documented options onto install.ps1's own switches. It
// deliberately accepts nothing else: an unknown flag is a typo, and forwarding a
// typo to PowerShell would produce a PowerShell error message instead of a
// usage message.
func parseArgs(args []string) (options, error) {
	for _, arg := range args {
		switch strings.ToLower(strings.TrimSpace(arg)) {
		case "", "--install":
			// The default; --install is accepted so the help text can name it.
		case "--help", "-h", "/?", "--h":
			return options{mode: modeHelp}, nil
		case "--version", "-v":
			return options{mode: modeVersion}, nil
		case "--verify", "--verify-only":
			return options{mode: modeInstall, passthrough: []string{"-VerifyOnly"}}, nil
		case "--uninstall":
			// -Force is not implied: the prompt is the last chance to stop
			// before a directory is deleted, and an installer that deletes
			// without asking is a worse default than one extra keystroke.
			return options{mode: modeInstall, passthrough: []string{"-Uninstall"}}, nil
		case "--add-to-path":
			return options{mode: modeInstall, passthrough: []string{"-AddToPath"}}, nil
		default:
			return options{}, fmt.Errorf("unknown option %q", arg)
		}
	}
	return options{mode: modeInstall}, nil
}

// payloadReady reports whether this build carries a staged payload or the
// committed placeholder. The marker is written by scripts/package.ps1 and is
// embedded through `all:payload`, which keeps dotfiles.
func payloadReady() bool {
	_, err := fs.Stat(payloadFS, payloadRoot+"/"+stagedMarker)
	return err == nil
}

// readPayload returns every embedded file except the marker, keyed by a
// forward-slash path relative to the staging root.
func readPayload() (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(payloadFS, payloadRoot, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(payloadRoot, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		// The staging marker is bookkeeping, and .gitkeep exists only so a
		// fresh checkout has something to embed. Neither is an installed file,
		// and neither is listed in SHA256SUMS.txt, so both are dropped here
		// rather than special-cased in the verification.
		if relative == stagedMarker || relative == keepAliveMarker {
			return nil
		}
		data, err := payloadFS.ReadFile(name)
		if err != nil {
			return err
		}
		files[relative] = data
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("the payload contains no files")
	}
	return files, nil
}

// parseSums reads SHA256SUMS.txt, which scripts/package.ps1 writes as
// "<hash>  <relative path in forward slashes>", the same shape `sha256sum` and
// pack.ps1 both use.
func parseSums(data []byte) ([]summaryLine, error) {
	var lines []summaryLine
	for _, raw := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Two spaces separate the hash from the path so a path containing a
		// single space still splits correctly.
		parts := strings.SplitN(trimmed, "  ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed line %q", trimmed)
		}
		hash := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		if len(hash) != sha256.Size*2 {
			return nil, fmt.Errorf("%q does not look like a sha256", hash)
		}
		if _, err := hex.DecodeString(hash); err != nil {
			return nil, fmt.Errorf("%q is not hexadecimal", hash)
		}
		if name == "" {
			return nil, fmt.Errorf("line %q has no path", trimmed)
		}
		lines = append(lines, summaryLine{Hash: strings.ToLower(hash), Path: name})
	}
	if len(lines) == 0 {
		return nil, errors.New("SHA256SUMS.txt is empty")
	}
	return lines, nil
}

// verifyPayload checks every file the summary names, and checks that the
// payload carries no extra file the summary does not. The second half matters:
// "all listed files match" is satisfied by a payload with the right files plus
// anything else somebody appended.
func verifyPayload(files map[string][]byte) error {
	data, present := files[sumsName]
	if !present {
		return fmt.Errorf("the payload does not carry %s, so nothing can be checked", sumsName)
	}
	sums, err := parseSums(data)
	if err != nil {
		return fmt.Errorf("%s: %w", sumsName, err)
	}

	// Which files to check is decided by the summary when it is complete, and
	// by the payload otherwise. scripts/package.ps1 regenerates the summary on
	// every run, so in a real release both sets agree; the fallback exists
	// because a summary that omits a file is itself the bug worth reporting.
	listed := make(map[string]string, len(sums))
	for _, line := range sums {
		listed[filepath.ToSlash(line.Path)] = line.Hash
	}

	var mismatches []string
	for name, want := range listed {
		data, present := files[name]
		if !present {
			mismatches = append(mismatches, "missing: "+name)
			continue
		}
		got := sha256.Sum256(data)
		if hex.EncodeToString(got[:]) != want {
			mismatches = append(mismatches, "hash differs: "+name)
		}
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		return fmt.Errorf("%d problem(s): %s", len(mismatches), strings.Join(mismatches, "; "))
	}

	// Extra files, ignoring the summary itself (it cannot list its own hash) and
	// the staging marker.
	for name := range files {
		if name == sumsName || name == stagedMarker {
			continue
		}
		if _, ok := listed[name]; !ok {
			mismatches = append(mismatches, "not listed in "+sumsName+": "+name)
		}
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		return fmt.Errorf("%d problem(s): %s", len(mismatches), strings.Join(mismatches, "; "))
	}
	return nil
}

// extractPayload writes the files under dest, refusing any path that would
// escape it. embed.FS cannot produce ".." today, but the check belongs here
// rather than in a comment: this function is the only thing standing between an
// archive and the user's filesystem, and a future payload source (a downloaded
// zip, say) would not get the same guarantee for free.
func extractPayload(files map[string][]byte, dest string) error {
	for name, data := range files {
		clean := filepath.Clean(filepath.FromSlash(name))
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("refusing to write %q: it escapes the target directory", name)
		}
		target := filepath.Join(dest, clean)
		if parent := filepath.Dir(target); parent != "" {
			if err := os.MkdirAll(parent, 0o755); err != nil {
				return err
			}
			// 0o755/0o644 are placeholders on Windows (NTFS inherits the
			// parent's ACL), so they are used for readability, not security.
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// runInstaller starts PowerShell over scripts\install.ps1 and waits.
//
// -ExecutionPolicy Bypass is scoped to this one process and changes no
// machine or user policy; the setup exe unpacks its own .ps1, so on a stock
// RemoteSigned machine the same rule that refuses a downloaded script applies
// here. Windows PowerShell is preferred over pwsh: it is present on every
// Windows install, while PowerShell 7 is an optional package. The returned
// error is only for "cannot start"; a non-zero installer exit code is passed
// straight through so the caller's exit code means what install.ps1 said.
func runInstaller(script string, extra []string, stdout, stderr *os.File) (int, error) {
	candidates := []string{"powershell.exe", "pwsh.exe"}
	var lastErr error
	for _, shell := range candidates {
		path, err := exec.LookPath(shell)
		if err != nil {
			lastErr = err
			continue
		}
		args := []string{"-NoProfile", "-NoLogo", "-ExecutionPolicy", "Bypass", "-File", script}
		args = append(args, extra...)
		command := exec.Command(path, args...)
		command.Stdin = os.Stdin
		command.Stdout = stdout
		command.Stderr = stderr
		err = command.Run()
		if err == nil {
			return exitOK, nil
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no PowerShell found on PATH")
	}
	return exitFailed, lastErr
}
