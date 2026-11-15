package ui

// The settings window's own behaviour, pinned.
//
// Two of these exist because of bugs that shipped: a sidebar tab that hit-tested
// and hovered but did nothing when clicked (settingsTabAppearance was missing
// from the switch), and a type ramp whose three steps rendered as two sizes
// because GDI snaps a requested cell height onto the face's own ladder.

import (
	"strings"
	"testing"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
)

// TestSettingsEveryTabSwitches is the guard the old hit-test could not be: it
// asks what happens *after* the click, not which control was under the cursor.
func TestSettingsEveryTabSwitches(t *testing.T) {
	for index := 0; index < settingsTabCount; index++ {
		app := &App{lang: i18n.LangZhCN}
		app.settingsTab = (index + 1) % settingsTabCount
		app.settingsApply(settingsTabControl(index), -1)
		if app.settingsTab != index {
			t.Fatalf("clicking sidebar tab %d left settingsTab at %d", index, app.settingsTab)
		}
	}
}

// TestSettingsTypeRampStaysOrdered measures the three ramps at four DPIs. The
// window's whole hierarchy is "the title is bigger than the row, which is bigger
// than the hint"; if two roles land on the same GDI rung that hierarchy is gone
// and the window reads as flat.

func TestSettingsTypeRampStaysOrdered(t *testing.T) {
	target, err := newCanvas(1, 1)
	if err != nil {
		t.Fatalf("canvas: %v", err)
	}
	defer target.close()

	sample := "状态提醒 Notifications"
	for _, name := range config.TextSizes {
		ramp := settingsTypeRamps[name]
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometrySized(dpi, name, 0, nil)
			width := func(size int) int { return target.measure(sample, geometry.font(size), 500) }
			if width(ramp.header) <= width(ramp.title) {
				t.Fatalf("%s @%d: header (%d) is not wider than the row title (%d)",
					name, dpi, width(ramp.header), width(ramp.title))
			}
			if width(ramp.title) <= width(ramp.hint) {
				t.Fatalf("%s @%d: row title (%d) is not wider than the hint (%d)",
					name, dpi, width(ramp.title), width(ramp.hint))
			}
			if width(ramp.hint) <= 0 || width(ramp.header) <= 0 {
				t.Fatalf("%s @%d: a role measures zero pixels", name, dpi)
			}
		}
	}
}

// TestSettingsTextSizeChangesTheWindow: the setting is only real if the layout
// follows it. Small < standard < large in both directions, at every DPI.

func TestSettingsHintsNeverLoseTextSilently(t *testing.T) {
	target, err := newCanvas(1, 1)
	if err != nil {
		t.Fatalf("canvas: %v", err)
	}
	defer target.close()

	for _, name := range config.TextSizes {
		ramp := settingsTypeRamps[name]
		for _, lang := range []i18n.Lang{i18n.LangZhCN, i18n.LangEn} {
			for tab := 0; tab < settingsTabCount; tab++ {
				rows := settingsRowsMeasuredFor(t, tab, lang, name)
				width := settingsGeometrySized(96, name, tab, nil).scale(settingsContentWidth)
				for _, row := range rows {
					// A note's paragraph lives in title; its hint is empty. It
					// is measured and drawn separately, and reserving one line
					// for a two-line note clips the second one - which is what
					// the first render of this window did.
					text, font := row.hint, ramp.hint
					if row.kind == rowNote {
						text, font = row.title, ramp.note
					}
					if text == "" {
						continue
					}
					lines := wrapHint(target, text, font, 400, width, settingsMaxLines(row.kind))
					if len(lines) != row.hintLines {
						t.Fatalf("%s/%s tab%d: %q reserves %d line(s) but draws %d",
							name, lang, tab, text, row.hintLines, len(lines))
					}
					ellipsised := strings.HasSuffix(lines[len(lines)-1], "…")
					if ellipsised && len(lines) != settingsMaxLines(row.kind) {
						t.Fatalf("%s/%s tab%d: %q was cut off on line %d of %d",
							name, lang, tab, text, len(lines), settingsMaxLines(row.kind))
					}
					// The drawn lines must still be the beginning of the
					// original text, in order.
					joined := strings.ReplaceAll(strings.Join(lines, ""), "…", "")
					if !strings.HasPrefix(strings.ReplaceAll(text, " ", ""), strings.ReplaceAll(joined, " ", "")) {
						t.Fatalf("%s/%s tab%d: wrapped text is not a prefix of %q (got %q)",
							name, lang, tab, text, joined)
					}
				}
			}
		}
	}
}

// TestSettingsFocusOrderIsComplete: Tab must reach every interactive control
// exactly once. A control missing from the order is unreachable by keyboard; one
// appearing twice makes Tab feel broken.

func TestSettingsKeyboardIsOptIn(t *testing.T) {
	// Pin the modifier: the focus walk reads Shift from the real keyboard, and a
	// test that silently walked backwards while a human held Shift failed here
	// once (focus index 12, visible true). The production reading is unchanged.
	previous := settingsShiftDown
	settingsShiftDown = func() bool { return false }
	defer func() { settingsShiftDown = previous }()

	off := &App{lang: i18n.LangZhCN}
	off.settingsGeometry = settingsGeometryFor(96, 0, settingsRowsMeasuredFor(t, 0, i18n.LangZhCN, textSizeStandard))
	if off.handleSettingsKey(wmKeyTab, 0) {
		t.Fatal("Tab was consumed although keyboard navigation is off")
	}
	if off.settingsFocus != 0 {
		t.Fatal("focus moved although keyboard navigation is off")
	}

	on := &App{lang: i18n.LangZhCN}
	on.options.Settings.SettingsKeyboard = true
	on.settingsGeometry = settingsGeometryFor(96, 0, settingsRowsMeasuredFor(t, 0, i18n.LangZhCN, textSizeStandard))
	if !on.handleSettingsKey(wmKeyTab, 0) {
		t.Fatal("Tab was not consumed although keyboard navigation is on")
	}
	if on.settingsFocus != 1 || !on.settingsFocusVisible {
		t.Fatalf("focus index %d, visible %v", on.settingsFocus, on.settingsFocusVisible)
	}
	if !on.handleSettingsKey(wmKeySpace, 0) {
		t.Fatal("Space was not consumed")
	}
}

// TestSettingsApplyPublishes: the appearance controls must reach the
// preferences callback. The dead-tab bug was one instance of a wider class -
// a control that draws, hit-tests and then changes nothing - and the callback
// is the only place that can tell "changed the model" from "changed the model
// and told anyone".

func TestSettingsApplyPublishes(t *testing.T) {
	cases := []struct {
		name    string
		control settingsControl
		index   int
		check   func(Settings) bool
	}{
		{"theme", settingsTheme, 1, func(s Settings) bool { return s.Theme == "paper" }},
		{"text size", settingsTextSize, 2, func(s Settings) bool { return s.SettingsTextSize == textSizeLarge }},
		{"hidden at launch", settingsHiddenAtLaunch, -1, func(s Settings) bool { return s.HiddenAtLaunch }},
		{"keyboard", settingsKeyboard, -1, func(s Settings) bool { return s.SettingsKeyboard }},
		{"tooltips", settingsTooltips, -1, func(s Settings) bool { return !s.SettingsTooltips }},
		{"per-source colour", settingsPerSource, -1, func(s Settings) bool { return s.PerSourceColor }},
	}
	for _, tc := range cases {
		published := 0
		app := &App{lang: i18n.LangZhCN}
		app.options.Settings = DefaultSettings()
		app.options.OnSettingsChanged = func(value Settings) {
			published++
			if !tc.check(value) {
				t.Fatalf("%s: published an unchanged value %+v", tc.name, value)
			}
		}
		app.settingsGeometry = settingsGeometryFor(96, 3, settingsRowsMeasuredFor(t, 3, i18n.LangZhCN, textSizeStandard))
		app.settingsApply(tc.control, tc.index)
		if published != 1 {
			t.Fatalf("%s: the preference callback ran %d times, want 1", tc.name, published)
		}
	}
}

// TestWrapHintKeepsTextTogether: the wrapper must not drop or reorder
// characters, and must not invent line breaks inside a Latin word it could have
// kept whole.

func TestWrapHintKeepsTextTogether(t *testing.T) {
	target, err := newCanvas(1, 1)
	if err != nil {
		t.Fatalf("canvas: %v", err)
	}
	defer target.close()

	cases := []struct {
		value string
		width int
	}{
		{"A prompt you are reading does not also pop a balloon", 200},
		{"正在阅读的请求不再额外弹气泡，但是那一行仍会标为未读", 120},
		{"supercalifragilisticexpialidocious", 40},
		{"", 100},
	}
	for _, tc := range cases {
		lines := wrapHint(target, tc.value, 14, 400, tc.width, 2)
		if tc.value == "" {
			if len(lines) != 0 {
				t.Fatalf("empty text produced %v", lines)
			}
			continue
		}
		if len(lines) == 0 || len(lines) > 2 {
			t.Fatalf("%q produced %d line(s)", tc.value, len(lines))
		}
		for _, line := range lines {
			if target.measure(line, 14, 400) > tc.width && !strings.HasSuffix(line, "…") {
				t.Fatalf("line %q is wider than %d", line, tc.width)
			}
		}
	}
}
