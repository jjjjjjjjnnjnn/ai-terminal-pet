package ui

import (
	"syscall"
	"unsafe"

	win "github.com/lxn/win"
)

type rgba struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

type canvasPoint struct {
	X int
	Y int
}

const (
	maxCanvasDimension = 4096
	maxCanvasPixels    = 16 * 1024 * 1024
)

type canvas struct {
	dc     win.HDC
	bitmap win.HBITMAP
	old    win.HGDIOBJ
	width  int
	height int
	stride int
	pixels []byte
}

func newCanvas(width, height int) (*canvas, error) {
	if width < 1 || height < 1 || width > maxCanvasDimension || height > maxCanvasDimension || width > maxCanvasPixels/height {
		return nil, syscall.EINVAL
	}
	dc := win.CreateCompatibleDC(0)
	if dc == 0 {
		return nil, syscall.EINVAL
	}
	header := win.BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(win.BITMAPINFOHEADER{})),
		BiWidth:       int32(width),
		BiHeight:      -int32(height),
		BiPlanes:      1,
		BiBitCount:    32,
		BiCompression: win.BI_RGB,
		BiSizeImage:   uint32(width * height * 4),
	}
	var bits unsafe.Pointer
	bitmap := win.CreateDIBSection(dc, &header, win.DIB_RGB_COLORS, &bits, 0, 0)
	if bitmap == 0 || bits == nil {
		win.DeleteDC(dc)
		return nil, syscall.EINVAL
	}
	old := win.SelectObject(dc, win.HGDIOBJ(bitmap))
	if old == 0 {
		win.DeleteObject(win.HGDIOBJ(bitmap))
		win.DeleteDC(dc)
		return nil, syscall.EINVAL
	}
	result := &canvas{
		dc:     dc,
		bitmap: bitmap,
		old:    old,
		width:  width,
		height: height,
		stride: width * 4,
		pixels: unsafe.Slice((*byte)(bits), width*height*4),
	}
	result.clear()
	return result, nil
}

func (c *canvas) close() {
	if c == nil || c.dc == 0 {
		return
	}
	win.SelectObject(c.dc, c.old)
	win.DeleteObject(win.HGDIOBJ(c.bitmap))
	win.DeleteDC(c.dc)
	c.dc = 0
	c.bitmap = 0
	c.pixels = nil
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

func (c *canvas) fillPolygon(points []canvasPoint, value rgba) {
	if len(points) < 3 || value.A == 0 {
		return
	}
	minimumX, minimumY := points[0].X, points[0].Y
	maximumX, maximumY := minimumX, minimumY
	for _, point := range points[1:] {
		if point.X < minimumX {
			minimumX = point.X
		}
		if point.X > maximumX {
			maximumX = point.X
		}
		if point.Y < minimumY {
			minimumY = point.Y
		}
		if point.Y > maximumY {
			maximumY = point.Y
		}
	}
	for y := minimumY; y <= maximumY; y++ {
		for x := minimumX; x <= maximumX; x++ {
			if pointInside(float64(x)+0.5, float64(y)+0.5, points) {
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

func (c *canvas) text(left, top, width, height int, value string, fontHeight, weight int, color rgba, flags uint32) {
	if value == "" || width <= 0 || height <= 0 {
		return
	}
	if fontHeight < 1 {
		fontHeight = 1
	}
	encoded, err := syscall.UTF16FromString(value)
	if err != nil || len(encoded) == 0 {
		return
	}
	font := cachedFont(fontHeight, weight)
	if font == 0 {
		return
	}
	oldFont := win.SelectObject(c.dc, win.HGDIOBJ(font))
	win.SetBkMode(c.dc, win.TRANSPARENT)
	win.SetTextColor(c.dc, win.RGB(color.R, color.G, color.B))
	bounds := win.RECT{Left: int32(left), Top: int32(top), Right: int32(left + width), Bottom: int32(top + height)}
	win.DrawTextEx(c.dc, &encoded[0], int32(len(encoded)-1), &bounds, flags, nil)
	if oldFont != 0 {
		win.SelectObject(c.dc, oldFont)
	}
	c.normalizeText(left, top, width, height, color)
}

// fontCache avoids CreateFontIndirect/DeleteObject churn: the 50ms
// animation timer repaints at 20FPS in waiting/error states, each frame
// issuing ~10 text() calls. Only a handful of distinct (height, weight)
// pairs exist, so a process-lifetime cache is bounded. Single-threaded
// use only (all present() calls run on the UI thread); no eviction.
var fontCache = make(map[uint64]win.HFONT)

func cachedFont(fontHeight, weight int) win.HFONT {
	key := uint64(uint32(fontHeight))<<32 | uint64(uint32(weight))
	if font, ok := fontCache[key]; ok {
		return font
	}
	face, _ := syscall.UTF16FromString("Microsoft YaHei UI")
	fontSpec := win.LOGFONT{
		LfHeight:         int32(fontHeight),
		LfWeight:         int32(weight),
		LfCharSet:        win.DEFAULT_CHARSET,
		LfOutPrecision:   win.OUT_TT_PRECIS,
		LfQuality:        win.ANTIALIASED_QUALITY,
		LfPitchAndFamily: win.DEFAULT_PITCH,
	}
	copy(fontSpec.LfFaceName[:], face)
	font := win.CreateFontIndirect(&fontSpec)
	if font == 0 {
		return 0
	}
	fontCache[key] = font
	return font
}

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

func (c *canvas) blend(x, y int, value rgba) {
	if x < 0 || y < 0 || x >= c.width || y >= c.height || value.A == 0 {
		return
	}
	offset := y*c.stride + x*4
	destinationA := float64(c.pixels[offset+3])
	sourceA := float64(value.A)
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

func (c *canvas) normalizeText(left, top, width, height int, color rgba) {
	maximum := maxInt(int(color.R), maxInt(int(color.G), int(color.B)))
	if maximum == 0 || color.A == 0 {
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
			alpha := clampChannel(float64(value) * 255 / float64(maximum))
			c.pixels[offset] = clampChannel(float64(color.B) * float64(alpha) / 255)
			c.pixels[offset+1] = clampChannel(float64(color.G) * float64(alpha) / 255)
			c.pixels[offset+2] = clampChannel(float64(color.R) * float64(alpha) / 255)
			c.pixels[offset+3] = alpha
		}
	}
}

func (c *canvas) alphaAt(x, y int) uint8 {
	if x < 0 || y < 0 || x >= c.width || y >= c.height {
		return 0
	}
	return c.pixels[y*c.stride+x*4+3]
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

func pointInside(x, y float64, points []canvasPoint) bool {
	inside := false
	previous := len(points) - 1
	for current := 0; current < len(points); current++ {
		currentX, currentY := float64(points[current].X), float64(points[current].Y)
		previousX, previousY := float64(points[previous].X), float64(points[previous].Y)
		if (currentY > y) != (previousY > y) && x < (previousX-currentX)*(y-currentY)/(previousY-currentY)+currentX {
			inside = !inside
		}
		previous = current
	}
	return inside
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
