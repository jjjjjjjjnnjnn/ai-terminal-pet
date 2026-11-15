//go:build darwin

package ui

import (
	"os"
	"path/filepath"
	"strings"
)

func openTypeFontDirs() []string {
	dirs := []string{
		"/System/Library/Fonts",
		"/System/Library/Fonts/Supplemental",
		"/Library/Fonts",
	}
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		dirs = append(dirs, filepath.Join(homeDir, "Library", "Fonts"))
	}
	return dirs
}

func scoreOpenTypeFont(path string, bold bool) int {
	lower := strings.ToLower(filepath.Base(path))
	score := 100
	switch {
	case strings.HasPrefix(lower, "sfns") && !strings.Contains(lower, "mono") && !strings.Contains(lower, "rounded") && !strings.Contains(lower, "italic"):
		score -= 45
	case strings.Contains(lower, "sfpro") && !strings.Contains(lower, "mono"):
		score -= 42
	case strings.Contains(lower, "helveticaneue") || lower == "helvetica.ttc":
		score -= 35
	case strings.Contains(lower, "arial") && !strings.Contains(lower, "narrow"):
		score -= 25
	case strings.Contains(lower, "pingfang"):
		score -= 20
	case strings.Contains(lower, "hiragino") && strings.Contains(lower, "sans"):
		score -= 15
	case strings.Contains(lower, "stheiti") || strings.Contains(lower, "notosanscjk"):
		score -= 10
	case strings.Contains(lower, "mono") || strings.Contains(lower, "serif") || strings.Contains(lower, "times"):
		score += 40
	}
	if strings.Contains(lower, "italic") || strings.Contains(lower, "oblique") || strings.Contains(lower, "condensed") {
		score += 80
	}
	hasBold := strings.Contains(lower, "bold") || strings.Contains(lower, "semibold") || strings.Contains(lower, "demibold") || strings.Contains(lower, "medium")
	if bold && hasBold {
		score -= 15
	} else if !bold && hasBold {
		score += 25
	}
	return score
}
