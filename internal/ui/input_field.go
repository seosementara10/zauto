package ui

import "strings"

// FindInputField locates an EditText by query, or the EditText below a matching label.
func FindInputField(resolver *Resolver, snap Snapshot, q FindQuery) *Resolved {
	if r := resolver.Find(snap, q); r != nil {
		if IsEditText(r.Element) {
			return r
		}
		if below := EditBelow(snap, r); below != nil {
			return below
		}
		return r
	}
	if len(q.Texts) > 0 {
		return FindEditBelowLabel(snap, resolver, q.Texts)
	}
	return nil
}

// IsEditText reports whether the element is an Android text input.
func IsEditText(elem Element) bool {
	cn := elem.ClassName
	return strings.Contains(cn, "EditText") || strings.Contains(cn, "AutoCompleteTextView")
}

// EditBelow returns the nearest enabled EditText below a label within a vertical band.
func EditBelow(snap Snapshot, label *Resolved) *Resolved {
	_, labelY := label.Center()
	var best *Resolved
	for _, elem := range snap.Elements {
		if !IsEditText(elem) || !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy < labelY-20 || cy > labelY+220 {
			continue
		}
		r := &Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		if best == nil {
			best = r
		} else {
			_, bestCY := best.Center()
			if cy < bestCY {
				best = r
			}
		}
	}
	return best
}

// FindEditBelowLabel finds a label by hints then returns EditBelow it.
func FindEditBelowLabel(snap Snapshot, resolver *Resolver, hints []string) *Resolved {
	label := resolver.Find(snap, FindQuery{Texts: hints})
	if label == nil {
		return nil
	}
	return EditBelow(snap, label)
}
