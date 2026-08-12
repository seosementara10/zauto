package post

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

var menuDrawerIndicatorTexts = []string{
	"Settings", "Pengaturan", "Setting & privacy", "Pengaturan & privasi",
	"Help & support", "Bantuan & dukungan", "Switch profile", "Ganti profil",
	"See more", "Lihat lainnya", "Meta Pay", "Privacy", "Privasi",
}

func menuDrawerOpen(e runtime.Exec, snap ui.Snapshot) bool {
	if e.Sess().Resolver.TextExists(snap, menuDrawerIndicatorTexts) {
		return true
	}
	menuRIDs := []string{
		"com.facebook.katana:id/scrollable_menu",
		":id/scrollable_menu",
		"com.facebook.katana:id/bookmarks_list",
	}
	for _, rid := range menuRIDs {
		if e.Sess().Resolver.Find(snap, ui.FindQuery{ResourceIDs: []string{rid}}) != nil {
			return true
		}
	}
	return false
}

func openMenuDrawer(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	menuQueries := []ui.FindQuery{
		{ContentDescs: []string{"Menu, Tab"}, PreferClickable: true, Prefer: "bottom"},
		{Texts: []string{"Menu", "Main menu", "Navigate to home screen menu"}, PreferClickable: true},
		{ResourceIDs: []string{"com.facebook.katana:id/assistive_navigation_icon"}, PreferClickable: true},
		{ContentDescs: []string{"Menu", "Main menu"}, PreferClickable: true, Prefer: "bottom"},
	}
	if err := e.PollTapObserve(observe, invalidate, menuQueries, timeoutSec); err != nil {
		return fmt.Errorf("menu button: %w", err)
	}

	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if menuDrawerOpen(e, snap) {
			e.Event("VERIFY menu drawer open")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("menu drawer did not open")
}
