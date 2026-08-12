package post

import (
	"strings"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// composerScreenOpen reports the full-screen create-post UI only.
// Do not treat the feed composer pill (Apa yang Anda pikirkan?) as an open composer screen.
func composerScreenOpen(e runtime.Exec, snap ui.Snapshot) bool {
	return composerScreenOpenSnap(e.Sess().Resolver, snap)
}

func composerScreenOpenSnap(_ *ui.Resolver, snap ui.Snapshot) bool {
	// Title must sit in the header band — avoid false positives from feed XML / stale nodes.
	return composerTitleInHeader(snap, state.FeedComposerScreenTexts)
}

func composerTitleInHeader(snap ui.Snapshot, titles []string) bool {
	want := map[string]struct{}{}
	for _, title := range titles {
		if n := ui.Normalize(title); n != "" {
			want[n] = struct{}{}
		}
	}
	if len(want) == 0 {
		return false
	}
	maxY := snapHeight(snap) * 25 / 100
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy > maxY {
			continue
		}
		for _, raw := range []string{elem.Text, elem.ContentDesc} {
			if raw == "" {
				continue
			}
			if _, ok := want[ui.Normalize(raw)]; ok {
				return true
			}
		}
	}
	return false
}

func snapHeight(snap ui.Snapshot) int {
	h := 0
	for _, elem := range snap.Elements {
		if elem.Bounds[3] > h {
			h = elem.Bounds[3]
		}
	}
	if h < 1000 {
		return 1600
	}
	return h
}

func galleryPickerOpen(e runtime.Exec, snap ui.Snapshot) bool {
	if e.Sess().Resolver.TextExists(snap, state.GalleryRecentTexts) {
		return true
	}
	if e.Sess().Resolver.TextExists(snap, state.GalleryImagePickerTexts) {
		return true
	}
	_, h := e.Sess().ScreenSize()
	minY := h * 15 / 100
	count := 0
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		cn := elem.ClassName
		if !strings.Contains(cn, "ImageView") && !strings.Contains(cn, "ImageButton") {
			continue
		}
		_, cy := elem.Center()
		if cy < minY {
			continue
		}
		if elem.Width() >= 80 && elem.Height() >= 80 {
			count++
		}
	}
	return count >= 4
}

func profileSwitcherOpen(e runtime.Exec, snap ui.Snapshot) bool {
	if e.Sess().Resolver.TextExists(snap, state.SeeAllProfilesTexts) {
		return true
	}
	if e.Sess().Resolver.TextExists(snap, state.SwitchProfileTexts) {
		return true
	}
	if e.Sess().Resolver.TextExists(snap, state.SeeAllProfilesTexts) {
		return true
	}
	return false
}

func composerTextContains(e runtime.Exec, snap ui.Snapshot, text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return true
	}
	if edit := ui.ComposerEditField(snap); edit != nil && ui.FieldTextContains(*edit, text) {
		return true
	}
	return e.Sess().Resolver.TextExists(snap, []string{text})
}

func imageAttachedInComposer(e runtime.Exec, snap ui.Snapshot) bool {
	removeTexts := []string{
		"Remove photo", "Remove image", "Hapus foto", "Hapus gambar",
		"Remove", "Hapus",
	}
	if e.Sess().Resolver.TextExists(snap, removeTexts) {
		return true
	}
	_, h := e.Sess().ScreenSize()
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		cn := elem.ClassName
		if !strings.Contains(cn, "ImageView") {
			continue
		}
		_, cy := elem.Center()
		if cy > h*75/100 {
			continue
		}
		if elem.Width() >= 100 && elem.Height() >= 100 {
			return true
		}
	}
	return false
}

func postSettingsScreenOpen(snap ui.Snapshot) bool {
	return composerTitleInHeader(snap, state.PostSettingsScreenTexts)
}
