//go:build windows

package main

import "syscall"

func hideConsoleWindow() {
	hwnd, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
	if hwnd == 0 {
		return
	}
	const swHide = 0
	_, _, _ = syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow").Call(hwnd, swHide)
}
