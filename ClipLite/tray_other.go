//go:build !windows

package main

type SystemTray struct {
	app      *App
	iconPath string
}

func NewSystemTray(app *App) *SystemTray {
	return &SystemTray{app: app}
}

func (t *SystemTray) Start() {}

func (t *SystemTray) Stop() {}
