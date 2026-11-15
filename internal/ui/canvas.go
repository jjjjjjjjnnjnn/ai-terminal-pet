package ui

// The pixel buffer and every drawing primitive the panel is made of.
//
// This file has no build tag, and that is handover §55.6 step 2: the twelve
// "rename-only" files (drawing_windows.go, drawing_compact_windows.go,
// settings_text.go, rows/mute/pulse/receipt/copy_toast/focus/tap/
// attention/approve) never call Win32 at all - they call *canvas. The type was
// Windows-only only because its GDI handles lived here. Now the buffer is the
// plain thing it always was (width, height, stride, RGBA bytes) and each
// platform supplies its own presentation: canvas_windows.go puts GDI behind these
// same methods, canvas_x11.go sends the bytes to an X server.
//
// What stays here is exactly what the drawing code depends on: the 16 primitives
// listed in §55.4, the blend arithmetic, the hit map, and the rounded-corner
// test. A platform implementation that changes these would change how the panel
// looks, which is the one thing a port must not do.

import (
	"syscall"
)

// rgba is the panel's colour type. Field order matters: canvas_windows.go stores
// BGRA in the DIB, so the byte order in memory is B,G,R,A while this struct reads
// R,G,B,A. The blend function below writes the bytes explicitly and is the only
// place that knows.
type rgba struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

const (
	// maxCanvasDimension and maxCanvasPixels bound one allocation. A session
	// list can grow, and a 4096x4096 cap is generous for a 200x400 panel while
	// still refusing a width*height that would overflow an int on a 32-bit
	// build.
	maxCanvasDimension = 4096
	maxCanvasPixels    = 16 * 1024 * 1024
)

// canvas is one reusable pixel buffer.
//
// The platform-specific fields are added by canvas_<goos>.go as a second
// declaration of the same struct? No - Go has no partial structs. So the buffer
// is a separate value the platform implementation embeds: see canvasBackend in
// each platform file, and note that every method here works on pixels only.
type canvas struct {
	backend canvasBackend
	width   int
	height  int
	stride  int
	pixels  []byte
	// alphaScale multiplies every colour's alpha as it is blended, so a whole
	// layer can be faded without rewriting the palette. The dock slide uses it
	// to fade the panel's contents out while its silhouette slides into the
	// screen edge. Reset to 1 when the layer is done.
	alphaScale float64
}

// newCanvas allocates one buffer and asks the platform to prepare whatever it
// needs to present it.
//
// The buffer's memory is the platform's, not this file's: on Windows the pixels
// have to *be* the DIB section's bits, because GDI rasterises text into them
// directly; on X11 they are a plain slice uploaded with PutImage. So the backend
// returns the slice rather than receiving one to fill.
func newCanvas(width, height int) (*canvas, error) {
	if width < 1 || height < 1 || width > maxCanvasDimension || height > maxCanvasDimension || width > maxCanvasPixels/height {
		return nil, syscall.EINVAL
	}
	result := &canvas{
		width:  width,
		height: height,
		stride: width * 4,
		// Zero would mean "fade everything to nothing", which is a nasty way to
		// fail for a field whose whole job is to default to no-op.
		alphaScale: 1,
	}
	pixels, err := result.backend.prepare(width, height)
	if err != nil {
		return nil, err
	}
	result.pixels = pixels
	return result, nil
}

// close releases whatever the platform prepared. Idempotent, because present()
// closes a surface when the geometry changes and a second close must not fault.
func (c *canvas) close() {
	if c == nil {
		return
	}
	c.backend.release()
	c.pixels = nil
}

// setAlphaScale sets the multiplier applied to every subsequent draw. Values
// outside 0..1 are clamped; 1 (the default) is a no-op.
func (c *canvas) setAlphaScale(scale float64) {
	if scale < 0 {
		scale = 0
	}
	if scale > 1 {
		scale = 1
	}
	c.alphaScale = scale
}

// scaledAlpha applies the current layer fade to a colour's alpha.
func (c *canvas) scaledAlpha(value uint8) float64 {
	if c.alphaScale >= 1 {
		return float64(value)
	}
	return float64(value) * c.alphaScale
}

func (c *canvas) clear() {
	for index := range c.pixels {
		c.pixels[index] = 0
	}
}

func (c *canvas) fillRect(left, top, width, height int, value rgba) {
	c.fillRounded(left, top, width, height, 0, value)
}

func (c *canvas) fillRounded(left, top, width, height, radius int, value rgba) {
	if width <= 0 || height <= 0 || value.A == 0 {
		return
	}
	if radius < 0 {
		radius = 0
	}
	if radius*2 > width {
		radius = width / 2
	}
	if radius*2 > height {
		radius = height / 2
	}
	right := left + width
	bottom := top + height
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			if c.insideRounded(x, y, left, top, right, bottom, radius) {
				c.blend(x, y, value)
			}
		}
	}
}

func (c *canvas) outlineRounded(left, top, width, height, radius, thickness int, value rgba) {
	if width <= 0 || height <= 0 || thickness <= 0 || value.A == 0 {
		return
	}
	if radius < thickness {
		radius = thickness
	}
	if radius*2 > width {
		radius = width / 2
	}
	if radius*2 > height {
		radius = height / 2
	}
	right := left + width
	bottom := top + height
	innerLeft := left + thickness
	innerTop := top + thickness
	innerRight := right - thickness
	innerBottom := bottom - thickness
	innerRadius := maxInt(radius-thickness, 0)
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			outer := c.insideRounded(x, y, left, top, right, bottom, radius)
			inner := c.insideRounded(x, y, innerLeft, innerTop, innerRight, innerBottom, innerRadius)
			if outer && !inner {
				c.blend(x, y, value)
			}
		}
	}
}

func (c *canvas) fillEllipse(left, top, width, height int, value rgba) {
	if width <= 0 || height <= 0 || value.A == 0 {
		return
	}
	radiusX := float64(width) / 2
	radiusY := float64(height) / 2
	centerX := float64(left) + radiusX
	centerY := float64(top) + radiusY
	for y := top; y < top+height; y++ {
		for x := left; x < left+width; x++ {
			dx := (float64(x) + 0.5 - centerX) / radiusX
			dy := (float64(y) + 0.5 - centerY) / radiusY
			if dx*dx+dy*dy <= 1 {
				c.blend(x, y, value)
			}
		}
	}
}

func (c *canvas) line(x1, y1, x2, y2, width int, value rgba) {
	if width < 1 {
		width = 1
	}
	minimumX := minInt(x1, x2) - width
	maximumX := maxInt(x1, x2) + width
	minimumY := minInt(y1, y2) - width
	maximumY := maxInt(y1, y2) + width
	radius := float64(width) / 2
	radiusSquared := radius * radius
	for y := minimumY; y <= maximumY; y++ {
		for x := minimumX; x <= maximumX; x++ {
			px := float64(x) + 0.5
			py := float64(y) + 0.5
			vx := float64(x2 - x1)
			vy := float64(y2 - y1)
			segment := vx*vx + vy*vy
			projection := 0.0
			if segment > 0 {
				projection = ((px-float64(x1))*vx + (py-float64(y1))*vy) / segment
				projection = mathClamp(projection, 0, 1)
			}
			nearestX := float64(x1) + projection*vx
			nearestY := float64(y1) + projection*vy
			deltaX := px - nearestX
			deltaY := py - nearestY
			if deltaX*deltaX+deltaY*deltaY <= radiusSquared {
				c.blend(x, y, value)
			}
		}
	}
}

// text draws one string into a box and then normalises the anti-aliased
// greyscale it produced into the requested colour (see normalizeText).
//
// The platform hook is required because text is the one primitive with no
// portable implementation: GDI rasterises it on Windows, and a POSIX build has
// to rasterise it itself. Everything else above is arithmetic on the buffer.
func (c *canvas) text(left, top, width, height int, value string, fontHeight, weight int, color rgba, flags uint32) {
	if value == "" || width <= 0 || height <= 0 {
		return
	}
	if fontHeight < 1 {
		fontHeight = 1
	}
	// The platform returns false when it cannot draw (no font, no glyphs).
	// Drawing nothing is the right failure: a panel missing one label still
	// shows the rest, while a panic would take the whole window down.
	if !c.backend.drawText(c, left, top, width, height, value, fontHeight, weight, color, flags) {
		return
	}
}

// measure returns the advance width of value in pixels at the given font.
func (c *canvas) measure(value string, fontHeight, weight int) int {
	if value == "" {
		return 0
	}
	if fontHeight < 1 {
		fontHeight = 1
	}
	return c.backend.measureText(value, fontHeight, weight)
}

// fitText trims value with a trailing ellipsis until it fits maxWidth.
//
// A layered window has no scrollbar and no tooltip, so text that does not fit
// is not merely clipped, it is lost. DT_END_ELLIPSIS cannot help in a
// DT_WORDBREAK box (the two flags are mutually exclusive), which is how long
// project names used to spill straight out of both sides of the strip.
func (c *canvas) fitText(value string, fontHeight, weight, maxWidth int) string {
	if value == "" || maxWidth <= 0 {
		return ""
	}
	if c.measure(value, fontHeight, weight) <= maxWidth {
		return value
	}
	const ellipsis = "…"
	runes := []rune(value)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + ellipsis
		if c.measure(candidate, fontHeight, weight) <= maxWidth {
			return candidate
		}
	}
	return ""
}

// insideRounded is the corner test every filled shape goes through.
func (c *canvas) insideRounded(x, y, left, top, right, bottom, radius int) bool {
	if x < left || x >= right || y < top || y >= bottom {
		return false
	}
	if radius == 0 {
		return true
	}
	cornerX := x
	cornerY := y
	if x < left+radius {
		cornerX = left + radius
	} else if x >= right-radius {
		cornerX = right - radius - 1
	}
	if y < top+radius {
		cornerY = top + radius
	} else if y >= bottom-radius {
		cornerY = bottom - radius - 1
	}
	dx := x - cornerX
	dy := y - cornerY
	return dx*dx+dy*dy <= radius*radius
}

// blend composites one source pixel over the destination.
//
// The pixel bytes are BGRA, which is what every layered-window API on Windows
// expects and what X11's ZPixmap with a 32-bit visual also gives for the
// little-endian models this project supports. Writing them explicitly is what
// keeps rgba's field order readable at every call site.
func (c *canvas) blend(x, y int, value rgba) {
	if x < 0 || y < 0 || x >= c.width || y >= c.height || value.A == 0 {
		return
	}
	offset := y*c.stride + x*4
	destinationA := float64(c.pixels[offset+3])
	sourceA := c.scaledAlpha(value.A)
	if sourceA <= 0 {
		return
	}
	outputA := sourceA + destinationA*(255-sourceA)/255
	if outputA <= 0 {
		c.pixels[offset] = 0
		c.pixels[offset+1] = 0
		c.pixels[offset+2] = 0
		c.pixels[offset+3] = 0
		return
	}
	weight := destinationA * (255 - sourceA) / 255
	blue := (float64(value.B)*sourceA + float64(c.pixels[offset])*weight) / outputA
	green := (float64(value.G)*sourceA + float64(c.pixels[offset+1])*weight) / outputA
	red := (float64(value.R)*sourceA + float64(c.pixels[offset+2])*weight) / outputA
	c.pixels[offset] = clampChannel(blue)
	c.pixels[offset+1] = clampChannel(green)
	c.pixels[offset+2] = clampChannel(red)
	c.pixels[offset+3] = clampChannel(outputA)
}

// normalizeText converts the anti-aliased greyscale a text rasteriser produced
// into the requested colour.
//
// Every platform's text is drawn as coverage, not colour: the glyphs land as
// white-on-nothing, and this walks the box turning non-zero luminance into the
// palette colour at the same alpha. That is why text is the one primitive whose
// platform hook hands the work back to shared code.
func (c *canvas) normalizeText(left, top, width, height int, color rgba) {
	maximum := maxInt(int(color.R), maxInt(int(color.G), int(color.B)))
	scale := c.alphaScale
	if maximum == 0 || color.A == 0 || scale <= 0 {
		return
	}
	for y := top; y < top+height; y++ {
		for x := left; x < left+width; x++ {
			if x < 0 || y < 0 || x >= c.width || y >= c.height {
				continue
			}
			offset := y*c.stride + x*4
			if c.pixels[offset+3] != 0 {
				continue
			}
			value := maxInt(int(c.pixels[offset]), maxInt(int(c.pixels[offset+1]), int(c.pixels[offset+2])))
			if value == 0 {
				continue
			}
			alpha := clampChannel(float64(value) * 255 / float64(maximum) * scale)
			c.pixels[offset] = clampChannel(float64(color.B) * float64(alpha) / 255)
			c.pixels[offset+1] = clampChannel(float64(color.G) * float64(alpha) / 255)
			c.pixels[offset+2] = clampChannel(float64(color.R) * float64(alpha) / 255)
			c.pixels[offset+3] = alpha
		}
	}
}

// hitMapInto writes the alpha channel into a caller-owned buffer so the
// 20 FPS alert animation does not allocate a fresh width*height slice on
// every frame. The buffer must be at least width*height bytes.
func (c *canvas) hitMapInto(result []byte) {
	if len(result) < c.width*c.height {
		return
	}
	result = result[:c.width*c.height]
	for index := 0; index < len(result); index++ {
		result[index] = c.pixels[index*4+3]
	}
}

func clampChannel(value float64) uint8 {
	if value <= 0 {
		return 0
	}
	if value >= 255 {
		return 255
	}
	return uint8(value + 0.5)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func mathClamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
