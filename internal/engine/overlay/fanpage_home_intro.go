package overlay

import (
	"fmt"
	"strings"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// DismissFanpageHomeIntro closes the "Beranda Khusus" carousel after switching to a fanpage (tap X top-left).
func DismissFanpageHomeIntro(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	e.Event("ACT dismiss fanpage_home_intro (tap X)")
	return e.ActUntilNotState(det, observe, state.UIFanpageHomeIntro, 15*time.Second, "fanpage home intro", func() error {
		return tapFanpageHomeIntroClose(e, observe)
	})
}

// DismissFanpageHomeIntroIfPresent polls briefly and dismisses the intro when detected.
func DismissFanpageHomeIntroIfPresent(e runtime.Exec, observe state.ObserveFn, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	det := state.NewDetector()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap, pkg, act := observe()
		d := det.Detect(snap, pkg, act)
		if d.State != state.UIFanpageHomeIntro || d.Confidence < state.VerifyMinConfidence {
			if !state.IsState(snap, pkg, state.UIFanpageHomeIntro) {
				return nil
			}
		}
		if err := tapFanpageHomeIntroClose(e, observe); err != nil {
			return err
		}
		time.Sleep(e.Sess().PollInterval())
	}
	snap, pkg, _ := observe()
	if state.IsState(snap, pkg, state.UIFanpageHomeIntro) {
		return fmt.Errorf("fanpage home intro not dismissed")
	}
	return nil
}

func tapFanpageHomeIntroClose(e runtime.Exec, observe state.ObserveFn) error {
	snap := e.Sess().ReadUI(false)
	w, h := e.Sess().ScreenSize()
	if r := findFanpageHomeIntroClose(e.Sess().Resolver, snap, w, h); r != nil {
		x, y := r.Center()
		if err := e.Sess().Client.Tap(x, y); err != nil {
			return err
		}
		e.Event("ACT tapped fanpage intro close X at (%d,%d) label=%q", x, y, r.Label)
		time.Sleep(e.Sess().PollInterval())
		return nil
	}

	maxY := h * 20 / 100
	if r := e.Sess().Resolver.Find(snap, ui.FindQuery{
		Texts:           state.FanpageHomeIntroSkipTexts,
		PreferClickable: true,
		Prefer:          "top",
		MaxCenterY:      maxY,
	}); r != nil {
		x, y := r.Center()
		if err := e.Sess().Client.Tap(x, y); err != nil {
			return err
		}
		e.Event("ACT tapped fanpage intro Lewati at (%d,%d)", x, y)
		time.Sleep(e.Sess().PollInterval())
		return nil
	}

	if x, y, ok := fanpageHomeIntroCloseFallback(w, h); ok {
		e.Event("ACT fanpage intro X not in hierarchy — tap header (%d,%d)", x, y)
		if err := e.Sess().Client.Tap(x, y); err != nil {
			return err
		}
		time.Sleep(e.Sess().PollInterval())
		return nil
	}

	return fmt.Errorf("fanpage home intro close control not found")
}

func findFanpageHomeIntroClose(resolver *ui.Resolver, snap ui.Snapshot, screenW, screenH int) *ui.Resolved {
	if resolver == nil {
		resolver = ui.NewDefaultResolver()
	}
	if !state.IsState(snap, "", state.UIFanpageHomeIntro) {
		return nil
	}
	if screenW <= 0 {
		screenW = 720
	}
	if screenH <= 0 {
		screenH = 1600
	}
	maxX := screenW * 28 / 100
	maxY := screenH * 20 / 100

	closeDescs := append([]string{}, state.OverlayCloseContentDescs...)
	closeDescs = append(closeDescs, "Navigate up", "Back", "Kembali")

	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Enabled || elem.Width() <= 0 {
			continue
		}
		cx, cy := elem.Center()
		if cx > maxX || cy > maxY {
			continue
		}
		if !fanpageIntroCloseCandidate(elem) {
			continue
		}
		for _, raw := range []string{elem.Text, elem.ContentDesc} {
			if raw == "" {
				continue
			}
			n := ui.Normalize(raw)
			for _, t := range state.OverlayCloseTexts {
				if raw == t || n == ui.Normalize(t) {
					r := &ui.Resolved{Element: elem, Label: strings.TrimSpace(raw), Bounds: elem.Bounds}
					if best == nil || cx < bestCenterX(best) {
						best = r
					}
				}
			}
			for _, want := range closeDescs {
				if want == "" {
					continue
				}
				if n == ui.Normalize(want) || strings.Contains(n, ui.Normalize(want)) {
					r := &ui.Resolved{Element: elem, Label: strings.TrimSpace(raw), Bounds: elem.Bounds}
					if best == nil || cx < bestCenterX(best) {
						best = r
					}
				}
			}
		}
	}
	return best
}

func fanpageIntroCloseCandidate(elem ui.Element) bool {
	if elem.Clickable {
		return true
	}
	cn := elem.ClassName
	return strings.Contains(cn, "ImageButton") ||
		strings.Contains(cn, "ImageView") ||
		strings.Contains(cn, "Button")
}

func fanpageHomeIntroCloseFallback(screenW, screenH int) (x, y int, ok bool) {
	if screenW <= 0 {
		screenW = 720
	}
	if screenH <= 0 {
		screenH = 1600
	}
	return screenW * 9 / 100, screenH * 8 / 100, true
}

func bestCenterX(r *ui.Resolved) int {
	if r == nil {
		return 0
	}
	x, _ := r.Center()
	return x
}
