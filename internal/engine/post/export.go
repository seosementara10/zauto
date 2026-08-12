package post

import (
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// Exported helpers for internal/engine/fanpage (keeps post ↔ fanpage one-way import).

func EnsureOnFeed(e runtime.Exec, observe state.ObserveFn, det *state.Detector, timeoutSec float64) error {
	return ensureOnFeed(e, observe, det, timeoutSec)
}

func OpenComposer(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	return openComposer(e, observe, invalidate, timeoutSec)
}

func CloseComposerIfOpen(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	return closeComposerIfOpen(e, observe, invalidate, timeoutSec)
}

func OpenMenuDrawer(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	return openMenuDrawer(e, observe, invalidate, timeoutSec)
}

func TypePostText(e runtime.Exec, observe state.ObserveFn, invalidate func(), text string) error {
	return typePostText(e, observe, invalidate, text)
}

func AttachImage(e runtime.Exec, observe state.ObserveFn, invalidate func(), localPath string, timeoutSec float64) error {
	return attachImage(e, observe, invalidate, localPath, timeoutSec)
}

func TapPublish(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	return tapPublish(e, observe, invalidate, timeoutSec)
}

func ComposerScreenOpen(e runtime.Exec, snap ui.Snapshot) bool {
	return composerScreenOpen(e, snap)
}

func PostSettingsScreenOpen(snap ui.Snapshot) bool {
	return postSettingsScreenOpen(snap)
}

func ProfileSwitcherOpen(e runtime.Exec, snap ui.Snapshot) bool {
	return profileSwitcherOpen(e, snap)
}

func OnFeedVisible(e runtime.Exec, snap ui.Snapshot) bool {
	return onFeedVisible(e, snap)
}

func PostTextVisibleOnFeed(e runtime.Exec, snap ui.Snapshot, text string) bool {
	return postTextVisibleOnFeed(e, snap, text)
}

func FeedFreshPostVisible(snap ui.Snapshot, screenH int) bool {
	return feedFreshPostVisible(snap, screenH)
}

func FeedPostPublishing(snap ui.Snapshot, screenH int) bool {
	return feedPostPublishing(snap, screenH)
}

func ScreenDims(e runtime.Exec) (w, h int) {
	return screenDims(e)
}

func MarkPublished(e runtime.Exec, category string) {
	markPublished(e, category)
}

func SnapHeight(snap ui.Snapshot) int {
	return snapHeight(snap)
}
