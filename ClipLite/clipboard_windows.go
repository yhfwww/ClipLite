//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	globalAlloc   = kernel32.NewProc("GlobalAlloc")
	globalLock    = kernel32.NewProc("GlobalLock")
	globalUnlock  = kernel32.NewProc("GlobalUnlock")
	globalFree    = kernel32.NewProc("GlobalFree")
	moveMemory    = kernel32.NewProc("RtlMoveMemory")

	user32           = syscall.NewLazyDLL("user32.dll")
	openClipboard    = user32.NewProc("OpenClipboard")
	closeClipboard   = user32.NewProc("CloseClipboard")
	emptyClipboard   = user32.NewProc("EmptyClipboard")
	getClipboardData = user32.NewProc("GetClipboardData")
	setClipboardData = user32.NewProc("SetClipboardData")
)

const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
)

func readClipboard() (string, error) {
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return "", nil
	}
	defer closeClipboard.Call()

	h, _, _ := getClipboardData.Call(CF_UNICODETEXT)
	if h == 0 {
		return "", nil
	}

	ptr, _, _ := globalLock.Call(h)
	if ptr == 0 {
		return "", nil
	}
	defer globalUnlock.Call(h)

	return toString((*uint16)(unsafe.Pointer(ptr))), nil
}

func toString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}
	var result []uint16
	for {
		c := *(*uint16)(unsafe.Pointer(ptr))
		if c == 0 {
			break
		}
		result = append(result, c)
		ptr = (*uint16)(unsafe.Add(unsafe.Pointer(ptr), 2))
	}
	return syscall.UTF16ToString(result)
}

func writeClipboard(text string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("writeClipboard panic: %v", r)
		}
	}()

	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("打开剪贴板失败")
	}
	defer closeClipboard.Call()

	emptyClipboard.Call()

	utf16 := syscall.StringToUTF16(text)
	size := len(utf16) * 2
	h, _, _ := globalAlloc.Call(GMEM_MOVEABLE, uintptr(size))
	if h == 0 {
		return fmt.Errorf("分配内存失败")
	}

	ptr, _, _ := globalLock.Call(h)
	if ptr == 0 {
		globalFree.Call(h)
		return fmt.Errorf("锁定内存失败")
	}

	moveMemory.Call(ptr, uintptr(unsafe.Pointer(&utf16[0])), uintptr(size))
	globalUnlock.Call(h)

	r, _, _ := setClipboardData.Call(CF_UNICODETEXT, h)
	if r == 0 {
		globalFree.Call(h)
		return fmt.Errorf("设置剪贴板数据失败")
	}

	return nil
}
