package connect

// The OpenCode check: three shapes, and the third one is what a user who has
// never cloned this repository actually has.
//
// opencode.json on the machine this was written on names a *directory*:
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
//
// Since the reporter is published as an npm package, the primary shape is a
// bare package name (or an object with a "package" field, which is the other
// documented form). That shape has no local file to inspect: OpenCode resolves
// it from its own package directory, so the evidence is the installed package's
// manifest rather than a path in the config.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OpenCodePackageName is the npm package this project publishes for OpenCode.
// One constant, because the snippet, the doctor line and the tests all name it.
const OpenCodePackageName = "opencode-ai-terminal-pet"

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
		// Zero hops: a package name. There is no file in the config to follow,
		// so the evidence is the package OpenCode installed for itself.
		if name := npmPackageName(entry); name != "" {
			return inspectOpenCodePackage(roots, finding, index, name)
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

// inspectOpenCodePackage reports the npm-package shape.
//
// The evidence is not the config directory. `opencode plugin add` installs the
// package into OpenCode's own npm store -- measured on 2.3.0 with a throwaway
// profile, the files land in
//
//	<cache>/opencode/npm/<name>@<tag>/<epoch-ms>/node_modules/<name>/
//
// where <cache> is $XDG_CACHE_HOME or ~/.cache (the latter on Windows too,
// which has no XDG but uses that path anyway). A manual `npm install` inside
// ~/.config/opencode puts it in that directory's node_modules instead, so both
// are checked; the store is searched first because it is what the CLI writes.
//
// "Not found in either" is a warn, not a fail: an install this tool cannot see
// is not proof that the plugin is missing, and calling a working setup broken
// sends the user to redo work that is already done.
func inspectOpenCodePackage(roots Roots, finding Finding, index int, name string) Finding {
	manifest, version := findOpenCodePackage(roots, name)
	if manifest == "" {
		return finding.warn("plugins[%d] -> %s is listed but no installed copy was found (run `opencode plugin add %s`)",
			index, name, name)
	}
	if version == "" {
		return finding.ok("plugins[%d] -> %s (installed in %s)", index, name, abbreviate(roots, filepath.Dir(manifest)))
	}
	return finding.ok("plugins[%d] -> %s@%s", index, name, version)
}

// findOpenCodePackage returns the manifest of an installed copy and the version
// it states. Patterns are searched in order and each one's matches are sorted,
// so with several past installs the newest timestamp wins.
func findOpenCodePackage(roots Roots, name string) (string, string) {
	for _, pattern := range openCodePackagePatterns(roots, name) {
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			continue
		}
		sort.Strings(matches)
		manifest := matches[len(matches)-1]
		raw, err := os.ReadFile(manifest)
		if err != nil {
			continue
		}
		var parsed struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(raw, &parsed) != nil {
			continue
		}
		return manifest, strings.TrimSpace(parsed.Version)
	}
	return "", ""
}

// openCodePackagePatterns lists where an installed copy can be, in the order
// they are worth looking at.
func openCodePackagePatterns(roots Roots, name string) []string {
	patterns := make([]string, 0, 3)
	for _, cache := range openCodeCacheRoots(roots) {
		patterns = append(patterns, filepath.Join(cache, "npm", name+"@*", "*", "node_modules", filepath.FromSlash(name), "package.json"))
	}
	if directory := openCodeConfigDir(roots); directory != "" {
		patterns = append(patterns, filepath.Join(directory, "node_modules", filepath.FromSlash(name), "package.json"))
	}
	return patterns
}

// openCodeCacheRoots is where OpenCode caches what it installs. $XDG_CACHE_HOME
// is honoured first because that is the XDG rule, and ~/.cache is the fallback
// that Windows uses as well (no XDG there; measured, not assumed).
func openCodeCacheRoots(roots Roots) []string {
	rootsOut := make([]string, 0, 2)
	if roots.XDGCacheHome != "" {
		rootsOut = append(rootsOut, filepath.Join(roots.XDGCacheHome, "opencode"))
	}
	if roots.Home != "" {
		rootsOut = append(rootsOut, filepath.Join(roots.Home, ".cache", "opencode"))
	}
	return rootsOut
}

// npmPackageName returns the package name an entry refers to, or "" when the
// entry is a path or a git spec instead.
//
// The distinction has to be positive, not "does it look like a path": a nil
// answer here falls through to the file/directory branches, and a wrong "yes"
// would make --doctor check a manifest nobody named. So the shapes this project
// itself writes are the only ones accepted - a bare name, a scoped name, either
// with an optional @version - and everything with a scheme, a slash in a
// non-scoped position (github:owner/repo), a drive letter or a leading dot is
// left to the path branches.
func npmPackageName(entry string) string {
	spec := strings.TrimSpace(entry)
	if spec == "" {
		return ""
	}
	if strings.Contains(spec, "://") || strings.ContainsRune(spec, '\\') {
		return "" // a URL, or a Windows path
	}
	if strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") || strings.HasPrefix(spec, "~") {
		return "" // a relative or home-relative path
	}
	if len(spec) > 1 && spec[1] == ':' {
		return "" // a Windows drive path: C:/...
	}
	name := spec
	if at := strings.LastIndex(spec, "@"); at > 0 {
		name = spec[:at] // drop @1.2.3 / @latest / @range
	}
	if strings.HasPrefix(name, "@") {
		if strings.Count(name, "/") != 1 {
			return ""
		}
	} else if strings.ContainsRune(name, '/') {
		return ""
	}
	if !isPackageName(name) {
		return ""
	}
	return name
}

// isPackageName accepts the characters npm allows in a name, plus the one slash
// a scoped name needs. It is a filter against nonsense, not a validator.
func isPackageName(name string) bool {
	if name == "" || name == "@" || strings.HasSuffix(name, "/") {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == '/', r == '@':
		default:
			return false
		}
	}
	return true
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
				switch value := element.(type) {
				case string:
					entries = append(entries, value)
				case map[string]any:
					// OpenCode also documents an object form,
					// {"package": "...", "options": {...}}. The package name is
					// what decides "is this entry ours"; the options are another
					// product's schema and are deliberately not read here.
					name, isString := value["package"].(string)
					if !isString || strings.TrimSpace(name) == "" {
						return nil, shapeWrongType
					}
					entries = append(entries, name)
				default:
					return nil, shapeWrongType
				}
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
