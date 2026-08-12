package overlay

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func DismissPasswordManager(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT close password_manager_sheet (tap X)")
	deadline := time.Now().Add(15 * time.Second)
	var backUsed bool

	for time.Now().Before(deadline) {
		snap, pkg, actName := observe()
		d := det.Detect(snap, pkg, actName)
		if d.State != state.UIPasswordManagerSheet || d.Confidence < state.VerifyMinConfidence {
			e.Event("VERIFY password_manager_sheet dismissed")
			return nil
		}

		// Slow dump — fast uiautomator dump often misses the sheet close icon.
		snap = e.Sess().ReadUI(false)
		w, h := e.Sess().ScreenSize()
		if r := findPasswordManagerClose(e.Sess().Resolver, snap, w, h); r != nil {
			x, y := r.Center()
			if err := e.Sess().Client.Tap(x, y); err != nil {
				e.Event("ACT close X tap error: %v", err)
			} else {
				e.Event("ACT tapped close X at (%d,%d) label=%q", x, y, r.Label)
				time.Sleep(e.Sess().PollInterval())
				continue
			}
		}

		tapped := false
		for _, q := range passwordManagerCloseQueries(h * 28 / 100) {
			if r := e.Sess().Resolver.Find(snap, q); r != nil {
				x, y := r.Center()
				if err := e.Sess().Client.Tap(x, y); err != nil {
					e.Event("ACT close query tap error: %v", err)
					continue
				}
				e.Event("ACT tapped close via query at (%d,%d) label=%q", x, y, r.Label)
				tapped = true
				break
			}
		}
		if tapped {
			time.Sleep(e.Sess().PollInterval())
			continue
		}

		if x, y, ok := passwordManagerCloseFallbackPoint(e.Sess().Resolver, snap, w, h); ok {
			e.Event("ACT close X not in hierarchy — tap relative header (%d,%d)", x, y)
			if err := e.Sess().Client.Tap(x, y); err != nil {
				e.Event("ACT relative close tap error: %v", err)
			} else {
				time.Sleep(e.Sess().PollInterval())
				continue
			}
		}

		if !backUsed {
			backUsed = true
			e.Event("ACT close X not found — fallback BACK")
			if err := e.Sess().Client.KeyEvent("BACK"); err != nil {
				return err
			}
			time.Sleep(e.Sess().PollInterval())
			continue
		}

		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("password manager sheet: not dismissed within timeout")
}

func passwordManagerCloseQueries(sheetMinY int) []ui.FindQuery {
	return []ui.FindQuery{
		{ContentDescs: state.PasswordManagerCloseContentDescs, Prefer: "right", MinCenterY: sheetMinY},
		{Texts: state.OverlayCloseTexts, Prefer: "right", MinCenterY: sheetMinY},
		{ContentDescs: state.OverlayCloseContentDescs, Prefer: "right", MinCenterY: sheetMinY},
		{ResourceIDs: []string{"close", "dismiss", "cancel", "negative", "nav_up"}, Prefer: "right", MinCenterY: sheetMinY},
	}
}
