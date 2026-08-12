package ui

import "strings"

// ComposerEditField finds the main text input on Facebook create-post screen.
func ComposerEditField(snap Snapshot) *Element {
	var best *Element
	for i := range snap.Elements {
		elem := &snap.Elements[i]
		if !elem.Enabled {
			continue
		}
		if !IsEditText(*elem) {
			continue
		}
		label := strings.ToLower(elem.Label())
		if label == "" && elem.ContentDesc == "" {
			continue
		}
		if best == nil || elem.Height() > best.Height() {
			best = elem
		}
	}
	return best
}
