package logout

import (
	"fmt"
	"time"

	"zauto/internal/config"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

var (
	menuDrawerIndicatorTexts = []string{
		"Settings", "Pengaturan", "Setting & privacy", "Pengaturan & privasi",
		"Help & support", "Bantuan & dukungan", "Bantuan", "Help",
		"Privacy", "Privasi", "See more", "Lihat lainnya",
		"Meta Pay", "Orders", "Pesanan", "Ad Center", "Pusat Iklan",
	}
	menuItemTexts = append([]string(nil), state.LogoutMenuItemTexts...)
)

func menuDrawerOpen(e runtime.Exec, snap ui.Snapshot) bool {
	if e.Sess().Resolver.TextExists(snap, menuItemTexts) {
		return true
	}
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

func openMenu(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	menuQueries := []ui.FindQuery{
		{Texts: []string{"Menu", "Main menu", "Navigate to home screen menu"}, PreferClickable: true},
		{ResourceIDs: []string{"com.facebook.katana:id/assistive_navigation_icon"}, PreferClickable: true},
		{ContentDescs: []string{"Menu", "Main menu"}, PreferClickable: true},
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
	snap := e.ReadSnap(observe)
	e.CaptureRecoveryArtifacts("menu_not_open", snap)
	return fmt.Errorf("menu drawer did not open")
}

func scrollMenuDrawer(e runtime.Exec, observe state.ObserveFn, invalidate func(), maxSwipes int) bool {
	w, h := e.Sess().ScreenSize()
	x := w / 4
	c := config.ScrollCoords{X1: x, Y1: h * 65 / 100, X2: x, Y2: h * 30 / 100, DurationMs: 280}

	for i := 0; i < maxSwipes; i++ {
		snap := e.ReadSnap(observe)
		if !menuDrawerOpen(e, snap) {
			e.Event("MENU scroll aborted: drawer closed (avoid feed scroll)")
			return false
		}
		if e.Sess().Resolver.TextExists(snap, menuItemTexts) {
			return true
		}
		_ = e.Sess().Client.Swipe(c.X1, c.Y1, c.X2, c.Y2, c.DurationMs)
		e.InvalidateObserve(invalidate)
		time.Sleep(e.Sess().PollInterval())
		snap = e.ReadSnap(observe)
		if e.Sess().Resolver.TextExists(snap, menuItemTexts) {
			return true
		}
	}
	snap := e.ReadSnap(observe)
	return e.Sess().Resolver.TextExists(snap, menuItemTexts)
}

func tapMenuItem(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	q := ui.FindQuery{Texts: menuItemTexts, PreferClickable: true, Prefer: "bottom"}
	return e.PollTapObserve(observe, invalidate, []ui.FindQuery{q}, timeoutSec)
}
