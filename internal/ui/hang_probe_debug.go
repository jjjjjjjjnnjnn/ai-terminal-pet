//go:build petdebug

package ui

import (
	"os"
	"strconv"
	"time"
)

// The diagnostic build's stall probe.
//
// Proving that the new clock still catches a real stall needs a real stall:
// QueryUnbiasedInterruptTime cannot be made to lie, and suspending the machine
// from a test is not an option. So this build lets the refresh handler block on
// purpose, and the watchdog then has to dump it.
//
//	PET_HANG_PROBE_MS=10000 go build -tags petdebug -o petdebug.exe ./cmd/ai-terminal-pet
//
// The variable is read once per call because this path only exists in a build
// nobody ships; the release build returns a constant zero and never looks.
const hangProbeEnvVar = "PET_HANG_PROBE_MS"

func (a *App) probeStallDelay() time.Duration {
	milliseconds, err := strconv.Atoi(os.Getenv(hangProbeEnvVar))
	if err != nil || milliseconds <= 0 {
		return 0
	}
	if milliseconds > 60_000 {
		milliseconds = 60_000
	}
	return time.Duration(milliseconds) * time.Millisecond
}
