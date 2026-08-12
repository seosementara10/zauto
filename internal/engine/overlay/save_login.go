package overlay

import (
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// DismissSaveLogin taps LAIN KALI on "Simpan info login" — shared by login and logout flows.
func DismissSaveLogin(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT tap LAIN KALI (dismiss save login)")
	return pollUntilDismissed(e, det, observe, pollDismissOpts{
		avoid:        state.UISaveLoginPrompt,
		timeout:      15 * time.Second,
		act:          func() error { return tapSaveLoginLater(e, observe) },
		doneEvent:    "VERIFY save login dismissed",
		missEvent:    "ACT LAIN KALI tap miss: %v",
		successEvent: "ACT LAIN KALI tapped — waiting sheet gone",
		errLabel:     "save login: LAIN KALI",
	})
}

func tapSaveLoginLater(e runtime.Exec, observe state.ObserveFn) error {
	queries := []ui.FindQuery{
		{Texts: []string{"LAIN KALI"}, PreferClickable: true, Prefer: "first"},
		{Texts: []string{"Lain Kali", "Lain kali"}, PreferClickable: true, Prefer: "left"},
		{Texts: state.SaveLoginLaterTexts, PreferClickable: true, Prefer: "left"},
		{Texts: state.SaveLoginLaterTexts, PreferClickable: true, Prefer: "bottom"},
		{Texts: state.SaveLoginLaterTexts, PreferClickable: false, Prefer: "left"},
	}
	return e.TapFirstQueryObserve(observe, nil, queries, 5)
}
