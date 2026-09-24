package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"ai-terminal-pet/internal/app"
)

func main() {
	debug.SetGCPercent(50)
	if err := app.Run(); err != nil {
		writeError(err)
		os.Exit(1)
	}
}

func writeError(value error) {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" || value == nil {
		return
	}
	directory := filepath.Join(root, "AI-Terminal-Pet")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return
	}
	path := filepath.Join(directory, "error.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	if info, statErr := file.Stat(); statErr == nil && info.Size() > 1<<20 {
		_ = file.Close()
		_ = os.Remove(path)
		file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return
		}
		defer file.Close()
	}
	_, _ = file.WriteString(time.Now().Format(time.RFC3339) + " " + value.Error() + "\n")
}
