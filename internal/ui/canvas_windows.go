//go:build windows

package ui

// The Windows backend: GDI behind the canvas.
//
// canvas.go owns the buffer and every drawing primitive; this file owns the three
// things only GDI can do - a device context to rasterise text into, a font cache,
// and the DIB section whose pixels become the canvas's pixels.
//
// The struct used to be the canvas itself (dc, bitmap, old HGDIOBJ, pixels all in
// one). Splitting it is handover §55.6 step 2: the twelve drawing files never
// touched those handles, they only called methods, so the type they need is a
// buffer plus a backend with three verbs.

import (
	"syscall"
	"unsafe"

	win "github.com/lxn/win"
)

// canvasBackend is the Windows side of one canvas.
type canvasBackend struct {
	dc     win.HDC
	bitmap win.HBITMAP
	old    win.HGDIOBJ
	armed  bool
}

// prepare creates the DIB section and returns its bits as the canvas buffer.
//
// Returning the slice rather than filling one is the whole reason for this
// split: text is rasterised by GDI into the bitmap's memory, so if the primitives
// blended into a separate allocation the glyphs would land somewhere else.
func (b *canvasBackend) prepare(width, height int) ([]byte, error) {
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
	b.dc = dc
	b.bitmap = bitmap
	b.old = old
	b.armed = true
	return unsafe.Slice((*byte)(bits), width*height*4), nil
}

// release gives the handles back. Idempotent so a resize that closes a surface
// twice cannot fault.
func (b *canvasBackend) release() {
	if !b.armed {
		return
	}
	win.SelectObject(b.dc, b.old)
	win.DeleteObject(win.HGDIOBJ(b.bitmap))
	win.DeleteDC(b.dc)
	b.dc = 0
	b.bitmap = 0
	b.armed = false
}

// drawText rasterises one string into the canvas box. GDI writes grey coverage
// into the bitmap in white, and canvas.normalizeText turns that into the
// requested palette colour.
func (b *canvasBackend) drawText(target *canvas, left, top, width, height int, value string, fontHeight, weight int, color rgba, flags uint32) bool {
	if !b.armed {
		return false
	}
	encoded, err := syscall.UTF16FromString(value)
	if err != nil || len(encoded) == 0 {
		return false
	}
	font := cachedFont(fontHeight, weight)
	if font == 0 {
		return false
	}
	oldFont := win.SelectObject(b.dc, win.HGDIOBJ(font))
	win.SetBkMode(b.dc, win.TRANSPARENT)
	win.SetTextColor(b.dc, win.RGB(255, 255, 255))
	bounds := win.RECT{Left: int32(left), Top: int32(top), Right: int32(left + width), Bottom: int32(top + height)}
	win.DrawTextEx(b.dc, &encoded[0], int32(len(encoded)-1), &bounds, flags, nil)
	if oldFont != 0 {
		win.SelectObject(b.dc, oldFont)
	}
	target.normalizeText(left, top, width, height, color)
	return true
}

// measureText returns the advance width in pixels.
func (b *canvasBackend) measureText(value string, fontHeight, weight int) int {
	if !b.armed {
		return 0
	}
	encoded, err := syscall.UTF16FromString(value)
	if err != nil || len(encoded) < 2 {
		return 0
	}
	font := cachedFont(fontHeight, weight)
	if font == 0 {
		return 0
	}
	oldFont := win.SelectObject(b.dc, win.HGDIOBJ(font))
	var size win.SIZE
	ok := win.GetTextExtentPoint32(b.dc, &encoded[0], int32(len(encoded)-1), &size)
	if oldFont != 0 {
		win.SelectObject(b.dc, oldFont)
	}
	if !ok {
		return 0
	}
	return int(size.CX)
}

// fontCache avoids CreateFontIndirect/DeleteObject churn: the 50ms animation
// timer repaints at 20FPS in waiting/error states, each frame issuing ~10 text()
// calls. Only a handful of distinct (height, weight) pairs exist, so a
// process-lifetime cache is bounded. Single-threaded use only (all present()
// calls run on the UI thread); no eviction.
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
