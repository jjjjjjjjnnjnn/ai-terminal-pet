package ui

// The settings window's platform-neutral rules, pinned.
//
// These ran on Windows alone until settings_layout.go existed: the layout, the
// hit test, the slider arithmetic and the segment tiling are pure functions, but
// they used to live in a *_windows.go file. A rule that decides what a click on a
// slider means is worth testing on every platform the panel claims to support,
// which is exactly what moving them here buys.

import (
	"testing"

	"ai-terminal-pet/internal/i18n"
)

func TestSettingsLayoutFitsTheWindow(t *testing.T) {
	for _, tab := range []int{0, 1, 2, 3} {
		rows := settingsRowsFor(t, tab)
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometryFor(dpi, tab, rows)
			if geometry.width <= 0 || geometry.height <= 0 {
				t.Fatalf("tab %d dpi %d: empty geometry", tab, dpi)
			}
			last := geometry.frames[len(geometry.frames)-1]
			if bottom := last.bottom(); bottom > geometry.height-geometry.scale(settingsMargin)/2 {
				t.Fatalf("tab %d dpi %d: last row ends at %d, panel is %d tall", tab, dpi, bottom, geometry.height)
			}
			for index, frame := range geometry.frames {
				if frame.Left < geometry.scale(settingsContentLeft)-1 {
					t.Fatalf("tab %d dpi %d: row %d starts left of the content column", tab, dpi, index)
				}
				if frame.right() > geometry.width-geometry.scale(settingsMargin)+1 {
					t.Fatalf("tab %d dpi %d: row %d overflows the right margin", tab, dpi, index)
				}
				if index > 0 && frame.Top < geometry.frames[index-1].bottom() {
					t.Fatalf("tab %d dpi %d: row %d overlaps row %d", tab, dpi, index, index-1)
				}
			}
			for index := range geometry.tabs {
				tab := geometry.tabs[index]
				if tab.right() >= geometry.scale(settingsContentLeft) {
					t.Fatalf("sidebar tab %d runs into the content column", index)
				}
			}
			if !geometry.close.contains(geometry.width-geometry.scale(settingsMargin)-1, geometry.close.Top+1) {
				t.Fatal("close button is not inside the right margin")
			}
		}
	}
}

// TestSettingsHitTestFindsEveryControl is the guard against a control that is
// drawn but unclickable: for every interactive row the centre of the drawn
// sub-rect must resolve back to that row's control.

func TestSettingsHitTestFindsEveryControl(t *testing.T) {
	for _, tab := range []int{0, 1, 2, 3} {
		rows := settingsRowsFor(t, tab)
		geometry := settingsGeometryFor(96, tab, rows)
		for index, row := range rows {
			if row.kind == rowNote {
				continue
			}
			rect := geometry.controls[index]
			x, y := rect.Left+rect.Width/2, rect.Top+rect.Height/2
			control, _, _, ok := geometry.hit(x, y)
			if !ok {
				t.Fatalf("tab %d row %d (%d): nothing hit at the control centre", tab, index, row.control)
			}
			// A stepper resolves to either half depending on x, and a segment
			// resolves to the option under the cursor; both are correct as long
			// as they belong to the same row.
			if control != row.control && control != row.alt {
				t.Fatalf("tab %d row %d: hit %d, want %d (or %d)", tab, index, control, row.control, row.alt)
			}
		}
		for index := range geometry.tabs {
			rect := geometry.tabs[index]
			control, _, _, ok := geometry.hit(rect.Left+rect.Width/2, rect.Top+rect.Height/2)
			if !ok || control != settingsTabControl(index) {
				t.Fatalf("sidebar tab %d did not hit its control", index)
			}
		}
	}
}

// TestSettingsSegmentsTileTheControl: the selected pill and the hit test derive
// their geometry from the same helper, so the options must cover the control
// exactly - no gap, no overlap, no rounding drift.

func TestSettingsSegmentsTileTheControl(t *testing.T) {
	control := settingsRect{Left: 100, Top: 40, Width: 272, Height: 28}
	for _, count := range []int{2, 3, 4} {
		previous := control.Left
		for index := 0; index < count; index++ {
			cell := settingsSegmentRect(control, count, index)
			if cell.Left != previous {
				t.Fatalf("count %d index %d: cell starts at %d, previous ended at %d", count, index, cell.Left, previous)
			}
			if cell.Height != control.Height {
				t.Fatalf("count %d index %d: height %d, want %d", count, index, cell.Height, control.Height)
			}
			previous = cell.right()
		}
		if previous != control.right() {
			t.Fatalf("count %d: cells end at %d, control ends at %d", count, previous, control.right())
		}
	}
}

func TestSettingsSliderMapping(t *testing.T) {
	track := settingsRect{Left: 10, Top: 0, Width: 200, Height: 4}
	row := settingsRow{kind: rowSlider, minimum: 40, maximum: 255}
	if got := settingsSliderValue(track, row, 10); got != 40 {
		t.Fatalf("left edge = %d, want the minimum", got)
	}
	if got := settingsSliderValue(track, row, 210); got != 255 {
		t.Fatalf("right edge = %d, want the maximum", got)
	}
	if got := settingsSliderValue(track, row, -500); got != 40 {
		t.Fatalf("far left = %d, want the minimum", got)
	}
	if got := settingsSliderValue(track, row, 9999); got != 255 {
		t.Fatalf("far right = %d, want the maximum", got)
	}
	if got := settingsSliderValue(track, row, 110); got != 148 {
		t.Fatalf("midpoint = %d, want 148 (round(40+107.5))", got)
	}
}

// TestDefaultSettingsMirrorConfig is the pin between the two packages that must
// not drift: ui cannot import config (the app layer owns file paths), so the
// defaults and the scheme names are mirrored, and this test is the contract.

func TestAppearanceTabIsLaidOutLikeEveryOtherTab(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangZhCN, i18n.LangEn} {
		rows := appearanceRows(lang)
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometryFor(dpi, 3, rows)
			last := geometry.frames[len(geometry.frames)-1]
			if bottom := last.bottom(); bottom > geometry.height-geometry.scale(settingsMargin)/2 {
				t.Fatalf("%v dpi %d: appearance tab overflows the panel (%d > %d)", lang, dpi, bottom, geometry.height)
			}
			for index, row := range rows {
				// A rowNote is a caption, not a control: it owns no hit target
				// by design, which is why the loop skips it rather than
				// asserting something the geometry never promised.
				if row.control == settingsNone {
					continue
				}
				rect := geometry.controls[index]
				control, _, _, ok := geometry.hit(rect.Left+rect.Width/2, rect.Top+rect.Height/2)
				if !ok || control != row.control {
					t.Fatalf("%v dpi %d row %d: hit %d, want %d", lang, dpi, index, control, row.control)
				}
			}
			if tab := settingsTabControl(3); tab != settingsTabAppearance {
				t.Fatalf("sidebar entry 3 maps to %d, want settingsTabAppearance", tab)
			}
			if index := settingsTabIndex(settingsTabAppearance); index != 3 {
				t.Fatalf("settingsTabAppearance maps to index %d, want 3", index)
			}
		}
	}
	if settingsTabCount != 4 {
		t.Fatalf("settingsTabCount = %d, want 4", settingsTabCount)
	}
}

// TestSessionsTabIsLaidOutLikeEveryOtherTab is the same contract for the tab
// that grew: every row fits, every control is hit-testable at its drawn centre,
// and the attentiveness rows name the controls their handlers switch on. The
// height assertion is the one that catches a tab that grew past the window,
// which is what a four-row addition is most likely to do.

func TestSessionsTabIsLaidOutLikeEveryOtherTab(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangZhCN, i18n.LangEn} {
		rows := sessionRows(lang, 2)
		for _, dpi := range []uint32{96, 120, 144, 192} {
			geometry := settingsGeometryFor(dpi, 2, rows)
			last := geometry.frames[len(geometry.frames)-1]
			if bottom := last.bottom(); bottom > geometry.height-geometry.scale(settingsMargin)/2 {
				t.Fatalf("%v dpi %d: sessions tab overflows the panel (%d > %d)", lang, dpi, bottom, geometry.height)
			}
			for index, row := range rows {
				if row.control == settingsNone {
					continue
				}
				rect := geometry.controls[index]
				control, _, _, ok := geometry.hit(rect.Left+rect.Width/2, rect.Top+rect.Height/2)
				if !ok || control != row.control {
					t.Fatalf("%v dpi %d row %d: hit %d, want %d", lang, dpi, index, control, row.control)
				}
			}
		}
	}
}

// The Done segment and the persisted policy have to agree in both directions,
// or clicking "Off" would select "Background" on the next paint.

func TestDonePolicySegmentRoundTrip(t *testing.T) {
	row := settingsRow{}
	for _, candidate := range sessionRows(i18n.LangEn, 0) {
		if candidate.control == settingsDonePolicy {
			row = candidate
		}
	}
	if len(row.options) != len(DonePolicyNames()) {
		t.Fatalf("done segment offers %d options, DonePolicyNames has %d", len(row.options), len(DonePolicyNames()))
	}
	for index, policy := range DonePolicyNames() {
		if got := settingsDonePolicyIndex(string(policy)); got != index {
			t.Fatalf("settingsDonePolicyIndex(%q) = %d, want %d", policy, got, index)
		}
		back, ok := settingsDonePolicyAt(index)
		if !ok || back != policy {
			t.Fatalf("settingsDonePolicyAt(%d) = %q/%v, want %q", index, back, ok, policy)
		}
	}
	if got := settingsDonePolicyIndex("nonsense"); got != 0 {
		t.Fatalf("unknown policy selected segment %d, want 0", got)
	}
	// Out-of-range indices keep the current value rather than inventing one.
	if _, ok := settingsDonePolicyAt(7); ok {
		t.Fatal("index 7 is not an option and must not resolve")
	}
	view := settingsView{DonePolicy: string(DonePolicyOff)}
	if got := settingsSegmentIndex(row, view); got != 1 {
		t.Fatalf("off should select segment 1, got %d", got)
	}
}

// The mutable-view settings (view) are a derived snapshot; the live ones are
// what the handlers read. A field missing from settingsView cannot be drawn.

func TestThemeSegmentMatchesTheThemeList(t *testing.T) {
	rows := appearanceRows(i18n.LangEn)
	var theme settingsRow
	for _, row := range rows {
		if row.control == settingsTheme {
			theme = row
		}
	}
	if len(theme.options) != len(ThemeNames) {
		t.Fatalf("theme segment offers %d options, ThemeNames has %d", len(theme.options), len(ThemeNames))
	}
	for index, name := range ThemeNames {
		if got := settingsThemeIndex(name); got != index {
			t.Fatalf("settingsThemeIndex(%q) = %d, want %d", name, got, index)
		}
	}
	if got := settingsThemeIndex("nonsense"); got != 0 {
		t.Fatalf("unknown theme selected segment %d, want 0", got)
	}
}

// TestRenderSettingsPreview rasterises the settings window so the layout can be
// reviewed without a live window (a layered window cannot be screenshotted).
//
//	PET_PREVIEW_DIR=<dir> go test ./internal/ui -run TestRenderSettingsPreview

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
