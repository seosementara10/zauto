package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// ScreenNote is an alias for runtime.ScreenNote (used by post/login flows).
type ScreenNote = runtime.ScreenNote

var screenLogMu sync.Mutex
var lastScreenLogBySerial = map[string]time.Time{}

// LogScreen emits one SCREEN line using the shared detector + UI hierarchy (same source as RECOVERY).
func (e *Executor) LogScreen(observe state.ObserveFn, where string, note ScreenNote) {
	snap, pkg, act := observe()
	d := state.NewDetector().Detect(snap, pkg, act)
	screen, composer, settings, promote := classifyScreenKind(snap, d)
	profile := note.Profile
	if profile == "" {
		profile = "unknown"
		if screen == "personal_feed" {
			profile = "personal"
		}
	}
	top := summarizeTopLabels(snap, screenHeight(e))
	e.event("SCREEN %s ui=%s screen=%s profile=%s composer=%t settings=%t promote=%t detail=%q top=%q pkg=%s",
		where, d.State, screen, profile, composer, settings, promote, note.Detail, top, pkg)
	_ = act
}

// LogScreenIfStale throttles repetitive SCREEN logs during wait loops.
func (e *Executor) LogScreenIfStale(observe state.ObserveFn, where string, note ScreenNote, minInterval time.Duration) {
	if minInterval <= 0 {
		minInterval = 5 * time.Second
	}
	screenLogMu.Lock()
	last := lastScreenLogBySerial[e.Session.Serial]
	if time.Since(last) < minInterval {
		screenLogMu.Unlock()
		return
	}
	lastScreenLogBySerial[e.Session.Serial] = time.Now()
	screenLogMu.Unlock()
	e.LogScreen(observe, where, note)
}

// CaptureFailure logs screen context then saves screenshot+hierarchy under captures/<serial>/.
// This is the single entry point — prefer this over ad-hoc Client.Screenshot to screenshots/.
func (e *Executor) CaptureFailure(label, where, detail string, observe state.ObserveFn, note ScreenNote) (screenshotPath, dumpPath string) {
	ctx := where
	if detail != "" {
		ctx = fmt.Sprintf("%s:%s", where, detail)
	}
	e.event("CAPTURE begin label=%s context=%s", label, ctx)
	e.LogScreen(observe, ctx, note)
	snap := e.readSnap(observe)
	if snap.XML == "" {
		snap = e.Session.ReadUI(true)
	}
	screenshotPath, dumpPath = e.captureRecoveryArtifacts(label, snap)
	if screenshotPath == "" && dumpPath == "" {
		e.event("CAPTURE failed label=%s context=%s (no artifacts written)", label, ctx)
	} else {
		e.event("CAPTURE done label=%s screenshot=%s hierarchy=%s", label, screenshotPath, dumpPath)
	}
	return screenshotPath, dumpPath
}

func classifyScreenKind(snap ui.Snapshot, d state.Detection) (screen string, composer, settings, promote bool) {
	promote = postPromoteSheetOpen(snap)
	settings = composerTitleInBand(snap, state.PostSettingsScreenTexts)
	composer = composerTitleInBand(snap, state.FeedComposerScreenTexts)
	switch {
	case promote:
		screen = "post_promote_sheet"
	case settings:
		screen = "post_settings"
	case composer:
		screen = "composer"
	case d.State == state.UISaveLoginPrompt,
		d.State == state.UIPostPromotePrompt,
		d.State == state.UIFanpageHomeIntro,
		d.State == state.UIPermission:
		screen = string(d.State)
	case feedHintsVisible(snap):
		screen = "personal_feed"
	case menuDrawerLikely(snap):
		screen = "menu_drawer"
	case d.State != state.UIUnknown && d.State != state.UILoading:
		screen = string(d.State)
	default:
		screen = "unknown"
	}
	return screen, composer, settings, promote
}

func composerTitleInBand(snap ui.Snapshot, titles []string) bool {
	want := map[string]struct{}{}
	for _, t := range titles {
		if n := ui.Normalize(t); n != "" {
			want[n] = struct{}{}
		}
	}
	maxY := snapMaxY(snap) * 25 / 100
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

func postPromoteSheetOpen(snap ui.Snapshot) bool {
	r := ui.NewDefaultResolver()
	return r.TextExists(snap, state.PostPromoteTitleTexts) &&
		r.TextExists(snap, state.PostPromoteLaterTexts)
}

func feedHintsVisible(snap ui.Snapshot) bool {
	return ui.NewDefaultResolver().TextExists(snap, state.LoggedInFeedHints)
}

func menuDrawerLikely(snap ui.Snapshot) bool {
	r := ui.NewDefaultResolver()
	drawerTexts := []string{
		"Settings", "Pengaturan", "Ganti profil", "Switch profile",
		"See all profiles", "Lihat semua profil", "Help & support", "Bantuan & dukungan",
	}
	return r.TextExists(snap, drawerTexts)
}

func snapMaxY(snap ui.Snapshot) int {
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

func screenHeight(e *Executor) int {
	_, h := e.Session.ScreenSize()
	if h <= 0 {
		h = 1600
	}
	return h
}

func summarizeTopLabels(snap ui.Snapshot, screenH int) string {
	if screenH <= 0 {
		screenH = snapMaxY(snap)
	}
	maxY := screenH * 30 / 100
	var labels []string
	seen := map[string]bool{}
	for _, elem := range snap.Elements {
		if !elem.Enabled {
			continue
		}
		_, cy := elem.Center()
		if cy > maxY {
			continue
		}
		raw := strings.TrimSpace(elem.Label())
		if raw == "" || len(raw) > 48 {
			continue
		}
		key := strings.ToLower(raw)
		if seen[key] {
			continue
		}
		seen[key] = true
		labels = append(labels, raw)
		if len(labels) >= 4 {
			break
		}
	}
	return strings.Join(labels, " | ")
}
