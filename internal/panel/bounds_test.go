package panel

import "testing"

func TestPanelBoundsMirrorStartX(t *testing.T) {
	var b panelBounds
	if got := b.mirrorStartX(); got != WindowX+WindowWidth+mirrorGap {
		t.Fatalf("default mirror start x = %d want %d", got, WindowX+WindowWidth+mirrorGap)
	}
	b.set(200, 50, 450, 700)
	if got := b.mirrorStartX(); got != 200+450+mirrorGap {
		t.Fatalf("live mirror start x = %d want %d", got, 200+450+mirrorGap)
	}
	snap := b.snapshot()
	if snap["mirror_start_x"] != 200+450+mirrorGap {
		t.Fatalf("snapshot mirror_start_x = %v", snap["mirror_start_x"])
	}
}
