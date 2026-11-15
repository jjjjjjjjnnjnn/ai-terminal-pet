//go:build windows

package connect

// How a URL path becomes a local path on Windows.
//
// Windows is the platform where a URL path and a file path disagree: a URL always
// carries forward slashes, and a file path here uses backslashes plus a drive
// letter. Both conversions live in this file so the portable half of the package
// never has to know which machine it is on (handover §55.7 warned about exactly
// this class of "looks fine, is not").

import "strings"

// toLocalPath turns "/C:/Users/x/y.js" into "C:\Users\x\y.js".
func toLocalPath(urlPath string) string {
	// A single leading slash before a drive letter is URL syntax, not a path
	// element.
	if len(urlPath) >= 3 && urlPath[0] == '/' && urlPath[2] == ':' {
		urlPath = urlPath[1:]
	}
	return strings.ReplaceAll(urlPath, "/", `\`)
}

// fromLocalPath turns a local path into the path component of a file URL.
func fromLocalPath(local string) string {
	return strings.ReplaceAll(local, `\`, "/")
}

// hasDriveLetter reports "C:...", the only non-slash-prefixed form Windows
// accepts as absolute.
func hasDriveLetter(value string) bool {
	return len(value) >= 2 && value[1] == ':'
}
