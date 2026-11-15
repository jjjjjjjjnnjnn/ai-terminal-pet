package ui

import (
	"fmt"
	"sort"
	"strings"
)

// LayoutPosition is one remembered (x, y) origin for a specific display layout.
type LayoutPosition struct {
	X int32
	Y int32
}

// Placement is the window geometry remembered between launches. The ui
// package deliberately does not import internal/config: the app layer owns the
// file paths and hands the UI a plain struct plus a persist callback.
type Placement struct {
	Placed   bool
	X        int32
	Y        int32
	Zoom     float64
	Expanded bool
	Layouts  map[string]LayoutPosition
}

// displayRect is one monitor's usable work area in virtual-desktop coordinates.
type displayRect struct {
	Left    int32
	Top     int32
	Right   int32
	Bottom  int32
	Primary bool
}

func (d displayRect) valid() bool {
	return d.Right > d.Left && d.Bottom > d.Top
}

func (d displayRect) contains(x, y int32) bool {
	return x >= d.Left && x < d.Right && y >= d.Top && y < d.Bottom
}

// layoutSignature builds a deterministic key from the active monitors' work
// areas so plugging or unplugging a display switches between saved origins
// without clobbering the disconnected configuration.
func layoutSignature(displays []displayRect) string {
	valid := make([]displayRect, 0, len(displays))
	for _, d := range displays {
		if d.valid() {
			valid = append(valid, d)
		}
	}
	if len(valid) == 0 {
		return ""
	}
	sort.Slice(valid, func(i, j int) bool {
		if valid[i].Left != valid[j].Left {
			return valid[i].Left < valid[j].Left
		}
		if valid[i].Top != valid[j].Top {
			return valid[i].Top < valid[j].Top
		}
		if valid[i].Right != valid[j].Right {
			return valid[i].Right < valid[j].Right
		}
		return valid[i].Bottom < valid[j].Bottom
	})
	parts := make([]string, len(valid))
	for i, d := range valid {
		parts[i] = fmt.Sprintf("%d,%d,%d,%d", d.Left, d.Top, d.Right, d.Bottom)
	}
	return strings.Join(parts, ";")
}

func primaryDisplay(displays []displayRect) (displayRect, bool) {
	var first displayRect
	hasFirst := false
	for _, d := range displays {
		if !d.valid() {
			continue
		}
		if d.Primary {
			return d, true
		}
		if !hasFirst {
			first = d
			hasFirst = true
		}
	}
	return first, hasFirst
}

func bestDisplayForPoint(displays []displayRect, cx, cy int32) (displayRect, bool) {
	var best displayRect
	var bestDist int64
	hasBest := false
	for _, d := range displays {
		if !d.valid() {
			continue
		}
		if d.contains(cx, cy) {
			return d, true
		}
		dx := distanceToInterval(cx, d.Left, d.Right)
		dy := distanceToInterval(cy, d.Top, d.Bottom)
		dist := dx*dx + dy*dy
		if !hasBest || dist < bestDist {
			best = d
			bestDist = dist
			hasBest = true
		}
	}
	return best, hasBest
}

func distanceToInterval(value, minVal, maxVal int32) int64 {
	if value < minVal {
		return int64(minVal - value)
	}
	if value >= maxVal {
		return int64(value - maxVal + 1)
	}
	return 0
}

func clampToDisplay(work displayRect, margin, width, height, x, y int32) (int32, int32) {
	minimumX := work.Left + margin
	minimumY := work.Top + margin
	maximumX := work.Right - width - margin
	maximumY := work.Bottom - height - margin
	outX, outY := x, y
	if maximumX < minimumX {
		outX = work.Left
	} else {
		outX = max32(minimumX, min32(outX, maximumX))
	}
	if maximumY < minimumY {
		outY = work.Top
	} else {
		outY = max32(minimumY, min32(outY, maximumY))
	}
	return outX, outY
}

// resolveInitialPlacement chooses the starting (x, y) for the current display
// configuration: an exact layout entry wins, then the global remembered
// position clamped onto the nearest display, then the primary right-edge
// anchor on a first run or --reset.
func resolveInitialPlacement(initial Placement, displays []displayRect, width, height int32, dpi uint32) (int32, int32, bool) {
	if len(displays) == 0 {
		return 0, 0, false
	}
	margin := int32(8 * dpi / 96)
	sig := layoutSignature(displays)
	if sig != "" && len(initial.Layouts) > 0 {
		if saved, ok := initial.Layouts[sig]; ok {
			if target, has := bestDisplayForPoint(displays, saved.X+width/2, saved.Y+height/2); has {
				cx, cy := clampToDisplay(target, margin, width, height, saved.X, saved.Y)
				return cx, cy, true
			}
		}
	}
	if initial.Placed {
		if target, has := bestDisplayForPoint(displays, initial.X+width/2, initial.Y+height/2); has {
			cx, cy := clampToDisplay(target, margin, width, height, initial.X, initial.Y)
			return cx, cy, true
		}
	}
	primary, ok := primaryDisplay(displays)
	if !ok {
		return 0, 0, false
	}
	anchorX := primary.Right - width - int32(24*dpi/96)
	anchorY := primary.Top + (primary.Bottom-primary.Top-height)/2
	cx, cy := clampToDisplay(primary, margin, width, height, anchorX, anchorY)
	return cx, cy, true
}

// cloneLayoutsWithCurrent copies all saved display-layout entries (preserving
// disconnected monitors) and updates the entry for the active layout signature.
func cloneLayoutsWithCurrent(existing map[string]LayoutPosition, sig string, x, y int32) map[string]LayoutPosition {
	if sig == "" && len(existing) == 0 {
		return nil
	}
	out := make(map[string]LayoutPosition, len(existing)+1)
	for k, v := range existing {
		if strings.TrimSpace(k) != "" {
			out[k] = v
		}
	}
	if sig != "" {
		out[sig] = LayoutPosition{X: x, Y: y}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func min32(left, right int32) int32 {
	if left < right {
		return left
	}
	return right
}

func max32(left, right int32) int32 {
	if left > right {
		return left
	}
	return right
}
