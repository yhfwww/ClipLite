//go:build windows

package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	hotkeyUser32      = syscall.NewLazyDLL("user32.dll")
	hotkeyKernel32    = syscall.NewLazyDLL("kernel32.dll")
	registerHotKey    = hotkeyUser32.NewProc("RegisterHotKey")
	unregisterHotKey  = hotkeyUser32.NewProc("UnregisterHotKey")
	hotkeyGetModule   = hotkeyKernel32.NewProc("GetModuleHandleW")
)

const (
	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_WIN      = 0x0008
	WM_HOTKEY    = 0x0312
)

type GlobalHotkey struct {
	hwnd     syscall.Handle
	hotkeyID int
	app      *App
	running  bool
	hInstance uintptr
}

func NewGlobalHotkey(app *App) *GlobalHotkey {
	return &GlobalHotkey{
		app:      app,
		hotkeyID: 1,
	}
}

func (h *GlobalHotkey) Register(hotkeyStr string) error {
	modifiers, vk, err := parseHotkey(hotkeyStr)
	if err != nil {
		return err
	}

	if h.running {
		unregisterHotKey.Call(uintptr(h.hwnd), uintptr(h.hotkeyID))
	}

	h.hInstance, _, _ = hotkeyGetModule.Call(0)
	h.app.log(fmt.Sprintf("[Hotkey] 模块句柄: 0x%x", h.hInstance))

	className, _ := syscall.UTF16PtrFromString("ClipLiteHotkeyClass")
	windowName, _ := syscall.UTF16PtrFromString("ClipLiteHotkey")

	wc := WNDCLASSEX{
		WndProc:   syscall.NewCallback(h.wndProc),
		Instance:  syscall.Handle(h.hInstance),
		ClassName: className,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))

	ret, _, err1 := registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		h.app.log(fmt.Sprintf("[Hotkey] 注册窗口类失败(可能已注册): %v", err1))
	}

	hwnd, _, err2 := createWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0,
		0, 0, 0, 0,
		0, 0, uintptr(h.hInstance), 0,
	)
	if hwnd == 0 {
		h.app.log(fmt.Sprintf("[Hotkey] 创建窗口失败: %v", err2))
		return fmt.Errorf("创建快捷键窗口失败: %v", err2)
	}
	h.hwnd = syscall.Handle(hwnd)
	h.app.log(fmt.Sprintf("[Hotkey] 窗口创建成功, hwnd=0x%x", hwnd))

	ret, _, err3 := registerHotKey.Call(
		uintptr(h.hwnd),
		uintptr(h.hotkeyID),
		uintptr(modifiers),
		uintptr(vk),
	)
	if ret == 0 {
		return fmt.Errorf("注册快捷键失败: %v (可能与其他程序冲突)", err3)
	}

	h.running = true
	go h.messageLoop()

	return nil
}

func (h *GlobalHotkey) wndProc(hwnd syscall.Handle, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	if msg == WM_HOTKEY {
		h.app.ShowWindow()
		return 0
	}
	ret, _, _ := defWindowProc.Call(uintptr(hwnd), uintptr(msg), wparam, lparam)
	return ret
}

func (h *GlobalHotkey) messageLoop() {
	defer func() {
		if r := recover(); r != nil {
			h.app.log(fmt.Sprintf("PANIC in hotkey messageLoop: %v", r))
		}
	}()

	var msg MSG
	for {
		ret, _, _ := getMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || int32(ret) == -1 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (h *GlobalHotkey) Unregister() {
	if h.running {
		unregisterHotKey.Call(uintptr(h.hwnd), uintptr(h.hotkeyID))
		if h.hwnd != 0 {
			postMessage.Call(uintptr(h.hwnd), uintptr(WM_QUIT), 0, 0)
		}
		h.running = false
		h.app.log("[Hotkey] 快捷键已注销")
	}
}

func parseHotkey(hotkeyStr string) (uint, uint, error) {
	var modifiers uint
	var vk uint

	parts := strings.Split(strings.TrimSpace(hotkeyStr), "+")
	for i, part := range parts {
		p := strings.TrimSpace(part)
		switch strings.ToLower(p) {
		case "ctrl", "control":
			modifiers |= MOD_CONTROL
		case "alt":
			modifiers |= MOD_ALT
		case "shift":
			modifiers |= MOD_SHIFT
		case "win", "windows":
			modifiers |= MOD_WIN
		default:
			if i == len(parts)-1 {
				vk = keyToVK(p)
				if vk == 0 {
					return 0, 0, fmt.Errorf("无法识别的键: %s", p)
				}
			}
		}
	}

	if vk == 0 {
		return 0, 0, fmt.Errorf("未指定按键")
	}

	return modifiers, vk, nil
}

func keyToVK(key string) uint {
	switch strings.ToUpper(key) {
	case "A": return 0x41
	case "B": return 0x42
	case "C": return 0x43
	case "D": return 0x44
	case "E": return 0x45
	case "F": return 0x46
	case "G": return 0x47
	case "H": return 0x48
	case "I": return 0x49
	case "J": return 0x4A
	case "K": return 0x4B
	case "L": return 0x4C
	case "M": return 0x4D
	case "N": return 0x4E
	case "O": return 0x4F
	case "P": return 0x50
	case "Q": return 0x51
	case "R": return 0x52
	case "S": return 0x53
	case "T": return 0x54
	case "U": return 0x55
	case "V": return 0x56
	case "W": return 0x57
	case "X": return 0x58
	case "Y": return 0x59
	case "Z": return 0x5A
	case "0": return 0x30
	case "1": return 0x31
	case "2": return 0x32
	case "3": return 0x33
	case "4": return 0x34
	case "5": return 0x35
	case "6": return 0x36
	case "7": return 0x37
	case "8": return 0x38
	case "9": return 0x39
	case "F1": return 0x70
	case "F2": return 0x71
	case "F3": return 0x72
	case "F4": return 0x73
	case "F5": return 0x74
	case "F6": return 0x75
	case "F7": return 0x76
	case "F8": return 0x77
	case "F9": return 0x78
	case "F10": return 0x79
	case "F11": return 0x7A
	case "F12": return 0x7B
	case "SPACE": return 0x20
	case "ENTER", "RETURN": return 0x0D
	case "TAB": return 0x09
	case "ESC", "ESCAPE": return 0x1B
	case "BACKSPACE": return 0x08
	case "DELETE": return 0x2E
	case "INSERT": return 0x2D
	case "HOME": return 0x24
	case "END": return 0x23
	case "PAGEUP": return 0x21
	case "PAGEDOWN": return 0x22
	case "UP": return 0x26
	case "DOWN": return 0x28
	case "LEFT": return 0x25
	case "RIGHT": return 0x27
	default: return 0
	}
}
