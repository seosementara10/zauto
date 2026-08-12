package monitor

import (
	"strconv"
	"strings"
)
// WindowTile is an exact scrcpy window placement (pixel-perfect, no gap).
type WindowTile struct {
	X, Y, W, H int
}

// ParseResolution parses "720x1600" or "720 x 1600".
func ParseResolution(s string) (w, h int) {
	s = strings.TrimSpace(strings.ReplaceAll(s, " ", ""))
	parts := strings.Split(strings.ToLower(s), "x")
	if len(parts) != 2 {
		return 720, 1600
	}
	w, _ = strconv.Atoi(parts[0])
	h, _ = strconv.Atoi(parts[1])
	if w <= 0 || h <= 0 {
		return 720, 1600
	}
	return w, h
}

// ContentSize returns scrcpy video size after --max-size scaling.
func ContentSize(screenW, screenH, maxSize int) (w, h int) {
	if maxSize <= 0 {
		maxSize = 480
	}
	long := screenW
	short := screenH
	if screenH > screenW {
		long, short = screenH, screenW
	}
	scale := float64(maxSize) / float64(long)
	w = int(float64(short)*scale + 0.5)
	h = int(float64(long)*scale + 0.5)
	if screenW > screenH {
		w = int(float64(long)*scale + 0.5)
		h = int(float64(short)*scale + 0.5)
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

// ComputeTiles places count windows in one row, edge-to-edge.
func ComputeTiles(screenW, screenH, maxSize, count, startX, startY int) []WindowTile {
	if count <= 0 {
		return nil
	}
	w, h := ContentSize(screenW, screenH, maxSize)
	tiles := make([]WindowTile, count)
	for i := 0; i < count; i++ {
		tiles[i] = WindowTile{
			X: startX + i*w,
			Y: startY,
			W: w,
			H: h,
		}
	}
	return tiles
}

// FarmOptions for panel-side mirrors: StartX is set by panel.MirrorStartX().
func FarmOptions(projectRoot string) Options {
	return Options{
		ProjectRoot: projectRoot,
		MaxSize:     480,
		StartY:      40,
	}
}
