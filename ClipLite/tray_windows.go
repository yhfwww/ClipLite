//go:build windows

package main

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	trayUser32         = syscall.NewLazyDLL("user32.dll")
	trayKernel32       = syscall.NewLazyDLL("kernel32.dll")
	trayShell32        = syscall.NewLazyDLL("shell32.dll")
	trayCreateMenu     = trayUser32.NewProc("CreatePopupMenu")
	trayAppendMenu     = trayUser32.NewProc("AppendMenuW")
	traySetMenuDefault = trayUser32.NewProc("SetMenuDefaultItem")
	trayTrackPopup     = trayUser32.NewProc("TrackPopupMenu")
	trayDestroyMenu    = trayUser32.NewProc("DestroyMenu")
	trayGetCursorPos   = trayUser32.NewProc("GetCursorPos")
	traySetForeground  = trayUser32.NewProc("SetForegroundWindow")
	trayShellNotify    = trayShell32.NewProc("Shell_NotifyIconW")

	setWindowsHookEx    = trayUser32.NewProc("SetWindowsHookExW")
	callNextHookEx      = trayUser32.NewProc("CallNextHookEx")
	unhookWindowsHookEx = trayUser32.NewProc("UnhookWindowsHookEx")
	getAsyncKeyState    = trayUser32.NewProc("GetAsyncKeyState")
	getModuleHandle     = trayKernel32.NewProc("GetModuleHandleW")

	registerClassEx  = trayUser32.NewProc("RegisterClassExW")
	createWindowEx   = trayUser32.NewProc("CreateWindowExW")
	defWindowProc    = trayUser32.NewProc("DefWindowProcW")
	loadImage        = trayUser32.NewProc("LoadImageW")
	getMessage       = trayUser32.NewProc("GetMessageW")
	translateMessage = trayUser32.NewProc("TranslateMessage")
	dispatchMessage  = trayUser32.NewProc("DispatchMessageW")
	postMessage      = trayUser32.NewProc("PostMessageW")
)

const (
	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	WM_TRAYICON  = 0x0400 + 0x1982
	WM_LBUTTONUP = 0x0202
	WM_RBUTTONUP = 0x0205
	WM_COMMAND   = 0x0111
	WM_DESTROY   = 0x0002
	WM_QUIT      = 0x0012

	TPM_LEFTALIGN = 0x0000

	MF_STRING    = 0x00000000
	MF_SEPARATOR = 0x00000800
	MF_DEFAULT   = 0x00001000

	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	VK_CONTROL     = 0x11
	VK_V           = 0x56

	WM_PASTE_DETECTED = 0x0400 + 0x1983
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             syscall.Handle
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            syscall.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
}

type POINT struct {
	X int32
	Y int32
}

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	WndProc       uintptr
	ClsExtra      int32
	WndExtra      int32
	Instance      syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	MenuName      *uint16
	ClassName     *uint16
	HIconSm       syscall.Handle
}

type MSG struct {
	HWnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type SystemTray struct {
	hwnd         syscall.Handle
	app          *App
	running      bool
	mu           sync.Mutex
	iconPath     string
	stopChan     chan struct{}
	hookHandle   uintptr
	kbHookProc   uintptr
	className    *uint16
	windowName   *uint16
	hInstance    uintptr
	lastPasteMS  int64
	pasteMu      sync.Mutex
}

func NewSystemTray(app *App) *SystemTray {
	className, _ := syscall.UTF16PtrFromString("ClipLiteTrayClass")
	windowName, _ := syscall.UTF16PtrFromString("ClipLiteTray")
	return &SystemTray{
		app:        app,
		stopChan:   make(chan struct{}),
		className:  className,
		windowName: windowName,
	}
}

func (t *SystemTray) Start() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		if r := recover(); r != nil {
			t.app.log(fmt.Sprintf("PANIC in tray Start: %v", r))
		}
	}()

	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.mu.Unlock()

	t.app.log("[Tray] 开始初始化托盘...")

	t.hInstance = t.getInstanceHandle()

	if !t.createWindow() {
		t.app.log("[Tray] 创建窗口失败")
		return
	}

	t.installKeyboardHook()

	t.addIcon()

	t.app.log("[Tray] 托盘初始化完成，进入消息循环")

	t.messageLoop()
}

func (t *SystemTray) getInstanceHandle() uintptr {
	ret, _, _ := getModuleHandle.Call(0)
	if ret != 0 {
		t.app.log(fmt.Sprintf("[Tray] 获取模块句柄成功: 0x%x", ret))
	} else {
		t.app.log("[Tray] 获取模块句柄失败，使用 0")
	}
	return ret
}

func (t *SystemTray) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running {
		return
	}
	t.running = false
	t.removeIcon()
	t.uninstallKeyboardHook()
	if t.hwnd != 0 {
		postMessage.Call(uintptr(t.hwnd), uintptr(WM_DESTROY), 0, 0)
	}
}

func (t *SystemTray) createWindow() bool {
	wc := WNDCLASSEX{
		CbSize:   uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:  syscall.NewCallback(t.wndProc),
		Instance: syscall.Handle(t.hInstance),
		ClassName: t.className,
	}

	ret, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		t.app.log(fmt.Sprintf("[Tray] 注册窗口类失败: %v", err))
		return false
	}

	hwnd, _, err := createWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(t.className)),
		uintptr(unsafe.Pointer(t.windowName)),
		0,
		0, 0, 0, 0,
		0, 0, t.hInstance, 0,
	)
	if hwnd == 0 {
		t.app.log(fmt.Sprintf("[Tray] 创建窗口失败: %v", err))
		return false
	}
	t.hwnd = syscall.Handle(hwnd)
	t.app.log(fmt.Sprintf("[Tray] 窗口创建成功, hwnd=0x%x", hwnd))
	return true
}

func (t *SystemTray) wndProc(hwnd syscall.Handle, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		switch lparam {
		case WM_LBUTTONUP:
			t.app.ShowWindow()
		case WM_RBUTTONUP:
			t.showMenu()
		}
		return 0
	case WM_COMMAND:
		switch wparam {
		case 1001:
			t.app.ShowWindow()
		case 1002:
			t.app.HideWindow()
		case 1003:
			t.app.ExitApp()
		}
		return 0
	case WM_PASTE_DETECTED:
		go t.handlePaste()
		return 0
	case WM_DESTROY:
		postMessage.Call(0, uintptr(WM_QUIT), 0, 0)
		return 0
	}
	ret, _, _ := defWindowProc.Call(uintptr(hwnd), uintptr(msg), wparam, lparam)
	return ret
}

func (t *SystemTray) installKeyboardHook() {
	t.kbHookProc = syscall.NewCallback(t.keyboardHook)

	hModule := t.hInstance
	ret, _, err := setWindowsHookEx.Call(
		uintptr(WH_KEYBOARD_LL),
		t.kbHookProc,
		hModule,
		0,
	)
	if ret == 0 {
		t.app.log(fmt.Sprintf("[Tray] 安装键盘钩子失败: %v (hModule=0x%x)", err, hModule))
		return
	}
	t.hookHandle = ret
	t.app.log(fmt.Sprintf("[Tray] 键盘钩子安装成功, handle=0x%x", ret))
}

func (t *SystemTray) uninstallKeyboardHook() {
	if t.hookHandle != 0 {
		unhookWindowsHookEx.Call(t.hookHandle)
		t.hookHandle = 0
		t.app.log("[Tray] 键盘钩子已卸载")
	}
}

func (t *SystemTray) keyboardHook(nCode int32, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 && wParam == WM_KEYDOWN {
		kb := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if kb.VkCode == VK_V {
			ctrlState, _, _ := getAsyncKeyState.Call(uintptr(VK_CONTROL))
			if ctrlState&0x8000 != 0 {
				now := time.Now().UnixMilli()
				t.pasteMu.Lock()
				if now-t.lastPasteMS < 500 {
					t.pasteMu.Unlock()
					goto NEXT
				}
				t.lastPasteMS = now
				t.pasteMu.Unlock()

				if t.hwnd != 0 {
					postMessage.Call(uintptr(t.hwnd), uintptr(WM_PASTE_DETECTED), 0, 0)
				}
			}
		}
	}
NEXT:
	ret, _, _ := callNextHookEx.Call(t.hookHandle, uintptr(nCode), wParam, lParam)
	return ret
}

func (t *SystemTray) handlePaste() {
	defer func() {
		if r := recover(); r != nil {
			t.app.log(fmt.Sprintf("PANIC in handlePaste: %v", r))
		}
	}()

	time.Sleep(100 * time.Millisecond)

	content, err := readClipboard()
	if err != nil {
		t.app.log(fmt.Sprintf("[Tray] 读取剪贴板失败: %v", err))
		return
	}
	if content == "" {
		return
	}

	t.app.RecordFromPaste(content)
}

func (t *SystemTray) addIcon() {
	if t.iconPath == "" {
		t.app.log("[Tray] 图标路径为空，跳过加载图标")
		return
	}

	execPath, _ := syscall.UTF16PtrFromString(t.iconPath)
	hIcon, _, err := loadImage.Call(
		0,
		uintptr(unsafe.Pointer(execPath)),
		1,
		16, 16,
		0x00000010,
	)
	if hIcon == 0 {
		t.app.log(fmt.Sprintf("[Tray] 加载图标失败: %v, 路径: %s", err, t.iconPath))
		return
	}

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = t.hwnd
	nid.UID = 1
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_TRAYICON
	nid.HIcon = syscall.Handle(hIcon)
	tip, _ := syscall.UTF16FromString("ClipLite - 剪贴板管理")
	copy(nid.SzTip[:], tip)

	ret, _, err := trayShellNotify.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		t.app.log(fmt.Sprintf("[Tray] 添加托盘图标失败: %v", err))
	} else {
		t.app.log("[Tray] 托盘图标添加成功")
	}
}

func (t *SystemTray) removeIcon() {
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = t.hwnd
	nid.UID = 1
	trayShellNotify.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

func (t *SystemTray) showMenu() {
	hMenu, _, _ := trayCreateMenu.Call()

	showText, _ := syscall.UTF16PtrFromString("显示主面板")
	trayAppendMenu.Call(hMenu, MF_STRING, 1001, uintptr(unsafe.Pointer(showText)))

	hideText, _ := syscall.UTF16PtrFromString("隐藏窗口")
	trayAppendMenu.Call(hMenu, MF_STRING, 1002, uintptr(unsafe.Pointer(hideText)))

	trayAppendMenu.Call(hMenu, MF_SEPARATOR, 0, 0)

	exitText, _ := syscall.UTF16PtrFromString("退出")
	trayAppendMenu.Call(hMenu, MF_STRING, 1003, uintptr(unsafe.Pointer(exitText)))

	traySetMenuDefault.Call(hMenu, 1001)

	var pt POINT
	trayGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	traySetForeground.Call(uintptr(t.hwnd))

	trayTrackPopup.Call(
		hMenu,
		TPM_LEFTALIGN,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		uintptr(t.hwnd),
		0,
	)

	trayDestroyMenu.Call(hMenu)
}

func (t *SystemTray) messageLoop() {
	var msg MSG
	for {
		ret, _, err := getMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || int32(ret) == -1 {
			t.app.log(fmt.Sprintf("[Tray] 消息循环退出, ret=%d, err=%v", ret, err))
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
