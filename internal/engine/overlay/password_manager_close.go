package overlay

import (
	"strings"

	"zauto/internal/state"
	"zauto/internal/ui"
)

// findPasswordManagerClose locates the top-right X on Google Password Manager bottom sheet.
func findPasswordManagerClose(resolver *ui.Resolver, snap ui.Snapshot, screenW, screenH int) *ui.Resolved {
	if resolver == nil {
		resolver = ui.NewDefaultResolver()
	}
	if screenW <= 0 {
		screenW = 720
	}
	if screenH <= 0 {
		screenH = 1600
	}
	sheetMinY := screenH * 25 / 100

	queries := passwordManagerCloseQueries(sheetMinY)
	for _, q := range queries {
		if r := resolver.Find(snap, q); r != nil {
			return r
		}
	}

	title := resolver.Find(snap, ui.FindQuery{Texts: state.PasswordManagerTitleTexts})
	if title == nil {
		title = resolver.Find(snap, ui.FindQuery{Texts: state.PasswordManagerSaveTitleTexts})
	}
	if title == nil {
		return nil
	}
	_, titleY := title.Center()
	minX := screenW * 65 / 100
	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Enabled || elem.Width() <= 0 {
			continue
		}
		cx, cy := elem.Center()
		if cy < titleY-100 || cy > titleY+100 {
			continue
		}
		if cx < minX {
			continue
		}
		w, h := elem.Width(), elem.Height()
		if w > 160 || h > 160 {
			continue
		}
		if !pmCloseCandidateClass(elem) {
			continue
		}
		r := &ui.Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		if best == nil {
			best = r
			continue
		}
		bestCX, _ := best.Center()
		if cx > bestCX {
			best = r
		}
	}
	return best
}

// passwordManagerCloseFallbackPoint returns a normalized tap near the header X when the
// close icon is visible but missing from the accessibility tree (common on GMS sheets).
func passwordManagerCloseFallbackPoint(resolver *ui.Resolver, snap ui.Snapshot, screenW, screenH int) (x, y int, ok bool) {
	if screenW <= 0 {
		screenW = 720
	}
	title := resolver.Find(snap, ui.FindQuery{Texts: state.PasswordManagerTitleTexts})
	if title == nil {
		title = resolver.Find(snap, ui.FindQuery{Texts: state.PasswordManagerSaveTitleTexts})
	}
	if title == nil {
		return 0, 0, false
	}
	_, titleY := title.Center()
	return screenW * 91 / 100, titleY, true
}

func pmCloseCandidateClass(elem ui.Element) bool {
	if elem.Clickable {
		return true
	}
	cn := elem.ClassName
	return strings.Contains(cn, "ImageButton") ||
		strings.Contains(cn, "ImageView") ||
		strings.Contains(cn, "Button")
}
