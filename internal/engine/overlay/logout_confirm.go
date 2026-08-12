package overlay

import (
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func TapLogoutConfirm(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT tap LOGOUT (confirm dialog after Keluar/LAIN KALI)")
	return pollUntilDismissed(e, det, observe, pollDismissOpts{
		avoid:        state.UILogoutConfirmPrompt,
		timeout:      15 * time.Second,
		act:          func() error { return TapLogoutConfirmButton(e, observe, nil, 3) },
		doneEvent:    "VERIFY logout confirm dismissed",
		missEvent:    "ACT LOGOUT tap miss: %v",
		successEvent: "ACT LOGOUT tapped — waiting dialog gone",
		errLabel:     "logout confirm: LOGOUT",
	})
}

// TapLogoutConfirmButton taps the LOGOUT button on the confirm dialog (also used by logout flow fallback).
func TapLogoutConfirmButton(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	queries := []ui.FindQuery{
		{Texts: state.LogoutConfirmButtonTexts, PreferClickable: true, Prefer: "right"},
		{Texts: []string{"LOGOUT"}, PreferClickable: true, Prefer: "right"},
		{Texts: state.LogoutConfirmButtonTexts, PreferClickable: true, Prefer: "bottom"},
	}
	return e.PollTapObserve(observe, invalidate, queries, timeoutSec)
}
