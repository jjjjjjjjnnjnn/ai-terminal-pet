package connect

// The OpenCode check: two hops, and the second one is the one that broke.
//
// opencode.json on this machine names a *directory*:
//
//	"plugins": ["file:///C:/Users/.../.config/opencode/plugins/terminal-pet"]
//
// and OpenCode resolves that to index.ts, which is a static re-export of
// reporters/opencode.js in the repository (handover §53.1). The failure that
// motivated this whole package (handover §52.3) lived on that second hop: the
// JSON entry was perfect, the directory was perfect, and the file it imported
// was not in the repository. Nothing but reading the import can see that.
//
// A "plugins" entry may also be a plain .js file URL, which is one hop. Both
// shapes are handled; anything else is reported as a shape we do not know
// rather than guessed at.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// openCodeConfigDir is where OpenCode keeps opencode.json:
// %USERPROFILE%\.config\opencode\opencode.json. This repository has that file
// on disk, and handover §53.1 records it as the file the local OpenCode
// actually reads.
func openCodeConfigDir(roots Roots) string {
	if roots.Home == "" {
		return ""
	}
	return filepath.Join(roots.Home, ".config", "opencode")
}

// InspectOpenCode reports whether OpenCode is wired to a reporter that still
// exists.
func InspectOpenCode(roots Roots) Finding {
	finding := Finding{Terminal: TerminalOpenCode}
	if roots.Home == "" {
		return finding.warn("cannot be checked (%s is not set)", homeVariable)
	}
	configPath := filepath.Join(openCodeConfigDir(roots), "opencode.json")
	loaded := loadConfig(configPath)
	switch loaded.state {
	case configMissing:
		return finding.warn("not installed (no %s)", abbreviate(roots, configPath))
	case configUnreadable:
		return finding.warn("cannot read %s: %s", abbreviate(roots, configPath), loaded.reason)
	case configMalformed:
		return finding.warn("%s is %s", abbreviate(roots, configPath), loaded.reason)
	}

	entries, listShape := stringList(loaded.object, "plugins", "plugin")
	if listShape == shapeAbsent {
		return finding.warn("not connected (no plugins[] in %s; run `ai-terminal-pet --connect=opencode`)", abbreviate(roots, configPath))
	}
	if listShape < 0 {
		return finding.warn("plugins entry in %s is not a list of strings", abbreviate(roots, configPath))
	}
	if len(entries) == 0 {
		return finding.warn("not connected (%s has an empty plugins list; run `ai-terminal-pet --connect=opencode`)", abbreviate(roots, configPath))
	}

	for index, entry := range entries {
		if !refersToReporter(entry) {
			continue
		}
		// One hop: the entry is the reporter itself.
		if target, err := FileURLToPath(expandUser(roots, entry)); err == nil && !isDirectory(target) {
			if exists(target) {
				return finding.ok("plugins[%d] -> %s", index, abbreviate(roots, target))
			}
			return finding.fail("plugins[%d] -> %s (missing)", index, abbreviate(roots, target))
		}
		// Two hops: the entry is a directory, and its entrypoint re-exports.
		return inspectOpenCodeEntrypoint(roots, finding, index, entry)
	}
	return finding.warn("not connected (plugins[] has %d entr%s, none of them ours; run `ai-terminal-pet --connect=opencode`)",
		len(entries), plural(len(entries)))
}

// inspectOpenCodeEntrypoint follows the directory entry to its entrypoint and
// then to the file that entrypoint imports.
func inspectOpenCodeEntrypoint(roots Roots, finding Finding, index int, entry string) Finding {
	directory, err := FileURLToPath(expandUser(roots, entry))
	if err != nil {
		return finding.warn("plugins[%d] -> %s cannot be read: %v", index, entry, err)
	}
	entrypoint := ""
	for _, name := range []string{"index.ts", "index.js", "index.mjs"} {
		candidate := filepath.Join(directory, name)
		if exists(candidate) {
			entrypoint = candidate
			break
		}
	}
	if entrypoint == "" {
		// A plugin directory with nothing to load. OpenCode would report that
		// the plugin failed to load and keep running, which is exactly the
		// class of silence --doctor exists to break.
		if exists(directory) {
			return finding.fail("plugins[%d] -> %s (no index.ts/index.js inside)", index, abbreviate(roots, directory))
		}
		return finding.fail("plugins[%d] -> %s (missing)", index, abbreviate(roots, directory))
	}
	raw, readErr := os.ReadFile(entrypoint)
	if readErr != nil {
		return finding.warn("cannot read %s: %v", abbreviate(roots, entrypoint), readErr)
	}
	imports := importedTargets(string(raw))
	if len(imports) == 0 {
		return finding.warn("%s imports nothing; the wrapper may be a stale copy", abbreviate(roots, entrypoint))
	}
	for _, reference := range imports {
		// Only file:// imports can be resolved from here. A package import
		// (OpenCode's own API) is not our business.
		if !strings.HasPrefix(strings.ToLower(reference), "file:") {
			continue
		}
		target, err := FileURLToPath(reference)
		if err != nil {
			return finding.warn("%s imports %s, which is not a usable path: %v", abbreviate(roots, entrypoint), reference, err)
		}
		if exists(target) {
			return finding.ok("%s -> %s", abbreviate(roots, entrypoint), abbreviate(roots, target))
		}
		return finding.fail("%s -> %s (missing - the repository file is gone, which is handover §52.3's failure mode)",
			abbreviate(roots, entrypoint), abbreviate(roots, target))
	}
	return finding.warn("%s imports no file:// target, so it cannot reach a reporter", abbreviate(roots, entrypoint))
}

// importedTargets returns every `from "<target>"` string in a wrapper, in
// source order. Built on scanSource, so a path quoted in a comment -- this
// project's own wrapper names reporters/opencode.js in three of them -- cannot
// be mistaken for the import that matters.
func importedTargets(source string) []string {
	tokens := scanSource(source)
	targets := make([]string, 0, 2)
	for index := 0; index+1 < len(tokens); index++ {
		if tokens[index].kind != tokenWord || !strings.EqualFold(tokens[index].value, "from") {
			continue
		}
		if tokens[index+1].kind != tokenString {
			continue
		}
		targets = append(targets, tokens[index+1].value)
	}
	return targets
}

func isIdentifierByte(value byte) bool {
	return value == '_' || value == '$' ||
		(value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || (value >= '0' && value <= '9')
}

// isDirectory reports whether a path that resolved is a directory, which is
// how the two hops are told apart before deciding which loop to run.
func isDirectory(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func plural(count int) string {
	if count == 1 {
		return "y"
	}
	return "ies"
}

// shape tells the caller what a configuration key looked like. The values are
// explicit rather than iota: "absent" must not be the zero value, because a
// key that was never looked at and a key that was not there are only the same
// answer by accident, and this is a diagnostic.
type shape int

const (
	shapeAbsent     shape = -1
	shapeWrongType  shape = -2
	shapeStringList shape = 1
)

// stringList reads a key that may be a list of strings or a single string.
// OpenCode's own documentation has used both "plugins" and "plugin", and this
// machine's file uses the plural form, so both are accepted.
func stringList(object map[string]any, keys ...string) ([]string, shape) {
	for _, key := range keys {
		value, ok := object[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			return []string{typed}, shapeStringList
		case []any:
			entries := make([]string, 0, len(typed))
			for _, element := range typed {
				text, ok := element.(string)
				if !ok {
					return nil, shapeWrongType
				}
				entries = append(entries, text)
			}
			return entries, shapeStringList
		default:
			return nil, shapeWrongType
		}
	}
	return nil, shapeAbsent
}

// ok/warn/fail keep every branch of the three inspectors one line and one
// shape, so nothing is returned half-built.
func (f Finding) ok(format string, args ...any) Finding {
	f.Status = StatusOK
	f.Detail = fmt.Sprintf(format, args...)
	return f
}

func (f Finding) warn(format string, args ...any) Finding {
	f.Status = StatusWarn
	f.Detail = fmt.Sprintf(format, args...)
	return f
}

func (f Finding) fail(format string, args ...any) Finding {
	f.Status = StatusFail
	f.Detail = fmt.Sprintf(format, args...)
	return f
}
