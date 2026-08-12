package post

import (
	"strings"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func onFeedVisible(e runtime.Exec, snap ui.Snapshot) bool {
	if e.Sess().Resolver.TextExists(snap, state.LoggedInFeedHints) {
		return true
	}
	return e.Sess().Resolver.TextExists(snap, state.FeedComposerTexts)
}

// feedPostPublishing reports an in-feed upload progress bar (common right after tapping Posting).
func feedPostPublishing(snap ui.Snapshot, screenH int) bool {
	if screenH <= 0 {
		screenH = 1600
	}
	minY := screenH * 12 / 100
	maxY := screenH * 75 / 100
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		if !strings.Contains(elem.ClassName, "ProgressBar") {
			continue
		}
		_, cy := elem.Center()
		if cy >= minY && cy <= maxY {
			return true
		}
	}
	return false
}

// feedFreshPostVisible reports a just-published post row on the home feed.
func feedFreshPostVisible(snap ui.Snapshot, screenH int) bool {
	if screenH <= 0 {
		screenH = 1600
	}
	minY := screenH * 18 / 100
	for _, elem := range snap.Elements {
		_, cy := elem.Center()
		if cy < minY {
			continue
		}
		raw := strings.ToLower(strings.TrimSpace(elem.Text))
		if raw == "" {
			raw = strings.ToLower(strings.TrimSpace(elem.ContentDesc))
		}
		if raw == "" {
			continue
		}
		if strings.Contains(raw, "baru saja") || strings.Contains(raw, "just now") {
			return true
		}
		if strings.Contains(raw, "sedang mengirim") || strings.Contains(raw, "sending") {
			return true
		}
		if strings.Contains(raw, "mengunggah") || strings.Contains(raw, "uploading") {
			return true
		}
	}
	return false
}

func feedPostVerified(e runtime.Exec, snap ui.Snapshot, screenH int, category string, content Content) (bool, string) {
	if !onFeedVisible(e, snap) {
		return false, ""
	}
	// Positive feed signals win over stale composer nodes left in the hierarchy after publish.
	if isPublished(e, category) {
		return true, "back on feed after publish"
	}
	if postTextVisibleOnFeed(e, snap, content.Text) {
		return true, "post text visible on feed"
	}
	if feedPostPublishing(snap, screenH) {
		return true, "feed post uploading (progress bar)"
	}
	if feedFreshPostVisible(snap, screenH) {
		return true, "fresh post on feed (Baru saja/loading)"
	}
	if composerScreenOpen(e, snap) {
		return false, ""
	}
	return false, ""
}
