package state

import (
	"strings"

	"zauto/internal/ui"
)

var systemPermissionPackages = []string{
	"com.android.permissioncontroller",
	"com.google.android.permissioncontroller",
}

var dialogShellClasses = []string{
	"AlertDialog", "Dialog", "BottomSheet", "ModalBottomSheet", "PopupWindow",
}

func IsSystemPermissionPackage(pkg string) bool {
	lower := strings.ToLower(pkg)
	for _, p := range systemPermissionPackages {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func HasSystemPermissionElement(snap ui.Snapshot) bool {
	for _, el := range snap.Elements {
		ep := strings.ToLower(el.Package)
		for _, p := range systemPermissionPackages {
			if strings.Contains(ep, p) {
				return true
			}
		}
	}
	return false
}

func SystemPermissionHint(snap ui.Snapshot, pkg string) string {
	if IsSystemPermissionPackage(pkg) {
		return "foreground_pkg:" + pkg
	}
	for _, el := range snap.Elements {
		ep := strings.ToLower(el.Package)
		for _, p := range systemPermissionPackages {
			if strings.Contains(ep, p) {
				return "element_pkg:" + el.Package
			}
		}
	}
	return ""
}

func DialogShellClass(snap ui.Snapshot) string {
	for _, el := range snap.Elements {
		cls := el.ClassName
		for _, shell := range dialogShellClasses {
			if strings.Contains(cls, shell) {
				return cls
			}
		}
	}
	return ""
}
