package ui

// EditTapPoint returns coordinates inside an EditText, biased toward the top edge.
// Center taps on fields just above the soft keyboard often hit the keyboard toolbar
// (e.g. Gboard settings gear) instead of the field.
func EditTapPoint(elem Element) (int, int) {
	x1, y1, x2, y2 := elem.Bounds[0], elem.Bounds[1], elem.Bounds[2], elem.Bounds[3]
	h := elem.Height()
	if h <= 0 {
		return elem.Center()
	}
	x := (x1 + x2) / 2
	y := y1 + h/4
	if y >= y2 {
		y = y1 + h/2
	}
	return x, y
}

func (r Resolved) EditTapPoint() (int, int) {
	return EditTapPoint(r.Element)
}
