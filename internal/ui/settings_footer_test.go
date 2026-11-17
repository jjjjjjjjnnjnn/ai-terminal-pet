package ui

import (
	"strings"
	"testing"

	"ai-terminal-pet/internal/i18n"
)

// TestSettingsFooterSitsUnderTheLastRow pins the geometry the footer needed: it
// is inside the window (a link clipped by the bottom edge is not a link), it is
// below every row, and its two links do not overlap - overlapping hit rects would
// make "Report an issue" unreachable from the right half of "Star on GitHub".
func TestSettingsFooterSitsUnderTheLastRow(t *testing.T) {
	for _, tab := range []int{0, 1, 2, 3} {
		rows := settingsRowsIn(t, tab, i18n.LangEn)
		geometry := settingsGeometryFor(96, tab, rows)
		if geometry.footer.Width <= 0 || geometry.footer.Height <= 0 {
			t.Fatalf("tab %d: no footer was laid out", tab)
		}
		if geometry.footer.Top < geometry.contentHeight {
			t.Fatalf("tab %d: the footer starts at %d, above the end of the rows (%d)",
				tab, geometry.footer.Top, geometry.contentHeight)
		}
		if geometry.footer.bottom() > geometry.height {
			t.Fatalf("tab %d: the footer ends at %d, past the window height %d",
				tab, geometry.footer.bottom(), geometry.height)
		}
		for index, frame := range geometry.frames {
			if frame.bottom() > geometry.footer.Top {
				t.Fatalf("tab %d: row %d ends below the footer", tab, index)
			}
		}
		star, issue := geometry.star, geometry.issue
		if star.right() > issue.Left {
			t.Fatalf("tab %d: the two footer links overlap (%d > %d)", tab, star.right(), issue.Left)
		}
		if issue.right() > geometry.footer.right() {
			t.Fatalf("tab %d: the second link runs past the content column", tab)
		}
	}
}

func TestSettingsFooterLinksHitTheirOwnControls(t *testing.T) {
	geometry := settingsGeometryFor(96, 0, settingsRowsIn(t, 0, i18n.LangEn))
	cases := []struct {
		name string
		rect settingsRect
		want settingsControl
	}{
		{"star", geometry.star, settingsStar},
		{"issue", geometry.issue, settingsIssue},
	}
	for _, tc := range cases {
		x := tc.rect.Left + tc.rect.Width/2
		y := tc.rect.Top + tc.rect.Height/2
		control, index, _, ok := geometry.hit(x, y)
		if !ok || control != tc.want || index != -1 {
			t.Fatalf("%s: hit(%d,%d) = %v,%d,%v; want %v", tc.name, x, y, control, index, ok, tc.want)
		}
	}
	// A point in the gap between the links must not resolve to either of them.
	gapX := geometry.star.right() + 2
	if geometry.star.right()+2 < geometry.issue.Left {
		control, _, _, ok := geometry.hit(gapX, geometry.star.Top+2)
		if ok && (control == settingsStar || control == settingsIssue) {
			t.Fatalf("the gap between the links hit %v", control)
		}
	}
}

// TestFooterLinksAreReachableByTab: a keyboard user has the same reason to want
// the bug tracker as a mouse user.
func TestFooterLinksAreReachableByTab(t *testing.T) {
	geometry := settingsGeometryFor(96, 0, settingsRowsIn(t, 0, i18n.LangEn))
	order := settingsFocusOrder(geometry)
	tail := order[len(order)-2:]
	if tail[0] != settingsStar || tail[1] != settingsIssue {
		t.Fatalf("the last two focus stops are %v, want star then issue", tail)
	}
}

// TestSubtitleCarriesTheBuild: a bug report that has to be chased for its build
// number is a bug report that does not get filed, and the header is where
// somebody copying the window already looks.
func TestSubtitleCarriesTheBuild(t *testing.T) {
	withVersion := settingsSubtitle(i18n.LangEn, "v9.9.9")
	if !strings.Contains(withVersion, "v9.9.9") {
		t.Fatalf("the subtitle lost the version: %q", withVersion)
	}
	if !strings.Contains(withVersion, i18n.T(i18n.LangEn, i18n.KeySettingsSubtitle)) {
		t.Fatalf("the subtitle lost its own label: %q", withVersion)
	}
	if bare := settingsSubtitle(i18n.LangEn, ""); strings.Contains(bare, "v") {
		t.Fatalf("an empty version invented one: %q", bare)
	}
}

// TestFooterDrawsBothLinks closes the loop from geometry to pixels: the labels
// the footer is supposed to draw actually reach the canvas.
func TestFooterDrawsBothLinks(t *testing.T) {
	geometry := settingsGeometryFor(96, 0, settingsRowsIn(t, 0, i18n.LangEn))
	target, err := newCanvas(geometry.width, geometry.height)
	if err != nil {
		t.Fatal(err)
	}
	defer target.close()
	drawSettings(target, settingsPaint{geometry: geometry, hover: settingsNone, hoverAt: -1, lang: i18n.LangEn})
	star := settingsFooterInk(target, geometry)
	if star < 20 {
		t.Fatalf("only %d light pixels in the footer band: the links were not drawn", star)
	}
	// Hovering one of them must be visible: the footer has no other affordance
	// saying a word is clickable.
	hovered, err := newCanvas(geometry.width, geometry.height)
	if err != nil {
		t.Fatal(err)
	}
	defer hovered.close()
	drawSettings(hovered, settingsPaint{geometry: geometry, hover: settingsStar, hoverAt: -1, lang: i18n.LangEn})
	if hovering := settingsFooterInk(hovered, geometry); hovering <= star {
		t.Fatalf("hovering the link drew no extra pixels (%d vs %d)", hovering, star)
	}
}

// settingsFooterInk counts light pixels inside the footer band.
func settingsFooterInk(c *canvas, geometry settingsGeometry) int {
	count := 0
	for y := geometry.footer.Top; y < geometry.footer.bottom() && y < c.height; y++ {
		for x := geometry.footer.Left; x < geometry.footer.right() && x < c.width; x++ {
			r, g, b, a := canvasPixel(c, x, y)
			if a >= 180 && maxInt(int(r), maxInt(int(g), int(b))) >= 150 {
				count++
			}
		}
	}
	return count
}
