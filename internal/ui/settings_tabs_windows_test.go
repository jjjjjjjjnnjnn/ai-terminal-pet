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
func TestSettingsTextSizeChangesTheWindow(t *testing.T) {
	for _, dpi := range []uint32{96, 144} {
		small := settingsGeometrySized(dpi, textSizeSmall, 2, nil)
		standard := settingsGeometrySized(dpi, textSizeStandard, 2, nil)
		large := settingsGeometrySized(dpi, textSizeLarge, 2, nil)
		if !(small.width < standard.width && standard.width < large.width) {
			t.Fatalf("@%d: widths are %d / %d / %d", dpi, small.width, standard.width, large.width)
		}
		if !(small.height < standard.height && standard.height < large.height) {
			t.Fatalf("@%d: heights are %d / %d / %d", dpi, small.height, standard.height, large.height)
		}
	}
}

// TestSettingsFittingShrinksForAShortScreen: the text size is a preference and
// the work area is a constraint. When they collide the window steps down rather
// than growing past the screen edge, because a control nobody can reach is worse
// than text one size smaller.
func TestSettingsFittingShrinksForAShortScreen(t *testing.T) {
	rows := settingsRowsMeasuredFor(t, 2, i18n.LangZhCN, textSizeLarge)
	roomy := settingsGeometryFitting(96, textSizeLarge, 2, rows, 2000)
	if roomy.textSize != textSizeLarge {
		t.Fatalf("a roomy screen should keep the chosen size, got %q", roomy.textSize)
	}
	for _, limit := range []int{800, 700, 600} {
		fitted := settingsGeometryFitting(96, textSizeLarge, 2, rows, limit)
		if fitted.height > limit && fitted.textSize != textSizeSmall {
			t.Fatalf("limit %d: kept %q at height %d", limit, fitted.textSize, fitted.height)
		}
	}
	tight := settingsGeometryFitting(96, textSizeLarge, 2, rows, 400)
	if tight.textSize != textSizeSmall {
		t.Fatalf("a screen nothing fits should still fall back to the smallest size, got %q", tight.textSize)
	}
}

// TestSettingsHintsNeverLoseTextSilently: a hint may be wrapped, and may end in
// an ellipsis only on its last allowed line. A hint that was cut off earlier -
// or a reserved height that does not match the drawn lines - is the failure this
// catches.
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
func TestSettingsFocusOrderIsComplete(t *testing.T) {
	for tab := 0; tab < settingsTabCount; tab++ {
		rows := settingsRowsMeasuredFor(t, tab, i18n.LangZhCN, textSizeStandard)
		geometry := settingsGeometryFor(96, tab, rows)
		seen := map[settingsControl]int{}
		for _, control := range settingsFocusOrder(geometry) {
			seen[control]++
		}
		if seen[settingsClose] != 1 {
			t.Fatalf("tab %d: close button appears %d times", tab, seen[settingsClose])
		}
		for index := 0; index < settingsTabCount; index++ {
			if seen[settingsTabControl(index)] != 1 {
				t.Fatalf("tab %d: sidebar tab %d appears %d times", tab, index, seen[settingsTabControl(index)])
			}
		}
		for _, row := range rows {
			if row.control == settingsNone || row.kind == rowNote {
				continue
			}
			if seen[row.control] != 1 {
				t.Fatalf("tab %d: control %d appears %d times in the focus order", tab, row.control, seen[row.control])
			}
		}
	}
}

// TestSettingsKeyboardIsOptIn: Tab is swallowed only when the user asked for it.
func TestSettingsKeyboardIsOptIn(t *testing.T) {
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
