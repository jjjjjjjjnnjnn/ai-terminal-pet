//go:build linux || darwin

package ui

import (
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type canvasBackend struct{}

func (b *canvasBackend) prepare(width, height int) ([]byte, error) {
	return make([]byte, width*height*4), nil
}

func (b *canvasBackend) release() {}

var (
	linuxFontsOnce    sync.Once
	linuxRegularFonts []*opentype.Font
	linuxBoldFonts    []*opentype.Font
	linuxFaceMu       sync.Mutex
	linuxFaceCache    = make(map[uint64]*compositeFontFace)
)

func loadLinuxFontChain() {
	linuxFontsOnce.Do(func() {
		dirs := openTypeFontDirs()

		var paths []string
		for _, root := range dirs {
			info, err := os.Stat(root)
			if err != nil || !info.IsDir() {
				continue
			}
			_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil || entry.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".ttf" || ext == ".otf" || ext == ".ttc" {
					paths = append(paths, path)
				}
				return nil
			})
		}

		loadOrdered := func(bold bool) []*opentype.Font {
			ranked := append([]string(nil), paths...)
			sort.Slice(ranked, func(i, j int) bool {
				si, sj := scoreOpenTypeFont(ranked[i], bold), scoreOpenTypeFont(ranked[j], bold)
				if si != sj {
					return si < sj
				}
				return ranked[i] < ranked[j]
			})
			var out []*opentype.Font
			seenFamilies := make(map[string]bool)
			for _, p := range ranked {
				if len(out) >= 8 {
					break
				}
				base := strings.ToLower(filepath.Base(p))
				// Keep candidate set fast & bounded.
				if strings.Contains(base, "italic") || strings.Contains(base, "oblique") {
					continue
				}
				f := parseFontFile(p)
				if f == nil {
					continue
				}
				if !seenFamilies[base] {
					seenFamilies[base] = true
					out = append(out, f)
				}
			}
			return out
		}

		linuxRegularFonts = loadOrdered(false)
		linuxBoldFonts = loadOrdered(true)
		if len(linuxBoldFonts) == 0 {
			linuxBoldFonts = linuxRegularFonts
		}
	})
}

func parseFontFile(path string) *opentype.Font {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}
	if strings.EqualFold(filepath.Ext(path), ".ttc") {
		col, err := opentype.ParseCollection(data)
		if err != nil || col.NumFonts() == 0 {
			return nil
		}
		f, err := col.Font(0)
		if err != nil {
			return nil
		}
		return f
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return nil
	}
	return f
}

type compositeFontFace struct {
	faces   []font.Face
	fonts   []*opentype.Font
	metrics font.Metrics
}

func (c *compositeFontFace) Close() error { return nil }

func (c *compositeFontFace) Metrics() font.Metrics {
	return c.metrics
}

func (c *compositeFontFace) faceForRune(r rune) font.Face {
	if len(c.faces) == 0 {
		return nil
	}
	var buf sfnt.Buffer
	for i, f := range c.faces {
		if i < len(c.fonts) && c.fonts[i] != nil {
			if idx, err := c.fonts[i].GlyphIndex(&buf, r); err == nil && idx != 0 {
				return f
			}
			continue
		}
		if _, ok := f.GlyphAdvance(r); ok {
			return f
		}
	}
	return c.faces[0]
}

func (c *compositeFontFace) Glyph(dot fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	f := c.faceForRune(r)
	if f == nil {
		return image.Rectangle{}, nil, image.Point{}, 0, false
	}
	return f.Glyph(dot, r)
}

func (c *compositeFontFace) GlyphBounds(r rune) (fixed.Rectangle26_6, fixed.Int26_6, bool) {
	f := c.faceForRune(r)
	if f == nil {
		return fixed.Rectangle26_6{}, 0, false
	}
	return f.GlyphBounds(r)
}

func (c *compositeFontFace) GlyphAdvance(r rune) (fixed.Int26_6, bool) {
	f := c.faceForRune(r)
	if f == nil {
		return 0, false
	}
	return f.GlyphAdvance(r)
}

func (c *compositeFontFace) Kern(r0, r1 rune) fixed.Int26_6 {
	if len(c.faces) == 0 {
		return 0
	}
	return c.faces[0].Kern(r0, r1)
}

// Wait: let's check opentype.Face type assertion cleanly via interface or sfnt.Font directly.
func cachedCompositeFace(fontHeight, weight int) *compositeFontFace {
	loadLinuxFontChain()
	bold := weight >= 600
	var bbit uint64
	if bold {
		bbit = 1
	}
	key := uint64(uint32(fontHeight))<<1 | bbit

	linuxFaceMu.Lock()
	defer linuxFaceMu.Unlock()
	if f, ok := linuxFaceCache[key]; ok {
		return f
	}
	sourceList := linuxRegularFonts
	if bold {
		sourceList = linuxBoldFonts
	}
	if len(sourceList) == 0 {
		return nil
	}
	sizePx := float64(fontHeight) * 0.84
	if sizePx < 6 {
		sizePx = 6
	}
	var faces []font.Face
	var fonts []*opentype.Font
	for _, ot := range sourceList {
		face, err := opentype.NewFace(ot, &opentype.FaceOptions{
			Size:    sizePx,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err == nil {
			faces = append(faces, face)
			fonts = append(fonts, ot)
		}
	}
	if len(faces) == 0 {
		return nil
	}
	// Pair source fonts directly inside wrapper so GlyphIndex works without private fields.
	wrapped := &compositeFontFace{
		faces:   faces,
		fonts:   fonts,
		metrics: faces[0].Metrics(),
	}
	linuxFaceCache[key] = wrapped
	return wrapped
}

func (b *canvasBackend) measureText(value string, fontHeight, weight int) int {
	face := cachedCompositeFace(fontHeight, weight)
	if face == nil {
		count := 0
		for range value {
			count++
		}
		return count * maxInt(fontHeight/2, 1)
	}
	return font.MeasureString(face, value).Ceil()
}

func (b *canvasBackend) drawText(target *canvas, left, top, width, height int, value string, fontHeight, weight int, color rgba, flags uint32) bool {
	face := cachedCompositeFace(fontHeight, weight)
	if face == nil || width <= 0 || height <= 0 {
		return false
	}

	singleLine := flags&dtSingleLine != 0
	wordBreak := flags&dtWordBreak != 0
	endEllipsis := flags&dtEndEllipsis != 0

	var lines []string
	if singleLine || !wordBreak {
		text := strings.ReplaceAll(value, "\r\n", " ")
		text = strings.ReplaceAll(text, "\n", " ")
		if endEllipsis && font.MeasureString(face, text).Ceil() > width {
			text = fitLinuxEllipsis(face, text, width)
		}
		lines = []string{text}
	} else {
		lines = wrapLinuxLines(face, value, width)
	}
	if len(lines) == 0 {
		return false
	}

	metrics := face.Metrics()
	lineHeight := metrics.Height.Ceil()
	if lineHeight < fontHeight {
		lineHeight = fontHeight
	}
	ascent := metrics.Ascent.Ceil()
	totalHeight := lineHeight * len(lines)
	startY := top
	if flags&dtVCenter != 0 && totalHeight < height {
		startY = top + (height-totalHeight)/2
	}

	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	drawer := font.Drawer{
		Dst:  mask,
		Src:  image.Opaque,
		Face: face,
	}
	for i, line := range lines {
		lineY := startY - top + i*lineHeight + ascent
		if lineY < 0 || lineY-ascent >= height {
			continue
		}
		lineWidth := font.MeasureString(face, line).Ceil()
		lineX := 0
		switch {
		case flags&dtCenter != 0:
			lineX = (width - lineWidth) / 2
		case flags&dtRight != 0:
			lineX = width - lineWidth
		}
		drawer.Dot = fixed.P(lineX, lineY)
		drawer.DrawString(line)
	}

	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			cov := mask.AlphaAt(px, py).A
			if cov == 0 {
				continue
			}
			a := uint8((uint32(color.A)*uint32(cov) + 127) / 255)
			if a == 0 {
				continue
			}
			target.blend(left+px, top+py, rgba{R: color.R, G: color.G, B: color.B, A: a})
		}
	}
	return true
}

func fitLinuxEllipsis(face font.Face, value string, maxWidth int) string {
	const ellipsis = "…"
	runes := []rune(value)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + ellipsis
		if font.MeasureString(face, candidate).Ceil() <= maxWidth {
			return candidate
		}
	}
	return ""
}

func wrapLinuxLines(face font.Face, value string, maxWidth int) []string {
	var out []string
	for _, para := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
		words := splitWrapTokens(para)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		for _, token := range words {
			candidate := token
			if cur != "" && needsSpaceJoin(cur, token) {
				candidate = cur + " " + token
			} else if cur != "" {
				candidate = cur + token
			}
			if cur == "" || font.MeasureString(face, candidate).Ceil() <= maxWidth {
				cur = candidate
			} else {
				out = append(out, cur)
				cur = token
			}
		}
		if cur != "" {
			out = append(out, cur)
		}
	}
	return out
}

func splitWrapTokens(line string) []string {
	var tokens []string
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}
	}
	for _, r := range line {
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		if r > 0x2E7F { // CJK and wide symbols break per rune.
			flush()
			tokens = append(tokens, string(r))
			continue
		}
		buf.WriteRune(r)
	}
	flush()
	return tokens
}

func needsSpaceJoin(prev, next string) bool {
	if prev == "" || next == "" {
		return false
	}
	last := rune(prev[len(prev)-1])
	first := []rune(next)[0]
	return last <= 0x7F && first <= 0x7F
}
