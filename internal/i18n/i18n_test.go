package i18n

import "testing"

func TestNormalizeSetting(t *testing.T) {
	cases := map[string]Setting{
		"":          SettingAuto,
		"auto":      SettingAuto,
		"AUTO":      SettingAuto,
		"en":        SettingEn,
		"EN":        SettingEn,
		"zh-CN":     SettingZhCN,
		"  zh-CN  ": SettingZhCN,
		"fr":        SettingAuto,
		"junk":      SettingAuto,
	}
	for input, want := range cases {
		if got := NormalizeSetting(input); got != want {
			t.Errorf("NormalizeSetting(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResolve(t *testing.T) {
	cases := []struct {
		setting Setting
		locale  string
		want    Lang
	}{
		{SettingAuto, "zh-CN", LangZhCN},
		{SettingAuto, "zh-TW", LangZhCN},
		{SettingAuto, "zh", LangZhCN},
		{SettingAuto, "zh_Hans_CN", LangZhCN},
		{SettingAuto, "en-US", LangEn},
		{SettingAuto, "ja-JP", LangEn},
		{SettingAuto, "", LangEn},
		{SettingAuto, "???", LangEn},
		{SettingEn, "zh-CN", LangEn},
		{SettingZhCN, "en-US", LangZhCN},
		{Setting(""), "zh-CN", LangZhCN},
		{Setting("fr"), "zh-CN", LangZhCN},
	}
	for _, tc := range cases {
		if got := Resolve(tc.setting, tc.locale); got != tc.want {
			t.Errorf("Resolve(%q, %q) = %q, want %q", tc.setting, tc.locale, got, tc.want)
		}
	}
}

func TestTFallback(t *testing.T) {
	if got := T(LangEn, KeyStateWaiting); got == "" || got == KeyStateWaiting {
		t.Errorf("T(en, state.waiting) = %q, want a label", got)
	}
	if got := T(LangZhCN, KeyStateWaiting); got != "等待确认" {
		t.Errorf("T(zh-CN, state.waiting) = %q, want 等待确认", got)
	}
	// Missing key in zh-CN falls back to English.
	if got := T(LangZhCN, "no.such.key"); got != "no.such.key" {
		t.Errorf("T(zh-CN, missing) = %q, want the key itself", got)
	}
	// Both dictionaries must carry the same key set: every English key
	// must exist in Chinese, otherwise the UI would silently flip
	// language mid-panel.
	for key := range dictEn {
		if _, ok := dictZhCN[key]; !ok {
			t.Errorf("dictZhCN missing key %q", key)
		}
	}
	for key := range dictZhCN {
		if _, ok := dictEn[key]; !ok {
			t.Errorf("dictEn missing key %q", key)
		}
	}
}
