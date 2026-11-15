package ui

import "math"

// Pure Cocoa/AppKit geometry, pixel, and key-translation helpers used by
// app_darwin.go. Keeping them untagged lets unit tests verify the math on
// Windows and Linux as well as macOS.

type nsPoint struct {
	X float64
	Y float64
}

type nsSize struct {
	Width  float64
	Height float64
}

type nsRect struct {
	Origin nsPoint
	Size   nsSize
}

const (
	darwinKeyCodeReturn    uint16 = 0x24
	darwinKeyCodeEnter     uint16 = 0x4C
	darwinKeyCodeEscape    uint16 = 0x35
	darwinKeyCodeHome      uint16 = 0x73
	darwinKeyCodeEnd       uint16 = 0x77
	darwinKeyCodeDownArrow uint16 = 0x7D
	darwinKeyCodeUpArrow   uint16 = 0x7E
	darwinKeyCodeEqual     uint16 = 0x18
	darwinKeyCodeMinus     uint16 = 0x1B
	darwinKeyCodeZero      uint16 = 0x1D
	darwinKeyCodePadPlus   uint16 = 0x45
	darwinKeyCodePadMinus  uint16 = 0x4E
	darwinKeyCodePadZero   uint16 = 0x52
)

// cocoaScreenToDisplayRect converts a Cocoa NSScreen rect (bottom-left origin)
// into a top-left Y-down displayRect using the primary screen's top edge.
func cocoaScreenToDisplayRect(frame nsRect, primaryTop float64) displayRect {
	left := int32(math.Round(frame.Origin.X))
	top := int32(math.Round(primaryTop - (frame.Origin.Y + frame.Size.Height)))
	width := int32(math.Round(frame.Size.Width))
	height := int32(math.Round(frame.Size.Height))
	return displayRect{
		Left:   left,
		Top:    top,
		Right:  left + width,
		Bottom: top + height,
	}
}

// topLeftToCocoaOrigin converts a top-left panel origin into Cocoa bottom-left
// window frame origin coordinates.
func topLeftToCocoaOrigin(x, y int32, height int, primaryTop float64) nsPoint {
	return nsPoint{
		X: float64(x),
		Y: primaryTop - float64(y) - float64(height),
	}
}

// cocoaPointToTopLeft converts a global Cocoa screen point (NSEvent.mouseLocation)
// into top-left screen coordinates matching Placement and dragState.
func cocoaPointToTopLeft(x, y, primaryTop float64) (int32, int32) {
	return int32(math.Round(x)), int32(math.Round(primaryTop - y))
}

// cocoaWindowPointToCanvas maps NSEvent.locationInWindow (bottom-left window
// coordinates) into top-left canvas pixel coordinates.
func cocoaWindowPointToCanvas(x, y float64, width, height int) (int, int) {
	if width <= 0 || height <= 0 {
		return 0, 0
	}
	cx := int(math.Floor(x))
	cy := height - 1 - int(math.Floor(y))
	return cx, cy
}

// bgraStraightToPremulRGBA converts canvas straight-alpha BGRA pixels into
// premultiplied RGBA bytes for Cocoa NSBitmapImageRep.
func bgraStraightToPremulRGBA(src []byte) []byte {
	out := make([]byte, len(src))
	for i := 0; i+3 < len(src); i += 4 {
		a := uint32(src[i+3])
		switch a {
		case 0:
			// Already zeroed.
		case 255:
			out[i] = src[i+2]   // R
			out[i+1] = src[i+1] // G
			out[i+2] = src[i]   // B
			out[i+3] = 255
		default:
			out[i] = uint8((uint32(src[i+2])*a + 127) / 255)
			out[i+1] = uint8((uint32(src[i+1])*a + 127) / 255)
			out[i+2] = uint8((uint32(src[i])*a + 127) / 255)
			out[i+3] = uint8(a)
		}
	}
	return out
}

// darwinKeyCodeToPanelKey maps a macOS virtual keyCode onto panel keyboard nav.
func darwinKeyCodeToPanelKey(keyCode uint16) panelKey {
	switch keyCode {
	case darwinKeyCodeEscape:
		return panelKeyEscape
	case darwinKeyCodeUpArrow:
		return panelKeyUp
	case darwinKeyCodeDownArrow:
		return panelKeyDown
	case darwinKeyCodeHome:
		return panelKeyHome
	case darwinKeyCodeEnd:
		return panelKeyEnd
	case darwinKeyCodeReturn, darwinKeyCodeEnter:
		return panelKeyEnter
	default:
		return panelKeyNone
	}
}

// darwinZoomAction checks Cmd+Plus / Cmd+Minus / Cmd+0 zoom shortcuts.
// Returns (delta, reset, handled).
func darwinZoomShortcut(keyCode uint16, commandHeld bool) (float64, bool, bool) {
	if !commandHeld {
		return 0, false, false
	}
	switch keyCode {
	case darwinKeyCodeEqual, darwinKeyCodePadPlus:
		return 0.1, false, true
	case darwinKeyCodeMinus, darwinKeyCodePadMinus:
		return -0.1, false, true
	case darwinKeyCodeZero, darwinKeyCodePadZero:
		return 0, true, true
	default:
		return 0, false, false
	}
}
