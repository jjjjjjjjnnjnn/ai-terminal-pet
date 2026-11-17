//go:build windows

package ui

// The click test needs *App, and App is per-platform: app_windows.go carries the
// settings window's state, app_linux.go and app_darwin.go do not. The layout,
// hit-test and drawing tests stay in the untagged file so they run everywhere;
// only the "does clicking the link reach the opener" half is Windows-only.
//
// It is worth keeping even though the window is Windows-only: the link is drawn
// from shared code and clicked from Windows code, and this is the seam between
// them.

import (
	"testing"

	"ai-terminal-pet/internal/i18n"
	"ai-terminal-pet/internal/project"
)

// TestFooterClickOpensTheRightURL is the assertion that makes the footer worth
// having: the control the hit test returns must reach the opener with the URL the
// label promises, and a nil opener must be a no-op rather than a crash.
func TestFooterClickOpensTheRightURL(t *testing.T) {
	var opened []string
	a := &App{options: Options{
		OpenExternal: func(url string) error {
			opened = append(opened, url)
			return nil
		},
	}}
	a.settingsGeometry = settingsGeometryFor(96, 0, settingsRowsIn(t, 0, i18n.LangEn))
	a.settingsApply(settingsStar, -1)
	a.settingsApply(settingsIssue, -1)
	if len(opened) != 2 {
		t.Fatalf("opened %d links, want 2: %v", len(opened), opened)
	}
	if opened[0] != project.RepoURL {
		t.Fatalf("the star link opened %q, want %q", opened[0], project.RepoURL)
	}
	if opened[1] != project.IssuesURL {
		t.Fatalf("the issue link opened %q, want %q", opened[1], project.IssuesURL)
	}

	// No opener configured: nothing happens, and in particular nothing panics.
	quiet := &App{}
	quiet.settingsApply(settingsStar, -1)
	quiet.settingsApply(settingsIssue, -1)
}
