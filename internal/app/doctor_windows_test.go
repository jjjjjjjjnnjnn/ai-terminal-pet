package app

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"ai-terminal-pet/internal/autostart"
	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/connect"
)

// doctorFixture builds a doctorEnv over a throwaway %LOCALAPPDATA%. Nothing
// here reads the machine's real install: every external answer - pipe, node,
// registry - is injected, so the test passes on a box with no pet installed.
type doctorFixture struct {
	root string
	env  doctorEnv
}

func newDoctorFixture(t *testing.T) *doctorFixture {
	t.Helper()
	root := t.TempDir()
	fixture := &doctorFixture{root: root}
	fixture.env = doctorEnv{
		localAppData: root,
		home:         filepath.Join(root, "profile"),
		goos:         "windows",
		goarch:       "amd64",
		executable:   filepath.Join(root, installDirectoryName, doctorMainExe),
		dial:         func(string, time.Duration) error { return nil },
		node: func(time.Duration) nodeResult {
			return nodeResult{Path: "C:/node/node.exe", Version: "v20.11.0"}
		},
		autostartStatus: func() (autostart.Status, error) { return autostart.Status{}, nil },
		soundSurface:    func() string { return "winmm PlaySoundW (built in)" },
		notifySurface:   func() string { return "tray balloon (Shell_NotifyIconW, built in)" },
	}
	return fixture
}

// install lays down the deployment set scripts/install.ps1 produces.
func (f *doctorFixture) install(t *testing.T) string {
	t.Helper()
	directory := filepath.Join(f.root, installDirectoryName)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range doctorInstallFiles {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("stub "+name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func (f *doctorFixture) runtime(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(f.root, installDirectoryName, "runtime"), 0o700); err != nil {
		t.Fatal(err)
	}
}

// connectAllThree lays down host configurations that are fully wired, so a test
// asserting "every check is ok" can hold the reporters line to the same
// standard as the rest without weakening it to a warn.
func (f *doctorFixture) connectAllThree(t *testing.T) {
	t.Helper()
	install := filepath.Join(f.root, installDirectoryName)
	// The hook binary is part of install.ps1's deployment set but not part of
	// doctorInstallFiles (which lists the three the overlay itself needs), so a
	// fully wired machine has to write it here.
	writeFixtureFile(t, filepath.Join(install, connect.HookBinaryName), "stub hook")

	repository := filepath.Join(f.root, "repo")
	reporter := filepath.Join(repository, "reporters", "opencode.js")
	writeFixtureFile(t, reporter, "module.exports = {}\n")
	writeFixtureFile(t, filepath.Join(f.root, "profile", ".config", "opencode", "opencode.json"),
		`{"plugins": ["`+fixtureFileURL(t, reporter)+`"]}`)

	hook := filepath.Join(install, connect.HookBinaryName)
	writeFixtureFile(t, filepath.Join(f.root, "profile", ".claude", "settings.json"), fixtureClaudeSettings(hook))

	writeFixtureFile(t, filepath.Join(f.root, "profile", ".gemini", "antigravity-cli", "settings.json"),
		`{"statusLine": {"command": "`+filepath.ToSlash(filepath.Join(install, connect.StatuslineBinaryName))+`"}}`)
	writeFixtureFile(t, filepath.Join(f.root, "profile", ".gemini", "config", "plugins", "pet", "hooks.json"),
		`{"pet": {"Termination": [{"command": "pet-hook.exe --mode=antigravity"}]}}`)
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func fixtureFileURL(t *testing.T, path string) string {
	t.Helper()
	value, err := connect.PathToFileURL(path)
	if err != nil {
		t.Fatalf("PathToFileURL(%q): %v", path, err)
	}
	return value
}

// fixtureClaudeSettings wires all eight events to the native hook, which is what
// a fully connected machine looks like.
func fixtureClaudeSettings(hook string) string {
	events := []string{
		"SessionStart", "UserPromptSubmit", "PreToolUse", "PostToolUse",
		"PermissionRequest", "Notification", "Stop", "SessionEnd",
	}
	var builder strings.Builder
	builder.WriteString(`{"hooks": {`)
	for index, event := range events {
		if index > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(`"` + event + `": [{"matcher": "*", "hooks": [{"type": "command", "command": ` +
			jsonCommand(`"`+hook+`" --mode=claude`) + `}]}]`)
	}
	builder.WriteString("}}")
	return builder.String()
}

// jsonCommand escapes one Windows path into a JSON string. The backslashes have
// to be doubled or the document is not JSON at all; a fixture that silently
// produces invalid JSON would let the check pass for the wrong reason.
func jsonCommand(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func checkNamed(t *testing.T, report DoctorReport, name string) DoctorCheck {
	for _, check := range report.Checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("report has no %q check; got %v", name, namesOf(report))
	return DoctorCheck{}
}

func namesOf(report DoctorReport) []string {
	names := make([]string, 0, len(report.Checks))
	for _, check := range report.Checks {
		names = append(names, check.Name)
	}
	return names
}

// A node that only answers on the second start is healthy: the cold start of
// the session is the slow one, and reporters must not be declared broken for
// it. The probe injects that shape directly, without spawning node.
func TestDoctorSlowFirstNodeStartIsOK(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	calls := 0
	fixture.env.node = func(time.Duration) nodeResult {
		calls++
		if calls == 1 {
			return nodeResult{Path: "C:/node/node.exe", TimedOut: true}
		}
		return nodeResult{Path: "C:/node/node.exe", Version: "v20.11.0", Note: "answered on the retry"}
	}
	check := checkNode(fixture.env)
	if check.Status != doctorOK {
		t.Fatalf("slow first start should be ok, got %q: %s", check.Status, check.Detail)
	}
	if calls != 2 {
		t.Fatalf("expected the probe to be retried once, got %d call(s)", calls)
	}
	if !strings.Contains(check.Detail, "v20.11.0") {
		t.Fatalf("detail does not carry the node version: %s", check.Detail)
	}
}

// A node that never answers stays a warning, not a failure: a user who only
// runs the native statusline has a working install without Node.
func TestDoctorNodeThatNeverAnswersWarns(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	fixture.env.node = func(time.Duration) nodeResult {
		return nodeResult{Path: "C:/node/node.exe", TimedOut: true}
	}
	check := checkNode(fixture.env)
	if check.Status != doctorWarn {
		t.Fatalf("a stalling node should warn, got %q: %s", check.Status, check.Detail)
	}
}

// A healthy install must be all-ok and exit 0, so a green run is not
// accidentally red because one probe is not injected.
func TestDoctorAllGreenExitsZero(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	fixture.connectAllThree(t)
	report := runDoctor(fixture.env)

	ok, warn, fail := report.Counts()
	if fail != 0 || warn != 0 {
		t.Fatalf("expected an all-ok report, got %d ok, %d warn, %d fail\n%s", ok, warn, fail, report.Text())
	}
	if got := report.ExitCode(); got != 0 {
		t.Fatalf("exit code = %d, want 0", got)
	}
	if want := len(report.Checks); ok != want {
		t.Fatalf("counted %d ok but the report has %d checks", ok, want)
	}
	for _, check := range report.Checks {
		if check.Status != doctorOK {
			t.Fatalf("%s is %q: %s", check.Name, check.Status, check.Detail)
		}
	}
}

// The empty-directory case: a machine where nothing is installed must say
// which files are missing, not just "fail".
func TestDoctorMissingInstallFailsWithNamedFiles(t *testing.T) {
	fixture := newDoctorFixture(t)
	report := runDoctor(fixture.env)
	if got := report.ExitCode(); got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
	install := checkNamed(t, report, "install files")
	if install.Status != doctorFail {
		t.Fatalf("install files = %q, want fail", install.Status)
	}
	for _, name := range doctorInstallFiles {
		if !strings.Contains(install.Detail, name) {
			t.Fatalf("detail does not name %s: %s", name, install.Detail)
		}
	}
	// One broken area must not take the pipe, node or autostart checks with it.
	if pipe := checkNamed(t, report, "pipe"); pipe.Status != doctorOK {
		t.Fatalf("pipe check was affected by the install failure: %q", pipe.Status)
	}
}

// LOCALAPPDATA unset must produce failures, not a panic.
func TestDoctorWithoutLocalAppDataDoesNotPanic(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.env.localAppData = ""

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("doctor panicked with an unset LOCALAPPDATA: %v", recovered)
		}
	}()
	report := runDoctor(fixture.env)
	if got := report.ExitCode(); got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
	environment := checkNamed(t, report, "environment")
	if environment.Status != doctorFail || !strings.Contains(environment.Detail, "LOCALAPPDATA") {
		t.Fatalf("environment = %+v, want a fail naming LOCALAPPDATA", environment)
	}
	install := checkNamed(t, report, "install files")
	if install.Status != doctorFail {
		t.Fatalf("install files = %q, want fail", install.Status)
	}
}

// A runtime path that exists but is not a directory is the deterministic
// stand-in for "not writable": the temp-file probe cannot even be attempted.
func TestDoctorRuntimeNotWritableFails(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	runtimePath := filepath.Join(fixture.root, installDirectoryName, "runtime")
	if err := os.WriteFile(runtimePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := runDoctor(fixture.env)
	check := checkNamed(t, report, "runtime")
	if check.Status != doctorFail {
		t.Fatalf("runtime = %q, want fail: %s", check.Status, check.Detail)
	}
	if got := report.ExitCode(); got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
}

// A corrupt config.json must be reported, not silently smoothed over by
// config.Load's fail-safe default. It is a warning because the overlay keeps
// running on defaults.
func TestDoctorCorruptConfigWarns(t *testing.T) {
	fixture := newDoctorFixture(t)
	directory := fixture.install(t)
	fixture.runtime(t)
	if err := os.WriteFile(filepath.Join(directory, config.FileName), []byte("{ not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := runDoctor(fixture.env)
	check := checkNamed(t, report, "config")
	if check.Status != doctorWarn {
		t.Fatalf("config = %q, want warn: %s", check.Status, check.Detail)
	}
	if !strings.Contains(check.Detail, "config.json") {
		t.Fatalf("detail does not name config.json: %s", check.Detail)
	}
	// Warnings alone must keep the exit code at 0.
	if got := report.ExitCode(); got != 0 {
		t.Fatalf("exit code = %d, want 0 for warnings-only", got)
	}
}

// A valid config must surface the values a user would be asked about in a
// bug report, rather than a bare "ok".
func TestDoctorReportsConfigValues(t *testing.T) {
	fixture := newDoctorFixture(t)
	directory := fixture.install(t)
	fixture.runtime(t)
	payload := "{\"autoApprove\":true,\"language\":\"zh-CN\",\"opacity\":120,\"zoom\":1}"
	if err := os.WriteFile(filepath.Join(directory, config.FileName), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	report := runDoctor(fixture.env)
	detail := checkNamed(t, report, "config").Detail
	for _, want := range []string{"autoApprove=true", "language=zh-CN", "opacity=120", "zoom="} {
		if !strings.Contains(detail, want) {
			t.Fatalf("config detail is missing %q: %s", want, detail)
		}
	}
}

// A Run value pointing at a deleted binary is the silent-at-logon failure
// scripts/install.ps1's stop-before-copy can produce. It must be a failure.
func TestDoctorDanglingAutostartTargetFails(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	fixture.env.autostartStatus = func() (autostart.Status, error) {
		return autostart.Status{Enabled: true, Target: filepath.Join(fixture.root, "gone", "AI-Terminal-Pet.exe")}, nil
	}
	report := runDoctor(fixture.env)
	check := checkNamed(t, report, "autostart")
	if check.Status != doctorFail {
		t.Fatalf("autostart = %q, want fail: %s", check.Status, check.Detail)
	}
	if !strings.Contains(check.Detail, "no longer exists") {
		t.Fatalf("detail must say the binary is gone: %s", check.Detail)
	}
	if got := report.ExitCode(); got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
}

// An enabled autostart entry that does point at a real file is ok.
func TestDoctorAutostartTargetPresentIsOK(t *testing.T) {
	fixture := newDoctorFixture(t)
	directory := fixture.install(t)
	fixture.runtime(t)
	fixture.env.autostartStatus = func() (autostart.Status, error) {
		return autostart.Status{Enabled: true, Target: filepath.Join(directory, doctorMainExe)}, nil
	}
	report := runDoctor(fixture.env)
	check := checkNamed(t, report, "autostart")
	if check.Status != doctorOK {
		t.Fatalf("autostart = %q, want ok: %s", check.Status, check.Detail)
	}
	if !strings.Contains(check.Detail, "bytes") {
		t.Fatalf("detail must carry the measured size: %s", check.Detail)
	}
}

// The stopped-overlay / no-node machine: warnings only, exit 0.
func TestDoctorStoppedOverlayAndMissingNodeAreWarnings(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	fixture.env.dial = func(string, time.Duration) error {
		return errors.New("The system cannot find the file specified.")
	}
	fixture.env.node = func(time.Duration) nodeResult {
		return nodeResult{Missing: true, Err: errors.New("not found")}
	}
	report := runDoctor(fixture.env)
	if got := report.ExitCode(); got != 0 {
		t.Fatalf("exit code = %d, want 0", got)
	}
	if got := checkNamed(t, report, "pipe").Status; got != doctorWarn {
		t.Fatalf("pipe = %q, want warn", got)
	}
	if got := checkNamed(t, report, "instance").Status; got != doctorWarn {
		t.Fatalf("instance = %q, want warn", got)
	}
	if got := checkNamed(t, report, "node").Status; got != doctorWarn {
		t.Fatalf("node = %q, want warn", got)
	}
}

// The line format is the contract users paste into issues.
func TestDoctorTextFormat(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	report := runDoctor(fixture.env)
	text := report.Text()

	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) != len(report.Checks)+1 {
		t.Fatalf("expected %d lines, got %d:\n%s", len(report.Checks)+1, len(lines), text)
	}
	linePattern := regexp.MustCompile("^\\[(ok|warn|fail)\\] [a-z][a-z ]*: .+$")
	for _, line := range lines[:len(lines)-1] {
		if !linePattern.MatchString(line) {
			t.Fatalf("line does not match the [ok]/[warn]/[fail] contract: %q", line)
		}
	}
	summary := regexp.MustCompile("^doctor: [0-9]+ ok, [0-9]+ warn, [0-9]+ fail$")
	if !summary.MatchString(lines[len(lines)-1]) {
		t.Fatalf("summary line does not match the contract: %q", lines[len(lines)-1])
	}
}

// One check failing to produce a value, even by panicking, must stay
// quarantined to its own line.
func TestDoctorCheckPanicIsContained(t *testing.T) {
	for _, test := range []struct {
		name  string
		check func() DoctorCheck
	}{
		{"panic", func() DoctorCheck { panic("boom") }},
		{"zero value", func() DoctorCheck { return DoctorCheck{} }},
	} {
		value := runCheck("quarantined", test.check)
		if value.Name != "quarantined" || value.Status != doctorFail || value.Detail == "" {
			t.Fatalf("%s: unexpected check value %+v", test.name, value)
		}
	}
}

// A dead pipe probe (nil or panicking) must be a warning, never a crash.
func TestDoctorPipeProbeFailureIsContained(t *testing.T) {
	if err := dialOnce(nil, "route pipe"); err == nil {
		t.Fatal("a nil probe must report an error")
	}
	panicErr := dialOnce(func(string, time.Duration) error { panic("boom") }, "route pipe")
	if panicErr == nil || !strings.Contains(panicErr.Error(), "panicked") {
		t.Fatalf("panicking probe = %v", panicErr)
	}
}

// --doctor must answer "will the sound be heard / the notification be seen" with
// the measured surface on a machine that has one.
func TestDoctorNamesTheSoundAndNotificationSurfaces(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	fixture.env.soundSurface = func() string { return "external player /usr/bin/paplay" }
	fixture.env.notifySurface = func() string { return "notify-send /usr/bin/notify-send" }
	report := runDoctor(fixture.env)
	for _, name := range []string{"audio", "notifications"} {
		check := checkNamed(t, report, name)
		if check.Status != doctorOK {
			t.Fatalf("%s = %q, want ok: %s", name, check.Status, check.Detail)
		}
	}
	if detail := checkNamed(t, report, "audio").Detail; !strings.Contains(detail, "paplay") {
		t.Fatalf("audio does not name the player: %s", detail)
	}
}

// A machine with neither is a warning, not a failure: the panel still works and
// the setting is still stored, so the exit code must stay 0.
func TestDoctorWarnsWhenNothingCanPlayOrShow(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.install(t)
	fixture.runtime(t)
	fixture.env.soundSurface = func() string { return "none (no pw-play, paplay or aplay in PATH)" }
	fixture.env.notifySurface = func() string { return "none (notify-send not found)" }
	report := runDoctor(fixture.env)
	for _, name := range []string{"audio", "notifications"} {
		if got := checkNamed(t, report, name).Status; got != doctorWarn {
			t.Fatalf("%s = %q, want warn", name, got)
		}
	}
	if got := report.ExitCode(); got != 0 {
		t.Fatalf("exit code = %d, warnings must not fail the run", got)
	}
}

// A probe that answers nothing at all is a warning too, never a panic and never
// a silent ok.
func TestDoctorSurfaceProbeWithoutAnAnswerWarns(t *testing.T) {
	fixture := newDoctorFixture(t)
	fixture.env.soundSurface = nil
	fixture.env.notifySurface = func() string { return "  " }
	if got := checkAudio(fixture.env).Status; got != doctorWarn {
		t.Fatalf("a nil probe = %q", got)
	}
	if got := checkNotifications(fixture.env).Status; got != doctorWarn {
		t.Fatalf("an empty answer = %q", got)
	}
}
