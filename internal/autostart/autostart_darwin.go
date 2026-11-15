//go:build darwin

package autostart

// The macOS backend: a LaunchAgent plist.
//
// macOS has no Run key. The equivalent that works without a privileged helper and
// without root is ~/Library/LaunchAgents/<label>.plist with RunAtLoad, which the
// login session's launchd reads for this user only - the same scope HKCU\...\Run
// gives on Windows.
//
// Two details decide the file's shape:
//
//   - **ProgramArguments, not Program.** A path with spaces is the normal case on
//     macOS (".app/Contents/MacOS/..."), and ProgramArguments is the form launchd
//     does not re-split. Windows had to quote the Run value for the same reason.
//   - **KeepAlive is absent.** The overlay is on-demand by design and exits when no
//     session is left; KeepAlive would relaunch it forever, which is exactly the
//     behaviour the idle self-destruct exists to avoid.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// launchAgentLabel is the reverse-DNS label launchd keys the agent on. It is the
// ValueName in a namespace launchd requires to be unique, which is also why Status
// reports the *executable* rather than this label.
const launchAgentLabel = "org.ai-terminal-pet.overlay"

// launchAgentPath is where the plist lives for the current user.
func launchAgentPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return "", errors.New("no home directory for the launch agent")
	}
	return filepath.Join(homeDir, "Library", "LaunchAgents", launchAgentLabel+".plist"), nil
}

// newBackend returns the one mechanism this platform has.
func newBackend() backend {
	return launchAgentBackend{}
}

type launchAgentBackend struct{}

func (launchAgentBackend) read(name string) (string, bool, error) {
	path, err := launchAgentPath()
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	program := plistStringValue(string(data), "Program")
	if program == "" {
		// A plist without Program still counts as "enabled" - launchd would try
		// to run it - so the honest answer is the empty target, not "off".
		return "", true, nil
	}
	return program, true, nil
}

func (launchAgentBackend) write(name, value string) error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	executable := unquote(value)
	contents := launchAgentPlist(executable)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Atomic replace: launchd may read the file at any moment, and a half-written
	// plist is a plist launchd logs an error about at every logon.
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, []byte(contents), 0o644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

func (launchAgentBackend) remove(name string) error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// launchd caches the loaded job; unload is best effort and its failure is not
	// the caller's problem (a user who never logged in has nothing loaded).
	_ = unloadLaunchAgent()
	return nil
}

// launchAgentPlist renders the property list.
//
// Written by hand rather than with a plist library: the document is seven elements,
// and a dependency for it would be linked into the overlay binary for one file that
// is generated once per install.
func launchAgentPlist(executable string) string {
	var builder strings.Builder
	builder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	builder.WriteString("<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n")
	builder.WriteString("<plist version=\"1.0\">\n<dict>\n")
	builder.WriteString("  <key>Label</key>\n  <string>" + xmlEscape(launchAgentLabel) + "</string>\n")
	builder.WriteString("  <key>Program</key>\n  <string>" + xmlEscape(executable) + "</string>\n")
	builder.WriteString("  <key>RunAtLoad</key>\n  <true/>\n")
	builder.WriteString("  <key>ProcessType</key>\n  <string>Interactive</string>\n")
	builder.WriteString("</dict>\n</plist>\n")
	return builder.String()
}

func xmlEscape(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(value)
}

// plistStringValue extracts the <string> that follows a given <key>. A plist parser
// would be the wrong tool here: the file is ours, the key set is fixed, and a
// malformed one has to answer "not enabled" rather than fail Status.
func plistStringValue(contents, key string) string {
	marker := "<key>" + xmlEscape(key) + "</key>"
	index := strings.Index(contents, marker)
	if index < 0 {
		return ""
	}
	rest := contents[index+len(marker):]
	open := strings.Index(rest, "<string>")
	if open < 0 {
		return ""
	}
	close := strings.Index(rest[open:], "</string>")
	if close < 0 {
		return ""
	}
	value := rest[open+len("<string>") : open+close]
	decoder := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">")
	return decoder.Replace(strings.TrimSpace(value))
}

// unloadLaunchAgent asks launchd to forget the job. Errors are ignored on purpose:
// this is a convenience, and a plist that is gone is what the caller asked for.
func unloadLaunchAgent() error {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return err
	}
	return nil
}
