package overlay

import (
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

// DismissLoginAccountFinder closes the "Cari akun saya" / phone-only recovery step blocking login.
func DismissLoginAccountFinder(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT dismiss login account finder")
	return pollUntilDismissed(e, det, observe, pollDismissOpts{
		avoid:        state.UILoginAccountFinderPrompt,
		timeout:      12 * time.Second,
		act:          func() error { return tapLoginAccountFinderDismiss(e) },
		doneEvent:    "VERIFY login account finder dismissed",
		missEvent:    "ACT account finder dismiss miss: %v",
		successEvent: "ACT account finder dismiss tapped",
		errLabel:     "login account finder",
	})
}

func tapLoginAccountFinderDismiss(e runtime.Exec) error {
	if err := e.WaitTapTexts(state.LoginAccountFinderDismissTexts, "bottom", 0, 0, 3); err == nil {
		return nil
	}
	return e.Sess().Client.KeyEvent("BACK")
}
