package post

import (
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

type ComposerVariant int

const (
	ComposerUnknown ComposerVariant = iota
	ComposerDirectPublish
	ComposerNextThenPublish
)

func (v ComposerVariant) String() string {
	switch v {
	case ComposerDirectPublish:
		return "DIRECT"
	case ComposerNextThenPublish:
		return "NEXT_THEN_PUBLISH"
	default:
		return "UNKNOWN"
	}
}

func composerVariantKey() string { return "composer_variant" }

func recordComposerVariant(e runtime.Exec, v ComposerVariant) {
	if v == ComposerUnknown {
		return
	}
	key := composerVariantKey()
	if prev, _ := e.Sess().Runtime[key].(string); prev == v.String() {
		return
	}
	e.Sess().Runtime[key] = v.String()
	e.Event("POST composer variant=%s", v)
}

func detectComposerVariant(resolver *ui.Resolver, snap ui.Snapshot, w, h int) ComposerVariant {
	if findPublishButton(snap, w, h) != nil {
		return ComposerDirectPublish
	}
	if resolver.TextExists(snap, state.PostComposerNewTitleTexts) {
		return ComposerNextThenPublish
	}
	if findComposerNextButton(snap, w, h) != nil {
		return ComposerNextThenPublish
	}
	if resolver.TextExists(snap, []string{"Buat postingan", "Create post", "Create Post"}) {
		return ComposerDirectPublish
	}
	return ComposerUnknown
}

func screenDims(e runtime.Exec) (w, h int) {
	w, h = e.Sess().ScreenSize()
	if w <= 0 {
		w = 720
	}
	if h <= 0 {
		h = 1600
	}
	return w, h
}

type composerActions struct {
	publish      *ui.Resolved
	next         *ui.Resolved
	keyboardDone *ui.Resolved
}

func scanComposerActions(snap ui.Snapshot, w, h int) composerActions {
	return composerActions{
		publish:      findPublishButton(snap, w, h),
		next:         findComposerNextButton(snap, w, h),
		keyboardDone: findKeyboardDoneButton(snap, w, h),
	}
}

// composerNeedsNextBeforePublish avoids tapping publish before Berikutnya on multi-step composers.
func composerNeedsNextBeforePublish(acts composerActions, screenH int) bool {
	if acts.next == nil {
		return false
	}
	// Fanpage composer: header Berikutnya only, no Posting until settings screen.
	if acts.publish == nil {
		return true
	}
	_, nextY := acts.next.Center()
	_, pubY := acts.publish.Center()
	// Bottom Berikutnya before bottom Posting (personal review).
	if nextY >= screenH*55/100 && pubY >= screenH*65/100 {
		return true
	}
	// Header Berikutnya before header Posting (fanpage / direct header flow).
	if nextY <= screenH*25/100 && pubY <= screenH*25/100 {
		return true
	}
	return false
}
