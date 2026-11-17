package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/home"
	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/settingscli"
)

// The terminal settings surface. Two commands share one package:
//
//	--settings            show every option and its value; edit them when a
//	                      terminal is attached
//	--set key=value ...   change one or more values and exit, scriptable
//
// They exist because Windows has a settings window and Linux/macOS do not
// (internal/app.SettingsWindowAvailable), and because "edit config.json by hand"
// fails silently: config.Normalize turns an unrecognised value back into the
// default, so a typo looks like a preference that refuses to stick.
//
// A running panel picks up either change on its own: internal/app/run.go stats
// config.json every 30 s and hot-reloads it.

func settingsConfigPath() (string, error) {
	directory := home.Directory()
	if directory == "" {
		return "", errors.New("this platform has no data directory, so there is no config.json to edit")
	}
	return home.Join("config.json"), nil
}

// settingsLanguage decides which language the editor speaks. It follows the
// stored preference, so someone who pinned Chinese is not handed an English
// table by the thing that lets them change the pin.
func settingsLanguage(value config.Config) i18n.Lang {
	return i18n.Resolve(i18n.NormalizeSetting(value.Language), i18n.SystemLocale())
}

// controlSettingsTerminal implements --settings where there is no settings
// window. With a terminal attached it edits; piped or redirected it prints the
// table and exits, so `ai-terminal-pet --settings | less` and CI both work.
//
// It deliberately does not start the panel: asking for settings is not asking
// for an overlay to appear, and on these platforms the panel would have no
// window to put them in.
func controlSettingsTerminal(extra []string) error {
	if len(extra) > 0 {
		return usageError("--settings takes no arguments (%s)", strings.Join(extra, " "))
	}
	path, err := settingsConfigPath()
	if err != nil {
		return err
	}
	value := config.Load(path)
	lang := settingsLanguage(value)
	fmt.Fprintf(os.Stdout, "%s\n\n", path)
	if !settingscli.IsTerminal(os.Stdin) {
		return settingscli.Render(os.Stdout, value, lang)
	}
	return settingscli.Run(os.Stdin, os.Stdout, value, lang, func(next config.Config) error {
		return config.Save(path, next)
	})
}

// controlSet applies one or more key=value pairs and writes the file once.
// Every pair is validated before anything is written, so a typo in the third
// argument cannot leave the first two applied.
func controlSet(command string, extra []string) error {
	pairs := make([]string, 0, len(extra)+1)
	if inline, ok := strings.CutPrefix(command, "--set="); ok {
		pairs = append(pairs, inline)
	}
	pairs = append(pairs, extra...)
	if len(pairs) == 0 {
		return usageError("--set needs at least one key=value (try: --set theme=paper)")
	}
	path, err := settingsConfigPath()
	if err != nil {
		return err
	}
	value := config.Load(path)
	applied := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		key, raw, ok := strings.Cut(pair, "=")
		if !ok {
			return usageError("--set expects key=value, got %q", pair)
		}
		key = strings.TrimSpace(key)
		if key == "mutedSessions" {
			// The only list-valued key, and the one place a "set the whole list"
			// field would be a way to lose somebody's mutes. Clearing is the
			// settings window's own action; single rows are muted from the panel.
			if !strings.EqualFold(strings.TrimSpace(raw), "clear") {
				return usageError("--set mutedSessions=clear is the only form; to mute one session, right-click its row in the panel")
			}
			removed := settingscli.ClearMutes(&value)
			applied = append(applied, fmt.Sprintf("mutedSessions cleared (%d entries)", removed))
			continue
		}
		stored, err := settingscli.SetValue(&value, key, raw)
		if err != nil {
			return usageError("%v", err)
		}
		applied = append(applied, fmt.Sprintf("%s = %s", key, stored))
	}
	if err := config.Save(path, value); err != nil {
		return err
	}
	for _, line := range applied {
		fmt.Fprintln(os.Stdout, line)
	}
	fmt.Fprintf(os.Stdout, "saved to %s\n", path)
	return nil
}
