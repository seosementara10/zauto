//go:build !windows

package panel

func desktopWindowBounds() (width, height, x, y int) {
	return WindowWidth, WindowHeightFallback, WindowX, WindowTop
}
