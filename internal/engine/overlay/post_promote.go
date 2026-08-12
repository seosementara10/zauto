package overlay

import (
	"fmt"
	"strings"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// DismissPostPromotePrompt checks "don't show again" and taps Lain Kali on the post-publish boost sheet.
func DismissPostPromotePrompt(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT dismiss post_promote (checkbox + Lain Kali)")
	return pollUntilDismissed(e, det, observe, pollDismissOpts{
		avoid:        state.UIPostPromotePrompt,
		timeout:      20 * time.Second,
		act:          func() error { return dismissPostPromoteOnce(e, observe) },
		doneEvent:    "VERIFY post promote dismissed",
		missEvent:    "ACT post promote dismiss miss: %v",
		successEvent: "ACT post promote checkbox + Lain Kali tapped",
		errLabel:     "post promote prompt",
	})
}

// DismissPostPromotePromptIfPresent polls briefly and dismisses the boost sheet when detected.
func DismissPostPromotePromptIfPresent(e runtime.Exec, observe state.ObserveFn, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	det := state.NewDetector()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap, pkg, act := observe()
		d := det.Detect(snap, pkg, act)
		if d.State != state.UIPostPromotePrompt || d.Confidence < state.VerifyMinConfidence {
			if !state.IsState(snap, pkg, state.UIPostPromotePrompt) {
				return nil
			}
		}
		if err := dismissPostPromoteOnce(e, observe); err != nil {
			return err
		}
		time.Sleep(e.Sess().PollInterval())
	}
	snap, pkg, _ := observe()
	if state.IsState(snap, pkg, state.UIPostPromotePrompt) {
		return fmt.Errorf("post promote prompt not dismissed")
	}
	return nil
}

func dismissPostPromoteOnce(e runtime.Exec, observe state.ObserveFn) error {
	snap := e.Sess().ReadUI(false)
	pkg := e.Sess().Client.ForegroundPackage()
	if !state.IsState(snap, pkg, state.UIPostPromotePrompt) {
		return fmt.Errorf("post promote prompt not visible")
	}
	_, h := e.Sess().ScreenSize()
	if h <= 0 {
		h = 1600
	}
	_ = tapPostPromoteDontShowAgain(e, snap, h)
	time.Sleep(e.Sess().PollInterval())
	return tapPostPromoteLater(e, observe)
}

func tapPostPromoteDontShowAgain(e runtime.Exec, snap ui.Snapshot, screenH int) bool {
	minY := screenH * 45 / 100
	if hit := findPostPromoteDontShowRow(snap, minY); hit != nil {
		x, y := hit.Center()
		e.Event("ACT tap post promote dont-show %q at (%d,%d)", hit.Label, x, y)
		if err := e.Sess().Client.Tap(x, y); err != nil {
			return false
		}
		return true
	}
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		_, cy := elem.Center()
		if cy < minY {
			continue
		}
		cn := elem.ClassName
		if strings.Contains(cn, "CheckBox") || strings.Contains(cn, "CompoundButton") {
			x, y := elem.Center()
			e.Event("ACT tap post promote checkbox at (%d,%d)", x, y)
			if err := e.Sess().Client.Tap(x, y); err != nil {
				return false
			}
			return true
		}
	}
	return false
}

func findPostPromoteDontShowRow(snap ui.Snapshot, minY int) *ui.Resolved {
	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy < minY {
			continue
		}
		label := strings.ToLower(elem.Label())
		matched := false
		for _, want := range state.PostPromoteDontShowTexts {
			if strings.Contains(label, strings.ToLower(want)) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		r := &ui.Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		if best == nil || (elem.Clickable && !best.Element.Clickable) {
			best = r
		}
	}
	if best != nil {
		return best
	}
	return ui.NewDefaultResolver().Find(snap, ui.FindQuery{
		Texts:           state.PostPromoteDontShowTexts,
		PreferClickable: true,
		Prefer:          "bottom",
		MinCenterY:      minY,
	})
}

func tapPostPromoteLater(e runtime.Exec, observe state.ObserveFn) error {
	snap := e.Sess().ReadUI(false)
	_, h := e.Sess().ScreenSize()
	if h <= 0 {
		h = 1600
	}
	minY := h * 50 / 100
	if hit := findPostPromoteLaterButton(snap, minY); hit != nil {
		x, y := hit.Center()
		e.Event("ACT tap post promote Lain Kali %q at (%d,%d)", hit.Label, x, y)
		if err := e.Sess().Client.Tap(x, y); err != nil {
			return err
		}
		return nil
	}
	queries := []ui.FindQuery{
		{Texts: []string{"Lain Kali", "Lain kali", "LAIN KALI"}, PreferClickable: true, Prefer: "bottom", MinCenterY: minY},
		{Texts: state.PostPromoteLaterTexts, PreferClickable: true, Prefer: "bottom", MinCenterY: minY},
		{Texts: state.PostPromoteLaterTexts, PreferClickable: true, Prefer: "left", MinCenterY: minY},
	}
	return e.TapFirstQueryObserve(observe, nil, queries, 5)
}

func findPostPromoteLaterButton(snap ui.Snapshot, minY int) *ui.Resolved {
	want := map[string]struct{}{}
	for _, label := range state.PostPromoteLaterTexts {
		if n := ui.Normalize(label); n != "" {
			want[n] = struct{}{}
		}
	}
	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Clickable || !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy < minY {
			continue
		}
		for _, raw := range []string{elem.Text, elem.ContentDesc} {
			if raw == "" {
				continue
			}
			if _, ok := want[ui.Normalize(raw)]; !ok {
				continue
			}
			r := &ui.Resolved{Element: elem, Label: strings.TrimSpace(raw), Bounds: elem.Bounds}
			if best == nil || cy > bestCenterY(best) {
				best = r
			}
		}
	}
	return best
}

func bestCenterY(r *ui.Resolved) int {
	if r == nil {
		return 0
	}
	_, y := r.Center()
	return y
}
