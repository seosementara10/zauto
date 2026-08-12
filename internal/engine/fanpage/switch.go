package fanpage

import (
	"fmt"
	"strings"
	"time"

	"zauto/internal/engine/overlay"
	"zauto/internal/engine/post"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/store"
	"zauto/internal/ui"
)

func switchToFanpage(e runtime.Exec, observe state.ObserveFn, invalidate func(), page store.Fanpage, timeoutSec float64) error {
	e.Event("ACT switch to fanpage fb_id=%s name=%q", page.FBPageID, page.Name)
	setActiveFanpage(e, page)
	logScreenContext(e, observe, "switch_fanpage:start", &page)

	snap := e.ReadSnap(observe)
	_, h := post.ScreenDims(e)
	check := verifyFanpageContext(e, snap, page, h)
	if check.OnFanpage && check.TargetMatch {
		rememberFanpageDisplayName(e, snap, page)
		logScreenContext(e, observe, "switch_fanpage:already", &page)
		logFanpageContextCheck(e, "switch_already", check)
		e.Event("VERIFY already on fanpage %s (%s)", page.FBPageID, check.Reason)
		return nil
	}
	if check.OnFanpage && check.Matched != nil && !check.TargetMatch {
		logFanpageContextCheck(e, "switch_wrong_fanpage", check)
		e.Event("SCREEN on fanpage %s (%q) but target is %s (%q) — switching",
			check.Matched.FBPageID, check.Matched.Name, page.FBPageID, page.Name)
	} else {
		logFanpageContextCheck(e, "switch_needed", check)
		e.Event("SCREEN not on target fanpage — opening profile switcher")
	}
	return performFanpageSwitch(e, observe, invalidate, page, timeoutSec)
}

func performFanpageSwitch(e runtime.Exec, observe state.ObserveFn, invalidate func(), page store.Fanpage, timeoutSec float64) error {
	det := state.NewDetector()
	if err := post.EnsureOnFeed(e, observe, det, timeoutSec); err != nil {
		return err
	}

	if err := post.OpenMenuDrawer(e, observe, invalidate, timeoutSec); err != nil {
		return err
	}

	if err := tapFanpageInMenuDrawer(e, observe, invalidate, page, timeoutSec); err != nil {
		e.Event("ACT fanpage not in menu — opening profile switcher (%v)", err)
		if err := openProfileSwitcher(e, observe, invalidate, timeoutSec); err != nil {
			capturePostFailure(e, observe, "fanpage_switcher", &page, err.Error())
			return err
		}
		if err := tapFanpageInList(e, observe, invalidate, page, timeoutSec); err != nil {
			return err
		}
	} else {
		confirmFanpageSwitchSheet(e, observe, invalidate, page, timeoutSec)
	}
	e.InvalidateObserve(invalidate)
	if err := overlay.DismissFanpageHomeIntroIfPresent(e, observe, 15*time.Second); err != nil {
		e.Event("ACT fanpage intro dismiss: %v", err)
	}

	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		_ = overlay.DismissFanpageHomeIntroIfPresent(e, observe, 2*time.Second)
		snap := e.ReadSnap(observe)
		_, h := post.ScreenDims(e)
		check := verifyFanpageContext(e, snap, page, h)
		if check.OnFanpage && check.TargetMatch && !fanpageSwitchSheetVisible(snap) {
			rememberFanpageDisplayName(e, snap, page)
			logScreenContext(e, observe, "switch_fanpage:done", &page)
			logFanpageContextCheck(e, "switch_done", check)
			e.Event("VERIFY switched to fanpage %s (%s)", page.FBPageID, check.Reason)
			return nil
		}
		logScreenContextIfStale(e, observe, "switch_fanpage:waiting", &page, 5*time.Second)
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("fanpage context not confirmed after switch: %s", page.FBPageID)
}

func ensureFanpageBeforeCompose(e runtime.Exec, observe state.ObserveFn, invalidate func(), page store.Fanpage, switchTimeout, composerTimeout float64) error {
	logScreenContext(e, observe, "ensure_fanpage:start", &page)
	if err := switchToFanpage(e, observe, invalidate, page, switchTimeout); err != nil {
		capturePostFailure(e, observe, "fanpage_switch", &page, err.Error())
		return err
	}
	if err := post.OpenComposer(e, observe, invalidate, composerTimeout); err != nil {
		capturePostFailure(e, observe, "fanpage_open_composer", &page, err.Error())
		return err
	}
	snap := e.ReadSnap(observe)
	_, h := post.ScreenDims(e)
	check := verifyFanpageContext(e, snap, page, h)
	if check.TargetMatch && (check.SignalBand == "composer" || composerAuthorMatchesDB(e, snap, page, h)) {
		logScreenContext(e, observe, "ensure_fanpage:composer_ok", &page)
		logFanpageContextCheck(e, "composer_ok", check)
		e.Event("VERIFY composer author fanpage %s", page.FBPageID)
		return nil
	}

	logScreenContext(e, observe, "ensure_fanpage:composer_wrong_author", &page)
	e.Event("ACT composer author mismatch — close composer then switch fanpage")
	if err := post.CloseComposerIfOpen(e, observe, invalidate, composerTimeout); err != nil {
		capturePostFailure(e, observe, "fanpage_close_composer", &page, err.Error())
		return err
	}
	if err := performFanpageSwitch(e, observe, invalidate, page, switchTimeout); err != nil {
		capturePostFailure(e, observe, "fanpage_force_switch", &page, err.Error())
		return err
	}
	if err := post.OpenComposer(e, observe, invalidate, composerTimeout); err != nil {
		capturePostFailure(e, observe, "fanpage_reopen_composer", &page, err.Error())
		return err
	}
	snap = e.ReadSnap(observe)
	check = verifyFanpageContext(e, snap, page, h)
	if !check.TargetMatch {
		logScreenContext(e, observe, "ensure_fanpage:still_personal", &page)
		logFanpageContextCheck(e, "composer_still_wrong", check)
		capturePostFailure(e, observe, "fanpage_composer_author", &page, "still personal profile")
		return fmt.Errorf("composer still on personal profile, not fanpage %s", page.FBPageID)
	}
	logScreenContext(e, observe, "ensure_fanpage:composer_ok_after_switch", &page)
	e.Event("VERIFY composer author fanpage %s after switch", page.FBPageID)
	return nil
}

func openProfileSwitcher(e runtime.Exec, observe state.ObserveFn, invalidate func(), timeoutSec float64) error {
	switchQueries := []ui.FindQuery{
		{Texts: state.SwitchProfileTexts, PreferClickable: true, Prefer: "top"},
		{ContentDescs: state.SwitchProfileTexts, PreferClickable: true, Prefer: "top"},
		{Texts: state.SeeAllProfilesTexts, PreferClickable: true, Prefer: "first"},
	}
	if err := e.PollTapObserve(observe, invalidate, switchQueries, timeoutSec); err == nil {
		if err := waitProfileSwitcher(e, observe, 4); err == nil {
			e.Event("ACT opened profile switcher")
			return nil
		}
		e.Event("ACT switch profile tap did not open switcher — trying profile row")
	}

	seeAll := []ui.FindQuery{
		{Texts: state.SeeAllProfilesTexts, PreferClickable: true, Prefer: "first"},
		{ContentDescs: state.SeeAllProfilesTexts, PreferClickable: true, Prefer: "first"},
	}
	if err := e.PollTapObserve(observe, invalidate, seeAll, 5); err == nil {
		if err := waitProfileSwitcher(e, observe, timeoutSec); err == nil {
			e.Event("ACT opened profile switcher via see-all")
			return nil
		}
	}

	// Fallback: tap profile name row at top of menu drawer.
	snap := e.ReadSnap(observe)
	_, h := e.Sess().ScreenSize()
	maxY := h * 35 / 100
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		_, cy := elem.Center()
		if cy > maxY {
			continue
		}
		label := strings.ToLower(elem.Label())
		if strings.Contains(label, "profile") || strings.Contains(label, "profil") {
			x, y := elem.Center()
			if err := e.Sess().Client.Tap(x, y); err != nil {
				return err
			}
			e.Event("ACT tapped profile row fallback")
			return waitProfileSwitcher(e, observe, timeoutSec)
		}
	}
	return fmt.Errorf("profile switcher not found in menu")
}

func waitProfileSwitcher(e runtime.Exec, observe state.ObserveFn, timeoutSec float64) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if post.ProfileSwitcherOpen(e, snap) {
			e.Event("VERIFY profile switcher open")
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("profile switcher not visible")
}

func tapFanpageInMenuDrawer(e runtime.Exec, observe state.ObserveFn, invalidate func(), page store.Fanpage, timeoutSec float64) error {
	catalog := accountFanpageCatalog(e, page)
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if hit := findFanpageMenuEntry(e.Sess().Resolver, snap, catalog, page); hit != nil {
			x, y := hit.Center()
			e.Event("ACT tap fanpage in menu %q at (%d,%d)", hit.Label, x, y)
			if err := e.Sess().Client.Tap(x, y); err != nil {
				return err
			}
			e.InvalidateObserve(invalidate)
			if matched := matchFanpageInCatalogList(hit.Label, catalog); matched != nil && !fanpageNumericID(hit.Label) {
				e.Sess().Runtime["fanpage_display_name"] = strings.TrimSpace(hit.Label)
			}
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("fanpage %q not in menu drawer", pageDisplayHint(page))
}

func findFanpageMenuEntry(resolver *ui.Resolver, snap ui.Snapshot, catalog []store.Fanpage, target store.Fanpage) *ui.Resolved {
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		raw := strings.TrimSpace(elem.ContentDesc)
		if raw == "" {
			raw = strings.TrimSpace(elem.Text)
		}
		if raw == "" {
			continue
		}
		lower := strings.ToLower(raw)
		if strings.Contains(lower, "beralih ke halaman") || strings.Contains(lower, "switch to your page") {
			if matched := matchFanpageInCatalogList(raw, catalog); matched != nil && matched.FBPageID == target.FBPageID {
				return &ui.Resolved{Element: elem, Label: raw, Bounds: elem.Bounds}
			}
		}
	}
	return findFanpageListEntry(resolver, snap, catalog, target)
}

func confirmFanpageSwitchSheet(e runtime.Exec, observe state.ObserveFn, invalidate func(), page store.Fanpage, timeoutSec float64) {
	waitSec := timeoutSec
	if waitSec > 12 {
		waitSec = 12
	}
	deadline := time.Now().Add(time.Duration(waitSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if !fanpageSwitchSheetVisible(snap) {
			e.Event("VERIFY fanpage switch sheet dismissed")
			return
		}
		time.Sleep(e.Sess().PollInterval())
	}
	snap := e.ReadSnap(observe)
	if !fanpageSwitchSheetVisible(snap) {
		return
	}
	e.Event("ACT fanpage switch sheet still open — tap Lihat Semua Opsi")
	seeAll := []ui.FindQuery{
		{Texts: []string{"Lihat Semua Opsi", "See all options", "See All Options"}, PreferClickable: true, Prefer: "first"},
		{ContentDescs: []string{"Lihat Semua Opsi", "See all options", "See All Options"}, PreferClickable: true, Prefer: "first"},
	}
	if err := e.PollTapObserve(observe, invalidate, seeAll, 5); err != nil {
		return
	}
	catalog := accountFanpageCatalog(e, page)
	deadline = time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		snap = e.ReadSnap(observe)
		if hit := findFanpageListEntry(e.Sess().Resolver, snap, catalog, page); hit != nil {
			x, y := hit.Center()
			e.Event("ACT tap fanpage in see-all list %q at (%d,%d)", hit.Label, x, y)
			_ = e.Sess().Client.Tap(x, y)
			e.InvalidateObserve(invalidate)
			return
		}
		time.Sleep(e.Sess().PollInterval())
	}
}

func fanpageSwitchSheetVisible(snap ui.Snapshot) bool {
	lower := strings.ToLower(snap.XML)
	if strings.Contains(lower, "beralih ke") || strings.Contains(lower, "switch to") {
		return true
	}
	return strings.Contains(lower, "lihat semua opsi") || strings.Contains(lower, "see all options")
}

func tapFanpageSwitchSheetRow(resolver *ui.Resolver, snap ui.Snapshot, catalog []store.Fanpage, target store.Fanpage) *ui.Resolved {
	if hit := findFanpageListEntry(resolver, snap, catalog, target); hit != nil {
		return hit
	}
	labels := fanpageLabelsFromCatalog(catalog, target)
	var best *ui.Resolved
	bestY := -1
	for _, label := range labels {
		if label == "" {
			continue
		}
		for _, elem := range snap.Elements {
			if !elem.Enabled {
				continue
			}
			raw := strings.TrimSpace(elem.Text)
			if raw == "" {
				raw = strings.TrimSpace(elem.ContentDesc)
			}
			if raw == "" || fanpageSwitchSheetTitle(raw) {
				continue
			}
			if raw != label && !strings.Contains(raw, label) {
				continue
			}
			matched := matchFanpageInCatalogList(raw, catalog)
			if matched == nil || matched.FBPageID != target.FBPageID {
				continue
			}
			_, cy := elem.Center()
			r := &ui.Resolved{Element: elem, Label: raw, Bounds: elem.Bounds}
			if best == nil || cy > bestY {
				best, bestY = r, cy
			}
		}
	}
	return best
}

func fanpageSwitchSheetTitle(label string) bool {
	lower := strings.ToLower(strings.TrimSpace(label))
	return strings.HasPrefix(lower, "beralih ke") || strings.HasPrefix(lower, "switch to")
}

func tapFanpageInList(e runtime.Exec, observe state.ObserveFn, invalidate func(), page store.Fanpage, timeoutSec float64) error {
	catalog := accountFanpageCatalog(e, page)
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	swipes := 0
	for time.Now().Before(deadline) {
		snap := e.ReadSnap(observe)
		if hit := findFanpageListEntry(e.Sess().Resolver, snap, catalog, page); hit != nil {
			x, y := hit.Center()
			e.Event("ACT tap fanpage %q at (%d,%d)", hit.Label, x, y)
			if err := e.Sess().Client.Tap(x, y); err != nil {
				return err
			}
			e.InvalidateObserve(invalidate)
			if matched := matchFanpageInCatalogList(hit.Label, catalog); matched != nil && !fanpageNumericID(hit.Label) {
				e.Sess().Runtime["fanpage_display_name"] = strings.TrimSpace(hit.Label)
			}
			return nil
		}
		if swipes < 4 {
			scrollProfileSwitcher(e)
			swipes++
			time.Sleep(e.Sess().PollInterval())
			continue
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("fanpage %q not found in profile list (UI name may differ from DB — set display name e.g. Nurhayati Fans)", pageDisplayHint(page))
}

func findFanpageListEntry(resolver *ui.Resolver, snap ui.Snapshot, catalog []store.Fanpage, target store.Fanpage) *ui.Resolved {
	if resolver == nil {
		resolver = ui.NewDefaultResolver()
	}
	labels := fanpageLabelsFromCatalog(catalog, target)
	for _, label := range labels {
		if label == "" {
			continue
		}
		q := ui.FindQuery{Texts: []string{label}, PreferClickable: true, Prefer: "first"}
		if r := resolver.Find(snap, q); r != nil {
			if matchFanpageInCatalogList(r.Label, catalog) != nil || r.Label == target.FBPageID {
				return r
			}
		}
		q.ContentDescs = []string{label}
		if r := resolver.Find(snap, q); r != nil {
			if matchFanpageInCatalogList(r.Label, catalog) != nil || r.Label == target.FBPageID {
				return r
			}
		}
	}
	for _, label := range labels {
		if label == "" || fanpageNumericID(label) {
			continue
		}
		if hit := findClickableLabelContains(snap, label); hit != nil {
			if matchFanpageInCatalogList(hit.Label, catalog) != nil && rawMatchesTargetPage(hit.Label, catalog, target) {
				return hit
			}
		}
	}
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable || elem.Width() <= 0 {
			continue
		}
		raw := strings.TrimSpace(elem.Label())
		if raw == "" || profileSwitcherNoise(raw) {
			continue
		}
		if matched := matchFanpageInCatalogList(raw, catalog); matched != nil && matched.FBPageID == target.FBPageID {
			return &ui.Resolved{Element: elem, Label: raw, Bounds: elem.Bounds}
		}
	}
	return findFanpageRowHeuristicForTarget(snap, catalog, target)
}

func findFanpageRowHeuristicForTarget(snap ui.Snapshot, catalog []store.Fanpage, target store.Fanpage) *ui.Resolved {
	if hit := findFanpageRowHeuristic(snap); hit != nil {
		if matched := matchFanpageInCatalogList(hit.Label, catalog); matched != nil && matched.FBPageID == target.FBPageID {
			return hit
		}
	}
	return nil
}

func rawMatchesTargetPage(raw string, catalog []store.Fanpage, target store.Fanpage) bool {
	_, score := scoreNameMatch(raw, target.Name, target.FBPageID)
	if score < 75 {
		return false
	}
	fp, _ := matchFanpageInCatalog(raw, catalog)
	return fp != nil && fp.FBPageID == target.FBPageID
}

func findClickableLabelContains(snap ui.Snapshot, label string) *ui.Resolved {
	want := strings.ToLower(strings.TrimSpace(label))
	if want == "" {
		return nil
	}
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		raw := strings.ToLower(strings.TrimSpace(elem.Label()))
		if raw != "" && strings.Contains(raw, want) {
			return &ui.Resolved{Element: elem, Label: elem.Label(), Bounds: elem.Bounds}
		}
	}
	return nil
}

func findFanpageRowHeuristic(snap ui.Snapshot) *ui.Resolved {
	var best *ui.Resolved
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		raw := strings.TrimSpace(elem.Label())
		if raw == "" || profileSwitcherNoise(raw) {
			continue
		}
		lower := strings.ToLower(raw)
		if !strings.Contains(lower, "fans") &&
			!strings.Contains(lower, "fanpage") &&
			!strings.Contains(lower, " halaman") &&
			!strings.HasSuffix(lower, " page") {
			continue
		}
		r := &ui.Resolved{Element: elem, Label: raw, Bounds: elem.Bounds}
		if best == nil {
			best = r
			continue
		}
		_, cy := elem.Center()
		_, bestCY := best.Center()
		if cy < bestCY {
			best = r
		}
	}
	return best
}

func profileSwitcherNoise(label string) bool {
	lower := strings.ToLower(strings.TrimSpace(label))
	switch lower {
	case "menu", "settings", "pengaturan", "help", "bantuan", "see all profiles", "lihat semua profil":
		return true
	}
	for _, t := range state.SwitchProfileTexts {
		if strings.EqualFold(label, t) {
			return true
		}
	}
	return false
}

func scrollProfileSwitcher(e runtime.Exec) {
	w, h := e.Sess().ScreenSize()
	if w <= 0 {
		w = 720
	}
	if h <= 0 {
		h = 1600
	}
	_ = e.Sess().Client.Swipe(w/2, h*65/100, w/2, h*35/100, 280)
}

func fanpageLabels(e runtime.Exec, page store.Fanpage) []string {
	catalog := accountFanpageCatalog(e, page)
	labels := fanpageLabelsFromCatalog(catalog, page)
	if e != nil {
		if dn, ok := pageDisplayNameFromRuntime(e); ok {
			labels = append([]string{dn}, labels...)
		}
	}
	return labels
}

func pageDisplayHint(page store.Fanpage) string {
	if name := strings.TrimSpace(page.Name); name != "" && !fanpageNumericID(name) {
		return name
	}
	return page.FBPageID
}

func rememberFanpageDisplayName(e runtime.Exec, snap ui.Snapshot, page store.Fanpage) {
	if dn, _ := e.Sess().Runtime["fanpage_display_name"].(string); dn != "" && !fanpageNumericID(dn) && !fanpageNameNoise(dn) {
		return
	}
	_, h := post.ScreenDims(e)
	if name := fanpageNameFromComposerArea(snap, h); name != "" && !fanpageNameNoise(name) {
		e.Sess().Runtime["fanpage_display_name"] = name
		e.Event("ACT fanpage display name=%q", name)
	}
}

func fanpageNameNoise(label string) bool {
	lower := strings.ToLower(strings.TrimSpace(label))
	switch lower {
	case "buka profil", "open profile", "lihat profil", "view profile",
		"menu", "facebook", "beranda", "home", "news feed",
		"insight baru", "new insight", "buat cerita", "create story":
		return true
	}
	if strings.HasPrefix(lower, "beralih ke") || strings.HasPrefix(lower, "switch to") {
		return true
	}
	if strings.Contains(lower, "apa yang anda pikirkan") ||
		strings.Contains(lower, "what's on your mind") {
		return true
	}
	return false
}

func fanpageNameFromComposerArea(snap ui.Snapshot, screenH int) string {
	if screenH <= 0 {
		screenH = post.SnapHeight(snap)
	}
	maxY := screenH * 45 / 100
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy > maxY {
			continue
		}
		raw := strings.TrimSpace(elem.Text)
		if raw == "" {
			raw = strings.TrimSpace(elem.ContentDesc)
		}
		if raw == "" || len(raw) < 3 || fanpageNameNoise(raw) {
			continue
		}
		if fanpageNumericID(raw) {
			continue
		}
		return raw
	}
	return ""
}

func fanpageContextVisible(e runtime.Exec, page store.Fanpage) (bool, ui.Snapshot) {
	snap := e.Sess().ReadUI(true)
	_, h := post.ScreenDims(e)
	check := verifyFanpageContext(e, snap, page, h)
	return check.OnFanpage && check.TargetMatch, snap
}

func composerAuthorMatchesDB(e runtime.Exec, snap ui.Snapshot, target store.Fanpage, screenH int) bool {
	check := verifyFanpageContext(e, snap, target, screenH)
	return check.TargetMatch && check.SignalBand == "composer"
}

func fanpageManageHintInHeader(snap ui.Snapshot, screenH int) bool {
	maxY := screenH * 35 / 100
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy > maxY {
			continue
		}
		for _, raw := range []string{elem.Text, elem.ContentDesc} {
			for _, hint := range state.FanpageFeedHints {
				if hint == "" {
					continue
				}
				if strings.Contains(raw, hint) {
					return true
				}
			}
		}
	}
	return false
}

func composerAuthorMatches(snap ui.Snapshot, names []string, screenH int) bool {
	if screenH <= 0 {
		screenH = post.SnapHeight(snap)
	}
	minY := screenH * 18 / 100
	maxY := screenH * 50 / 100
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy < minY || cy > maxY {
			continue
		}
		for _, name := range names {
			if name == "" {
				continue
			}
			if strings.Contains(elem.Text, name) || strings.Contains(elem.ContentDesc, name) {
				return true
			}
		}
	}
	return false
}

func fanpageNumericID(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func verifyFanpagePosted(e runtime.Exec, content post.Content, page store.Fanpage, timeoutSec float64) error {
	observe, _ := e.CachedObserve()
	_, h := post.ScreenDims(e)
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		_ = overlay.DismissPostPromotePromptIfPresent(e, observe, 2*time.Second)
		snap := e.ReadSnap(observe)
		composerOpen := post.ComposerScreenOpen(e, snap) || post.PostSettingsScreenOpen(snap)

		if composerOpen {
			if post.FeedPostPublishing(snap, h) {
				logScreenContext(e, observe, "verify_fanpage:uploading", &page)
				e.Event("VERIFY fanpage post uploading")
				return nil
			}
			logScreenContextIfStale(e, observe, "verify_fanpage:composer_still_open", &page, 5*time.Second)
			time.Sleep(e.Sess().PollInterval())
			continue
		}

		if content.Text != "" && post.PostTextVisibleOnFeed(e, snap, content.Text) {
			logScreenContext(e, observe, "verify_fanpage:text_visible", &page)
			e.Event("VERIFY fanpage post text visible")
			return nil
		}
		if post.FeedFreshPostVisible(snap, h) || post.FeedPostPublishing(snap, h) {
			logScreenContext(e, observe, "verify_fanpage:fresh_post", &page)
			e.Event("VERIFY fanpage fresh post on feed")
			return nil
		}
		if check := verifyFanpageContext(e, snap, page, h); check.OnFanpage && check.TargetMatch {
			logScreenContext(e, observe, "verify_fanpage:feed_ok", &page)
			logFanpageContextCheck(e, "verify_ok", check)
			e.Event("VERIFY fanpage feed after publish (%s)", check.Reason)
			return nil
		}
		logScreenContextIfStale(e, observe, "verify_fanpage:waiting", &page, 5*time.Second)
		time.Sleep(e.Sess().PollInterval())
	}
	capturePostFailure(e, observe, "fanpage_verify_timeout", &page, page.FBPageID)
	return fmt.Errorf("fanpage post not verified for %s", page.FBPageID)
}

func fanpagePostAlreadyVisible(e runtime.Exec, observe state.ObserveFn, page store.Fanpage, content post.Content) bool {
	snap := e.ReadSnap(observe)
	h := screenDimsHeight(e)
	if content.Text != "" && post.PostTextVisibleOnFeed(e, snap, content.Text) {
		check := verifyFanpageContext(e, snap, page, h)
		if check.OnFanpage && check.TargetMatch {
			logFanpageContextCheck(e, "skip_already_visible", check)
			return true
		}
	}
	return isFanpagePublished(e, page.FBPageID)
}

func screenDimsHeight(e runtime.Exec) int {
	_, h := post.ScreenDims(e)
	return h
}
