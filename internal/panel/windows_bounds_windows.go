//go:build windows

package panel

import (
	"syscall"
	"unsafe"
)

var procSystemParametersInfo = syscall.NewLazyDLL("user32.dll").NewProc("SystemParametersInfoW")

const spiGetWorkArea = 0x0030

type workAreaRect struct {
	Left, Top, Right, Bottom int32
}

// desktopWindowBounds returns panel width/height and top-left position on the primary monitor.
func desktopWindowBounds() (width, height, x, y int) {
	var area workAreaRect
	_, _, _ = procSystemParametersInfo.Call(
		uintptr(spiGetWorkArea),
		0,
		uintptr(unsafe.Pointer(&area)),
		0,
	)
	workW := int(area.Right - area.Left)
	workH := int(area.Bottom - area.Top)
	if workH < 480 {
		workH = WindowHeightFallback
	}

	height = workH - 16
	if height < 560 {
		height = WindowHeightFallback
	}

	width = int(float64(workW) * 0.58)
	if width < 720 {
		width = 720
	}
	if width > 1600 {
		width = 1600
	}
	if width > workW-32 {
		width = workW - 32
	}

	x = int(area.Left) + (workW-width)/2
	if x < int(area.Left)+8 {
		x = int(area.Left) + 8
	}
	y = int(area.Top) + 8
	return width, height, x, y
}
