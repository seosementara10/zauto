package panel

import "sync"

// panelBounds tracks the live Wails window position for mirror placement.
type panelBounds struct {
	mu     sync.RWMutex
	x      int
	y      int
	width  int
	height int
	live   bool
}

func (b *panelBounds) set(x, y, width, height int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.x = x
	b.y = y
	b.width = width
	b.height = height
	b.live = true
}

func (b *panelBounds) hasLive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.live && b.width > 0
}

func (b *panelBounds) mirrorStartX() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.live && b.width > 0 {
		return b.x + b.width + mirrorGap
	}
	return WindowX + WindowDefaultWidth + mirrorGap
}

func (b *panelBounds) snapshot() map[string]int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	mirrorX := WindowX + WindowDefaultWidth + mirrorGap
	if b.live && b.width > 0 {
		mirrorX = b.x + b.width + mirrorGap
	}
	return map[string]int{"x": b.x, "y": b.y, "width": b.width, "height": b.height, "mirror_start_x": mirrorX}
}
