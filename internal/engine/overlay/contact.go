package overlay

import (
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func SkipLocationServices(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT skip location services → Lewati")
	return e.ActUntilNotState(det, observe, state.UILocationServicesPrompt, 12*time.Second, "location services", func() error {
		q := ui.FindQuery{Texts: state.PostLoginSkipTexts, PreferClickable: true, Prefer: "top"}
		return e.TapFirstQuery([]ui.FindQuery{q}, 5)
	})
}

func SkipContactFollow(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT skip contact → Lewati (may need 2 taps: screen + confirm)")
	return e.ActUntilNotState(det, observe, state.UIContactFollowPrompt, 15*time.Second, "contact follow", func() error {
		return tapContactSkipLewati(e, 3)
	})
}

func tapContactSkipLewati(e runtime.Exec, maxAttempts int) error {
	q := ui.FindQuery{Texts: state.ContactFollowSkipTexts, PreferClickable: true, Prefer: "first"}
	for i := 0; i < maxAttempts; i++ {
		if err := e.TapFirstQuery([]ui.FindQuery{q}, 4); err != nil {
			if i == 0 {
				return err
			}
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return nil
}
