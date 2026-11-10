package app

// --doctor answers "what is not wired up on this machine?" in one pass.
//
// The overlay fails silently by design: a reporter that never reaches the
// pipe, an install missing protocol.js, or a Run entry pointing at a binary
// that was uninstalled all look the same from the outside - the pet simply
// does not appear. Collecting the measured value for each of those into one
// line is the difference between a bug report that can be read and one that
// has to be guessed at.
//
// Every check is independent: a broken install directory must not stop the
// pipe, node or autostart checks from running. The point is to list
// everything that is wrong in a single pass, not to stop at the first item.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	winio "github.com/Microsoft/go-winio"

	"ai-terminal-pet/internal/autostart"
	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/pipeserver"
)

const (
	doctorOK   = "ok"
	doctorWarn = "warn"
	doctorFail = "fail"

	// These names mirror scripts/install.ps1's deployment set. They are literals
	// rather than values derived from the running executable, because the
	// whole point of the check is to notice when this process is not the
	// thing that was installed.
	installDirectoryName = "AI-Terminal-Pet"
	doctorMainExe        = "AI-Terminal-Pet.exe"
	doctorStatuslineExe  = "pet-statusline.exe"
	doctorProtocolFile   = "protocol.js"

	// A diagnostic must not become the slow part of the user's day. A missing
	// listener fails instantly on Windows; the deadline only covers a server
	// that accepted a connection and then stopped responding.
	doctorDialTimeout = 250 * time.Millisecond
	// node is spawned by every reporter, so its cold start is something the
	// user already feels. 300ms covers a warm Node and keeps a hung or
	// AV-throttled one from turning --doctor into a coffee break.
	doctorNodeTimeout = 300 * time.Millisecond
)

// doctorInstallFiles is the set scripts/install.ps1 deploys into the install
// directory. Sight of any of them missing is reported as a failure because
// the installer itself treats an incomplete deployment as one.
var doctorInstallFiles = []string{doctorMainExe, doctorStatuslineExe, doctorProtocolFile}

// DoctorCheck is one line of the report: what was inspected, the verdict and
// the measured value. Detail carries the evidence, so "warn" never has to be
// taken on trust.
type DoctorCheck struct {
	Name   string
	Status string
	Detail string
}

// DoctorReport is the whole run. Checks keep the order they were run in,
// which is the order a human reads them in.
type DoctorReport struct {
	Checks []DoctorCheck
}

// DoctorExitCode is 1 when any check failed and 0 otherwise. Warnings are
// not a failure: a machine with the overlay stopped and Node absent is a
// perfectly usable install.
func (r DoctorReport) ExitCode() int {
	if _, _, fail := r.Counts(); fail > 0 {
		return 1
	}
	return 0
}

// Counts returns the tally printed on the last line.
func (r DoctorReport) Counts() (ok int, warn int, fail int) {
	for _, check := range r.Checks {
		switch check.Status {
		case doctorOK:
			ok++
		case doctorWarn:
			warn++
		default:
			// An unknown status can only come from a broken check. Counting
			// it as a failure keeps the summary honest instead of quietly
			// shortening the total.
			fail++
		}
	}
	return ok, warn, fail
}

// Text renders the human-readable report: one line per check and a summary
// last, which is what a user pastes into a bug report.
func (r DoctorReport) Text() string {
	ok, warn, fail := r.Counts()
	var builder strings.Builder
	for _, check := range r.Checks {
		fmt.Fprintf(&builder, "[%s] %s: %s\n", check.Status, check.Name, oneLine(check.Detail))
	}
	fmt.Fprintf(&builder, "doctor: %d ok, %d warn, %d fail\n", ok, warn, fail)
	return builder.String()
}

// RunDoctor inspects the machine this process is running on.
func RunDoctor() DoctorReport {
	return runDoctor(defaultDoctorEnv())
}

// doctorEnv is the seam the tests use. Every external fact -- the install
// tree, the pipe, the registry, Node -- arrives through here, so a unit test
// never has to depend on what happens to be installed on the test machine.
type doctorEnv struct {
	localAppData  string
	executable    string
	executableErr error
	goos          string
	goarch        string
	// dial reports whether a named pipe listener answered.
	dial func(path string, timeout time.Duration) error
	// node reports the outcome of one "node --version" probe.
	node func(timeout time.Duration) nodeResult
	// autostartStatus reads the HKCU Run entry.
	autostartStatus func() (autostart.Status, error)
}

func defaultDoctorEnv() doctorEnv {
	executable, executableErr := os.Executable()
	return doctorEnv{
		localAppData:    os.Getenv("LOCALAPPDATA"),
		executable:      executable,
		executableErr:   executableErr,
		goos:            runtime.GOOS,
		goarch:          runtime.GOARCH,
		dial:            probePipe,
		node:            probeNode,
		autostartStatus: func() (autostart.Status, error) { return autostart.New().Status() },
	}
}

// probePipe connects once and closes immediately. Nothing is written: the
// server treats an immediate EOF as an empty payload and never calls its
// update callback, so a doctor run cannot change what the overlay displays.
func probePipe(path string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	connection, err := winio.DialPipeContext(ctx, path)
	if err != nil {
		return err
	}
	return connection.Close()
}

// nodeResult is one node probe. Missing and TimedOut are kept apart because
// they mean opposite things to the user: "install Node" versus "Node is
// installed but did not answer".
type nodeResult struct {
	Path    string
	Version string
	// Note records something the user should know about a successful probe,
	// such as it only answering on the retry.
	Note     string
	Missing  bool
	TimedOut bool
	Err      error
}

func probeNode(timeout time.Duration) nodeResult {
	path, err := exec.LookPath("node")
	if err != nil {
		return nodeResult{Missing: true, Err: err}
	}
	return runNodeVersion(path, timeout)
}

func runNodeVersion(path string, timeout time.Duration) nodeResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var output bytes.Buffer
	command := exec.CommandContext(ctx, path, "--version")
	command.Stdout = &output
	// --doctor is reachable from a shortcut as well as a console; hiding the
	// child console keeps the probe from flashing a black window.
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	runErr := command.Run()
	version := strings.TrimSpace(output.String())
	if runErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nodeResult{Path: path, TimedOut: true, Err: runErr}
		}
		return nodeResult{Path: path, Err: runErr}
	}
	return nodeResult{Path: path, Version: version}
}

// runDoctor runs every check behind a recover wall. The checks are written
// not to panic; the wall exists so that a future one that does still produces
// a failure line instead of taking the other eight checks down with it.
func runDoctor(env doctorEnv) DoctorReport {
	pipePath := pipeserver.PipePathForUser()
	connectErr := dialOnce(env.dial, pipePath)
	checks := []DoctorCheck{
		runCheck("environment", func() DoctorCheck { return checkEnvironment(env) }),
		runCheck("install files", func() DoctorCheck { return checkInstallFiles(env) }),
		runCheck("version", func() DoctorCheck { return checkVersion(env) }),
		runCheck("runtime", func() DoctorCheck { return checkRuntime(env) }),
		runCheck("config", func() DoctorCheck { return checkConfig(env) }),
		runCheck("pipe", func() DoctorCheck { return checkPipe(pipePath, connectErr) }),
		runCheck("instance", func() DoctorCheck { return checkInstance(connectErr) }),
		runCheck("node", func() DoctorCheck { return checkNode(env) }),
		runCheck("autostart", func() DoctorCheck { return checkAutoStart(env) }),
	}
	return DoctorReport{Checks: checks}
}

func runCheck(name string, check func() DoctorCheck) (result DoctorCheck) {
	result = DoctorCheck{Name: name, Status: doctorFail, Detail: "check did not produce a result"}
	defer func() {
		if recovered := recover(); recovered != nil {
			result = DoctorCheck{Name: name, Status: doctorFail, Detail: fmt.Sprintf("check panicked: %v", recovered)}
		}
	}()
	result = check()
	result.Name = name
	if result.Status != doctorOK && result.Status != doctorWarn && result.Status != doctorFail {
		// A check that returned no verdict must not vanish into the
		// summary as a silent non-failure.
		result.Status = doctorFail
		if strings.TrimSpace(result.Detail) == "" {
			result.Detail = "check returned no verdict"
		}
	}
	return result
}

func dialOnce(dial func(string, time.Duration) error, path string) (err error) {
	if dial == nil {
		return errors.New("pipe probe is unavailable")
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("pipe probe panicked: %v", recovered)
		}
	}()
	return dial(path, doctorDialTimeout)
}

func checkEnvironment(env doctorEnv) DoctorCheck {
	detail := fmt.Sprintf("%s/%s (compile time)", env.goos, env.goarch)
	if env.goos != "windows" {
		return DoctorCheck{Status: doctorFail, Detail: detail + "; this build was not compiled for Windows"}
	}
	if env.localAppData == "" {
		return DoctorCheck{Status: doctorFail, Detail: detail + "; windows=true; LOCALAPPDATA is not set"}
	}
	info, err := os.Stat(env.localAppData)
	switch {
	case err != nil:
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf("%s; windows=true; LOCALAPPDATA=%s: %v", detail, env.localAppData, err)}
	case !info.IsDir():
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf("%s; windows=true; LOCALAPPDATA=%s is not a directory", detail, env.localAppData)}
	}
	return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("%s; windows=true; LOCALAPPDATA=%s (exists)", detail, env.localAppData)}
}

func checkInstallFiles(env doctorEnv) DoctorCheck {
	if env.localAppData == "" {
		return DoctorCheck{Status: doctorFail, Detail: "LOCALAPPDATA is not set, so the install directory cannot be located"}
	}
	directory := filepath.Join(env.localAppData, installDirectoryName)
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		// Name every file the installer would have dropped here. A missing
		// directory is the same failure as an empty one, and the files are
		// what the user actually has to restore.
		reason := fmt.Sprintf("install directory %s: %v", directory, err)
		if err == nil {
			reason = fmt.Sprintf("%s exists but is not a directory", directory)
		}
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf(
			"running from %s; %s; missing: %s",
			executableLabel(env), reason, strings.Join(doctorInstallFiles, ", "))}
	}
	parts := make([]string, 0, len(doctorInstallFiles))
	missing := make([]string, 0, len(doctorInstallFiles))
	for _, name := range doctorInstallFiles {
		fileInfo, statErr := os.Stat(filepath.Join(directory, name))
		switch {
		case statErr != nil:
			missing = append(missing, name)
			parts = append(parts, name+"=missing")
		case fileInfo.IsDir():
			missing = append(missing, name)
			parts = append(parts, name+"=directory")
		default:
			parts = append(parts, fmt.Sprintf("%s=%d bytes", name, fileInfo.Size()))
		}
	}
	detail := fmt.Sprintf("running from %s; %s; %s", executableLabel(env), directory, strings.Join(parts, ", "))
	if len(missing) > 0 {
		return DoctorCheck{Status: doctorFail, Detail: detail + "; missing: " + strings.Join(missing, ", ")}
	}
	return DoctorCheck{Status: doctorOK, Detail: detail}
}

// checkVersion is deliberately honest about what it cannot know. The version
// is a compile-time constant with no copy inside the installed binary, so
// comparing two builds would mean executing the other one. Rather than claim
// a match it cannot verify, the check reports the installed file's observable
// facts (size and LastWriteTime) and warns when the running process is not the
// installed binary.
func checkVersion(env doctorEnv) DoctorCheck {
	detail := fmt.Sprintf("this build is %s", Version)
	if env.executableErr != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("%s; os.Executable() failed: %v", detail, env.executableErr)}
	}
	if env.localAppData == "" {
		return DoctorCheck{Status: doctorWarn, Detail: detail + "; LOCALAPPDATA is not set, so the installed copy cannot be inspected"}
	}
	installed := filepath.Join(env.localAppData, installDirectoryName, doctorMainExe)
	info, err := os.Stat(installed)
	if err != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("%s; %s cannot be inspected: %v", detail, installed, err)}
	}
	if samePath(env.executable, installed) {
		return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("%s; running binary IS the installed one (%d bytes, modified %s)", detail, info.Size(), info.ModTime().Format(time.RFC3339))}
	}
	// Read-only comparison stops at the file's own metadata. Say which way the
	// sizes point so the warning is actionable rather than noise.
	relation := "sizes differ, so the install is probably an older build"
	if selfInfo, selfErr := os.Stat(env.executable); selfErr == nil && selfInfo.Size() == info.Size() {
		relation = "sizes match, so it is probably the same build"
	}
	return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf(
		"%s; cannot compare versions read-only: installed %s is %d bytes, modified %s (%s); this process runs from %s",
		detail, installed, info.Size(), info.ModTime().Format(time.RFC3339), relation, env.executable)}
}

func checkRuntime(env doctorEnv) DoctorCheck {
	if env.localAppData == "" {
		return DoctorCheck{Status: doctorWarn, Detail: "LOCALAPPDATA is not set, so runtime/ cannot be located"}
	}
	directory := filepath.Join(env.localAppData, installDirectoryName, "runtime")
	info, err := os.Stat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return DoctorCheck{Status: doctorWarn, Detail: directory + " does not exist yet; the first reporter push creates it"}
	}
	if err != nil {
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf("%s: %v", directory, err)}
	}
	if !info.IsDir() {
		return DoctorCheck{Status: doctorFail, Detail: directory + " exists but is not a directory; snapshots cannot be written"}
	}
	// Actually create and delete a file. Checking an ACL or the read-only
	// attribute would answer a different question from the one that matters:
	// can the overlay write a snapshot here right now.
	temporary, writeErr := os.CreateTemp(directory, config.TempPrefix+"doctor-*.tmp")
	if writeErr != nil {
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf("%s is not writable: %v", directory, writeErr)}
	}
	name := temporary.Name()
	closeErr := temporary.Close()
	removeErr := os.Remove(name)
	writable := "writable (created and removed a temp file)"
	if closeErr != nil || removeErr != nil {
		writable = fmt.Sprintf("writable, but the probe file could not be cleaned up (%v)", errors.Join(closeErr, removeErr))
	}
	entries, readErr := os.ReadDir(directory)
	if readErr != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("%s is %s; could not list snapshots: %v", directory, writable, readErr)}
	}
	snapshots := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			snapshots++
		}
	}
	return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("%s is %s; %d *.json snapshot(s) present", directory, writable, snapshots)}
}

// checkConfig reads the two files with the same functions the overlay uses,
// and additionally parses them raw, because config.Load is fail-safe by
// design and cannot tell "absent" from "corrupt". A corrupt file is a warning
// rather than a failure: the overlay explicitly starts on defaults, so nothing
// here is broken - the user's edits are just being ignored without saying so.
func checkConfig(env doctorEnv) DoctorCheck {
	if env.localAppData == "" {
		return DoctorCheck{Status: doctorWarn, Detail: "LOCALAPPDATA is not set, so config.json / state.json cannot be located"}
	}
	directory := filepath.Join(env.localAppData, installDirectoryName)
	configPath := filepath.Join(directory, config.FileName)
	statePath := filepath.Join(directory, config.StateFileName)

	notes := make([]string, 0, 4)
	corrupt := false

	data, err := os.ReadFile(configPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		notes = append(notes, "config.json absent (defaults in use)")
	case err != nil:
		corrupt = true
		notes = append(notes, "config.json unreadable: "+err.Error())
	default:
		var raw config.Config
		if parseErr := json.Unmarshal(data, &raw); parseErr != nil {
			corrupt = true
			notes = append(notes, "config.json is not valid JSON ("+parseErr.Error()+"); the overlay silently falls back to defaults")
		}
	}
	settings := config.Load(configPath).Normalize()
	notes = append(notes, fmt.Sprintf("autoApprove=%t language=%s opacity=%d sound=%t notifications=%t",
		settings.AutoApprove, settings.Language, settings.Opacity, settings.Sound, settings.Notifications))

	stateData, err := os.ReadFile(statePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		notes = append(notes, "state.json absent (window placement not remembered yet)")
	case err != nil:
		corrupt = true
		notes = append(notes, "state.json unreadable: "+err.Error())
	default:
		var raw config.State
		if parseErr := json.Unmarshal(stateData, &raw); parseErr != nil {
			corrupt = true
			notes = append(notes, "state.json is not valid JSON ("+parseErr.Error()+"); the overlay falls back to the default placement")
		}
	}
	placement := config.LoadState(statePath).Normalize()
	notes = append(notes, fmt.Sprintf("zoom=%.2f placed=%t x=%d y=%d expanded=%t",
		placement.Zoom, placement.Placed, placement.X, placement.Y, placement.Expanded))

	detail := configPath + "; " + strings.Join(notes, "; ")
	if corrupt {
		return DoctorCheck{Status: doctorWarn, Detail: detail}
	}
	return DoctorCheck{Status: doctorOK, Detail: detail}
}

func checkPipe(path string, connectErr error) DoctorCheck {
	if connectErr == nil {
		return DoctorCheck{Status: doctorOK, Detail: path + "; accepted a connection, the overlay is listening"}
	}
	return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("%s; no listener: %v - the overlay is not running", path, connectErr)}
}

// checkInstance does not probe the single-instance mutex on purpose.
// CreateMutex would create the object if it did not exist, and touching the
// live overlay's handle is exactly the kind of side effect a diagnostic must
// not have. A listening pipe is the observable consequence of the same
// instance, so that is the evidence used here - and it is named in the output
// so the claim can be judged.
func checkInstance(connectErr error) DoctorCheck {
	if connectErr == nil {
		return DoctorCheck{Status: doctorOK, Detail: "a live instance is listening on the pipe (mutex not probed, see checkPipe)"}
	}
	return DoctorCheck{Status: doctorWarn, Detail: "no live instance: nothing answers on the pipe; the mutex was not probed because opening it could disturb a real instance"}
}

func checkNode(env doctorEnv) DoctorCheck {
	if env.node == nil {
		return DoctorCheck{Status: doctorWarn, Detail: "node probe is unavailable"}
	}
	result := env.node(doctorNodeTimeout)
	if result.TimedOut {
		// The first node start of a session routinely takes longer than the
		// warm one (cold page cache, AV scanning the executable). Retrying once
		// at the same budget separates "slow first start" from "node is going
		// to stall every reporter hook"; warning about the former would be a
		// false alarm on a healthy machine, which is worse than no check.
		retry := env.node(doctorNodeTimeout)
		if !retry.TimedOut && retry.Err == nil && !retry.Missing {
			retry.Note = "the first start missed the " + doctorNodeTimeout.String() + " deadline, the retry answered"
			result = retry
		}
	}
	switch {
	case result.Missing:
		return DoctorCheck{Status: doctorWarn, Detail: "node not found in PATH; reporters/*.js need Node.js (the native pet-statusline.exe does not)"}
	case result.TimedOut:
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("node found at %s but --version did not answer within %s; a cold start under an AV scanner looks like this", result.Path, doctorNodeTimeout)}
	case result.Err != nil:
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("node found at %s but --version failed: %v", result.Path, result.Err)}
	}
	detail := fmt.Sprintf("%s (%s)", result.Version, result.Path)
	if result.Note != "" {
		detail += "; " + result.Note
	}
	return DoctorCheck{Status: doctorOK, Detail: detail}
}

// checkAutoStart is one of the highest-value checks here: scripts/install.ps1 stops
// the overlay before every upgrade, and if the Run value is not repaired the
// entry points at a path that no longer exists and fails at logon without
// saying anything.
func checkAutoStart(env doctorEnv) DoctorCheck {
	if env.autostartStatus == nil {
		return DoctorCheck{Status: doctorWarn, Detail: "autostart probe is unavailable"}
	}
	status, err := env.autostartStatus()
	if err != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("cannot read HKCU Run: %v", err)}
	}
	if !status.Enabled {
		return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("off (no %s value in HKCU Run; a reporter, the tray or --autostart=on starts it)", autostart.ValueName)}
	}
	if strings.TrimSpace(status.Target) == "" {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("on but the %s value holds no path", autostart.ValueName)}
	}
	info, statErr := os.Stat(status.Target)
	if statErr != nil {
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf("on -> %s, but that binary no longer exists (%v): autostart would fail silently at logon", status.Target, statErr)}
	}
	if info.IsDir() {
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf("on -> %s, but that is a directory, not an executable", status.Target)}
	}
	return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("on -> %s (%d bytes)", status.Target, info.Size())}
}

func executableLabel(env doctorEnv) string {
	if env.executableErr != nil {
		return fmt.Sprintf("<unknown: %v>", env.executableErr)
	}
	if env.executable == "" {
		return "<unknown>"
	}
	return env.executable
}

func samePath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

// oneLine keeps the report greppable: one check per line, no matter what text
// an error carried.
func oneLine(detail string) string {
	if strings.TrimSpace(detail) == "" {
		return "(no detail)"
	}
	return strings.Join(strings.Fields(detail), " ")
}
