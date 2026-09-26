package ui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/session"
)

// TestRenderPreview rasterises the panel to PNG so the layout can be reviewed
// without a live window. Opt-in: set PET_PREVIEW_DIR to a directory and run
//
//	go test ./internal/ui -run TestRenderPreview
//
// It is a development aid, not an assertion - the point is to be able to look
// at the picture after changing geometry, which the probe cannot do.
func TestRenderPreview(t *testing.T) {
	dir := os.Getenv("PET_PREVIEW_DIR")
	if dir == "" {
		t.Skip("set PET_PREVIEW_DIR to render panel previews")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	desktop := color.RGBA{R: 0x20, G: 0x22, B: 0x2b, A: 0xff}

	waiting := []session.Session{
		{Source: "claude-hook", SessionID: "a", Project: `C:\Users\me\work\ai-terminal-pet`, State: session.StateWaitingPermission, Reason: "Bash(go test)", UpdatedAt: now.Add(-2 * time.Minute)},
		{Source: "opencode", SessionID: "b", Project: `C:\Users\me\work\webapp`, State: session.StateWorking, Reason: "editing src/app.tsx", UpdatedAt: now.Add(-6 * time.Second)},
		{Source: "antigravity-statusline", SessionID: "c", Project: `C:\Users\me\work\docs`, State: session.StateThinking, UpdatedAt: now.Add(-40 * time.Second)},
	}
	busy := []session.Session{
		{Source: "opencode", SessionID: "b", Project: `C:\Users\me\work\webapp`, State: session.StateWorking, Reason: "editing src/app.tsx", UpdatedAt: now.Add(-6 * time.Second)},
		{Source: "opencode", SessionID: "d", Project: `C:\Users\me\work\api`, State: session.StateToolUse, Reason: "running migrations", UpdatedAt: now.Add(-70 * time.Second)},
	}
	broken := []session.Session{
		{Source: "claude-hook", SessionID: "e", Project: `C:\Users\me\work\infra`, State: session.StateError, Reason: "exit status 1", UpdatedAt: now.Add(-11 * time.Minute)},
	}
	many := make([]session.Session, 0, 9)
	for index := 0; index < 9; index++ {
		state := session.StateWorking
		if index == 1 {
			state = session.StateWaitingPermission
		}
		many = append(many, session.Session{
			Source:    "opencode",
			SessionID: string(rune('a' + index)),
			Project:   `C:\Users\me\work\proj-` + string(rune('a'+index)),
			State:     state,
			Reason:    "compiling",
			UpdatedAt: now.Add(-time.Duration(index) * time.Minute),
		})
	}

	cases := []struct {
		name     string
		items    []session.Session
		expanded bool
		pressed  bool
		hovered  bool
		scale    int
	}{
		{"compact-waiting", waiting, false, false, false, 3},
		{"compact-waiting-hover", waiting, false, false, true, 3},
		{"compact-busy", busy, false, false, false, 3},
		{"compact-error", broken, false, false, false, 3},
		{"compact-empty", nil, false, false, false, 3},
		{"expanded-waiting", waiting, true, false, false, 2},
		{"expanded-many", many, true, false, false, 2},
		{"expanded-empty", nil, true, false, false, 2},
	}
	for _, tc := range cases {
		width := compactWidth
		height := compactHeight
		if tc.expanded {
			width = expandedWidth
			height = expandedPanelHeight(len(tc.items))
		}
		target, err := newCanvas(width, height)
		if err != nil {
			t.Fatalf("%s: canvas: %v", tc.name, err)
		}
		drawVerticalPanel(target, tc.items, now, tc.expanded, tc.pressed, tc.hovered)
		img := compositeOver(target, desktop)
		target.close()
		path := filepath.Join(dir, tc.name+".png")
		writePNG(t, path, upscale(img, tc.scale))
		t.Logf("wrote %s (%dx%d)", path, width, height)
	}
}

func compositeOver(c *canvas, background color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			offset := y*c.stride + x*4
			alpha := float64(c.pixels[offset+3]) / 255
			inverse := 1 - alpha
			blend := func(source byte, destination byte) uint8 {
				value := float64(source) + float64(destination)*inverse
				if value > 255 {
					value = 255
				}
				return uint8(value + 0.5)
			}
			img.Set(x, y, color.RGBA{
				R: blend(c.pixels[offset+2], background.R),
				G: blend(c.pixels[offset+1], background.G),
				B: blend(c.pixels[offset], background.B),
				A: 255,
			})
		}
	}
	return img
}

func upscale(source *image.RGBA, factor int) *image.RGBA {
	if factor <= 1 {
		return source
	}
	bounds := source.Bounds()
	result := image.NewRGBA(image.Rect(0, 0, bounds.Dx()*factor, bounds.Dy()*factor))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			value := source.RGBAAt(x, y)
			for dy := 0; dy < factor; dy++ {
				for dx := 0; dx < factor; dx++ {
					result.Set(x*factor+dx, y*factor+dy, value)
				}
			}
		}
	}
	return result
}

func writePNG(t *testing.T, path string, img *image.RGBA) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}
