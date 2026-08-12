package overlay

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func LocaleSetupError(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	idx := e.DeviceIndex()
	var texts []string
	var choice string
	if idx%2 == 0 {
		texts = state.LocaleSetupTryAgainTexts
		choice = "try_again"
	} else {
		texts = state.LocaleSetupContinueEnglishTexts
		choice = "continue_english"
	}
	e.Event("ACT locale_setup choice=%s device_index=%d", choice, idx)

	return e.ActUntilNotState(det, observe, state.UILocaleSetupError, 15*time.Second, "locale setup", func() error {
		q := ui.FindQuery{Texts: texts, PreferClickable: true, Prefer: "first"}
		if err := e.TapFirstQuery([]ui.FindQuery{q}, 12); err != nil {
			return fmt.Errorf("locale setup %s: %w", choice, err)
		}
		return nil
	})
}
