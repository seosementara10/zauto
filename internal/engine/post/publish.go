package post

import (
	"zauto/internal/engine/runtime"
)

// publishToFeed opens composer, fills content, publishes, and verifies on current profile feed.
func publishToFeed(e runtime.Exec, category string, content Content, composerTimeout, verifyTimeout float64) error {
	observe, invalidate := e.CachedObserve()

	if err := openComposer(e, observe, invalidate, composerTimeout); err != nil {
		e.CaptureFailure("post_open_composer", "open_composer", err.Error(), observe, runtime.ScreenNote{})
		return err
	}
	e.InvalidateObserve(invalidate)

	if err := typePostText(e, observe, invalidate, content.Text); err != nil {
		return err
	}
	if content.ImagePath != "" {
		if err := attachImage(e, observe, invalidate, content.ImagePath, composerTimeout); err != nil {
			e.CaptureFailure("post_attach_image", "attach_image", err.Error(), observe, runtime.ScreenNote{})
			return err
		}
	}
	if err := tapPublish(e, observe, invalidate, composerTimeout); err != nil {
		e.CaptureFailure("post_publish", "publish", err.Error(), observe, runtime.ScreenNote{})
		return err
	}
	markPublished(e, category)
	return verifyPosted(e, category, content, verifyTimeout)
}
