package ui

import "strings"

// FieldBoundsMatch reports whether two elements refer to the same on-screen field.
func FieldBoundsMatch(a, b Element) bool {
	if a.ResourceID != "" && a.ResourceID == b.ResourceID {
		return true
	}
	ax, ay := a.Center()
	bx, by := b.Center()
	dx, dy := ax-bx, ay-by
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx <= 40 && dy <= 40
}

// FieldHasExpectedValue reports whether the field already shows the intended value once
// (not duplicated or polluted with extra characters).
func FieldHasExpectedValue(field Element, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	got := fieldDisplayText(field)
	if got == "" {
		return false
	}
	gotNorm := strings.ToLower(got)
	wantNorm := strings.ToLower(value)
	if gotNorm == wantNorm {
		return true
	}
	if !strings.Contains(gotNorm, wantNorm) {
		return false
	}
	if len(gotNorm) > len(wantNorm)+4 {
		return false
	}
	if strings.Contains(gotNorm, wantNorm+wantNorm) {
		return false
	}
	return true
}

// FieldNeedsClear reports whether a field should be cleared before typing the expected value.
func FieldNeedsClear(field Element, value string) bool {
	// Masked password fields expose label in content-desc, not the typed value.
	if field.Password {
		return false
	}
	got := fieldDisplayText(field)
	if got == "" {
		return false
	}
	if FieldHasExpectedValue(field, value) {
		return false
	}
	return true
}

func fieldDisplayText(field Element) string {
	got := strings.TrimSpace(field.Text)
	if got == "" && !field.Password {
		got = strings.TrimSpace(field.ContentDesc)
	}
	return got
}

// FieldTextContains reports whether the field shows all or part of the expected value.
func FieldTextContains(field Element, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	got := strings.TrimSpace(field.Text)
	if got == "" {
		got = strings.TrimSpace(field.ContentDesc)
	}
	if got == "" {
		return false
	}
	gotNorm := strings.ToLower(got)
	wantNorm := strings.ToLower(value)
	if strings.Contains(gotNorm, wantNorm) {
		return true
	}
	if len(wantNorm) > 8 {
		return strings.Contains(gotNorm, wantNorm[:8])
	}
	return false
}

// FindEditAtBounds locates an EditText near the given bounds.
func FindEditAtBounds(snap Snapshot, ref Element) *Element {
	for i := range snap.Elements {
		elem := &snap.Elements[i]
		if !IsEditText(*elem) || !elem.Enabled {
			continue
		}
		if FieldBoundsMatch(*elem, ref) {
			return elem
		}
	}
	return nil
}
