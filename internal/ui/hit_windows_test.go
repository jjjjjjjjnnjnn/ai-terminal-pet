package ui

import "testing"

func TestClassifyHit(t *testing.T) {
	// 4x3 canvas: opaque block at x in [1,3), y in [1,2).
	hit := make([]byte, 4*3)
	hit[1*4+1] = 255
	hit[1*4+2] = 19 // below threshold
	if !classifyHit(1, 1, 4, 3, hit) {
		t.Fatal("opaque pixel must hit")
	}
	if classifyHit(2, 1, 4, 3, hit) {
		t.Fatal("sub-threshold alpha must miss")
	}
	if classifyHit(0, 0, 4, 3, hit) {
		t.Fatal("transparent pixel must miss")
	}
	// Regression: X must be checked against width, not len(hit).
	// Window wider than canvas (zoom-out/DPI-down) used to escape the
	// slice (Y*width+X >= len) and panic the process with no error.log.
	wide, high := 28, 133
	wideHit := make([]byte, wide*high)
	if classifyHit(39, 132, wide, high, wideHit) {
		t.Fatal("out-of-canvas X must miss, not panic")
	}
	if classifyHit(39, 170, wide, high, wideHit) {
		t.Fatal("out-of-canvas corner must miss, not panic")
	}
	if classifyHit(-1, 0, wide, high, wideHit) || classifyHit(0, -1, wide, high, wideHit) {
		t.Fatal("negative coords must miss")
	}
	if classifyHit(0, 0, 0, 0, nil) || classifyHit(0, 0, 4, 3, nil) {
		t.Fatal("empty geometry must miss")
	}
}
