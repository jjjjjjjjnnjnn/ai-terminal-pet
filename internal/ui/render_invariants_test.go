package ui

import (
	"testing"
	"time"

	"ai-terminal-pet/internal/config"
	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/session"
)

// The panel's appearance has exactly one automated guarantee today: `go test`
// proves the *layout* (rects, rows, hit boxes) and nothing proves the *pixels*.
// docs/ROADMAP.md §6 said so out loud: "UI has no automated visual regression;
// rely on --doctor, the probes and a human looking at TestRenderPreview.
//
// That is a real gap for two of the three platforms. Windows' rasteriser has
// been looked at thousands of times by whoever runs the preview test. The Linux
// X11 and macOS AppKit panels are rasterised by a different, pure-Go code path
// that nobody on this project can look at on real hardware (docs/PLATFORMS.md),
// so "the panel draws something recognisable" is currently an assumption.
//
// This file turns that assumption into assertions - of *invariants*, never of
// exact pixels. GDI and the OpenType rasteriser antialias differently, ship
// different default fonts and hint differently, so pixel equality would only
// ever mean "the same machine". What must hold everywhere is the shape of a
// correct render:
//
//   - the frame is the size the layout says, opaque in the middle, and fully
//     transparent outside the rounded corners;
//   - the panel body is painted (not a transparent hole);
//   - text really was rasterised - there are light pixels where words go, which
//     is what a missing font stack silently loses;
//   - the session state's colour is actually on screen, in the hue the palette
//     declares, for each state;
//   - different states and different languages actually look different, so a
//     renderer that ignores its input cannot pass;
//   - the two shipped palettes are actually different, and the light one is
//     actually light (the "muted text on a near-white fill" mistake that
//     internal/ui/theme_test.go guards at the colour level, guarded here at the
//     pixel level).
//
// Thresholds are deliberately loose bands, not magic numbers copied from one
// machine: `go test -run TestRenderedPanel -v` logs what was actually measured,
// so moving a band is a decision made from a number rather than a guess.

const (
	// Measured 2026-11-17 on Windows/slate/zh, which is the GDI rasteriser:
	// compact waiting paints 93.8% of the frame with 659 light pixels, compact
	// empty 93.0% with 553, expanded three rows 98.6% with 2512. The bands are
	// deliberately far below those numbers - the two platforms nobody can look
	// at rasterise text with a different font stack and antialias more heavily -
	// while still failing loudly for the two states that matter: nothing
	// painted, or no text at all. `-v` logs what each platform measured.
	minOpaqueRatio = 0.80
	// Measured 93.0%-98.6% across the four cases on both rasterisers; 70% leaves
	// room for a platform whose fill is thinner without letting a low-alpha wash
	// stand in for a panel.
	minSolidRatio = 0.70
	minInkPixels  = 40
	// The state ring, the state word and the jump arrow are the only places a
	// state hue appears; Windows measures 319-913 such pixels depending on the
	// state. A correct render cannot produce zero.
	minStateHuePixels = 20
	// Two states that differ only in the state word still differ in thousands of
	// pixels (7.7% zh-versus-en is the smallest gap measured); 1% catches "the
	// renderer ignored its input" without demanding the same glyph shapes on
	// every platform.
	minDistinctRatio = 0.01
)

// previewNow is a fixed clock: a test that depends on the wall clock is a test
// that fails at midnight.
var previewNow = time.Date(2026, 11, 17, 9, 0, 0, 0, time.UTC)

// invariantSessions is the smallest set of rows that exercises every visual
// element the assertions care about: a waiting row (state hue + countdown), a
// working row and an idle row.
func invariantSessions() []session.Session {
	return []session.Session{
		{
			Source: "claude-hook", SessionID: "wait", Project: `C:\work\ai-terminal-pet`,
			State: session.StateWaitingPermission, Reason: "Bash(go test)",
			UpdatedAt: previewNow.Add(-2 * time.Minute),
		},
		{
			Source: "opencode", SessionID: "work", Project: `C:\work\webapp`,
			State: session.StateWorking, Reason: "editing src/app.tsx",
			UpdatedAt: previewNow.Add(-6 * time.Second),
		},
		{
			Source: "antigravity-statusline", SessionID: "idle", Project: `C:\work\notes`,
			State: session.StateIdle, UpdatedAt: previewNow.Add(-40 * time.Second),
		},
	}
}

// renderInvariantFrame draws one panel state onto its own canvas. The canvas is
// closed by the caller through t.Cleanup so a failing assertion still releases
// the platform buffer.
func renderInvariantFrame(t *testing.T, state panelState) *canvas {
	t.Helper()
	if state.Now.IsZero() {
		state.Now = previewNow
	}
	if state.Sliver == 0 {
		state.Sliver = dockedSliver
	}
	if state.Dock == 0 {
		state.Dock = dockRight
	}
	if state.Lang == "" {
		state.Lang = i18n.LangZhCN
	}
	width, height := compactWidth, compactHeight
	if state.Expanded {
		width, height = expandedWidth, expandedPanelHeight(len(state.Items))
	}
	frame, err := newCanvas(width, height)
	if err != nil {
		t.Fatalf("canvas %dx%d: %v", width, height, err)
	}
	t.Cleanup(frame.close)
	drawVerticalPanel(frame, state)
	return frame
}

// frameStats is what the assertions read: how much of the frame was painted,
// how much of it is genuinely filled rather than a faint wash, how many light
// pixels there are (text and glyphs), and the alpha at the rounded corners.
//
// "Filled" is a ratio and not a single pixel on purpose: on Windows the empty
// compact strip happens to put a translucent hint box under its centre pixel
// (alpha 215) while the same frame is 238 at the centre on Linux. Sampling one
// point would encode where a glyph landed under one platform's font metrics.
type frameStats struct {
	opaque  int
	solid   int
	ink     int
	corners [4]uint8
}

func canvasPixel(c *canvas, x, y int) (r, g, b, a uint8) {
	offset := y*c.stride + x*4
	if offset < 0 || offset+3 >= len(c.pixels) {
		return 0, 0, 0, 0
	}
	return c.pixels[offset+2], c.pixels[offset+1], c.pixels[offset], c.pixels[offset+3]
}

func statsFor(c *canvas) frameStats {
	stats := frameStats{}
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			r, g, b, a := canvasPixel(c, x, y)
			if a == 0 {
				continue
			}
			stats.opaque++
			if a >= 200 {
				stats.solid++
			}
			if a >= 180 && maxInt(int(r), maxInt(int(g), int(b))) >= 150 {
				stats.ink++
			}
		}
	}
	stats.corners = [4]uint8{
		canvasPixelAlpha(c, 0, 0),
		canvasPixelAlpha(c, c.width-1, 0),
		canvasPixelAlpha(c, 0, c.height-1),
		canvasPixelAlpha(c, c.width-1, c.height-1),
	}
	return stats
}

func canvasPixelAlpha(c *canvas, x, y int) uint8 {
	_, _, _, a := canvasPixel(c, x, y)
	return a
}

// countHuePixels counts pixels whose channels satisfy match. Hue predicates are
// written as channel *differences* because text is stored premultiplied: a
// coloured glyph at 60% coverage has its channels scaled, so an equality test
// against the palette colour would miss the very pixels it is looking for.
func countHuePixels(c *canvas, match func(r, g, b, a uint8) bool) int {
	count := 0
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			if match(canvasPixel(c, x, y)) {
				count++
			}
		}
	}
	return count
}

// differingRatio is the fraction of pixels that differ by more than a
// perceptible amount (8 per channel) between two same-sized frames.
func differingRatio(a, b *canvas) float64 {
	if a.width != b.width || a.height != b.height {
		return 1
	}
	total := a.width * a.height
	if total == 0 {
		return 0
	}
	differing := 0
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			ar, ag, ab, aa := canvasPixel(a, x, y)
			br, bg, bb, ba := canvasPixel(b, x, y)
			delta := maxInt(absoluteInt(int(ar)-int(br)), maxInt(absoluteInt(int(ag)-int(bg)), absoluteInt(int(ab)-int(bb))))
			if delta > 8 || absoluteInt(int(aa)-int(ba)) > 8 {
				differing++
			}
		}
	}
	return float64(differing) / float64(total)
}

func absoluteInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// TestRenderedPanelKeepsItsShape is the core invariant: the frame has the
// geometry the layout declares, it is painted, and it contains text.
func TestRenderedPanelKeepsItsShape(t *testing.T) {
	cases := []struct {
		name  string
		state panelState
		// wantWidth/wantHeight are the layout's own numbers, restated here so a
		// silent geometry change has to be made in two places on purpose.
		wantWidth, wantHeight int
	}{
		{
			name:       "compact waiting",
			state:      panelState{Items: invariantSessions()},
			wantWidth:  compactWidth,
			wantHeight: compactHeight,
		},
		{
			name:       "expanded three rows",
			state:      panelState{Items: invariantSessions(), Expanded: true},
			wantWidth:  expandedWidth,
			wantHeight: expandedPanelHeight(3),
		},
		{
			name:       "compact empty",
			state:      panelState{},
			wantWidth:  compactWidth,
			wantHeight: compactHeight,
		},
		{
			name:       "expanded one row",
			state:      panelState{Items: invariantSessions()[:1], Expanded: true},
			wantWidth:  expandedWidth,
			wantHeight: expandedPanelHeight(1),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frame := renderInvariantFrame(t, tc.state)
			if frame.width != tc.wantWidth || frame.height != tc.wantHeight {
				t.Fatalf("frame is %dx%d, layout says %dx%d", frame.width, frame.height, tc.wantWidth, tc.wantHeight)
			}
			stats := statsFor(frame)
			area := frame.width * frame.height
			ratio := float64(stats.opaque) / float64(area)
			solidRatio := float64(stats.solid) / float64(area)
			t.Logf("%s: %d/%d pixels painted (%.1f%%), %.1f%% opaque, %d light pixels",
				tc.name, stats.opaque, area, ratio*100, solidRatio*100, stats.ink)

			// A rounded rect leaves the four corner pixels fully transparent.
			// Getting this wrong means the corner radius stopped being applied,
			// and the frame becomes a hard rectangle on the user's desktop.
			for index, alpha := range stats.corners {
				if alpha != 0 {
					t.Errorf("corner %d has alpha %d, want 0 (rounded corners are no longer transparent)", index, alpha)
				}
			}
			if ratio < minOpaqueRatio {
				t.Errorf("only %.1f%% of the frame is painted (want >= %.0f%%): the body is missing", ratio*100, minOpaqueRatio*100)
			}
			// Painted with a real fill, not a faint wash: a panel that lost its
			// fill colour would still be "opaque" if a shadow or a border
			// covered the frame at low alpha.
			if solidRatio < minSolidRatio {
				t.Errorf("only %.1f%% of the frame is fully opaque (want >= %.0f%%): the body fill is missing or too faint",
					solidRatio*100, minSolidRatio*100)
			}
			// This is the assertion that catches a missing font stack: the
			// OpenType path draws nothing when it cannot find a font, and the
			// panel then looks like a coloured slab with no words in it.
			if stats.ink < minInkPixels {
				t.Errorf("only %d light pixels (want >= %d): no text was rasterised - if this machine has no fonts installed, the panel really would show no words (see internal/ui/canvas_fonts_*.go)",
					stats.ink, minInkPixels)
			}
		})
	}
}

// TestStateColourReachesTheScreen checks the palette's promise at the pixel
// level: a waiting session is amber, a failed one is red, a working one is
// green. Each predicate is written as channel *differences* so it survives the
// antialiasing (and the premultiplied text storage) that every rasteriser does
// differently.
func TestStateColourReachesTheScreen(t *testing.T) {
	cases := []struct {
		name    string
		session session.Session
		match   func(r, g, b, a uint8) bool
	}{
		{
			name:    "waiting is amber",
			session: session.Session{Source: "claude-hook", SessionID: "w", State: session.StateWaitingPermission, Reason: "Bash(go test)", UpdatedAt: previewNow.Add(-time.Minute)},
			match: func(r, g, b, a uint8) bool {
				return a >= 120 && r >= 90 && int(r)-int(b) >= 40 && int(g)-int(b) >= 10
			},
		},
		{
			name:    "error is red",
			session: session.Session{Source: "claude-hook", SessionID: "e", State: session.StateError, Reason: "exit status 1", UpdatedAt: previewNow.Add(-time.Minute)},
			match: func(r, g, b, a uint8) bool {
				return a >= 120 && r >= 90 && int(r)-int(g) >= 40
			},
		},
		{
			name:    "working is green",
			session: session.Session{Source: "opencode", SessionID: "g", State: session.StateWorking, Reason: "editing", UpdatedAt: previewNow.Add(-time.Minute)},
			match: func(r, g, b, a uint8) bool {
				return a >= 120 && g >= 90 && int(g)-int(r) >= 30 && int(g)-int(b) >= 10
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frame := renderInvariantFrame(t, panelState{Items: []session.Session{tc.session}})
			found := countHuePixels(frame, tc.match)
			t.Logf("%s: %d pixels carry that hue", tc.name, found)
			if found < minStateHuePixels {
				t.Errorf("only %d pixels carry the %s hue (want >= %d): the state colour never reached the canvas", found, tc.name, minStateHuePixels)
			}
		})
	}
}

// TestStatesAndLanguagesLookDifferent is the "did the renderer read its input"
// assertion. Two frames that should be visibly different must differ; a
// renderer that ignores the session list or the language would otherwise pass
// every per-frame invariant above.
func TestStatesAndLanguagesLookDifferent(t *testing.T) {
	waiting := renderInvariantFrame(t, panelState{Items: []session.Session{
		{Source: "claude-hook", SessionID: "w", State: session.StateWaitingPermission, Reason: "Bash", UpdatedAt: previewNow},
	}})
	working := renderInvariantFrame(t, panelState{Items: []session.Session{
		{Source: "opencode", SessionID: "g", State: session.StateWorking, Reason: "editing", UpdatedAt: previewNow},
	}})
	idle := renderInvariantFrame(t, panelState{Items: []session.Session{
		{Source: "opencode", SessionID: "i", State: session.StateIdle, UpdatedAt: previewNow},
	}})
	english := renderInvariantFrame(t, panelState{
		Items: []session.Session{{Source: "claude-hook", SessionID: "w", State: session.StateWaitingPermission, Reason: "Bash", UpdatedAt: previewNow}},
		Lang:  i18n.LangEn,
	})

	pairs := []struct {
		name string
		a, b *canvas
	}{
		{"waiting vs working", waiting, working},
		{"waiting vs idle", waiting, idle},
		{"working vs idle", working, idle},
		{"zh vs en", waiting, english},
	}
	for _, pair := range pairs {
		ratio := differingRatio(pair.a, pair.b)
		t.Logf("%s: %.2f%% of pixels differ", pair.name, ratio*100)
		if ratio < minDistinctRatio {
			t.Errorf("%s differ in %.2f%% of pixels (want >= %.0f%%): the renderer ignored its input",
				pair.name, ratio*100, minDistinctRatio*100)
		}
	}
}

// TestThemesAreActuallyDifferent guards the palette switch at the pixel level:
// the light theme must paint a light body, and the shipped theme must paint a
// dark one. theme_test.go already checks the *colours* are readable; this checks
// that applyAppearance reaches the rasteriser at all.
func TestThemesAreActuallyDifferent(t *testing.T) {
	t.Cleanup(func() { applyAppearance(config.ThemeSlate, false) })

	applyAppearance(config.ThemeSlate, false)
	slate := renderInvariantFrame(t, panelState{Items: invariantSessions()})
	slateLuminance := luminance(unpremultipliedCenter(slate))

	applyAppearance(config.ThemePaper, false)
	paper := renderInvariantFrame(t, panelState{Items: invariantSessions()})
	paperLuminance := luminance(unpremultipliedCenter(paper))

	t.Logf("centre luminance: slate %.2f, paper %.2f", slateLuminance, paperLuminance)
	if slateLuminance > 0.20 {
		t.Errorf("slate's panel centre has luminance %.2f (want <= 0.20): the dark theme painted a light body", slateLuminance)
	}
	if paperLuminance < 0.60 {
		t.Errorf("paper's panel centre has luminance %.2f (want >= 0.60): the light theme painted a dark body", paperLuminance)
	}
}

// TestTheBreathIsTheOnlyThingMotionControls is the contract behind
// config.animations: with the switch off, a clock tick may not change one
// pixel; with it on, the waiting frame must actually breathe. Without the
// first half "reduce motion" would be a switch that only stopped the timer and
// left a frozen halo at whatever phase it happened to be drawn in.
func TestTheBreathIsTheOnlyThingMotionControls(t *testing.T) {
	later := previewNow.Add(300 * time.Millisecond)
	waiting := func(still bool, at time.Time) panelState {
		return panelState{Items: invariantSessions()[:1], Now: at, Still: still}
	}

	moving := renderInvariantFrame(t, waiting(false, previewNow))
	movingLater := renderInvariantFrame(t, waiting(false, later))
	if differingRatio(moving, movingLater) < 0.005 {
		t.Errorf("two frames 300 ms apart are identical (%.4f): the waiting disc is not breathing, so the switch has nothing to turn off",
			differingRatio(moving, movingLater))
	}

	still := renderInvariantFrame(t, waiting(true, previewNow))
	stillLater := renderInvariantFrame(t, waiting(true, later))
	if ratio := differingRatio(still, stillLater); ratio != 0 {
		t.Errorf("%.4f of the pixels changed while animations are off: the disc must be drawn at rest", ratio)
	}
}

// unpremultipliedCenter reads the middle pixel as a straight colour, so the
// luminance comparison is about hue and not about how translucent the panel is.
// Text pixels are stored premultiplied; the panel body is not, but normalising
// both cases keeps this helper honest if the fill colour ever changes.
func unpremultipliedCenter(c *canvas) rgba {
	r, g, b, a := canvasPixel(c, c.width/2, c.height/2)
	if a == 0 || a == 255 {
		return rgba{r, g, b, a}
	}
	scale := func(value uint8) uint8 {
		scaled := int(value) * 255 / int(a)
		if scaled > 255 {
			return 255
		}
		return uint8(scaled)
	}
	return rgba{scale(r), scale(g), scale(b), a}
}
