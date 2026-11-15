//go:build !windows

package i18n

// The POSIX side of the locale lookup, tested directly: normalizePOSIXLocale is
// pure, so these cases run on every platform rather than only on the machine
// that happens to have a Chinese locale configured.

import "testing"

func TestNormalizePOSIXLocale(t *testing.T) {
	cases := map[string]string{
		"en_US.UTF-8":      "en-US",
		"zh_CN.UTF-8":      "zh-CN",
		"zh_CN":            "zh-CN",
		"zh_CN.utf8":       "zh-CN",
		"de_DE.UTF-8@euro": "de-DE",
		"zh_TW.UTF-8":      "zh-TW",
		"en_GB.ISO-8859-1": "en-GB",
		"C":                "",
		"POSIX":            "",
		"":                 "",
		"  ja_JP.UTF-8 ":   "ja-JP",
		"fr_FR.UTF-8:fr":   "fr-FR",
	}
	for input, want := range cases {
		if got := normalizePOSIXLocale(input); got != want {
			t.Errorf("normalizePOSIXLocale(%q) = %q, want %q", input, got, want)
		}
	}
}

// The two platforms have to report the same string for the same language, or a
// zh_CN machine would render Chinese on Linux and English on macOS: the whole
// point of a shared NormalizeLang.
func TestNormalizePOSIXLocaleAgreesWithWindowsNames(t *testing.T) {
	for _, pair := range [][2]string{
		{"zh_CN.UTF-8", "zh-CN"},
		{"en_US.UTF-8", "en-US"},
	} {
		if got := normalizePOSIXLocale(pair[0]); got != pair[1] {
			t.Fatalf("%q normalized to %q, Windows would report %q", pair[0], got, pair[1])
		}
	}
}
