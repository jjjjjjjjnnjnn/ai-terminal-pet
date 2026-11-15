//go:build linux

package autostart

// The Linux backend: an XDG autostart desktop entry.
//
// This is the mechanism that works across desktop environments without a
// privileged helper: ~/.config/autostart/<name>.desktop is read by GNOME, KDE,
// XFCE and every other XDG-compliant session at login, for this user only - the
// same scope HKCU\...\Run gives on Windows and a LaunchAgent gives on macOS.
//
// Deliberately NOT a systemd user unit: a `systemd --user` service needs the user's
// session bus and either lingers after logoff (which a pet must not) or needs
// `systemctl --user enable` run from a live session. The desktop entry is a file
// that either exists or does not, which is exactly the semantics Enable/Disable/
// Sync already have.
//
// The two keys that matter:
//
//	Exec            the command line. A path with spaces must be quoted, the same
//	                reason Windows quotes the Run value.
//	X-GNOME-Autostart-enabled  what some desktop filesystems read to grey the entry
//	                out without deleting it; written as true so a user who copies
//	                the file elsewhere does not inherit a disabled entry.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// desktopEntryName is the file name, without the -autostart prefix some
// environments expect.
const desktopEntryName = "ai-terminal-pet.desktop"

// desktopEntryPath is ~/.config/autostart/<name>.desktop, honouring
// XDG_CONFIG_HOME when it is absolute.
func desktopEntryPath() (string, error) {
	if root := os.Getenv("XDG_CONFIG_HOME"); root != "" && isRooted(root) {
		return posixJoin(root, "autostart", desktopEntryName), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return "", errors.New("no home directory for the autostart entry")
	}
	return posixJoin(homeDir, ".config", "autostart", desktopEntryName), nil
}

// newBackend returns the one mechanism this platform has.
func newBackend() backend {
	return desktopEntryBackend{}
}

type desktopEntryBackend struct{}

func (desktopEntryBackend) read(name string) (string, bool, error) {
	path, err := desktopEntryPath()
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
	// A [Desktop Entry] group is required; a file without one is not an entry any
	// session will read, so it is reported as absent rather than as a broken one.
	if !strings.Contains(string(data), "[Desktop Entry]") {
		return "", false, nil
	}
	exec := desktopEntryValue(string(data), "Exec")
	if exec == "" {
		return "", true, nil
	}
	// Exec may carry arguments; the target is the first token, unquoted.
	return firstCommandToken(exec), true, nil
}

func (desktopEntryBackend) write(name, value string) error {
	path, err := desktopEntryPath()
	if err != nil {
		return err
	}
	executable := unquote(value)
	contents := desktopEntry(executable)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, []byte(contents), 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

func (desktopEntryBackend) remove(name string) error {
	path, err := desktopEntryPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// desktopEntry renders the file. Hidden is false so the entry shows up in a
// desktop environment's own startup list: a background service that hides itself
// from the one place a user can turn it off is a bad neighbour.
func desktopEntry(executable string) string {
	var builder strings.Builder
	builder.WriteString("[Desktop Entry]\n")
	builder.WriteString("Type=Application\n")
	builder.WriteString("Name=AI Terminal Pet\n")
	builder.WriteString("Comment=Floating status overlay for AI coding terminals\n")
	builder.WriteString("Exec=" + quoteDesktopExec(executable) + "\n")
	builder.WriteString("Terminal=false\n")
	builder.WriteString("X-GNOME-Autostart-enabled=true\n")
	builder.WriteString("Hidden=false\n")
	return builder.String()
}

// quoteDesktopExec quotes the executable for the Exec key. The desktop entry spec
// allows double quotes and requires the backslash, dollar and backtick characters
// inside them to be escaped.
func quoteDesktopExec(executable string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`, "`", "\\`").Replace(executable)
	return `"` + escaped + `"`
}

// desktopEntryValue reads one key's value from a desktop entry, stopping at the next
// key or group. Blank lines and comments are skipped, which is what a hand-edited
// file looks like.
func desktopEntryValue(contents, key string) string {
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "[") {
			continue
		}
		name, value, found := strings.Cut(trimmed, "=")
		if !found || strings.TrimSpace(name) != key {
			continue
		}
		return strings.TrimSpace(value)
	}
	return ""
}

// firstCommandToken returns the first argument of a command line, unquoted.
func firstCommandToken(command string) string {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return ""
	}
	if trimmed[0] == '"' {
		if end := strings.Index(trimmed[1:], `"`); end >= 0 {
			return strings.ReplaceAll(trimmed[1:1+end], `\"`, `"`)
		}
		return trimmed
	}
	fields := strings.Fields(trimmed)
	return fields[0]
}

// isRooted reports whether a value starts with "/" or a drive letter, independent
// of the host's own path rules.
func isRooted(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '/' {
		return true
	}
	return len(value) >= 2 && value[1] == ':'
}

// posixJoin joins with forward slashes.
func posixJoin(elements ...string) string {
	result := ""
	for _, element := range elements {
		if element == "" {
			continue
		}
		if result == "" {
			result = strings.TrimRight(element, "/")
			continue
		}
		result += "/" + strings.Trim(element, "/")
	}
	return result
}
