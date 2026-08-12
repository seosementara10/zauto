//go:build windows

package monitor

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	swpNoZOrder     = 0x0004
	swpNoActivate   = 0x0010
	swpShowWindow   = 0x0040
	enumWindowsStop = 0
)

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	procEnumWindows        = user32.NewProc("EnumWindows")
	procGetWindowTextW     = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procIsWindowVisible    = user32.NewProc("IsWindowVisible")
	procSetWindowPos       = user32.NewProc("SetWindowPos")
)

func windowTitle(hwnd syscall.Handle) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if n <= 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func isWindowVisible(hwnd syscall.Handle) bool {
	r, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return r != 0
}

func setWindowRect(hwnd syscall.Handle, tile WindowTile) bool {
	if hwnd == 0 {
		return false
	}
	r, _, _ := procSetWindowPos.Call(
		uintptr(hwnd),
		0,
		uintptr(tile.X),
		uintptr(tile.Y),
		uintptr(tile.W),
		uintptr(tile.H),
		uintptr(swpNoZOrder|swpNoActivate|swpShowWindow),
	)
	return r != 0
}

type windowMatch struct {
	hwnd  syscall.Handle
	title string
}

func findScrcpyWindow(serial string, hpNum int) syscall.Handle {
	short := serial
	if len(short) > 8 {
		short = serial[len(serial)-8:]
	}
	exact := ""
	if hpNum > 0 {
		exact = fmt.Sprintf("HP%d ...%s", hpNum, short)
	}
	var found syscall.Handle
	enumCallback := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		if !isWindowVisible(hwnd) {
			return 1
		}
		title := windowTitle(hwnd)
		if title == "" {
			return 1
		}
		if exact != "" && title == exact {
			found = hwnd
			return enumWindowsStop
		}
		if strings.Contains(title, short) && strings.HasPrefix(title, "HP") {
			found = hwnd
			return enumWindowsStop
		}
		return 1
	})
	procEnumWindows.Call(enumCallback, 0)
	return found
}

// ScrcpyWindowRunning reports whether a visible scrcpy window exists for the device.
func ScrcpyWindowRunning(serial string) bool {
	return findScrcpyWindow(serial, 0) != 0
}

// RelayoutScrcpyWindows moves existing scrcpy windows to new tiles without restarting.
// Return value is per-serial success (same order as serials).
func RelayoutScrcpyWindows(serials []string, tiles []WindowTile, hpNums map[string]int) []bool {
	out := make([]bool, len(serials))
	for i, serial := range serials {
		if i >= len(tiles) {
			break
		}
		hp := hpNums[serial]
		if hp <= 0 {
			hp = i + 1
		}
		hwnd := findScrcpyWindow(serial, hp)
		if hwnd == 0 {
			continue
		}
		out[i] = setWindowRect(hwnd, tiles[i])
	}
	return out
}
