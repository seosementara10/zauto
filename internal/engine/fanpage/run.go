package fanpage

import (
	"fmt"
	"log"
	"strings"

	"zauto/internal/config"
	"zauto/internal/engine/post"
	"zauto/internal/engine/runtime"
)

// Run posts to one or more fanpages linked to the logged-in account.
func Run(e runtime.Exec, action config.Action) error {
	targets, err := ResolveFanpageTargets(e, action)
	if err != nil {
		return err
	}
	content, err := post.ResolveContent(e, action)
	if err != nil {
		return err
	}
	composerTimeout := action.ParamFloat("composer_timeout_sec", 15)
	verifyTimeout := action.ParamFloat("verify_timeout_sec", 45)
	switchTimeout := action.ParamFloat("switch_timeout_sec", 20)

	observe, invalidate := e.CachedObserve()

	for i, page := range targets {
		e.Event("POST fanpage %d/%d fb_id=%s %s", i+1, len(targets), page.FBPageID, content.Summary())
		logScreenContext(e, observe, "fanpage_post:start", &page)

		if isFanpagePublished(e, page.FBPageID) || fanpagePostAlreadyVisible(e, observe, page, content) {
			e.Event("POST skip fanpage page=%s — already published or visible on feed", page.FBPageID)
			markFanpagePublished(e, page.FBPageID)
			continue
		}

		if err := ensureFanpageBeforeCompose(e, observe, invalidate, page, switchTimeout, composerTimeout); err != nil {
			return fmt.Errorf("fanpage %s composer: %w", page.FBPageID, err)
		}
		e.InvalidateObserve(invalidate)
		if err := post.TypePostText(e, observe, invalidate, content.Text); err != nil {
			return err
		}
		if content.ImagePath != "" {
			if err := post.AttachImage(e, observe, invalidate, content.ImagePath, composerTimeout); err != nil {
				return fmt.Errorf("fanpage %s image: %w", page.FBPageID, err)
			}
		}
		logScreenContext(e, observe, "fanpage_post:before_publish", &page)
		if err := post.TapPublish(e, observe, invalidate, composerTimeout); err != nil {
			capturePostFailure(e, observe, "fanpage_publish", &page, err.Error())
			return fmt.Errorf("fanpage %s publish: %w", page.FBPageID, err)
		}
		markFanpagePublished(e, page.FBPageID)
		if err := verifyFanpagePosted(e, content, page, verifyTimeout); err != nil {
			e.Event("POST ok fanpage fb_id=%s (published, verify soft-fail: %v)", page.FBPageID, err)
		} else {
			e.Event("POST ok fanpage fb_id=%s", page.FBPageID)
		}
	}

	mode := strings.ToLower(action.ParamString("fanpage_mode", "single"))
	log.Printf("[%s] facebook_fanpage_post: ok pages=%d mode=%s", e.Sess().Serial, len(targets), mode)
	return nil
}
