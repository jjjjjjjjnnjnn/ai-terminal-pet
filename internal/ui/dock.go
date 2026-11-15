package ui

import (
	"math"
	"time"
)

// dockTolerance is how close to a work-area edge counts as hugging it. It is
// generous on purpose: the default anchor leaves a margin, and a user who
// drags the strip up against an edge means to dock it there.
const dockTolerance = 48

// dockedSliver is how much of the panel stays on screen when it is fully
// tucked, in logical units at 96 DPI. Narrow enough to be out of the way, wide
// enough to aim at without hunting for it.
const dockedSliver = 10

const (
	slideOutDuration = 160 * time.Millisecond
	slideInDuration  = 240 * time.Millisecond
)

// dockedOrigin is where a docked frame sits when it is fully out: the gap it
// settled at, measured from the edge it hugs.
func dockedOrigin(dock dockSide, edge, gap int32, width int) int32 {
	switch dock {
	case dockRight:
		return edge - gap - int32(width)
	case dockLeft:
		return edge + gap
	}
	return 0
}

// tuckedOrigin is where a frame sits when it has slid all the way into its
// edge, leaving exactly sliver units on screen.
func tuckedOrigin(dock dockSide, edge int32, width, sliver int) int32 {
	switch dock {
	case dockRight:
		return edge - int32(sliver)
	case dockLeft:
		return edge - int32(width-sliver)
	}
	return 0
}

// slideBetween interpolates the frame origin partway through the dock slide.
func slideBetween(untucked, tucked int32, tuck float64) int32 {
	switch {
	case tuck <= 0:
		return untucked
	case tuck >= 1:
		return tucked
	}
	return untucked + int32(math.Round(float64(tucked-untucked)*tuck))
}

// slideTuck is the tuck value after elapsed time of a slide from -> to, easing
// out so the panel leaves promptly and settles gently.
func slideTuck(from, to float64, elapsed time.Duration) float64 {
	duration := slideOutDuration
	if to > from {
		duration = slideInDuration
	}
	if elapsed <= 0 {
		return from
	}
	if elapsed >= duration {
		return to
	}
	progress := float64(elapsed) / float64(duration)
	eased := 1 - math.Pow(1-progress, 3)
	return from + (to-from)*eased
}
