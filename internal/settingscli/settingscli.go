// Package settingscli is the settings surface for platforms without a settings
// window.
//
// Windows has a four-tab window (internal/ui/settings_windows.go). Linux and
// macOS do not: a toolkit there would mean cgo, and docs/ROADMAP.md §5 rules out
// a "degraded" port, so until now `--settings` on those platforms only started
// the panel and printed guidance. That left one real gap: every option was
// reachable only by hand-editing config.json, and a typo in that file is how a
// preference silently stops existing (config.Normalize rewrites what it cannot
// parse, so a misspelled value comes back as the default with no complaint).
//
// This package is the replacement, and it is deliberately a *terminal* surface:
// no raw mode, no alternate screen, no framework. It answers three questions:
//
//   - what can I set, and what is it now?      (Render)
//   - set exactly this one thing, and fail loudly on a typo  (SetValue)
//   - walk me through the list                 (Run)
//
// Everything except the last one is pure, which is what lets the interesting
// parts be tested without a terminal.
package settingscli

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/project"
)

// Kind is how a field is read and written from one line of text.
type Kind int

const (
	// Toggle is a boolean: on/off, true/false, yes/no, 1/0.
	Toggle Kind = iota
	// Choice is one of a fixed list, by value or by its 1-based index.
	Choice
	// Number is an integer inside [Min, Max].
	Number
)

// Field is one editable key. Label and Hint are i18n keys, not sentences: the
// labels here deliberately reuse the Win32 settings window's own strings, so the
// two surfaces cannot call the same option two different things.
type Field struct {
	Key      string
	Label    string
	Hint     string
	Kind     Kind
	Choices  []string
	Min, Max int
	Get      func(config.Config) string
	Set      func(*config.Config, string) error
}

// Fields is the complete, ordered list. It is a function, not a package-level
// variable, because callers may sort or annotate the result.
//
// `mutedSessions` is deliberately absent: it is a list of per-session keys that
// only makes sense next to the rows that created them, and a "set the whole list
// by hand" field would be a way to lose somebody's mutes. The settings window's
// one useful action on it - clear them all - is offered as ClearMutes below.
func Fields() []Field {
	onOff := func(get func(config.Config) bool, set func(*config.Config, bool)) (func(config.Config) string, func(*config.Config, string) error) {
		return func(value config.Config) string { return FormatBool(get(value)) },
			func(value *config.Config, raw string) error {
				parsed, err := ParseBool(raw)
				if err != nil {
					return err
				}
				set(value, parsed)
				return nil
			}
	}
	toggle := func(key, label, hint string, get func(config.Config) bool, set func(*config.Config, bool)) Field {
		read, write := onOff(get, set)
		return Field{Key: key, Label: label, Hint: hint, Kind: Toggle, Get: read, Set: write}
	}
	number := func(key, label, hint string, minimum, maximum int, get func(config.Config) int, set func(*config.Config, int)) Field {
		return Field{
			Key: key, Label: label, Hint: hint, Kind: Number, Min: minimum, Max: maximum,
			Get: func(value config.Config) string { return strconv.Itoa(get(value)) },
			Set: func(value *config.Config, raw string) error {
				parsed, err := ParseNumber(raw, minimum, maximum)
				if err != nil {
					return err
				}
				set(value, parsed)
				return nil
			},
		}
	}
	choice := func(key, label, hint string, choices []string, get func(config.Config) string, set func(*config.Config, string)) Field {
		return Field{
			Key: key, Label: label, Hint: hint, Kind: Choice, Choices: choices,
			Get: get,
			Set: func(value *config.Config, raw string) error {
				parsed, err := matchChoice(raw, choices)
				if err != nil {
					return err
				}
				set(value, parsed)
				return nil
			},
		}
	}

	return []Field{
		toggle("notifications", i18n.KeySettingsNotify, i18n.KeySettingsNotifyHint,
			func(value config.Config) bool { return value.Notifications },
			func(value *config.Config, on bool) { value.Notifications = on }),
		toggle("sound", i18n.KeySettingsSound, i18n.KeySettingsSoundHint,
			func(value config.Config) bool { return value.Sound },
			func(value *config.Config, on bool) { value.Sound = on }),
		number("soundVolume", i18n.KeySettingsVolume, i18n.KeySettingsVolumeHint, 0, 100,
			func(value config.Config) int { return value.SoundVolume },
			func(value *config.Config, parsed int) { value.SoundVolume = parsed }),
		choice("soundScheme", i18n.KeySettingsScheme, i18n.KeySettingsSchemeHint, config.SoundSchemes,
			func(value config.Config) string { return value.SoundScheme },
			func(value *config.Config, parsed string) { value.SoundScheme = parsed }),
		toggle("soundOnWaiting", i18n.KeySettingsSoundWaiting, i18n.KeySettingsSoundWaitingHint,
			func(value config.Config) bool { return value.SoundOnWaiting },
			func(value *config.Config, on bool) { value.SoundOnWaiting = on }),
		toggle("soundOnError", i18n.KeySettingsSoundError, i18n.KeySettingsSoundErrorHint,
			func(value config.Config) bool { return value.SoundOnError },
			func(value *config.Config, on bool) { value.SoundOnError = on }),
		toggle("notifyOnFinish", i18n.KeySettingsSoundFinish, i18n.KeySettingsSoundFinishHint,
			func(value config.Config) bool { return value.NotifyOnFinish },
			func(value *config.Config, on bool) { value.NotifyOnFinish = on }),
		choice("donePolicy", i18n.KeySettingsDonePolicy, i18n.KeySettingsDonePolicyHint,
			config.DonePolicies,
			func(value config.Config) string { return value.DonePolicy },
			func(value *config.Config, parsed string) { value.DonePolicy = parsed }),
		toggle("autoApprove", i18n.KeySettingsAutoApprove, i18n.KeySettingsAutoApproveHint,
			func(value config.Config) bool { return value.AutoApprove },
			func(value *config.Config, on bool) { value.AutoApprove = on }),
		number("autoApproveTimeoutSec", i18n.KeySettingsApproveTimeout, i18n.KeySettingsApproveTimeoutHint,
			config.MinAutoApproveTimeoutSec, config.MaxAutoApproveTimeoutSec,
			func(value config.Config) int { return value.AutoApproveTimeoutSec },
			func(value *config.Config, parsed int) { value.AutoApproveTimeoutSec = parsed }),
		toggle("alwaysOnTop", i18n.KeySettingsTopmost, i18n.KeySettingsTopmostHint,
			func(value config.Config) bool { return value.AlwaysOnTop },
			func(value *config.Config, on bool) { value.AlwaysOnTop = on }),
		toggle("edgeRetract", i18n.KeySettingsRetract, i18n.KeySettingsRetractHint,
			func(value config.Config) bool { return value.EdgeRetract },
			func(value *config.Config, on bool) { value.EdgeRetract = on }),
		number("opacity", i18n.KeySettingsOpacity, i18n.KeySettingsOpacityHint, config.MinOpacity, config.MaxOpacity,
			func(value config.Config) int { return value.Opacity },
			func(value *config.Config, parsed int) { value.Opacity = parsed }),
		choice("theme", i18n.KeySettingsTheme, i18n.KeySettingsThemeHint, config.Themes,
			func(value config.Config) string { return value.Theme },
			func(value *config.Config, parsed string) { value.Theme = parsed }),
		toggle("perSourceColor", i18n.KeySettingsPerSource, i18n.KeySettingsPerSourceHint,
			func(value config.Config) bool { return value.PerSourceColor },
			func(value *config.Config, on bool) { value.PerSourceColor = on }),
		toggle("hiddenAtLaunch", i18n.KeySettingsHiddenAtLaunch, i18n.KeySettingsHiddenAtLaunchHint,
			func(value config.Config) bool { return value.HiddenAtLaunch },
			func(value *config.Config, on bool) { value.HiddenAtLaunch = on }),
		toggle("foregroundAware", i18n.KeySettingsForegroundAware, i18n.KeySettingsForegroundAwareHint,
			func(value config.Config) bool { return value.ForegroundAware },
			func(value *config.Config, on bool) { value.ForegroundAware = on }),
		number("idleMinutes", i18n.KeySettingsIdleMinutes, i18n.KeySettingsIdleMinutesHint, 0, config.MaxIdleMinutes,
			func(value config.Config) int { return value.IdleMinutes },
			func(value *config.Config, parsed int) { value.IdleMinutes = parsed }),
		toggle("quietCurrent", i18n.KeySettingsQuietCurrent, i18n.KeySettingsQuietCurrentHint,
			func(value config.Config) bool { return value.QuietCurrent },
			func(value *config.Config, on bool) { value.QuietCurrent = on }),
		toggle("quietSubagents", i18n.KeySettingsQuietSubagents, i18n.KeySettingsQuietSubagentsHint,
			func(value config.Config) bool { return value.QuietSubagents },
			func(value *config.Config, on bool) { value.QuietSubagents = on }),
		choice("language", i18n.KeySettingsLanguage, i18n.KeySettingsLanguageHint, Languages,
			func(value config.Config) string { return value.Language },
			func(value *config.Config, parsed string) { value.Language = parsed }),
	}
}

// Languages are the pinned display languages; "auto" follows the host locale.
var Languages = []string{"auto", string(i18n.LangEn), string(i18n.LangZhCN)}

// FormatBool is how a toggle is printed, so --settings, the editor and the
// confirmation line after --set all say "on" or "off".
func FormatBool(value bool) string {
	if value {
		return "on"
	}
	return "off"
}

// ParseBool accepts the spellings a person actually types. It is strict on
// purpose: "maybe" is a typo, and guessing at it is how a preference silently
// becomes the default.
func ParseBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "on", "true", "yes", "y", "1", "enable", "enabled":
		return true, nil
	case "off", "false", "no", "n", "0", "disable", "disabled":
		return false, nil
	}
	return false, fmt.Errorf("%q is not on or off", strings.TrimSpace(raw))
}

// ParseNumber reads an integer and refuses anything outside [minimum, maximum]
// rather than clamping: a value the user did not choose is not a value they can
// trust to still be there next launch.
func ParseNumber(raw string, minimum, maximum int) (int, error) {
	trimmed := strings.TrimSpace(raw)
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", trimmed)
	}
	if parsed < minimum || parsed > maximum {
		return 0, fmt.Errorf("%d is outside %d..%d", parsed, minimum, maximum)
	}
	return parsed, nil
}

// matchChoice accepts a value or its 1-based index, case-insensitively.
func matchChoice(raw string, choices []string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if index, err := strconv.Atoi(trimmed); err == nil {
		if index < 1 || index > len(choices) {
			return "", fmt.Errorf("%d is not one of 1..%d", index, len(choices))
		}
		return choices[index-1], nil
	}
	for _, choice := range choices {
		if strings.EqualFold(choice, trimmed) {
			return choice, nil
		}
	}
	return "", fmt.Errorf("%q is not one of %s", trimmed, strings.Join(choices, ", "))
}

// SetValue applies one "key=value" pair to a config. It answers with the value
// as stored (normalised, e.g. "PAPER" becomes "paper"), which is what the caller
// prints back so the user sees exactly what landed.
func SetValue(value *config.Config, key, raw string) (string, error) {
	fields := Fields()
	for _, field := range fields {
		if field.Key != key {
			continue
		}
		if err := field.Set(value, raw); err != nil {
			return "", fmt.Errorf("%s: %w", key, err)
		}
		return field.Get(*value), nil
	}
	return "", fmt.Errorf("unknown key %q (known keys: %s)", key, strings.Join(Keys(fields), ", "))
}

// Keys lists the editable keys in table order.
func Keys(fields []Field) []string {
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		keys = append(keys, field.Key)
	}
	return keys
}

// ClearMutes empties the per-session mute list, the one action the settings
// window offers for it. It returns how many entries were dropped so the caller
// can say something truthful about what happened.
func ClearMutes(value *config.Config) int {
	removed := len(value.MutedSessions)
	value.MutedSessions = nil
	return removed
}

// Render prints the whole table: one line per key, its current value, and a
// short localised label. It is the non-interactive form of --settings and the
// thing the interactive editor redraws.
func Render(w io.Writer, value config.Config, lang i18n.Lang, version string) error {
	fields := Fields()
	width := 0
	labels := make([]string, len(fields))
	for index, field := range fields {
		labels[index] = i18n.T(lang, field.Label)
		if len(labels[index]) > width {
			width = len(labels[index])
		}
	}
	keyWidth := 0
	for _, field := range fields {
		if len(field.Key) > keyWidth {
			keyWidth = len(field.Key)
		}
	}
	if _, err := fmt.Fprintf(w, "%s\n\n", i18n.T(lang, i18n.KeySettingsTitle)); err != nil {
		return err
	}
	for index, field := range fields {
		line := fmt.Sprintf("  %2d  %-*s  %-*s  %s\n",
			index+1, keyWidth, field.Key, width, labels[index], field.Get(value))
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	// The same two links the Win32 window draws in its footer, printed rather
	// than clickable: a terminal can be piped, and launching a browser from a
	// program that was asked for a table would be a surprise.
	if _, err := fmt.Fprintf(w, "\n%s  %s\n%s  %s\n",
		i18n.T(lang, i18n.KeySettingsStar), project.RepoURL,
		i18n.T(lang, i18n.KeySettingsIssue), project.IssuesURL); err != nil {
		return err
	}
	if version != "" {
		if _, err := fmt.Fprintf(w, "v%s\n", version); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\n%s\n", helpLine(lang))
	return err
}

func helpLine(lang i18n.Lang) string {
	if lang == i18n.LangZhCN {
		return "用 --set key=value 改一个值（例：--set theme=paper）；没有设置窗口时这是唯一入口，运行中的面板 30 秒内会重新读配置。"
	}
	return "Use --set key=value to change one value (e.g. --set theme=paper). With no settings window this is the only entry point; a running panel re-reads the file within 30 s."
}

// Run is the interactive editor. It reads one line at a time - a field number,
// "s" to save, "q" to quit, or "?" to redraw - so it needs no terminal
// capabilities and can be driven by a script in a test.
//
// Saving is the caller's job: this package does not know where config.json is,
// which is also why it can be tested without touching a disk.
func Run(in io.Reader, out io.Writer, value config.Config, lang i18n.Lang, version string, save func(config.Config) error) error {
	reader := bufio.NewReader(in)
	if err := Render(out, value, lang, version); err != nil {
		return err
	}
	changed := false
	for {
		if _, err := fmt.Fprint(out, "\n> "); err != nil {
			return err
		}
		line, err := reader.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			// EOF (or a broken pipe) without input: quit without saving, which
			// is the same answer a user gives by closing the terminal.
			fmt.Fprintln(out)
			return nil
		}
		answer := strings.TrimSpace(line)
		switch {
		case answer == "":
			continue
		case strings.EqualFold(answer, "s"):
			if !changed {
				fmt.Fprintln(out, "nothing changed")
				return nil
			}
			if err := save(value.Normalize()); err != nil {
				return err
			}
			fmt.Fprintln(out, "saved")
			return nil
		case strings.EqualFold(answer, "q"):
			if changed {
				fmt.Fprintln(out, "quit without saving: the edits above were not written")
			} else {
				fmt.Fprintln(out, "quit, nothing changed")
			}
			return nil
		case answer == "?":
			if err := Render(out, value, lang, version); err != nil {
				return err
			}
			continue
		}
		index, convErr := strconv.Atoi(answer)
		fields := Fields()
		if convErr != nil || index < 1 || index > len(fields) {
			fmt.Fprintf(out, "%q is not a number from 1 to %d (s saves, q quits, ? redraws)\n", answer, len(fields))
			continue
		}
		applied, err := promptField(reader, out, &value, fields[index-1], lang)
		if err != nil {
			return err
		}
		if applied {
			changed = true
		}
	}
}

func promptField(reader *bufio.Reader, out io.Writer, value *config.Config, field Field, lang i18n.Lang) (bool, error) {
	label := i18n.T(lang, field.Label)
	if _, err := fmt.Fprintf(out, "%s [%s] (now %s): ", label, choicesFor(field), field.Get(*value)); err != nil {
		return false, err
	}
	raw, err := reader.ReadString('\n')
	if err != nil && strings.TrimSpace(raw) == "" {
		return false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		fmt.Fprintln(out, "unchanged")
		return false, nil
	}
	if _, err := SetValue(value, field.Key, raw); err != nil {
		fmt.Fprintf(out, "not applied: %v\n", err)
		return false, nil
	}
	fmt.Fprintf(out, "%s = %s\n", field.Key, field.Get(*value))
	return true, nil
}

func choicesFor(field Field) string {
	switch field.Kind {
	case Toggle:
		return "on/off"
	case Number:
		return fmt.Sprintf("%d..%d", field.Min, field.Max)
	case Choice:
		quoted := make([]string, len(field.Choices))
		for index, choice := range field.Choices {
			quoted[index] = fmt.Sprintf("%d=%s", index+1, choice)
		}
		return strings.Join(quoted, " ")
	}
	return ""
}

// SortedKeys is the stable list the tools and tests use; the table order is the
// product's own, so it is only for messages that must not depend on it.
func SortedKeys(fields []Field) []string {
	keys := Keys(fields)
	sort.Strings(keys)
	return keys
}
