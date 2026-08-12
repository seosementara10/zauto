package ui

import (
	"sort"
)

const minLoginFieldHeight = 28

// LoginFormEdits returns enabled Facebook login EditText fields, top-to-bottom.
// Used when label-based lookup fails (e.g. soft keyboard open and labels scrolled off-screen).
func LoginFormEdits(snap Snapshot) []Element {
	var edits []Element
	for _, elem := range snap.Elements {
		if !isLoginEdit(elem) {
			continue
		}
		edits = append(edits, elem)
	}
	sort.Slice(edits, func(i, j int) bool {
		_, yi := edits[i].Center()
		_, yj := edits[j].Center()
		if yi != yj {
			return yi < yj
		}
		return edits[i].Bounds[0] < edits[j].Bounds[0]
	})
	return edits
}

func isLoginEdit(elem Element) bool {
	if !elem.Enabled || elem.Height() < minLoginFieldHeight {
		return false
	}
	if !IsEditText(elem) {
		return false
	}
	return isLoginPackageEdit(elem)
}
