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

func primaryMonitorWorkArea() (left, top, right, bottom int) {
	var area workAreaRect
	_, _, _ = procSystemParametersInfo.Call(
		uintptr(spiGetWorkArea),
		0,
		uintptr(unsafe.Pointer(&area)),
		0,
	)
	return int(area.Left), int(area.Top), int(area.Right), int(area.Bottom)
}

// desktopWindowBounds returns panel width/height and top-left position on the primary monitor.
func desktopWindowBounds() (width, height, x, y int) {
	left, top, right, bottom := primaryMonitorWorkArea()
	workW := right - left
	workH := bottom - top
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

	x = left + (workW-width)/2
	if x < left+8 {
		x = left + 8
	}
	y = top + 8
	return width, height, x, y
}
