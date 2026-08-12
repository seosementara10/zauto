package post

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"zauto/internal/engine/overlay"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func ensureOnFeed(e runtime.Exec, observe state.ObserveFn, det *state.Detector, timeoutSec float64) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap, pkg, act := observe()
		d := det.Detect(snap, pkg, act)
		if d.State == state.UILoggedIn && d.Confidence >= state.VerifyMinConfidence {
			e.Event("VERIFY on feed logged_in confidence=%.0f%%", d.Confidence*100)
			return nil
		}
		if e.Sess().Resolver.TextExists(snap, state.LoggedInFeedHints) {
			e.Event("VERIFY on feed via hints")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("not on logged-in feed before post")
}

func openComposer(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	_, h := e.Sess().ScreenSize()
	queries := []ui.FindQuery{
		{Texts: state.FeedComposerTexts, PreferClickable: true, Prefer: "top", MaxCenterY: h * 40 / 100},
		{Texts: state.FeedComposerTexts, PreferClickable: false, Prefer: "top", MaxCenterY: h * 40 / 100},
		{ContentDescs: state.FeedComposerTexts, PreferClickable: true, Prefer: "top", MaxCenterY: h * 40 / 100},
	}
	if err := e.PollTapObserve(observe, invalidate, queries, timeoutSec); err != nil {
		return fmt.Errorf("composer entry not found: %w", err)
	}
	e.Event("ACT composer opened")
	return waitComposerScreen(e, observe, invalidate, timeoutSec)
}

func waitComposerScreen(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		e.InvalidateObserve(invalidate)
		snap := e.ReadSnap(observe)
		if composerScreenOpen(e, snap) {
			e.Event("VERIFY composer screen open")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("create post screen not visible")
}

func typePostText(e runtime.Exec, observe state.ObserveFn, invalidate func(), text string) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	snap := e.ReadSnap(observe)
	var ref *ui.Element
	if edit := ui.ComposerEditField(snap); edit != nil {
		ref = edit
	} else {
		if err := e.PollTapObserve(observe, invalidate, []ui.FindQuery{
			{Texts: state.FeedComposerTexts, PreferClickable: true, Prefer: "first"},
		}, 3); err != nil {
			return fmt.Errorf("composer field not found: %w", err)
		}
		snap = e.ReadSnap(observe)
		ref = ui.ComposerEditField(snap)
		if ref == nil {
			return fmt.Errorf("composer edit field not found")
		}
	}
	x, y := ref.Center()
	if err := e.Sess().Client.Tap(x, y); err != nil {
		return err
	}
	e.InvalidateObserve(invalidate)
	if err := waitComposerFieldReady(e, observe, *ref, 3*time.Second); err != nil {
		return err
	}
	if err := e.Sess().Client.InputText(text); err != nil {
		return fmt.Errorf("input post text: %w", err)
	}
	e.InvalidateObserve(invalidate)
	if err := waitComposerTextPopulated(e, observe, text, 8*time.Second); err != nil {
		return err
	}
	e.Event("ACT typed post text len=%d", len(text))
	return nil
}

func waitComposerFieldReady(e runtime.Exec, observe state.ObserveFn, ref ui.Element, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if field := ui.FindEditAtBounds(snap, ref); field != nil && (field.Focused || field.Enabled) {
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("composer field not ready")
}

func waitComposerTextPopulated(e runtime.Exec, observe state.ObserveFn, text string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if composerTextContains(e, snap, text) {
			e.Event("VERIFY composer text populated")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("composer text not visible after input")
}

func attachImage(e runtime.Exec, observe state.ObserveFn, invalidate func(), localPath string, timeoutSec float64) error {
	if localPath == "" {
		return nil
	}
	remoteName := "zauto_" + filepath.Base(localPath)
	remote := "/sdcard/Pictures/zauto/" + remoteName
	if _, err := e.Sess().Client.Shell("mkdir", "-p", "/sdcard/Pictures/zauto"); err != nil {
		return err
	}
	if err := e.Sess().Client.PushFile(localPath, remote); err != nil {
		return fmt.Errorf("push image: %w", err)
	}
	_ = e.Sess().Client.ScanMediaFile(remote)
	e.Event("ACT image pushed to %s", remote)

	photoQueries := []ui.FindQuery{
		{Texts: state.PostPhotoButtonTexts, PreferClickable: true, Prefer: "bottom"},
		{ContentDescs: state.PostPhotoButtonTexts, PreferClickable: true, Prefer: "bottom"},
		{ResourceIDs: []string{"photo", "gallery", "media_picker"}, PreferClickable: true, Prefer: "bottom"},
	}
	if err := e.PollTapObserve(observe, invalidate, photoQueries, timeoutSec); err != nil {
		return fmt.Errorf("photo button not found: %w", err)
	}
	if err := waitGalleryPicker(e, observe, timeoutSec); err != nil {
		return fmt.Errorf("gallery picker not open: %w", err)
	}

	_ = e.PollTapObserve(observe, invalidate, []ui.FindQuery{
		{Texts: state.GalleryRecentTexts, PreferClickable: true, Prefer: "first"},
	}, 4)

	if err := tapFirstGalleryThumbnail(e, observe, invalidate); err != nil {
		return fmt.Errorf("gallery image not selected: %w", err)
	}
	if err := waitImageAttached(e, observe, timeoutSec); err != nil {
		return fmt.Errorf("image attachment not verified: %w", err)
	}
	e.Event("ACT image selected from gallery")
	return nil
}

func waitGalleryPicker(e runtime.Exec, observe state.ObserveFn, timeoutSec float64) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if galleryPickerOpen(e, snap) {
			e.Event("VERIFY gallery picker open")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("gallery picker not visible")
}

func waitImageAttached(e runtime.Exec, observe state.ObserveFn, timeoutSec float64) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if !galleryPickerOpen(e, snap) && imageAttachedInComposer(e, snap) {
			e.Event("VERIFY image attached in composer")
			return nil
		}
		if !galleryPickerOpen(e, snap) && composerScreenOpen(e, snap) {
			e.Event("VERIFY back on composer after image pick")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("image attachment not confirmed")
}

func tapFirstGalleryThumbnail(e runtime.Exec, observe state.ObserveFn, invalidate func()) error {
	snap := e.ReadSnap(observe)
	_, h := e.Sess().ScreenSize()
	minY := h * 20 / 100
	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		cn := elem.ClassName
		if !strings.Contains(cn, "ImageView") && !strings.Contains(cn, "ImageButton") {
			continue
		}
		cx, cy := elem.Center()
		if cy < minY {
			continue
		}
		if elem.Width() < 80 || elem.Height() < 80 {
			continue
		}
		r := &ui.Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		if best == nil {
			best = r
			continue
		}
		bestCX, bestCY := best.Center()
		if cy < bestCY || (cy == bestCY && cx < bestCX) {
			best = r
		}
	}
	if best == nil {
		return fmt.Errorf("no gallery thumbnail found")
	}
	x, y := best.Center()
	e.Event("ACT tap gallery thumb at (%d,%d)", x, y)
	if err := e.Sess().Client.Tap(x, y); err != nil {
		return err
	}
	e.InvalidateObserve(invalidate)
	return nil
}

func verifyPosted(e runtime.Exec, category string, content Content, timeoutSec float64) error {
	observe, invalidate := e.CachedObserve()
	_, h := screenDims(e)
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		e.InvalidateObserve(invalidate)
		_ = dismissInAppWebViewSnap(e, observe, invalidate, 0.5)
		_ = overlay.DismissPostPromotePromptIfPresent(e, observe, 2*time.Second)
		snap := e.ReadSnap(observe)
		if ok, label := feedPostVerified(e, snap, h, category, content); ok {
			logScreenContext(e, observe, "verify_post_ok:"+label, nil)
			e.Event("VERIFY %s", label)
			return nil
		}
		logScreenContextIfStale(e, observe, "verify_post_waiting", nil, 5*time.Second)
		time.Sleep(e.Sess().PollInterval())
	}
	logScreenContext(e, observe, "verify_post_timeout", nil)
	capturePostFailure(e, observe, "post_verify_timeout", nil, "personal")
	if isPublished(e, category) {
		e.Event("VERIFY post accepted after publish tap (observe timeout)")
		return nil
	}
	return fmt.Errorf("post not verified within timeout")
}

func postTextVisibleOnFeed(e runtime.Exec, snap ui.Snapshot, text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return true
	}
	if e.Sess().Resolver.TextExists(snap, []string{text}) {
		return true
	}
	if prefix := verifyTextPrefix(text); prefix != "" {
		return e.Sess().Resolver.TextExists(snap, []string{prefix})
	}
	return false
}

func verifyTextPrefix(text string) string {
	text = strings.TrimSpace(text)
	if len(text) <= 30 {
		return text
	}
	return strings.TrimSpace(text[:30])
}
