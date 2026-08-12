package state

import (
	"strings"

	"zauto/internal/ui"
)

var facebookPackages = []string{"com.facebook.katana", "com.facebook.lite", "com.facebook.orca"}

// IsFacebookPackage reports whether pkg is a known Facebook app package.
func IsFacebookPackage(pkg string) bool {
	lower := strings.ToLower(strings.TrimSpace(pkg))
	for _, p := range facebookPackages {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// IsIMEPackage reports whether pkg belongs to a soft-keyboard IME (Gboard, etc.).
func IsIMEPackage(pkg string) bool {
	p := strings.ToLower(strings.TrimSpace(pkg))
	return strings.Contains(p, "inputmethod") || strings.Contains(p, "gboard")
}

// IMEPackageContext reports whether the foreground or hierarchy shows an IME package.
func IMEPackageContext(snap ui.Snapshot, pkg string) bool {
	if IsIMEPackage(pkg) {
		return true
	}
	for _, elem := range snap.Elements {
		if IsIMEPackage(elem.Package) {
			return true
		}
	}
	return false
}

// IsIMEForeground reports whether the foreground app is an IME (full-screen settings or keyboard).
func IsIMEForeground(pkg string) bool {
	return IsIMEPackage(pkg)
}
