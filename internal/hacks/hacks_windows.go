//go:build windows
// +build windows

package hacks

import (
	"runtime"
	"syscall"
	"unsafe"
)

// setClipboardWindows sets the clipboard content on Windows
func setClipboardWindows(text string) error {
	user32 := syscall.NewLazyDLL("user32.dll")
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	openClipboard := user32.NewProc("OpenClipboard")
	closeClipboard := user32.NewProc("CloseClipboard")
	emptyClipboard := user32.NewProc("EmptyClipboard")
	setClipboardData := user32.NewProc("SetClipboardData")
	globalAlloc := kernel32.NewProc("GlobalAlloc")
	globalLock := kernel32.NewProc("GlobalLock")
	globalUnlock := kernel32.NewProc("GlobalUnlock")

	const (
		CF_TEXT       = 1
		GMEM_MOVEABLE = 0x0002
	)

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return syscall.GetLastError()
	}
	defer closeClipboard.Call()

	// Empty clipboard
	emptyClipboard.Call()

	// Convert string to UTF-16 and get byte length
	utf16Text := syscall.StringToUTF16(text)
	size := len(utf16Text) * 2

	// Allocate global memory
	hMem, _, _ := globalAlloc.Call(GMEM_MOVEABLE, uintptr(size))
	if hMem == 0 {
		return syscall.GetLastError()
	}

	// Lock memory and copy text
	pMem, _, _ := globalLock.Call(hMem)
	if pMem == 0 {
		return syscall.GetLastError()
	}

	// Copy UTF-16 data to allocated memory
	dst := (*[1 << 20]uint16)(unsafe.Pointer(pMem))
	copy(dst[:len(utf16Text)], utf16Text)

	globalUnlock.Call(hMem)

	// Set clipboard data
	ret, _, _ = setClipboardData.Call(CF_TEXT, hMem)
	if ret == 0 {
		return syscall.GetLastError()
	}

	return nil
}

// StoreWinClipboard stores the letter "w" in clipboard on Windows, does nothing on other OS
func StoreWinClipboard() error {
	if runtime.GOOS == "windows" {
		return setClipboardWindows("w")
	}
	// Do nothing on non-Windows systems
	return nil
}
