package app

// --doctor answers "what is not wired up on this machine?" in one pass.

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
	"time"

	"ai-terminal-pet/internal/autostart"
	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/connect"
	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/pipeserver"
	"ai-terminal-pet/internal/ui"
)

const (
	doctorOK   = "ok"
	doctorWarn = "warn"
	doctorFail = "fail"

	installDirectoryName = "AI-Terminal-Pet"
	doctorMainExe        = "AI-Terminal-Pet.exe"
	doctorStatuslineExe  = "pet-statusline.exe"
	doctorProtocolFile   = "protocol.js"

	doctorDialTimeout = 250 * time.Millisecond
	doctorNodeTimeout = 300 * time.Millisecond
)

var doctorInstallFiles = []string{doctorMainExe, doctorStatuslineExe, doctorProtocolFile}

// DoctorCheck is one line of the report: what was inspected, the verdict and
// the measured value.
type DoctorCheck struct {
	Name   string
	Status string
	Detail string
}

// DoctorReport is the whole run.
type DoctorReport struct {
	Checks []DoctorCheck
}

// ExitCode returns 1 when any check failed and 0 otherwise.
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
			fail++
		}
	}
	return ok, warn, fail
}

// Text renders the human-readable report.
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

type doctorEnv struct {
	localAppData    string
	xdgDataHome     string
	home            string
	executable      string
	executableErr   error
	goos            string
	goarch          string
	dial            func(path string, timeout time.Duration) error
	node            func(timeout time.Duration) nodeResult
	autostartStatus func() (autostart.Status, error)
	// soundSurface and notifySurface name what the `sound` and `notifications`
	// settings actually reach on this machine. They are injected for the same
	// reason the pipe and node probes are: a test must be able to describe a
	// machine that has neither, on a machine that has both.
	soundSurface  func() string
	notifySurface func() string
}

func defaultDoctorEnv() doctorEnv {
	executable, executableErr := os.Executable()
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil || homeDir == "" {
		homeDir = os.Getenv("USERPROFILE")
	}
	return doctorEnv{
		localAppData:    os.Getenv("LOCALAPPDATA"),
		xdgDataHome:     os.Getenv("XDG_DATA_HOME"),
		home:            homeDir,
		executable:      executable,
		executableErr:   executableErr,
		goos:            runtime.GOOS,
		goarch:          runtime.GOARCH,
		dial:            probePipe,
		node:            probeNode,
		autostartStatus: func() (autostart.Status, error) { return autostart.New().Status() },
		soundSurface:    ui.SoundSurface,
		notifySurface:   ui.NotificationSurface,
	}
}

func (e doctorEnv) dataDirectory() string {
	if e.goos == "windows" {
		if e.localAppData == "" {
			return ""
		}
		return filepath.Join(e.localAppData, installDirectoryName)
	}
	return home.DirectoryFor(e.goos, e.localAppData, e.xdgDataHome, e.home)
}

func (e doctorEnv) mainBinaryName() string {
	if e.goos == "windows" {
		return doctorMainExe
	}
	return "AI-Terminal-Pet"
}

func (e doctorEnv) installFileList() []string {
	if e.goos == "windows" {
		return doctorInstallFiles
	}
	return []string{"AI-Terminal-Pet", "pet-statusline", doctorProtocolFile}
}

func (e doctorEnv) missingDataDirDetail(purpose string) string {
	if e.goos == "windows" {
		if purpose == "" {
			return "LOCALAPPDATA is not set, so the install directory cannot be located"
		}
		return "LOCALAPPDATA is not set, so " + purpose + " cannot be located"
	}
	if purpose == "" {
		return "data directory could not be resolved (HOME/XDG_DATA_HOME unset)"
	}
	return "data directory could not be resolved, so " + purpose + " cannot be located"
}

type nodeResult struct {
	Path     string
	Version  string
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
	hideChildConsole(command)
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
		runCheck("reporters", func() DoctorCheck { return checkReporters(env) }),
		runCheck("autostart", func() DoctorCheck { return checkAutoStart(env) }),
		runCheck("alerts", func() DoctorCheck { return checkAlerts(env) }),
		runCheck("audio", func() DoctorCheck { return checkAudio(env) }),
		runCheck("notifications", func() DoctorCheck { return checkNotifications(env) }),
	}
	return DoctorReport{Checks: checks}
}

func checkReporters(env doctorEnv) DoctorCheck {
	roots := connect.Roots{
		Home:          env.home,
		LocalAppData:  env.localAppData,
		XDGDataHome:   env.xdgDataHome,
		DataDir:       env.dataDirectory(),
		Executable:    env.executable,
		ExecutableErr: env.executableErr,
	}
	findings := connect.Inspect(roots)
	status := doctorOK
	switch connect.Worst(findings) {
	case connect.StatusFail:
		status = doctorFail
	case connect.StatusWarn:
		status = doctorWarn
	}
	return DoctorCheck{Status: status, Detail: connect.Summary(findings)}
}

func checkAlerts(env doctorEnv) DoctorCheck {
	directory := env.dataDirectory()
	if directory == "" {
		return DoctorCheck{Status: doctorWarn, Detail: env.missingDataDirDetail("runtime/alerts.json")}
	}
	path := filepath.Join(directory, "alerts.json")
	record := readAlertReport(directory)
	if record == nil {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return DoctorCheck{Status: doctorOK, Detail: "no alert receipt yet (" + path + "); the overlay writes one on its first scan"}
		}
		return DoctorCheck{Status: doctorWarn, Detail: path + " is present but cannot be parsed; the next overlay scan overwrites it"}
	}
	detail := fmt.Sprintf("last scan %s, presence=%s, balloon=%s", record.At.Format(time.RFC3339), record.Presence, emptyAsNone(record.Balloon))
	if record.Skipped != "" {
		detail += ", skipped=" + record.Skipped
	}
	if len(record.Sessions) > 0 {
		decisions := make([]string, 0, len(record.Sessions))
		for _, item := range record.Sessions {
			decisions = append(decisions, item.Kind+"/"+item.Decision)
		}
		detail += ", decisions=" + strings.Join(decisions, "; ")
	}
	return DoctorCheck{Status: doctorOK, Detail: detail}
}

// checkAudio answers "will `sound` be audible here", and checkNotifications the
// same question for the alert surface. Both are warnings when the answer is
// "nothing": the install is healthy and the panel works, but a setting the user
// can see doing nothing is exactly the shape of gap --doctor exists to name.
func checkAudio(env doctorEnv) DoctorCheck {
	return surfaceCheck("sound", env.soundSurface)
}

func checkNotifications(env doctorEnv) DoctorCheck {
	return surfaceCheck("notification", env.notifySurface)
}

// surfaceCheck is the one shape both questions have: a name, or a leading
// "none" that already carries the reason and the way out.
func surfaceCheck(name string, probe func() string) DoctorCheck {
	if probe == nil {
		return DoctorCheck{Status: doctorWarn, Detail: name + " surface probe is unavailable"}
	}
	surface := strings.TrimSpace(probe())
	switch {
	case surface == "":
		return DoctorCheck{Status: doctorWarn, Detail: "the platform did not name a " + name + " surface"}
	case strings.HasPrefix(surface, "none"):
		return DoctorCheck{Status: doctorWarn, Detail: surface}
	}
	return DoctorCheck{Status: doctorOK, Detail: surface}
}

func emptyAsNone(value string) string {
	if value == "" {
		return "none"
	}
	return value
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
	switch env.goos {
	case "windows":
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
	case "linux", "darwin":
		dir := env.dataDirectory()
		if dir == "" {
			return DoctorCheck{Status: doctorFail, Detail: detail + "; " + env.missingDataDirDetail("")}
		}
		extra := ""
		if env.goos == "linux" {
			display := strings.TrimSpace(os.Getenv("DISPLAY"))
			if display == "" {
				extra = "; DISPLAY unset (headless)"
			} else {
				extra = "; DISPLAY=" + display
			}
		}
		return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("%s; dataDir=%s%s", detail, dir, extra)}
	default:
		return DoctorCheck{Status: doctorFail, Detail: detail + "; unsupported operating system"}
	}
}

func checkInstallFiles(env doctorEnv) DoctorCheck {
	directory := env.dataDirectory()
	if directory == "" {
		return DoctorCheck{Status: doctorFail, Detail: env.missingDataDirDetail("")}
	}
	files := env.installFileList()
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		reason := fmt.Sprintf("install directory %s: %v", directory, err)
		if err == nil {
			reason = fmt.Sprintf("%s exists but is not a directory", directory)
		}
		return DoctorCheck{Status: doctorFail, Detail: fmt.Sprintf(
			"running from %s; %s; missing: %s",
			executableLabel(env), reason, strings.Join(files, ", "))}
	}
	parts := make([]string, 0, len(files))
	missing := make([]string, 0, len(files))
	for _, name := range files {
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

func checkVersion(env doctorEnv) DoctorCheck {
	detail := fmt.Sprintf("this build is %s", Version)
	if env.executableErr != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("%s; os.Executable() failed: %v", detail, env.executableErr)}
	}
	directory := env.dataDirectory()
	if directory == "" {
		return DoctorCheck{Status: doctorWarn, Detail: detail + "; " + env.missingDataDirDetail("the installed copy")}
	}
	installed := filepath.Join(directory, env.mainBinaryName())
	info, err := os.Stat(installed)
	if err != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("%s; %s cannot be inspected: %v", detail, installed, err)}
	}
	if samePath(env.executable, installed, env.goos) {
		return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("%s; running binary IS the installed one (%d bytes, modified %s)", detail, info.Size(), info.ModTime().Format(time.RFC3339))}
	}
	relation := "sizes differ, so the install is probably an older build"
	if selfInfo, selfErr := os.Stat(env.executable); selfErr == nil && selfInfo.Size() == info.Size() {
		relation = "sizes match, so it is probably the same build"
	}
	return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf(
		"%s; cannot compare versions read-only: installed %s is %d bytes, modified %s (%s); this process runs from %s",
		detail, installed, info.Size(), info.ModTime().Format(time.RFC3339), relation, env.executable)}
}

func checkRuntime(env doctorEnv) DoctorCheck {
	base := env.dataDirectory()
	if base == "" {
		return DoctorCheck{Status: doctorWarn, Detail: env.missingDataDirDetail("runtime/")}
	}
	directory := filepath.Join(base, "runtime")
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

func checkConfig(env doctorEnv) DoctorCheck {
	directory := env.dataDirectory()
	if directory == "" {
		return DoctorCheck{Status: doctorWarn, Detail: env.missingDataDirDetail("config.json / state.json")}
	}
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
		retry := env.node(doctorNodeTimeout)
		if !retry.TimedOut && retry.Err == nil && !retry.Missing {
			retry.Note = "the first start missed the " + doctorNodeTimeout.String() + " deadline, the retry answered"
			result = retry
		}
	}
	switch {
	case result.Missing:
		return DoctorCheck{Status: doctorWarn, Detail: "node not found in PATH; reporters/*.js need Node.js (the native pet-statusline does not)"}
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

func checkAutoStart(env doctorEnv) DoctorCheck {
	if env.autostartStatus == nil {
		return DoctorCheck{Status: doctorWarn, Detail: "autostart probe is unavailable"}
	}
	status, err := env.autostartStatus()
	if err != nil {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("cannot read autostart: %v", err)}
	}
	if !status.Enabled {
		if env.goos == "windows" {
			return DoctorCheck{Status: doctorOK, Detail: fmt.Sprintf("off (no %s value in HKCU Run; a reporter, the tray or --autostart=on starts it)", autostart.ValueName)}
		}
		return DoctorCheck{Status: doctorOK, Detail: "off (run --autostart=on to enable logon autostart)"}
	}
	if strings.TrimSpace(status.Target) == "" {
		return DoctorCheck{Status: doctorWarn, Detail: fmt.Sprintf("on but %s holds no path", autostart.ValueName)}
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

func samePath(left, right, goos string) bool {
	if left == "" || right == "" {
		return false
	}
	if goos == "windows" || goos == "darwin" {
		return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
	}
	return filepath.Clean(left) == filepath.Clean(right)
}

func oneLine(detail string) string {
	if strings.TrimSpace(detail) == "" {
		return "(no detail)"
	}
	return strings.Join(strings.Fields(detail), " ")
}
