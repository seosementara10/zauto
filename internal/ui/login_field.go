package ui

import "strings"

// FindLoginEmailField locates the email/phone EditText on the Facebook login form.
func FindLoginEmailField(resolver *Resolver, snap Snapshot, hints, resourceIDs []string) *Resolved {
	if resolver == nil {
		resolver = NewDefaultResolver()
	}
	q := FindQuery{Texts: hints, ContentDescs: hints, ResourceIDs: resourceIDs}
	if r := FindInputField(resolver, snap, q); r != nil && !r.Element.Password {
		return r
	}
	for _, elem := range snap.Elements {
		if !IsEditText(elem) || !elem.Enabled || elem.Password {
			continue
		}
		if !isLoginPackageEdit(elem) {
			continue
		}
		desc := Normalize(elem.ContentDesc)
		for _, h := range hints {
			if desc != "" && strings.Contains(desc, Normalize(h)) {
				return &Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
			}
		}
	}
	edits := LoginFormEdits(snap)
	for _, elem := range edits {
		if !elem.Password {
			return &Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		}
	}
	if len(edits) > 0 {
		return &Resolved{Element: edits[0], Label: edits[0].Label(), Bounds: edits[0].Bounds}
	}
	return nil
}

// FindLoginPasswordField locates the password EditText on the Facebook login form.
func FindLoginPasswordField(resolver *Resolver, snap Snapshot, hints, resourceIDs []string) *Resolved {
	if resolver == nil {
		resolver = NewDefaultResolver()
	}
	q := FindQuery{Texts: hints, ContentDescs: hints, ResourceIDs: resourceIDs}
	if r := FindInputField(resolver, snap, q); r != nil {
		return r
	}
	for _, elem := range snap.Elements {
		if !IsEditText(elem) || !elem.Enabled {
			continue
		}
		if !isLoginPackageEdit(elem) {
			continue
		}
		if elem.Password {
			return &Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		}
		desc := Normalize(elem.ContentDesc)
		for _, h := range hints {
			if desc != "" && strings.Contains(desc, Normalize(h)) {
				return &Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
			}
		}
	}
	edits := LoginFormEdits(snap)
	for _, elem := range edits {
		if elem.Password {
			return &Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		}
	}
	if len(edits) > 1 {
		e := edits[1]
		return &Resolved{Element: e, Label: e.Label(), Bounds: e.Bounds}
	}
	return nil
}

func isLoginPackageEdit(elem Element) bool {
	if elem.Package == "" {
		return true
	}
	p := strings.ToLower(elem.Package)
	return strings.Contains(p, "com.facebook.katana") ||
		strings.Contains(p, "com.facebook.lite") ||
		strings.Contains(p, "com.facebook.orca")
}
