package post

import (
	"log"

	"zauto/internal/config"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

// Run posts text and/or image to the personal account feed (Beranda).
func Run(e runtime.Exec, action config.Action) error {
	content, err := ResolveContent(e, action)
	if err != nil {
		return err
	}
	e.Event("POST start personal %s", content.Summary())

	composerTimeout := action.ParamFloat("composer_timeout_sec", 15)
	verifyTimeout := action.ParamFloat("verify_timeout_sec", 45)
	category := postCategoryForAction(action)

	if isPublished(e, category) {
		e.Event("POST skip re-post — already published this run")
		_ = verifyPosted(e, category, content, verifyTimeout)
		e.Event("POST ok personal %s", content.Summary())
		log.Printf("[%s] facebook_auto_post: ok (already published)", e.Sess().Serial)
		return nil
	}

	det := state.NewDetector()
	observe, _ := e.CachedObserve()
	if err := ensureOnFeed(e, observe, det, composerTimeout); err != nil {
		return err
	}
	if err := publishToFeed(e, category, content, composerTimeout, verifyTimeout); err != nil {
		if isPublished(e, category) {
			e.Event("POST ok personal after publish %s", content.Summary())
			log.Printf("[%s] facebook_auto_post: ok (published, verify skipped)", e.Sess().Serial)
			return nil
		}
		return err
	}
	e.Event("POST ok personal %s", content.Summary())
	log.Printf("[%s] facebook_auto_post: ok", e.Sess().Serial)
	return nil
}
