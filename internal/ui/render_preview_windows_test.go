package ui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-terminal-pet/internal/i18n"
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
	idle := []session.Session{
		{Source: "opencode", SessionID: "f", Project: `C:\Users\me\work\notes`, State: session.StateIdle, UpdatedAt: now.Add(-20 * time.Minute)},
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
		tuck     float64
		scale    int
		lang     i18n.Lang
		// approve attaches a live auto-approve countdown to the first
		// waiting session in items, so the Allow (Ns) row + bar renders.
		approve bool
		// crop is how much of the frame is on screen, in logical units. A
		// tucked panel is mostly past the screen edge, and the rasterizer
		// still draws the whole frame - cropping here is what makes the
		// preview show what the user actually sees.
		crop int
		// marks attaches the per-row badges: unread marks on the first row,
		// muted on the second. The two are different claims about a row (it
		// needs you / it does not speak), so they get their own previews.
		marks bool
		// hoverRow is which row the pointer is over, -1 for none.
		hoverRow int
	}{
		{"compact-waiting", waiting, false, false, false, 0, 3, i18n.LangZhCN, false, 0, false, -1},
		{"compact-waiting-hover", waiting, false, false, true, 0, 3, i18n.LangZhCN, false, 0, false, -1},
		{"compact-busy", busy, false, false, false, 0, 3, i18n.LangZhCN, false, 0, false, -1},
		{"compact-error", broken, false, false, false, 0, 3, i18n.LangZhCN, false, 0, false, -1},
		{"compact-empty", nil, false, false, false, 0, 3, i18n.LangZhCN, false, 0, false, -1},
		{"compact-waiting-en", waiting, false, false, false, 0, 3, i18n.LangEn, false, 0, false, -1},
		{"docked-tuck-half", busy, false, false, false, 0.5, 3, i18n.LangZhCN, false, 33, false, -1},
		{"docked-tuck-full", busy, false, false, false, 1, 3, i18n.LangZhCN, false, 10, false, -1},
		{"docked-tuck-full-idle", idle, false, false, false, 1, 3, i18n.LangZhCN, false, 10, false, -1},
		{"expanded-waiting", waiting, true, false, false, 0, 2, i18n.LangZhCN, false, 0, false, -1},
		{"expanded-waiting-en", waiting, true, false, false, 0, 2, i18n.LangEn, false, 0, false, -1},
		{"expanded-approve-zh", waiting, true, false, false, 0, 2, i18n.LangZhCN, true, 0, false, -1},
		{"expanded-approve-en", waiting, true, false, false, 0, 2, i18n.LangEn, true, 0, false, -1},
		{"expanded-many", many, true, false, false, 0, 2, i18n.LangZhCN, false, 0, false, -1},
		{"expanded-empty", nil, true, false, false, 0, 2, i18n.LangZhCN, false, 0, false, -1},
		// The alert layer's two visible states: a row waiting on the user that
		// was announced while they were elsewhere (unread), and a row the user
		// silenced (muted). Neither changes the geometry, so both can be
		// reviewed against the existing expanded previews.
		{"expanded-unread", waiting, true, false, false, 0, 2, i18n.LangZhCN, false, 0, true, -1},
		{"expanded-unread-en", waiting, true, false, false, 0, 2, i18n.LangEn, false, 0, true, -1},
		// A hovered row is the only affordance saying the row itself is a
		// target, which is what makes the per-row context menu discoverable.
		{"expanded-row-hover", waiting, true, false, true, 0, 2, i18n.LangZhCN, false, 0, false, 1},
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
		var countdown map[string]approvalView
		if tc.approve {
			for _, current := range tc.items {
				if current.State == session.StateWaitingPermission {
					countdown = map[string]approvalView{
						approvalSessionKey(current.Source, current.SessionID, current.PID): {RemainingSec: 3, Fraction: 0.6},
					}
					break
				}
			}
		}
		drawVerticalPanel(target, panelState{
			Items:     tc.items,
			Now:       now,
			Expanded:  tc.expanded,
			Pressed:   tc.pressed,
			Hovered:   tc.hovered,
			Lang:      tc.lang,
			Countdown: countdown,
			Unread:    unreadFor(tc.items, tc.marks),
			Muted:     mutedFor(tc.items, tc.marks),
			HoverRow:  tc.hoverRow,
			Tuck:      tc.tuck,
			Sliver:    dockedSliver,
			Dock:      dockRight,
		})
		img := compositeOver(target, desktop)
		target.close()
		if tc.crop > 0 && tc.crop < width {
			img = cropImage(img, tc.crop)
		}
		path := filepath.Join(dir, tc.name+".png")
		writePNG(t, path, upscale(img, tc.scale))
		t.Logf("wrote %s (%dx%d)", path, width, height)
	}
}

// unreadFor and mutedFor attach the per-row marks to a preview. Only the first
// two rows are marked, because a preview that badges every row would not show
// what a single mark looks like next to an unmarked neighbour - which is the
// thing the eye has to be able to tell apart on the real panel.
func unreadFor(items []session.Session, marks bool) *unreadSet {
	if !marks {
		return nil
	}
	set := newUnreadSet()
	if len(items) > 0 {
		set.mark([]string{session.Identity(items[0])})
	}
	return set
}

func mutedFor(items []session.Session, marks bool) *muteSet {
	if !marks {
		return nil
	}
	set := newMuteSet(nil)
	if len(items) > 1 {
		set.set(items[1], true)
	}
	return set
}

// cropImage keeps the leftmost logicalWidth columns, which is the part of a
// right-docked frame that is still on screen.
func cropImage(source *image.RGBA, logicalWidth int) *image.RGBA {
	width := minInt(logicalWidth, source.Bounds().Dx())
	return source.SubImage(image.Rect(0, 0, width, source.Bounds().Dy())).(*image.RGBA)
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
