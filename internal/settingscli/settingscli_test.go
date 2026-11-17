package settingscli

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/project"
)

// TestEveryPersistedPreferenceIsEditable is the guard that matters most: a new
// config key that never reaches this table is a key Linux and macOS users can
// only set by hand-editing JSON. The exclusion list is explicit, so adding a key
// to internal/config fails here until somebody decides which side it belongs on.
func TestEveryPersistedPreferenceIsEditable(t *testing.T) {
	// Deliberately not editable from a terminal:
	//   mutedSessions - a list of per-session keys; see ClearMutes.
	//   settingsTextSize / settingsKeyboard / settingsTooltips - they configure a
	//     settings *window*, and this surface has none.
	//   x / y / placed / layouts - window placement, owned by the panel.
	excluded := []string{
		"mutedSessions",
		"settingsTextSize", "settingsKeyboard", "settingsTooltips",
		"x", "y", "placed", "layouts",
	}
	editable := Keys(Fields())
	for _, key := range excluded {
		if contains(editable, key) {
			t.Errorf("%q is in the editable list and the exclusion list; pick one", key)
		}
	}
	all := configKeys(t)
	for _, key := range all {
		if contains(excluded, key) {
			continue
		}
		if !contains(editable, key) {
			t.Errorf("config key %q has no settings entry: either add a Field or add it to the exclusion list, with the reason", key)
		}
	}
	for _, key := range editable {
		if !contains(all, key) {
			t.Errorf("settings entry %q names a config key that does not exist", key)
		}
	}
}

// configKeys reads the json names off config.Config by reflection, so the list
// follows the struct instead of being retyped.
func configKeys(t *testing.T) []string {
	t.Helper()
	var keys []string
	value := reflect.TypeOf(config.Config{})
	for index := 0; index < value.NumField(); index++ {
		name := strings.Split(value.Field(index).Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestSetValueAcceptsWhatPeopleType(t *testing.T) {
	cases := []struct {
		key   string
		raw   string
		check func(config.Config) bool
		want  string
	}{
		{"notifications", "off", func(c config.Config) bool { return !c.Notifications }, "off"},
		{"notifications", "TRUE", func(c config.Config) bool { return c.Notifications }, "on"},
		{"notifications", "yes", func(c config.Config) bool { return c.Notifications }, "on"},
		{"soundVolume", "0", func(c config.Config) bool { return c.SoundVolume == 0 }, "0"},
		{"soundScheme", "2", func(c config.Config) bool { return c.SoundScheme == config.SoundSchemes[1] }, config.SoundSchemes[1]},
		{"soundScheme", "PULSE", func(c config.Config) bool { return c.SoundScheme == "pulse" }, "pulse"},
		{"theme", "paper", func(c config.Config) bool { return c.Theme == config.ThemePaper }, config.ThemePaper},
		{"donePolicy", "always", func(c config.Config) bool { return c.DonePolicy == config.DonePolicyAlways }, config.DonePolicyAlways},
		{"opacity", "40", func(c config.Config) bool { return c.Opacity == config.MinOpacity }, "40"},
		{"idleMinutes", "0", func(c config.Config) bool { return c.IdleMinutes == 0 }, "0"},
		{"language", "zh-cn", func(c config.Config) bool { return c.Language == string(i18n.LangZhCN) }, string(i18n.LangZhCN)},
		{"autoApproveTimeoutSec", "60", func(c config.Config) bool { return c.AutoApproveTimeoutSec == 60 }, "60"},
	}
	for _, tc := range cases {
		value := config.Default()
		got, err := SetValue(&value, tc.key, tc.raw)
		if err != nil {
			t.Fatalf("%s=%s: %v", tc.key, tc.raw, err)
		}
		if !tc.check(value) {
			t.Fatalf("%s=%s: the value was accepted but did not land", tc.key, tc.raw)
		}
		if got != tc.want {
			t.Fatalf("%s=%s: answered %q, want %q", tc.key, tc.raw, got, tc.want)
		}
	}
}

func TestSetValueRefusesRatherThanGuessing(t *testing.T) {
	cases := []struct{ key, raw, message string }{
		{"notifications", "maybe", "not on or off"},
		{"soundVolume", "loud", "not a number"},
		{"soundVolume", "101", "outside 0..100"},
		{"opacity", "39", "outside"},
		{"autoApproveTimeoutSec", "0", "outside 1..60"},
		{"theme", "neon", "not one of"},
		{"soundScheme", "5", "not one of 1.."},
		{"language", "fr", "not one of"},
		{"donePolicy", "sometimes", "not one of"},
		{"noSuchKey", "1", "unknown key"},
	}
	for _, tc := range cases {
		value := config.Default()
		before := value
		if _, err := SetValue(&value, tc.key, tc.raw); err == nil {
			t.Fatalf("%s=%s was accepted", tc.key, tc.raw)
		} else if !strings.Contains(err.Error(), tc.message) {
			t.Fatalf("%s=%s: error %q does not mention %q", tc.key, tc.raw, err, tc.message)
		}
		if !reflect.DeepEqual(value, before) {
			t.Fatalf("%s=%s modified the config despite refusing", tc.key, tc.raw)
		}
	}
}

func TestAnUnknownKeyListsTheKnownOnes(t *testing.T) {
	value := config.Default()
	_, err := SetValue(&value, "theme2", "paper")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "theme") {
		t.Fatalf("the error should list the known keys, got %q", err)
	}
}

func TestRenderPrintsEveryKeyWithItsValue(t *testing.T) {
	value := config.Default()
	value.Theme = config.ThemePaper
	var out strings.Builder
	if err := Render(&out, value, i18n.LangEn, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, field := range Fields() {
		if !strings.Contains(text, field.Key) {
			t.Fatalf("%q is missing from the table:\n%s", field.Key, text)
		}
	}
	if !strings.Contains(text, config.ThemePaper) {
		t.Fatalf("the current theme is missing from the table:\n%s", text)
	}
	if strings.Contains(text, "设置") {
		t.Fatalf("the English table contains Chinese labels:\n%s", text)
	}
}

func TestRenderSpeaksTheRequestedLanguage(t *testing.T) {
	var chinese strings.Builder
	if err := Render(&chinese, config.Default(), i18n.LangZhCN, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(chinese.String(), i18n.T(i18n.LangZhCN, i18n.KeySettingsNotify)) {
		t.Fatalf("the Chinese table does not use the Chinese label:\n%s", chinese.String())
	}
}

// scripted drives Run with a fixed list of answers and records what it saved.
func scripted(t *testing.T, answers string) (config.Config, bool, string) {
	t.Helper()
	saved := false
	var kept config.Config
	var out strings.Builder
	value := config.Default()
	err := Run(strings.NewReader(answers), &out, value, i18n.LangEn, "1.2.3", func(next config.Config) error {
		saved = true
		kept = next
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return kept, saved, out.String()
}

func TestRunSavesWhatWasEdited(t *testing.T) {
	kept, saved, out := scripted(t, "1\noff\ns\n")
	if !saved {
		t.Fatalf("nothing was saved:\n%s", out)
	}
	if kept.Notifications {
		t.Fatal("notifications was still on after being set off")
	}
	if !strings.Contains(out, "notifications = off") || !strings.Contains(out, "saved") {
		t.Fatalf("the transcript does not say what happened:\n%s", out)
	}
}

func TestRunQuitsWithoutSaving(t *testing.T) {
	kept, saved, out := scripted(t, "1\noff\nq\n")
	if saved {
		t.Fatal("q saved the file")
	}
	if !reflect.DeepEqual(kept, config.Config{}) {
		t.Fatalf("nothing should have been handed to save, got %+v", kept)
	}
	if !strings.Contains(out, "not written") {
		t.Fatalf("quitting after an edit must say the edit was not written:\n%s", out)
	}
}

func TestRunKeepsGoingAfterABadValue(t *testing.T) {
	kept, saved, out := scripted(t, "2\nmaybe\n2\noff\ns\n")
	if !saved {
		t.Fatalf("the second, valid answer should still be saved:\n%s", out)
	}
	if kept.Sound {
		t.Fatal("sound is still on")
	}
	if !strings.Contains(out, "not applied") {
		t.Fatalf("the invalid answer must be reported:\n%s", out)
	}
}

func TestRunWithoutEditsSavesNothing(t *testing.T) {
	_, saved, out := scripted(t, "s\n")
	if saved {
		t.Fatal("s with no edit must not rewrite the file")
	}
	if !strings.Contains(out, "nothing changed") {
		t.Fatalf("expected 'nothing changed':\n%s", out)
	}
}

func TestRunQuitsAtEOFWithoutSaving(t *testing.T) {
	_, saved, _ := scripted(t, "1\noff\n")
	if saved {
		t.Fatal("EOF closed the editor and saved")
	}
}

func TestClearMutesReportsWhatItRemoved(t *testing.T) {
	value := config.Default()
	value.MutedSessions = []string{"claude-hook|a", "opencode|b"}
	if removed := ClearMutes(&value); removed != 2 {
		t.Fatalf("removed %d, want 2", removed)
	}
	if len(value.MutedSessions) != 0 {
		t.Fatalf("the list still has %d entries", len(value.MutedSessions))
	}
}

// TestRenderPrintsTheSupportLinks is the terminal half of the settings footer:
// the same two links the Win32 window draws, printed instead of clickable, plus
// the build string a bug report needs.
func TestRenderPrintsTheSupportLinks(t *testing.T) {
	var out strings.Builder
	if err := Render(&out, config.Default(), i18n.LangEn, "9.9.9"); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{project.RepoURL, project.IssuesURL, "v9.9.9"} {
		if !strings.Contains(text, want) {
			t.Fatalf("%q is missing from the table:\n%s", want, text)
		}
	}
	// No version supplied means no line, not a bare "v".
	var bare strings.Builder
	if err := Render(&bare, config.Default(), i18n.LangEn, ""); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(bare.String(), "\nv\n") {
		t.Fatalf("an empty version printed a bare v:\n%s", bare.String())
	}
}
