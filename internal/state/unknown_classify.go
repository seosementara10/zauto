package state

import (
	"strings"

	"zauto/internal/ui"
)

var systemWindowPackages = []string{
	"com.android.systemui",
	"com.google.android.gms",
	"com.android.settings",
}

// ClassifyUnknown assigns a sub-kind when the primary state is unknown or uncertain.
func ClassifyUnknown(snap ui.Snapshot, pkg string, inv Investigation) UnknownKind {
	if inv.Probes[UIError] >= InvestigateMinConfidence {
		return UnknownKindError
	}
	if inv.Probes[UILoading] >= 0.4 {
		return UnknownKindLoading
	}
	if inv.Probes[UIKeyboardSettings] >= InvestigateMinConfidence {
		return UnknownKindOverlay
	}
	if IsSystemPermissionPackage(pkg) || HasSystemPermissionElement(snap) {
		return UnknownKindPermission
	}
	if shell := DialogShellClass(snap); shell != "" {
		return UnknownKindOverlay
	}
	if IsFacebookPackage(pkg) {
		return UnknownKindAppScreen
	}
	for _, el := range snap.Elements {
		if IsFacebookPackage(el.Package) {
			return UnknownKindAppScreen
		}
	}
	lower := strings.ToLower(pkg)
	for _, p := range systemWindowPackages {
		if strings.Contains(lower, p) {
			return UnknownKindSystemWindow
		}
	}
	if IsIMEPackage(pkg) {
		return UnknownKindOverlay
	}
	if pkg != "" {
		return UnknownKindSystemWindow
	}
	return UnknownKindAppScreen
}

// EnrichUnknownDetection classifies unknown detections for recovery routing.
func EnrichUnknownDetection(d Detection, snap ui.Snapshot, pkg string, inv Investigation) Detection {
	if d.State != UIUnknown && d.Confidence >= VerifyMinConfidence {
		d.UnknownKind = UnknownKindNone
		return d
	}
	d.UnknownKind = ClassifyUnknown(snap, pkg, inv)
	return d
}
