package main

import (
	"fmt"
	"sync"
	"time"
)

type ClipboardMonitor struct {
	app         *App
	running     bool
	mu          sync.Mutex
	lastContent string
	stopChan    chan struct{}
}

func NewClipboardMonitor(app *App) *ClipboardMonitor {
	return &ClipboardMonitor{
		app:      app,
		stopChan: make(chan struct{}),
	}
}

func (m *ClipboardMonitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			m.app.log(fmt.Sprintf("PANIC in clipboard monitor: %v", r))
		}
	}()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			content, err := readClipboard()
			if err != nil {
				continue
			}
			if content != "" {
				m.mu.Lock()
				m.lastContent = content
				m.mu.Unlock()
			}
		case <-m.stopChan:
			return
		}
	}
}

func (m *ClipboardMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		m.running = false
		close(m.stopChan)
	}
}

func (m *ClipboardMonitor) SetClipboard(content string) error {
	m.mu.Lock()
	m.lastContent = content
	m.mu.Unlock()

	return writeClipboard(content)
}

func (m *ClipboardMonitor) GetLastContent() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastContent
}
