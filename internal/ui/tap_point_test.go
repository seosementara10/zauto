package ui

import "testing"

func TestEditTapPointBiasesToTop(t *testing.T) {
	elem := Element{Bounds: [4]int{64, 981, 584, 1023}}
	x, y := EditTapPoint(elem)
	cx, cy := elem.Center()
	if x != cx {
		t.Fatalf("x=%d want center x=%d", x, cx)
	}
	if y >= cy {
		t.Fatalf("tap y=%d should be above center y=%d", y, cy)
	}
	if y < elem.Bounds[1] || y >= elem.Bounds[3] {
		t.Fatalf("tap y=%d outside bounds %+v", y, elem.Bounds)
	}
}
