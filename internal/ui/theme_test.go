package ui

import (
	"math"
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

// luminance is the standard relative-luminance formula, used here only to ask
// "is this text lighter or darker than its background than it is readable".
// A WCAG 4.5:1 ratio is deliberately NOT asserted: these are 8-12 px labels on a
// translucent panel, and a ratio measured against a hypothetical desktop would
// be a stronger claim than the product can back up.
func luminance(c rgba) float64 {
	channel := func(value uint8) float64 {
		v := float64(value) / 255
		if v <= 0.0396 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
}

func contrastAgainst(text, background rgba) float64 {
	lighter, darker := luminance(text), luminance(background)
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

// TestEveryThemeKeepsItsTextReadableOnItsOwnPanel is the test that would have
// caught the obvious light-theme mistake: reusing the dark palette's "muted"
// (light grey) on a near-white fill makes the small print vanish.
func TestEveryThemeKeepsItsTextReadableOnItsOwnPanel(t *testing.T) {
	for _, name := range ThemeNames {
		p := themePalette(name)
		for label, pair := range map[string][2]rgba{
			"primary":   {p.Primary, p.Panel},
			"secondary": {p.Secondary, p.Panel},
			"muted":     {p.Muted, p.Panel},
			"hint":      {p.Hint, p.Panel},
		} {
			if ratio := contrastAgainst(pair[0], pair[1]); ratio < 3.0 {
				t.Errorf("%s: %s on panel is %.2f:1, too low to read", name, label, ratio)
			}
		}
		// Every state hue has to stand out from the panel it is drawn on: it is
		// the one signal the strip exists for.
		for label, colour := range map[string]rgba{
			"waiting": p.Waiting, "error": p.Error, "working": p.Working,
			"tool": p.Tool, "thinking": p.Thinking, "idle": p.Idle,
		} {
			if ratio := contrastAgainst(colour, p.Panel); ratio < 2.5 {
				t.Errorf("%s: state %s on panel is %.2f:1", name, label, ratio)
			}
		}
	}
}

// TestThemesActuallyDiffer keeps a theme from being a switch that changes
// nothing: slate and paper must not share their panel or primary text.
func TestThemesActuallyDiffer(t *testing.T) {
	slate, paper := themePalette(themeSlate), themePalette(themePaper)
	if slate.Panel == paper.Panel {
		t.Fatal("slate and paper share a panel colour, so the theme does nothing")
	}
	if luminance(slate.Panel) >= luminance(paper.Panel) {
		t.Fatal("paper is not lighter than slate")
	}
	contrast := themePalette(themeContrast)
	if contrast.Panel.R != 0 || contrast.Panel.G != 0 || contrast.Panel.B != 0 {
		t.Fatal("the contrast theme's panel is not black")
	}
	// The palette in use is slate unless something sets it otherwise, which is
	// what keeps every existing screenshot true.
	if colors != slatePalette {
		t.Fatal("the process-wide palette is not slate")
	}
}

// TestNormalizeThemeFallsBackToSlate pins the same contract as
// config.ValidTheme: an unknown name must not reach themePalette as a zero
// palette (transparent black would make the whole panel invisible).
func TestNormalizeThemeFallsBackToSlate(t *testing.T) {
	cases := map[string]string{
		"":              themeSlate,
		"  SLATE ":      themeSlate,
		"PAPER":         themePaper,
		"contrast":      themeContrast,
		"high-contrast": themeSlate,
		"nonsense":      themeSlate,
	}
	for input, want := range cases {
		if got := normalizeTheme(input); got != want {
			t.Errorf("normalizeTheme(%q) = %q, want %q", input, got, want)
		}
	}
	for _, name := range ThemeNames {
		if normalizeTheme(name) != name {
			t.Errorf("%q is not its own normalised form", name)
		}
	}
}

// TestThemeAndProvenanceMoveTheFrameKey is the regression guard for the same
// class of bug the frozen wait duration had: the redraw guard compares keys, so
// anything drawn has to be in the key.
func TestThemeAndProvenanceMoveTheFrameKey(t *testing.T) {
	now := time.Unix(1_000_020, 0)
	base := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 2, true, false, false, 0, false, now, i18n.LangEn, -1, themeSlate, false, 0, 0, -1)
	themed := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 2, true, false, false, 0, false, now, i18n.LangEn, -1, themePaper, false, 0, 0, -1)
	perSource := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 2, true, false, false, 0, false, now, i18n.LangEn, -1, themeSlate, true, 0, 0, -1)
	if base == themed {
		t.Fatal("a theme change must repaint")
	}
	if base == perSource {
		t.Fatal("turning provenance colours on must repaint")
	}
	// An unknown theme name normalises to slate, so it must NOT repaint.
	same := frameKey(compactWidth, compactHeight, 7, session.StateWaitingPermission, 2, true, false, false, 0, false, now, i18n.LangEn, -1, "nonsense", false, 0, 0, -1)
	if base != same {
		t.Fatal("an unknown theme name normalised to slate but the key still moved")
	}
}

// TestSourceColorsAreDistinctAndNeutralForUnknown pins the provenance palette:
// the four known terminals must be distinguishable from each other, and an
// unlisted reporter must get the neutral colour rather than a guess.
func TestSourceColorsAreDistinctAndNeutralForUnknown(t *testing.T) {
	seen := map[rgba]string{}
	for _, source := range []string{"claude-hook", "opencode", "antigravity-statusline", "deepseek-harness"} {
		colour := sourceColor(source)
		if previous, ok := seen[colour]; ok {
			t.Fatalf("%q and %q share a provenance colour", source, previous)
		}
		seen[colour] = source
	}
	if got := sourceColor("some-new-cli"); got != colors.Secondary {
		t.Fatalf("unknown source colour = %v, want the neutral Secondary", got)
	}
	if got := sourceColor("CLAUDE-HOOK"); got != sourceColor("claude-hook") {
		t.Fatal("source matching must be case-insensitive")
	}
}
