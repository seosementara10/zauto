package overlay

import (
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

// DismissSoftKeyboard hides the soft keyboard (BACK). Underlying screen stays classified by detector.
func DismissSoftKeyboard(e runtime.Exec) {
	_ = e.Sess().Client.KeyEvent("BACK")
	time.Sleep(e.Sess().PollInterval())
}

// KeyboardSettingsBlocking reports whether detector classifies Gboard/IME settings as blocking.
func KeyboardSettingsBlocking(e runtime.Exec) bool {
	snap := e.Sess().ReadUI(true)
	pkg := e.Sess().Client.ForegroundPackage()
	return state.IsState(snap, pkg, state.UIKeyboardSettings)
}

// DismissKeyboardSettings closes Gboard Setelan via BACK (registered overlay handler).
func DismissKeyboardSettings(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT dismiss keyboard settings overlay")
	return pollUntilDismissed(e, det, observe, pollDismissOpts{
		avoid:        state.UIKeyboardSettings,
		timeout:      10 * time.Second,
		act:          func() error { DismissSoftKeyboard(e); return nil },
		doneEvent:    "VERIFY keyboard settings dismissed",
		missEvent:    "ACT keyboard settings BACK miss: %v",
		successEvent: "ACT keyboard settings BACK — waiting overlay gone",
		errLabel:     "keyboard settings",
	})
}

// DismissKeyboardSettingsIfPresent proactively backs out when detector sees keyboard settings.
func DismissKeyboardSettingsIfPresent(e runtime.Exec) {
	for i := 0; i < 2; i++ {
		if !KeyboardSettingsBlocking(e) {
			return
		}
		e.Event("ACT dismiss keyboard settings (proactive)")
		DismissSoftKeyboard(e)
	}
}

// ShouldMarkPasswordTyped reports whether password input likely succeeded (detector: not IME-blocked).
func ShouldMarkPasswordTyped(e runtime.Exec) bool {
	DismissKeyboardSettingsIfPresent(e)
	snap := e.Sess().ReadUI(true)
	pkg := e.Sess().Client.ForegroundPackage()
	return !state.IMEBlocksInput(snap, pkg)
}

// KeyboardSettingsOpen is a fast post-tap check via detector.
func KeyboardSettingsOpen(e runtime.Exec) bool {
	return KeyboardSettingsBlocking(e)
}
