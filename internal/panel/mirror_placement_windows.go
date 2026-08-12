//go:build windows

package panel

// mirrorStartXHeadless places mirrors on the right edge of the primary work area
// (browser/headless mode — no live Wails window to anchor beside).
func mirrorStartXHeadless(deviceCount, tileW, fallback int) int {
	if deviceCount <= 0 || tileW <= 0 {
		return fallback
	}
	left, _, right, _ := primaryMonitorWorkArea()
	totalW := deviceCount * tileW
	startX := right - totalW - mirrorGap
	minX := left + mirrorGap
	if startX < minX {
		startX = minX
	}
	return startX
}
