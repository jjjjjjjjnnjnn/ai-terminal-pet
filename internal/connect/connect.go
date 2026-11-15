// Package connect answers the half of the contract the overlay cannot see:
// on this machine, which terminals are actually wired to us, and are the files
// they point at still there?
//
// It exists because a broken wiring looks exactly like an idle terminal. The
// reporter is fire-and-forget by design (AGENTS rule 1), so a stale path in
// opencode.json, a wrapper whose import target moved, or a deleted
// pet-statusline.exe all produce the same visible outcome: no session row.
// --doctor already inspects the install directory, the pipe, Node and the
// autostart entry; it inspects nothing on the other side of the contract, the
// host's own configuration (ROADMAP §E6 recorded that gap when the check set
// was written).
//
// Two consumers, one implementation:
//   - internal/app renders Inspect() as the --doctor "reporters" line;
//   - --connect renders Snippets() as paste-ready text and writes nothing.
//
// The package holds no Win32 and no window: every external fact arrives
// through Roots, so a test can build three configuration trees in a temp
// directory and drive each branch without touching the machine it runs on.
// That is also why it can be lifted into the macOS build unchanged when E3
// gets there: the only platform-shaped part is where the hosts happen to keep
// their files, and that is a Roots value, not a syscall.
package connect

import (
	"os"
	"strings"

	pethome "ai-terminal-pet/internal/home"
)

// The three hosts this project ships reporters for. The names are also the
// values --connect accepts, so they are literals rather than an enum with a
// String method: assigning one to the other is the point.
const (
	TerminalOpenCode    = "opencode"
	TerminalClaude      = "claude"
	TerminalAntigravity = "antigravity"
)

// TerminalNames is the order --connect prints and --doctor reports in.
// Documented order, not map order: a report that reshuffles between runs is
// unreadable in a bug thread.
func TerminalNames() []string {
	return []string{TerminalOpenCode, TerminalClaude, TerminalAntigravity}
}

// Status is one terminal's verdict.
//
// The zero value is deliberately not a status a constructor ever produces: a
// Finding that forgot to set one must not read as healthy. String maps it to
// "fail" for the same reason --doctor counts an unknown status as a failure --
// a summary that quietly shortens itself is worse than a loud one.
type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return string(StatusOK)
	case StatusWarn:
		return string(StatusWarn)
	default:
		return string(StatusFail)
	}
}

// rank exists so Worst() can compare verdicts without a second table.
func (s Status) rank() int {
	switch s {
	case StatusOK:
		return 0
	case StatusWarn:
		return 1
	default:
		return 2
	}
}

// Finding is one terminal's line: who was inspected, the verdict, and the
// measured value. Detail always carries evidence -- a "warn" the user cannot
// check is a warn they will not trust.
type Finding struct {
	Terminal string
	Status   Status
	Detail   string
}

// Line renders one finding the way both consumers want it:
// "opencode: ok - plugins[0] -> ~/.config/.../reporters/opencode.js".
func (f Finding) Line() string {
	return f.Terminal + ": " + f.Status.String() + " - " + f.Detail
}

// Roots is every external fact these checks need. Splitting it out keeps
// Inspect a pure function of its input: the seven config shapes below are
// all covered by tests that build them under t.TempDir().
type Roots struct {
	// Home is the user profile (%USERPROFILE%). The three hosts keep their
	// configuration under it, and it is also the abbreviation source: details
	// say "~/.claude/settings.json", not the full path, so a report line stays
	// readable when it is pasted into an issue.
	Home string
	// LocalAppData is where the Windows installer puts the binaries (%LOCALAPPDATA%).
	// A native reporter reference is resolved against this prefix because
	// that is the one copy every reporter is meant to run from.
	LocalAppData string
	// XDGDataHome holds $XDG_DATA_HOME when resolving Linux installs without an
	// explicit DataDir.
	XDGDataHome string
	// DataDir overrides the resolved data/install directory (home.Directory()).
	DataDir string
	// Executable is os.Executable(). Snippets uses it to find a repository
	// checkout when the overlay was launched straight out of bin/ -- OpenCode
	// loads reporters/opencode.js from the checkout, which is the one path an
	// install directory cannot supply.
	Executable string
	// ExecutableErr is what os.Executable() returned if it failed. Snippets
	// degrades to a marker instead of guessing a repository path.
	ExecutableErr error
}

// DefaultRoots reads the three values from the environment. Home falls back to
// USERPROFILE when UserHomeDir fails, because every host here keeps its
// configuration under a profile directory and a machine without one has no
// configuration to inspect anyway -- reporting "not installed" is the honest
// answer, and it stays a warn.
func DefaultRoots() Roots {
	executable, executableErr := os.Executable()
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
	}
	return Roots{
		Home:          homeDir,
		LocalAppData:  os.Getenv("LOCALAPPDATA"),
		XDGDataHome:   os.Getenv("XDG_DATA_HOME"),
		DataDir:       pethome.Directory(),
		Executable:    executable,
		ExecutableErr: executableErr,
	}
}

// Inspect runs all three checks. Order is fixed and unrelated to severity:
// the report reads the same way on every machine, so a diff between two
// --doctor outputs only shows what actually changed.
func Inspect(roots Roots) []Finding {
	return []Finding{
		InspectOpenCode(roots),
		InspectClaude(roots),
		InspectAntigravity(roots),
	}
}

// Worst is the verdict --doctor reports for the whole line: one broken
// terminal is not cancelled out by two healthy ones.
func Worst(findings []Finding) Status {
	worst := StatusOK
	for _, finding := range findings {
		if finding.Status.rank() > worst.rank() {
			worst = finding.Status
		}
	}
	return worst
}

// Summary joins the findings into the single Detail line --doctor prints.
// Findings are separated by "; " and the doctor's oneLine() collapses any
// remaining whitespace, so the whole set stays on one line.
func Summary(findings []Finding) string {
	lines := make([]string, 0, len(findings))
	for _, finding := range findings {
		lines = append(lines, finding.Line())
	}
	return strings.Join(lines, "; ")
}
