//go:build !windows

package i18n

// The POSIX locale lookup: LANG, LC_ALL and LC_MESSAGES, in that order of
// precedence (LC_ALL wins over everything, LC_MESSAGES is the most specific
// category and some shells set only that).
//
// Measured formats, from the three places a value can come from:
//
//	en_US.UTF-8   Linux, the common case (UTF-8 suffix)
//	zh_CN.UTF-8   the locale this project's Chinese UI is selected by
//	zh_CN         a value with no codeset, as macOS often stores it
//	C             the "C"/"POSIX" pseudo-locale, which means "no locale": "" here
//
// The separator between language and territory is rewritten from '_' to '-'
// because NormalizeLang compares a prefix and the rest of this project's code
// speaks in BCP 47 names ("zh-CN"). Getting this wrong does not break anything
// loudly - it falls back to English - which is exactly why it has a test.

import (
	"os"
	"strings"
)

func systemLocale() (string, error) {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			continue
		}
		// LANGUAGE may be a colon-separated priority list; the first entry is
		// the one the user put first.
		if index := strings.IndexByte(value, ':'); index >= 0 {
			value = value[:index]
		}
		return normalizePOSIXLocale(value), nil
	}
	return "", nil
}

// normalizePOSIXLocale turns "zh_CN.UTF-8@euro" into "zh-CN": the codeset after
// '.' and the modifier after '@' carry no language information, and the
// territory separator has to become a hyphen for the rest of the code.
func normalizePOSIXLocale(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	// "C" and "POSIX" are the standard "unset" names.
	if trimmed == "C" || trimmed == "POSIX" || trimmed == "c" {
		return ""
	}
	if index := strings.IndexByte(trimmed, '.'); index >= 0 {
		trimmed = trimmed[:index]
	}
	if index := strings.IndexByte(trimmed, '@'); index >= 0 {
		trimmed = trimmed[:index]
	}
	trimmed = strings.ReplaceAll(trimmed, "_", "-")
	lowered := strings.ToLower(trimmed)
	// "en-us" -> "en-US" like the Windows name, so the two platforms report
	// the same string for the same machine. The language part is two letters
	// (or three for some ISO 639-2 codes), so the separator is found rather
	// than assumed to sit at a fixed offset.
	if index := strings.IndexByte(lowered, '-'); index > 0 {
		return lowered[:index] + "-" + strings.ToUpper(trimmed[index+1:])
	}
	return lowered
}
