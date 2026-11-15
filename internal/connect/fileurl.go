package connect

// The file URL encoder/decoder.
//
// OpenCode loads a plugin by URL, and the one form known to work on this machine
// is the one sitting in %USERPROFILE%\.config\opencode\plugins\terminal-pet\
// index.ts (handover §53.1):
//
//	file:///C:/Users/rongj/Desktop/%E9%A1%B9%E7%9B%AE%E5%BC%80%E5%8F%91/ai-terminal-pet/reporters/opencode.js
//
// Two properties are load-bearing and neither is obvious:
//
//  1. three slashes. net/url's URL.String() only produces "file:///C:/..." when
//     the path it is given starts with a slash, so a naive URL{Scheme:"file",
//     Path:"C:/x"} emits "file:C:/x" -- which looks fine in a console and fails
//     in the host.
//  2. percent-encoding. The repository lives under a non-ASCII directory
//     (项目开发), and the loader takes the raw URL.
//
// This machine's working index.ts is the golden sample for both properties, and
// fileurl_test.go asserts it byte for byte rather than assuming.
//
// Written by hand rather than with net/url: this package is linked into the
// overlay binary, and net/url drags in net/netip and the URL parser for what is
// here a byte-escape loop. The measured difference is in handover §56.10.

import (
	"errors"
	"fmt"
	"strings"
)

// PathToFileURL renders an absolute local path as the file URL a host can load.
// Both separators are accepted on input; output is always forward slashes,
// because a backslash inside a URL is an escape, not a separator.
func PathToFileURL(local string) (string, error) {
	trimmed := strings.TrimSpace(local)
	if trimmed == "" {
		return "", errors.New("empty path")
	}
	slashed := strings.ReplaceAll(trimmed, `\`, "/")
	if strings.HasPrefix(slashed, "//") {
		// \\server\share\x -> file://server/share/x: the server name is a host
		// component and cannot be percent-encoded into the path.
		rest := strings.TrimPrefix(slashed, "//")
		host, tail, found := strings.Cut(rest, "/")
		if !found || host == "" {
			return "", fmt.Errorf("%q has no share after the server name", local)
		}
		return "file://" + escapeURLPart(host) + "/" + escapeURLPath(tail), nil
	}
	if !strings.HasPrefix(slashed, "/") {
		// A drive-letter path ("C:/x") or a bare relative one. A relative path
		// cannot become a file URL the host resolves from its own working
		// directory, so refusing is the honest answer.
		if !hasDriveLetter(slashed) {
			return "", fmt.Errorf("%q is not absolute", local)
		}
		slashed = "/" + slashed
	}
	return "file://" + escapeURLPath(slashed), nil
}

// hasDriveLetter reports "C:...", the only non-slash-prefixed form Windows
// accepts as absolute.
func hasDriveLetter(value string) bool {
	return len(value) >= 2 && value[1] == ':'
}

// FileURLToPath is the inverse of PathToFileURL. It also accepts the shapes a
// user may have pasted by hand: percent-encoded drive letters, mixed
// separators, surrounding quotes, and a bare path that was never a URL.
func FileURLToPath(ref string) (string, error) {
	trimmed := unquote(strings.TrimSpace(ref))
	if trimmed == "" {
		return "", errors.New("empty reference")
	}
	if !strings.HasPrefix(strings.ToLower(trimmed), "file:") {
		if strings.Contains(trimmed, "://") {
			return "", fmt.Errorf("%q is a URL with a scheme other than file:", trimmed)
		}
		// Already a path: the common shape for the JS reporters.
		return strings.ReplaceAll(trimmed, "/", `\`), nil
	}
	rest := trimmed[len("file:"):]
	if strings.HasPrefix(rest, "//") {
		rest = rest[2:]
		if strings.HasPrefix(rest, "/") {
			// file:///C:/x -- the three-slash form.
			return decodedPathToWindows(rest), nil
		}
		host, tail, found := strings.Cut(rest, "/")
		if !found {
			return "", fmt.Errorf("%q is a file URL with a host but no path", trimmed)
		}
		// file://server/share/x -> \\server\share\x
		decoded, err := percentDecode(tail)
		if err != nil {
			return "", err
		}
		return `\\` + host + `\` + strings.ReplaceAll(decoded, "/", `\`), nil
	}
	// file:C:/x -- how a naive encoder writes it.
	return decodedPathToWindows(rest), nil
}

// decodedPathToWindows turns a URL path into a Windows path: "/C:/Users/x"
// becomes "C:\Users\x". A single leading slash before a drive letter is URL
// syntax, not a path element.
func decodedPathToWindows(escaped string) string {
	decoded, err := percentDecode(escaped)
	if err != nil {
		// A malformed escape is not a path. Returning the raw text keeps the
		// caller's existence check honest: it will not find it, and the finding
		// says what the config pointed at.
		decoded = escaped
	}
	if len(decoded) >= 3 && decoded[0] == '/' && decoded[2] == ':' {
		decoded = decoded[1:]
	}
	return strings.ReplaceAll(decoded, "/", `\`)
}

// unreserved is RFC 3986's "never needs escaping" set, plus the two separators a
// file path needs to stay readable: ':' for the drive letter, '/' for structure.
func isUnreserved(value byte) bool {
	switch {
	case value >= 'a' && value <= 'z':
		return true
	case value >= 'A' && value <= 'Z':
		return true
	case value >= '0' && value <= '9':
		return true
	}
	switch value {
	case '-', '.', '_', '~', '/', ':':
		return true
	}
	return false
}

// escapeURLPath percent-encodes one path, byte by byte. Byte-wise rather than
// rune-wise because that is what a URL carries: 项目 becomes %E9%A1%B9...
func escapeURLPath(value string) string {
	var builder strings.Builder
	builder.Grow(len(value) * 3)
	for index := 0; index < len(value); index++ {
		character := value[index]
		if isUnreserved(character) {
			builder.WriteByte(character)
			continue
		}
		builder.WriteByte('%')
		builder.WriteByte(hexDigit(character >> 4))
		builder.WriteByte(hexDigit(character & 0x0f))
	}
	return builder.String()
}

// escapeURLPart is escapeURLPath for a host component. There is no separate
// implementation: a server name is plain ASCII in practice, and one escaper is
// one place for a bug to be. The name exists so the call site's intent is
// readable.
func escapeURLPart(value string) string {
	return escapeURLPath(value)
}

func hexDigit(value byte) byte {
	if value < 10 {
		return '0' + value
	}
	return 'A' + value - 10
}

// percentDecode reverses escapeURLPath, accepting lowercase escapes too because
// hand-edited configuration uses them.
func percentDecode(value string) (string, error) {
	var builder strings.Builder
	builder.Grow(len(value))
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character != '%' {
			builder.WriteByte(character)
			continue
		}
		if index+2 >= len(value) {
			return "", fmt.Errorf("truncated escape at offset %d", index)
		}
		high, okHigh := unhex(value[index+1])
		low, okLow := unhex(value[index+2])
		if !okHigh || !okLow {
			return "", fmt.Errorf("bad escape %q at offset %d", value[index:index+3], index)
		}
		builder.WriteByte(high<<4 | low)
		index += 2
	}
	return builder.String(), nil
}

func unhex(value byte) (byte, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	}
	return 0, false
}

// unquote strips one layer of matching quotes, which is how every host on
// Windows writes a path containing a space.
func unquote(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// baseName is the last element of either separator style, used when a finding
// names the file it landed on rather than the whole path.
func baseName(value string) string {
	slashed := strings.ReplaceAll(value, `\`, "/")
	if index := strings.LastIndexByte(slashed, '/'); index >= 0 {
		return slashed[index+1:]
	}
	return slashed
}
