package fanpage

import (
	"fmt"
	"sort"
	"strings"

	"zauto/internal/engine/post"
	"zauto/internal/engine/runtime"
	"zauto/internal/store"
	"zauto/internal/ui"
)

// FanpageContextResult compares visible UI signals against the account fanpage catalog in DB.
type FanpageContextResult struct {
	OnFanpage   bool
	TargetMatch bool
	Matched     *store.Fanpage
	UISignal    string
	SignalBand  string
	MatchKind   string
	Reason      string
	DBCompare   string
}

type fanpageUISignal struct {
	Text string
	Band string
}

// verifyFanpageContext reads the current screen and checks it against DB fanpages for this account.
func verifyFanpageContext(e runtime.Exec, snap ui.Snapshot, target store.Fanpage, screenH int) FanpageContextResult {
	catalog := accountFanpageCatalog(e, target)
	return verifyFanpageContextWithCatalog(e, snap, target, catalog, screenH)
}

func verifyFanpageContextWithCatalog(e runtime.Exec, snap ui.Snapshot, target store.Fanpage, catalog []store.Fanpage, screenH int) FanpageContextResult {
	if screenH <= 0 {
		screenH = post.SnapHeight(snap)
	}
	manageHint := fanpageManageHintInHeader(snap, screenH)
	signals := scanFanpageUISignals(snap, screenH)

	best := FanpageContextResult{Reason: "no_db_match"}
	bestScore := -1

	for _, sig := range signals {
		for i := range catalog {
			fp := &catalog[i]
			kind, nameScore := scoreNameMatch(sig.Text, fp.Name, fp.FBPageID)
			if nameScore <= 0 {
				continue
			}
			total := nameScore + scoreFanpageMatch(sig.Band, kind, manageHint)
			if total <= bestScore || !fanpageContextPass(nameScore, sig.Band, manageHint) {
				continue
			}
			bestScore = total
			best = FanpageContextResult{
				OnFanpage:   true,
				TargetMatch: fp.FBPageID == target.FBPageID,
				Matched:     fp,
				UISignal:    sig.Text,
				SignalBand:  sig.Band,
				MatchKind:   kind,
				Reason:      fanpageReason(sig.Band, kind, manageHint),
			}
		}
	}

	if best.OnFanpage {
		best.DBCompare = formatFanpageDBCompare(best, target, catalog)
		return best
	}

	if manageHint {
		best.Reason = "manage_hint_without_db_name"
	} else if len(signals) == 0 {
		best.Reason = "no_fanpage_ui_signals"
	} else {
		best.Reason = "ui_not_in_db_catalog"
	}
	best.DBCompare = formatFanpageDBCompare(best, target, catalog)
	if e != nil {
		if personal := personalAccountName(e); personal != "" {
			for _, sig := range signals {
				if namesFuzzyMatch(sig.Text, personal) {
					best.Reason = "personal_profile_name"
					best.DBCompare = fmt.Sprintf("ui=%q looks like personal account %q, not in fanpage DB", sig.Text, personal)
					break
				}
			}
		}
	}
	return best
}

func scoreFanpageMatch(band, kind string, manageHint bool) int {
	score := 0
	switch band {
	case "header":
		score += 40
	case "composer":
		score += 35
	case "feed":
		score += 20
	}
	switch kind {
	case "id_exact":
		score += 50
	case "name_exact":
		score += 40
	case "name_fuzzy":
		score += 25
	}
	if manageHint {
		score += 15
	}
	return score
}

func fanpageReason(band, kind string, manageHint bool) string {
	reason := band + "_" + kind
	if manageHint {
		reason += "+manage_page"
	}
	return reason
}

func formatFanpageDBCompare(r FanpageContextResult, target store.Fanpage, catalog []store.Fanpage) string {
	matchedID := ""
	matchedName := ""
	if r.Matched != nil {
		matchedID = r.Matched.FBPageID
		matchedName = r.Matched.Name
	}
	return fmt.Sprintf(
		"ui=%q band=%s db_target=%q target_id=%s matched=%q matched_id=%s kind=%s on_fanpage=%t target_ok=%t catalog=%d",
		r.UISignal, r.SignalBand, target.Name, target.FBPageID,
		matchedName, matchedID, r.MatchKind, r.OnFanpage, r.TargetMatch, len(catalog),
	)
}

func scanFanpageUISignals(snap ui.Snapshot, screenH int) []fanpageUISignal {
	type band struct {
		name string
		minY int
		maxY int
	}
	bands := []band{
		{name: "header", maxY: screenH * 28 / 100},
		{name: "composer", minY: screenH * 18 / 100, maxY: screenH * 50 / 100},
		{name: "feed", minY: screenH * 12 / 100, maxY: screenH * 45 / 100},
	}
	seen := map[string]bool{}
	var out []fanpageUISignal
	for _, b := range bands {
		for _, elem := range snap.Elements {
			if !elem.Enabled {
				continue
			}
			_, cy := elem.Center()
			if cy < b.minY || (b.maxY > 0 && cy > b.maxY) {
				continue
			}
			for _, raw := range []string{elem.Text, elem.ContentDesc} {
				raw = strings.TrimSpace(raw)
				if raw == "" || fanpageNameNoise(raw) {
					continue
				}
				key := b.name + "|" + strings.ToLower(raw)
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, fanpageUISignal{Text: raw, Band: b.name})
			}
		}
	}
	return out
}

func scoreNameMatch(uiText, dbName, dbID string) (kind string, score int) {
	uiText = strings.TrimSpace(uiText)
	if uiText == "" {
		return "", 0
	}
	if fanpageNumericID(uiText) {
		if uiText == dbID {
			return "id_exact", 100
		}
		return "", 0
	}
	if namesEquivalent(uiText, dbName) {
		return "name_exact", 90
	}
	uiTokens := tokenSet(normalizeFanpageName(uiText))
	dbTokens := tokenSet(normalizeFanpageName(dbName))
	if tokenSetsEqual(uiTokens, dbTokens) {
		return "name_fuzzy", 85
	}
	if len(uiTokens) >= 2 && len(dbTokens) >= 2 {
		if tokenSubset(uiTokens, dbTokens) || tokenSubset(dbTokens, uiTokens) {
			return "name_fuzzy", 75
		}
	}
	na, nb := normalizeFanpageName(uiText), normalizeFanpageName(dbName)
	if strings.Contains(na, nb) || strings.Contains(nb, na) {
		return "name_fuzzy", 60
	}
	for t := range uiTokens {
		if len(t) >= 5 && dbTokens[t] {
			return "name_fuzzy", 45
		}
	}
	return "", 0
}

func matchFanpageInCatalog(uiText string, catalog []store.Fanpage) (*store.Fanpage, string) {
	uiText = strings.TrimSpace(uiText)
	if uiText == "" {
		return nil, ""
	}
	var best *store.Fanpage
	bestKind := ""
	bestScore := 0
	for i := range catalog {
		kind, score := scoreNameMatch(uiText, catalog[i].Name, catalog[i].FBPageID)
		if score > bestScore {
			bestScore = score
			best = &catalog[i]
			bestKind = kind
		}
	}
	return best, bestKind
}

func namesEquivalent(a, b string) bool {
	a = normalizeFanpageName(a)
	b = normalizeFanpageName(b)
	return a != "" && a == b
}

func namesFuzzyMatch(a, b string) bool {
	aTokens := tokenSet(normalizeFanpageName(a))
	bTokens := tokenSet(normalizeFanpageName(b))
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return false
	}
	if tokenSetsEqual(aTokens, bTokens) {
		return true
	}
	if len(aTokens) >= 2 && len(bTokens) >= 2 {
		if tokenSubset(aTokens, bTokens) || tokenSubset(bTokens, aTokens) {
			return true
		}
	}
	// Shared distinctive token (e.g. "nurhayati") — caller must disambiguate via full catalog scoring.
	for t := range aTokens {
		if len(t) >= 5 && bTokens[t] {
			return true
		}
	}
	na, nb := normalizeFanpageName(a), normalizeFanpageName(b)
	return strings.Contains(na, nb) || strings.Contains(nb, na)
}

func normalizeFanpageName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		switch r {
		case '.', ',', '!', '?', ':', ';', '-', '_', '(', ')', '[', ']', '"', '\'':
			return ' '
		default:
			return r
		}
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

func tokenSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, t := range strings.Fields(s) {
		if len(t) >= 2 {
			out[t] = true
		}
	}
	return out
}

func tokenSetsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func tokenSubset(sub, sup map[string]bool) bool {
	if len(sub) == 0 {
		return false
	}
	for k := range sub {
		if !sup[k] {
			return false
		}
	}
	return true
}

func personalAccountName(e runtime.Exec) string {
	if e == nil {
		return ""
	}
	if name, _ := e.Sess().Runtime["account_name"].(string); name != "" {
		return name
	}
	return ""
}

func accountFanpageCatalog(e runtime.Exec, target store.Fanpage) []store.Fanpage {
	if e == nil {
		return []store.Fanpage{target}
	}
	catalog, err := AccountFanpages(e.Sess(), e.Ctx())
	if err != nil || len(catalog) == 0 {
		return []store.Fanpage{target}
	}
	return catalog
}

func logFanpageContextCheck(e runtime.Exec, where string, check FanpageContextResult) {
	e.Event("FANPAGE %s %s reason=%s", where, check.DBCompare, check.Reason)
}

// fanpageContextStrict keeps a bool API for legacy callers; requires target DB match.
func fanpageContextStrict(e runtime.Exec, snap ui.Snapshot, target store.Fanpage, screenH int) (bool, string) {
	check := verifyFanpageContext(e, snap, target, screenH)
	return check.OnFanpage && check.TargetMatch, check.Reason
}

func fanpageLabelsFromCatalog(catalog []store.Fanpage, page store.Fanpage) []string {
	var out []string
	seen := map[string]bool{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, fp := range catalog {
		if fp.FBPageID == page.FBPageID {
			add(fp.Name)
			add(fp.FBPageID)
		}
	}
	add(page.Name)
	add(page.FBPageID)
	sort.Strings(out)
	return out
}

func pageDisplayNameFromRuntime(e runtime.Exec) (string, bool) {
	if e == nil {
		return "", false
	}
	dn, _ := e.Sess().Runtime["fanpage_display_name"].(string)
	dn = strings.TrimSpace(dn)
	if dn == "" || fanpageNumericID(dn) || fanpageNameNoise(dn) {
		return "", false
	}
	return dn, true
}

func catalogLabelsForUI(catalog []store.Fanpage) []string {
	var out []string
	seen := map[string]bool{}
	for _, fp := range catalog {
		for _, s := range []string{fp.Name, fp.FBPageID} {
			s = strings.TrimSpace(s)
			if s == "" || seen[s] {
				continue
			}
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// matchFanpageInCatalogList is used by profile switcher list scanning.
func fanpageContextPass(nameScore int, band string, manageHint bool) bool {
	if nameScore >= 75 {
		return true
	}
	if manageHint && nameScore >= 45 && band == "header" {
		return true
	}
	return false
}

func matchFanpageInCatalogList(label string, catalog []store.Fanpage) *store.Fanpage {
	label = strings.TrimSpace(label)
	var best *store.Fanpage
	bestScore := 0
	for i := range catalog {
		_, score := scoreNameMatch(label, catalog[i].Name, catalog[i].FBPageID)
		if score > bestScore {
			bestScore = score
			best = &catalog[i]
		}
	}
	if bestScore < 75 {
		return nil
	}
	return best
}

func uiResolverFindFanpage(resolver *ui.Resolver, snap ui.Snapshot, catalog []store.Fanpage) *ui.Resolved {
	if resolver == nil {
		resolver = ui.NewDefaultResolver()
	}
	for _, fp := range catalog {
		for _, q := range []ui.FindQuery{
			{Texts: []string{fp.Name}, PreferClickable: true, Prefer: "first"},
			{Texts: []string{fp.FBPageID}, PreferClickable: true, Prefer: "first"},
		} {
			if r := resolver.Find(snap, q); r != nil {
				return r
			}
		}
	}
	for _, elem := range snap.Elements {
		if !elem.Enabled || !elem.Clickable {
			continue
		}
		raw := strings.TrimSpace(elem.Label())
		if raw == "" {
			continue
		}
		if matchFanpageInCatalogList(raw, catalog) != nil {
			return &ui.Resolved{Element: elem, Label: raw, Bounds: elem.Bounds}
		}
	}
	return nil
}
