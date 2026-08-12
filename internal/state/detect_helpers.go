package state

import "zauto/internal/ui"

// DetectScreen classifies the current screen (single entry for overlay/login/recovery checks).
func DetectScreen(snap ui.Snapshot, pkg, activity string) Detection {
	return NewDetector().Detect(snap, pkg, activity)
}

// IsState reports whether the detector classifies the screen as want with sufficient confidence.
func IsState(snap ui.Snapshot, pkg string, want UIState) bool {
	d := DetectScreen(snap, pkg, "")
	return d.State == want && d.Confidence >= VerifyMinConfidence
}

// LoginFormReady reports whether the Facebook login form is present (edits or labeled fields).
func LoginFormReady(resolver *ui.Resolver, snap ui.Snapshot) bool {
	if len(ui.LoginFormEdits(snap)) >= 2 {
		return true
	}
	if resolver == nil {
		resolver = ui.NewDefaultResolver()
	}
	hasPassword := resolver.Find(snap, ui.FindQuery{Texts: LoginPasswordFieldTexts}) != nil
	hasEmail := resolver.Find(snap, ui.FindQuery{Texts: LoginEmailFieldTexts}) != nil
	hasLoginBtn := resolver.Find(snap, ui.FindQuery{Texts: LoginButtonTexts, PreferClickable: true}) != nil
	return hasPassword && (hasEmail || hasLoginBtn)
}

// IMEBlocksInput reports whether IME is foreground or keyboard settings overlay is up.
func IMEBlocksInput(snap ui.Snapshot, pkg string) bool {
	if IsIMEForeground(pkg) {
		return true
	}
	return IsState(snap, pkg, UIKeyboardSettings)
}
