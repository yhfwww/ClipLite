//go:build !windows

package main

import "fmt"

type GlobalHotkey struct {
	app *App
}

func NewGlobalHotkey(app *App) *GlobalHotkey {
	return &GlobalHotkey{app: app}
}

func (h *GlobalHotkey) Register(hotkeyStr string) error {
	fmt.Printf("快捷键功能仅在 Windows 上可用: %s\n", hotkeyStr)
	return nil
}

func (h *GlobalHotkey) Unregister() {}
