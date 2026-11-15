//go:build !windows && !linux && !darwin

package ui

import (
	"errors"

	"ai-terminal-pet/internal/store"
)

// App is a non-window placeholder on platforms without an active native GUI
// layer (macOS currently runs CLI subcommands cleanly).
type App struct {
	values  *store.Store
	options Options
}

func New(values *store.Store, options Options) *App {
	return &App{values: values, options: options}
}

func (a *App) Refresh() {}

func (a *App) Activate() {}

func (a *App) Close() {}

func (a *App) QueueSettings(value Settings) {
	a.options.Settings = value
}

func (a *App) Run() error {
	return errors.New("window layer is unavailable on this platform")
}
