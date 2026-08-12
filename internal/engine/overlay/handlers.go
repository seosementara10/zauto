package overlay

import (
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

// Handler executes one classified overlay state (permission, save login, etc.).
type Handler func(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error

var handlers = map[state.UIState]Handler{
	state.UIPermission:             HandlePermission,
	state.UIPasswordManagerSheet:   DismissPasswordManager,
	state.UISaveLoginPrompt:        DismissSaveLogin,
	state.UILoginAccountFinderPrompt: DismissLoginAccountFinder,
	state.UILogoutConfirmPrompt:    TapLogoutConfirm,
	state.UIContactFollowPrompt:    SkipContactFollow,
	state.UILocationServicesPrompt: SkipLocationServices,
	state.UILocaleSetupError:       LocaleSetupError,
	state.UISavedProfileScreen:     RemoveSavedProfile,
	state.UIFanpageHomeIntro:       DismissFanpageHomeIntro,
	state.UIPostPromotePrompt:      DismissPostPromotePrompt,
	state.UIKeyboardSettings:       DismissKeyboardSettings,
}

// Dispatch runs the overlay handler for a resolved detection.
func Dispatch(e runtime.Exec, det *state.Detector, observe state.ObserveFn, d state.Detection) error {
	h, ok := handlers[d.State]
	if !ok {
		return e.ErrUnhandledState(d)
	}
	e.Event("DISPATCH state=%s confidence=%.0f%%", d.State, d.Confidence*100)
	return h(e, det, observe)
}
