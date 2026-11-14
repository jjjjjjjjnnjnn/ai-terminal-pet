//go:build !petdebug

package ui

// The shipping build: nothing is traced, because there is no repro to serve and
// one log line per exit decision would be a file write the documentation says
// does not happen. See exit_trace.go for why only a diagnostic build writes.
func traceExit(string) {}

// traceSettings is the settings-window half of the same contract: the debug
// build records open/hide, the shipping build does not.
func traceSettings(string) {}
