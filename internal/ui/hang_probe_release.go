//go:build windows && !petdebug

package ui

import "time"

// The shipping build: a refresh never blocks on purpose, so the stall probe is
// a constant. See hang_probe_debug.go for the diagnostic version and hang.go for
// why the watchdog's clock changed at all.
func (a *App) probeStallDelay() time.Duration { return 0 }
