package ui

// Measuring, not asserting: how long does one settings raster take on this
// machine at the sizes the user actually runs?
//
// docs/HANDOVER.md §51.3 listed three candidates for "the settings window goes
// away". 1.0.8 proved candidate 1 (the idle self-destruct quitting the process
// underneath a window on the same UI thread) and fixed it. Candidates 2 and 3
// (a stalled message pump, a DPI clamp that puts the window off the work area)
// are still open, and the honest first step is to measure the paint instead of
// guessing. This file is that measurement. It is a report, not a gate, so it
// prints and never fails on a slow machine - a threshold here would be a
// flaky test on a busy CI runner, which is worse than no test.

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
)

func TestMeasureSettingsPaint(t *testing.T) {
	if testing.Short() {
		t.Skip("measurement runs only when asked")
	}
	sizes := []struct {
		name     string
		textSize string
	}{
		{"standard", textSizeStandard},
		{"large", textSizeLarge},
	}
	view := settingsView{
		Notifications: true, AlwaysOnTop: true, EdgeRetract: true,
		Sound: true, SoundOnWaiting: true, SoundOnError: true, NotifyOnFinish: true,
		Opacity: 235, Volume: 85, ApproveTimeout: 5,
		Scheme: config.SoundSchemeChime, Language: "zh-CN",
		Theme: config.ThemeContrast,
	}
	rowsForTab := func(tab int, lang i18n.Lang) []settingsRow {
		switch tab {
		case 1:
			return soundRows(lang)
		case 2:
			return sessionRows(lang, 0)
		case 3:
			return appearanceRows(lang)
		default:
			return generalRows(lang)
		}
	}
	// A measuring canvas is what turns hint text into a line count; without it
	// the geometry is short by however many wrapped lines the hints need.
	measure, err := newCanvas(1, 1)
	if err != nil {
		t.Fatalf("measure canvas: %v", err)
	}
	defer measure.close()

	for _, size := range sizes {
		fonts := settingsTypeRamps[size.textSize]
		for _, tab := range []int{0, 1, 2, 3} {
			lang := i18n.LangZhCN
			probe := settingsGeometrySized(150, size.textSize, tab, nil)
			rows := measureSettingsRows(measure, rowsForTab(tab, lang), fonts, probe.scale(settingsContentWidth))
			geometry := settingsGeometryFitting(150, size.textSize, tab, rows, 1600)
			target, err := newCanvas(geometry.width, geometry.height)
			if err != nil {
				t.Fatalf("%s tab %d: canvas: %v", size.name, tab, err)
			}
			paint := settingsPaint{geometry: geometry, view: view, lang: lang, hover: settingsTheme}
			drawSettings(target, paint) // warm-up: fonts and the DIB
			const runs = 20
			start := time.Now()
			for index := 0; index < runs; index++ {
				drawSettings(target, paint)
			}
			perFrame := time.Since(start) / runs
			t.Logf("%-9s tab=%d %4dx%-4d %8.3f ms/frame", size.name, tab, geometry.width, geometry.height, float64(perFrame.Microseconds())/1000)
			target.close()
		}
	}
}
