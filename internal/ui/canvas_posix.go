//go:build !windows && !linux && !darwin

package ui

// The compile-only canvas for platforms whose window layer does not exist yet.
//
// What it is: a real buffer. prepare returns an ordinary slice, so every
// primitive in canvas.go, every hit map, and every pixel-level test works here
// exactly as it does on Windows. That is not a courtesy - it is what lets
// `GOOS=darwin go vet ./internal/ui` type-check the 2,927 lines of drawing code
// that handover §55.4 called the most valuable part of this port.
//
// What it is not: a renderer. Text is the one primitive with no portable
// implementation (GDI rasterises it on Windows), so drawText returns false and
// measureText estimates. Both are stated here rather than hidden in a helper:
//
//   - a panel drawn with this backend has shapes and no words;
//   - its layout is laid out for an estimate, so labels can overlap.
//
// Neither can reach a user, because the platform entry point refuses to start the
// overlay where there is no window layer: see internal/app's
// "no window layer on this platform" error. When the X11 backend lands (canvas_x11.go)
// it replaces this file's drawText with a real glyph rasteriser and deletes the
// estimate below; the file name is the reminder that this is the piece still
// missing.

// canvasBackend is empty here: there is nothing to prepare.
type canvasBackend struct{}

// prepare hands back a plain RGBA buffer.
func (b *canvasBackend) prepare(width, height int) ([]byte, error) {
	return make([]byte, width*height*4), nil
}

// release has nothing to free.
func (b *canvasBackend) release() {}

// drawText draws nothing and says so.
//
// Returning false rather than painting a placeholder is the honest answer: the
// caller (canvas.text) then skips normalizeText, so the buffer keeps whatever
// the shapes left instead of gaining a fake block where a label should be.
func (b *canvasBackend) drawText(target *canvas, left, top, width, height int, value string, fontHeight, weight int, color rgba, flags uint32) bool {
	return false
}

// measureText estimates an advance width from the rune count.
//
// An estimate is the wrong answer for layout and it is the only answer available
// without a font: the real implementation measures glyph advances. It is here
// because the alternative - returning 0 - collapses every wrapped paragraph into
// one line and looks like a layout bug rather than a missing renderer.
func (b *canvasBackend) measureText(value string, fontHeight, weight int) int {
	count := 0
	for range value {
		count++
	}
	// Half an em per rune is the average of a proportional font; bold is
	// slightly wider, and the weight is a divisor on 400 (regular).
	perRune := fontHeight / 2
	if weight >= 600 {
		perRune = fontHeight*3/5 + 1
	}
	if perRune < 1 {
		perRune = 1
	}
	return count * perRune
}
