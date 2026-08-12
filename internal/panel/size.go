package panel

// Default panel window — resizable up to full screen; mirrors open to the right of the panel.
const (
	WindowWidth          = 1100
	WindowMinWidth       = 360
	WindowDefaultWidth   = 1100
	WindowHeightFallback = 720
	WindowX              = 40
	WindowTop            = 40
	mirrorGap            = 10
)

// MirrorStartX is the left edge of the first scrcpy window (panel right + gap).
func MirrorStartX() int { return WindowX + WindowDefaultWidth + mirrorGap }
