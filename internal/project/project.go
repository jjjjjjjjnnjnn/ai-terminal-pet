// Package project holds the outward-facing facts about this project that more
// than one surface has to print: the links a user is sent to, and nothing else.
//
// It exists because the settings surfaces are three (the Win32 window, the
// terminal editor, and the README) and a URL typed three times is a URL that
// eventually differs in one of them. Version lives in internal/app, next to the
// constant the binary reports; these are the two links that point at where the
// project actually lives.
package project

const (
	// RepoURL is the repository: the star button, the README, the releases.
	RepoURL = "https://github.com/jjjjjjjjnnjnn/ai-terminal-pet"
	// IssuesURL opens a new issue directly, so the footer's second link is one
	// click from a report rather than a click and a search.
	IssuesURL = RepoURL + "/issues/new"
)
