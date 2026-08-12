package login

import (
	"fmt"
	"time"

	"zauto/internal/engine/overlay"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

const runtimePasswordTypedKey = "login_password_typed"

// ResetFillState clears per-attempt login fill flags (call once at start of FillFields).
func ResetFillState(e runtime.Exec) {
	e.Sess().Runtime[runtimePasswordTypedKey] = false
}

// EnsureFormReady dismisses blocking overlays and waits for the standard login form.
func EnsureFormReady(e runtime.Exec) error {
	snap := e.Sess().ReadUI(true)
	if state.LoginFormReady(e.Sess().Resolver, snap) {
		return nil
	}
	det := state.NewDetector()
	observe, _ := e.CachedObserve()
	snap, pkg, act := observe()
	d := det.Detect(snap, pkg, act)
	if d.State == state.UIKeyboardSettings && d.Confidence >= state.VerifyMinConfidence {
		if err := overlay.DismissKeyboardSettings(e, det, observe); err != nil {
			return err
		}
	}
	if d.State == state.UILoginAccountFinderPrompt && d.Confidence >= state.VerifyMinConfidence {
		if err := overlay.DismissLoginAccountFinder(e, det, observe); err != nil {
			return err
		}
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		snap = e.Sess().ReadUI(true)
		pkg = e.Sess().Client.ForegroundPackage()
		if state.IsState(snap, pkg, state.UIKeyboardSettings) {
			overlay.DismissKeyboardSettingsIfPresent(e)
		}
		if state.LoginFormReady(e.Sess().Resolver, snap) {
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("login form not ready")
}
