package post

import (
	"fmt"
	"strings"
	"time"

	"zauto/internal/engine/overlay"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func tapPublish(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	_ = dismissInAppWebViewSnap(e, observe, invalidate, 2)
	if err := tapFinalPublish(e, observe, invalidate, timeoutSec); err != nil {
		return err
	}
	e.Event("ACT tapped publish")
	logScreenContext(e, observe, "after_publish_tap", nil)
	// No composer/feed check here — tap succeeded; verifyPosted handles outcome.
	e.InvalidateObserve(invalidate)
	_ = dismissInAppWebViewSnap(e, observe, invalidate, 3)
	if err := overlay.DismissPostPromotePromptIfPresent(e, observe, 12*time.Second); err != nil {
		e.Event("ACT post promote dismiss soft-fail: %v", err)
	}
	return nil
}

func tapFinalPublish(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	dismissKeyboardSnap(e, observe, invalidate)
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		if dismissInAppWebViewSnap(e, observe, invalidate, 0.5) {
			continue
		}
		snap := e.ReadSnap(observe)
		w, h := screenDims(e)
		recordComposerVariant(e, detectComposerVariant(e.Sess().Resolver, snap, w, h))
		acts := scanComposerActions(snap, w, h)

		if acts.keyboardDone != nil {
			if err := tapResolved(e, acts.keyboardDone, observe, invalidate, "dismiss keyboard"); err != nil {
				return err
			}
			continue
		}
		if acts.next != nil && composerNeedsNextBeforePublish(acts, h) {
			if err := tapComposerNextAndWait(e, acts.next, observe, invalidate, 8); err != nil {
				e.Event("ACT composer next retry: %v", err)
			}
			continue
		}
		if acts.publish != nil {
			x, y := acts.publish.Center()
			e.Event("ACT tap publish %q at (%d,%d)", acts.publish.Label, x, y)
			if err := e.Sess().Client.Tap(x, y); err != nil {
				return err
			}
			e.InvalidateObserve(invalidate)
			return nil
		}
		if acts.next != nil {
			if err := tapComposerNextAndWait(e, acts.next, observe, invalidate, 8); err != nil {
				e.Event("ACT composer next retry: %v", err)
			}
			continue
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("final publish button not found")
}

func tapResolved(e runtime.Exec, hit *ui.Resolved, observe state.ObserveFn, invalidate func(), label string) error {
	// Tap the resolved element's bounds center — not hard-coded screen coordinates.
	x, y := hit.Center()
	e.Event("ACT %s %q at (%d,%d)", label, hit.Label, x, y)
	if err := e.Sess().Client.Tap(x, y); err != nil {
		return err
	}
	e.InvalidateObserve(invalidate)
	return nil
}

func tapComposerNextAndWait(e runtime.Exec, hit *ui.Resolved, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	if err := tapResolved(e, hit, observe, invalidate, "tap composer Berikutnya"); err != nil {
		return err
	}
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		e.InvalidateObserve(invalidate)
		snap := e.ReadSnap(observe)
		w, h := screenDims(e)
		if findPublishButton(snap, w, h) != nil || postSettingsScreenOpen(snap) {
			e.Event("VERIFY publish screen ready after Berikutnya")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("publish button not ready after Berikutnya")
}

func findComposerNextButton(snap ui.Snapshot, screenW, screenH int) *ui.Resolved {
	if screenH <= 0 {
		screenH = 1600
	}
	if screenW <= 0 {
		screenW = 720
	}
	// Fanpage composer: Berikutnya in header (top-right).
	if hit := findExactLabelOpts(snap, state.PostComposerNextTexts, exactLabelOpts{
		minY:        screenH * 5 / 100,
		maxY:        screenH * 25 / 100,
		minX:        screenW * 55 / 100,
		preferRight: true,
	}); hit != nil {
		return hit
	}
	// Personal NEXT review: Berikutnya at bottom.
	return findExactLabel(snap, state.PostComposerNextTexts, screenH*65/100, screenH)
}

func findPublishButton(snap ui.Snapshot, screenW, screenH int) *ui.Resolved {
	if hit := findHeaderPublishButton(snap, screenW, screenH); hit != nil {
		return hit
	}
	return findBottomPublishButton(snap, screenH)
}

// findHeaderPublishButton — DIRECT composer: match text/content-desc "Posting" in header (top-right).
func findHeaderPublishButton(snap ui.Snapshot, screenW, screenH int) *ui.Resolved {
	return findExactLabelOpts(snap, state.PostFinalPublishTexts, exactLabelOpts{
		minY:        screenH * 5 / 100,
		maxY:        screenH * 25 / 100,
		minX:        screenW * 55 / 100,
		preferRight: true,
	})
}

// findBottomPublishButton — NEXT review screen: match wide clickable "Posting" button at bottom.
func findBottomPublishButton(snap ui.Snapshot, screenH int) *ui.Resolved {
	if screenH <= 0 {
		screenH = 1600
	}
	return findExactLabelOpts(snap, state.PostFinalPublishTexts, exactLabelOpts{
		minY:         screenH * 65 / 100,
		minWidth:     120,
		minHeight:    36,
		preferBottom: true,
	})
}

var keyboardDoneLabels = []string{"Selesai", "Done", "Selesai edit", "Done editing"}

func findKeyboardDoneButton(snap ui.Snapshot, screenW, screenH int) *ui.Resolved {
	return findExactLabelOpts(snap, keyboardDoneLabels, exactLabelOpts{
		minY:        screenH * 5 / 100,
		maxY:        screenH * 25 / 100,
		minX:        screenW * 55 / 100,
		preferRight: true,
	})
}

func isFinalPublishLabel(label string) bool {
	n := ui.Normalize(label)
	for _, want := range state.PostFinalPublishTexts {
		if ui.Normalize(want) == n {
			return true
		}
	}
	switch n {
	case "post", "posting", "publish", "kirim", "berbagi", "bagikan", "share":
		return true
	default:
		return false
	}
}

func isKeyboardDoneLabel(label string) bool {
	switch ui.Normalize(label) {
	case "selesai", "done", "selesai edit", "done editing":
		return true
	default:
		return false
	}
}

type exactLabelOpts struct {
	minY, maxY       int
	minX, maxX       int
	minWidth, minHeight int
	preferRight      bool
	preferBottom     bool
}

func findExactLabel(snap ui.Snapshot, labels []string, minY, maxY int, maxX ...int) *ui.Resolved {
	opts := exactLabelOpts{minY: minY, maxY: maxY}
	if len(maxX) > 0 {
		opts.maxX = maxX[0]
	}
	return findExactLabelOpts(snap, labels, opts)
}

// findExactLabelOpts locates a clickable element by exact text/content-desc match (not fuzzy, not fixed x/y tap).
func findExactLabelOpts(snap ui.Snapshot, labels []string, opts exactLabelOpts) *ui.Resolved {
	want := map[string]struct{}{}
	for _, label := range labels {
		if n := ui.Normalize(label); n != "" {
			want[n] = struct{}{}
		}
	}
	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Clickable || !elem.Enabled || elem.Width() <= 0 {
			continue
		}
		cx, cy := elem.Center()
		if opts.minY > 0 && cy < opts.minY {
			continue
		}
		if opts.maxY > 0 && cy > opts.maxY {
			continue
		}
		if opts.minX > 0 && cx < opts.minX {
			continue
		}
		if opts.maxX > 0 && cx > opts.maxX {
			continue
		}
		if opts.minWidth > 0 && elem.Width() < opts.minWidth {
			continue
		}
		if opts.minHeight > 0 && elem.Height() < opts.minHeight {
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
			if best == nil {
				best = r
				continue
			}
			bestCX, bestCY := best.Center()
			switch {
			case opts.preferBottom && cy > bestCY:
				best = r
			case opts.preferRight && cx > bestCX:
				best = r
			case !opts.preferBottom && !opts.preferRight && cy < bestCY:
				best = r
			}
		}
	}
	return best
}

func dismissKeyboardSnap(e runtime.Exec, observe state.ObserveFn, invalidate func()) {
	snap := e.ReadSnap(observe)
	w, h := screenDims(e)
	if done := findKeyboardDoneButton(snap, w, h); done != nil {
		_ = tapResolved(e, done, observe, invalidate, "dismiss keyboard")
		waitKeyboardDismissed(e, observe, 3*time.Second)
		return
	}
	if !keyboardLikelyVisible(snap) {
		return
	}
	_ = e.Sess().Client.KeyEvent("BACK")
	e.InvalidateObserve(invalidate)
	waitKeyboardDismissed(e, observe, 3*time.Second)
}

func keyboardLikelyVisible(snap ui.Snapshot) bool {
	for _, elem := range snap.Elements {
		if ui.IsEditText(elem) && elem.Focused {
			return true
		}
	}
	return false
}

func waitKeyboardDismissed(e runtime.Exec, observe state.ObserveFn, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	w, h := screenDims(e)
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if findKeyboardDoneButton(snap, w, h) == nil && !keyboardLikelyVisible(snap) {
			return
		}
		time.Sleep(e.Sess().PollInterval())
	}
}

// closeComposerIfOpen backs out of composer or post-settings until feed-level UI is visible.
func closeComposerIfOpen(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if !composerScreenOpen(e, snap) && !postSettingsScreenOpen(snap) {
			e.Event("VERIFY composer closed")
			return nil
		}
		e.Event("ACT BACK close composer/settings")
		if err := e.Sess().Client.KeyEvent("BACK"); err != nil {
			return err
		}
		e.InvalidateObserve(invalidate)
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("composer still open after BACK")
}

func dismissInAppWebViewSnap(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) bool {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if !looksLikeInAppWebViewSnap(snap, e) {
			return false
		}
		w, h := screenDims(e)
		closeLabels := append([]string{}, state.OverlayCloseContentDescs...)
		closeLabels = append(closeLabels, "Tutup", "Close")
		if hit := findExactLabel(snap, closeLabels, 0, h*18/100, w*25/100); hit != nil {
			_ = tapResolved(e, hit, observe, invalidate, "dismiss in-app webview")
			e.Event("ACT dismissed in-app webview (Tutup)")
			return true
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return false
}

func looksLikeInAppWebViewSnap(snap ui.Snapshot, e runtime.Exec) bool {
	if e.Sess().Resolver.TextExists(snap, state.LoggedInFeedHints) {
		return false
	}
	if composerScreenOpen(e, snap) {
		return false
	}
	w, h := screenDims(e)
	return findExactLabel(snap, []string{"Tutup", "Close"}, 0, h*18/100, w*25/100) != nil
}
