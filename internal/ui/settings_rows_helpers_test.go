package ui

// Row-list fixtures shared by the settings tests, on every platform.
//
// They build the production row lists and then MEASURE them, which is the only
// way a layout test can exercise real hint heights: wrapping depends on a text
// measurer, and the canvas is the one the window uses. Moving them out of
// settings_windows_test.go is what lets the layout tests run off Windows.

import (
	"testing"

	"ai-terminal-pet/internal/i18n"
)

func settingsRowsFor(t *testing.T, tab int) []settingsRow {
	t.Helper()
	return settingsRowsIn(t, tab, i18n.LangZhCN)
}

func settingsRowsIn(t *testing.T, tab int, lang i18n.Lang) []settingsRow {
	return settingsRowsMeasuredFor(t, tab, lang, textSizeStandard)
}

// settingsRowsMeasuredFor is the production path: build the rows, then measure
// how many lines each hint needs at this text size and column width.
func settingsRowsMeasuredFor(t *testing.T, tab int, lang i18n.Lang, textSize string) []settingsRow {
	t.Helper()
	var rows []settingsRow
	switch tab {
	case 1:
		rows = soundRows(lang)
	case 2:
		rows = sessionRows(lang, 0)
	case 3:
		rows = appearanceRows(lang)
	default:
		rows = generalRows(lang)
	}
	target, err := newCanvas(1, 1)
	if err != nil {
		t.Fatalf("measure canvas: %v", err)
	}
	defer target.close()
	probe := settingsGeometrySized(96, textSize, tab, nil)
	return measureSettingsRows(target, rows, settingsTypeRamps[textSize], probe.scale(settingsContentWidth))
}
