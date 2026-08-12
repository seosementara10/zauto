//go:build windows

package panel

import (
	"syscall"
	"unsafe"
)

var (
	procFindWindowW        = syscall.NewLazyDLL("user32.dll").NewProc("FindWindowW")
	procSetForegroundWindow = syscall.NewLazyDLL("user32.dll").NewProc("SetForegroundWindow")
	procShowWindow         = syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
)

const swRestore = 9

// FocusPanelWindow brings an existing zauto Panel window to the foreground.
func FocusPanelWindow() bool {
	title, _ := syscall.UTF16PtrFromString("zauto Panel")
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		return false
	}
	procShowWindow.Call(hwnd, swRestore)
	procSetForegroundWindow.Call(hwnd)
	return true
}
