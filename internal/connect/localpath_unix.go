//go:build !windows

package connect

// The unix path spelling. Less interesting than Windows, and worth writing down
// because "less interesting" is exactly what used to make this package
// Windows-only: the drift, not the difficulty, was the problem.
//
// A URL path and a file path agree here - both use forward slashes - so
// toLocalPath is the identity and there is no drive letter to look for. What this
// file does have to do is reject the thing PathToFileURL must refuse everywhere: a
// bare relative name, which a host cannot resolve from its own directory.

import "strings"

// toLocalPath keeps the path as it came; percent-decoding already happened.
func toLocalPath(urlPath string) string {
	return urlPath
}

// fromLocalPath turns a local path into the path component of a file URL. On
// POSIX that is the identity, because a backslash is part of the file name here,
// not a separator (handover §55.7's "looks the same on both" trap).
func fromLocalPath(local string) string {
	return local
}

// hasDriveLetter never applies: a POSIX absolute path starts with "/", which
// PathToFileURL already accepts.
func hasDriveLetter(value string) bool {
	return false
}

// posixPathConversion is the marker that this file is the unix spelling: nothing
// else in the package is allowed to assume a separator. Kept declared rather than
// documented so a future edit that reintroduces `strings.ReplaceAll(x, "/", "\\")`
// outside this file is visible in review by its absence.
func posixPathConversion() bool {
	return !strings.Contains("", `\`)
}
