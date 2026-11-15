//go:build linux

package ui

import (
	"os"
	"path/filepath"
	"strings"
)

func openTypeFontDirs() []string {
	homeDir, _ := os.UserHomeDir()
	var dirs []string
	dirs = append(dirs, "/usr/share/fonts", "/usr/local/share/fonts")
	if homeDir != "" {
		dirs = append(dirs, filepath.Join(homeDir, ".local", "share", "fonts"), filepath.Join(homeDir, ".fonts"))
	}
	// WSL2 mounts host Windows fonts here when Linux lacks CJK fonts.
	dirs = append(dirs, "/mnt/c/Windows/Fonts")
	return dirs
}

func scoreOpenTypeFont(path string, bold bool) int {
	lower := strings.ToLower(filepath.Base(path))
	score := 100
	switch {
	case strings.Contains(lower, "ubuntusans") && !strings.Contains(lower, "mono"):
		score -= 40
	case strings.Contains(lower, "dejavusans") && !strings.Contains(lower, "mono"):
		score -= 35
	case strings.Contains(lower, "notosans") && !strings.Contains(lower, "mono") && !strings.Contains(lower, "cjk"):
		score -= 38
	case strings.Contains(lower, "liberationsans"):
		score -= 30
	case strings.Contains(lower, "segoeui"):
		score -= 20
	case strings.Contains(lower, "notosanscjk") || strings.Contains(lower, "wenquanyi") || strings.Contains(lower, "msyh") || strings.Contains(lower, "simhei"):
		score -= 10
	case strings.Contains(lower, "mono") || strings.Contains(lower, "serif"):
		score += 40
	}
	if strings.Contains(lower, "italic") || strings.Contains(lower, "oblique") {
		score += 80
	}
	hasBold := strings.Contains(lower, "bold") || strings.Contains(lower, "semibold") || strings.Contains(lower, "demibold")
	if bold && hasBold {
		score -= 15
	} else if !bold && hasBold {
		score += 25
	}
	if strings.HasPrefix(path, "/mnt/c/") {
		score += 50 // Prefer native Linux fonts before WSL Windows fallback
	}
	return score
}
