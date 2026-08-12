package overlay

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func RemoveSavedProfile(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT saved_profile → ⋯ menu → Hapus Profil dari perangkat ini")
	return e.ActUntilNotState(det, observe, state.UISavedProfileScreen, 20*time.Second, "saved profile", func() error {
		return removeSavedProfileFromDevice(e)
	})
}

func removeSavedProfileFromDevice(e runtime.Exec) error {
	_, h := e.Sess().ScreenSize()
	menuQueries := []ui.FindQuery{
		{ContentDescs: state.SavedProfileMenuContentDescs, PreferClickable: true, Prefer: "top", MaxCenterY: h * 45 / 100},
		{ResourceIDs: []string{"overflow", "more_options", "profile_menu", "button_icon"}, PreferClickable: true, Prefer: "top", MaxCenterY: h * 45 / 100},
	}
	if err := e.TapFirstQuery(menuQueries, 10); err != nil {
		return fmt.Errorf("profile ⋯ menu: %w", err)
	}
	time.Sleep(e.Sess().PollInterval())

	removeQ := ui.FindQuery{Texts: state.RemoveProfileFromDeviceTexts, PreferClickable: true, Prefer: "first"}
	if err := e.TapFirstQuery([]ui.FindQuery{removeQ}, 10); err != nil {
		return fmt.Errorf("hapus profil dari perangkat: %w", err)
	}
	time.Sleep(e.Sess().PollInterval())

	confirmQ := ui.FindQuery{Texts: state.RemoveProfileConfirmTexts, PreferClickable: true, Prefer: "bottom"}
	_ = e.TapFirstQuery([]ui.FindQuery{confirmQ}, 6)
	time.Sleep(e.Sess().PollInterval())
	return nil
}
