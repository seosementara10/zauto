package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zauto/internal/state"
	"zauto/internal/ui"
)

const recoveryMaxStages = 5

// runRecoveryEngine executes one recovery stage per unknown encounter (1–5), then STOP with diagnostics.
func (e *Executor) runRecoveryEngine(
	det *state.Detector,
	observe state.ObserveFn,
	mem *state.DeviceMemory,
	invalidate func(),
) (state.Detection, bool, error) {
	mem.UnknownCount++
	stage := mem.UnknownCount

	var stageErr error
	switch stage {
	case 1:
		stageErr = e.recoveryStageRescan(invalidate)
	case 2:
		stageErr = e.recoveryStageRefreshObserve(invalidate)
	case 3:
		stageErr = e.recoveryStageInspectForeground()
	case 4:
		stageErr = e.recoveryStageVisualDetect()
	default:
		snap, pkg, act := observe()
		inv := det.Investigate(snap, pkg, act)
		d, err := e.recoveryStageDiagnosticsStop(mem, inv, snap, pkg)
		return d, true, err
	}

	snap, pkg, act := observe()
	inv := det.Investigate(snap, pkg, act)
	mem.SetDetection(inv.Detection)

	e.event("RECOVERY unknown_count=%d stage=%d/%d kind=%s method=%s",
		stage, stage, recoveryMaxStages, inv.Detection.UnknownKind, inv.Method)

	if resolved, ok := e.tryResolvedDetection(inv.Detection); ok {
		mem.ResetUnknownCount()
		return resolved, false, nil
	}

	if stage == 4 {
		if d, ok := e.scanKnownTextCatalogs(det, snap, pkg, act); ok {
			mem.SetDetection(d)
			mem.ResetUnknownCount()
			return d, false, nil
		}
	}

	if stage >= recoveryMaxStages {
		d, err := e.recoveryStageDiagnosticsStop(mem, inv, snap, pkg)
		return d, true, err
	}

	if stageErr != nil {
		e.event("RECOVERY stage=%d incomplete: %v", stage, stageErr)
	}
	return inv.Detection, false, stageErr
}

func (e *Executor) tryResolvedDetection(d state.Detection) (state.Detection, bool) {
	if d.State == state.UIUnknown {
		return d, false
	}
	if d.Confidence >= state.VerifyMinConfidence {
		return d, true
	}
	if d.State == state.UIPermission && d.Confidence >= state.InvestigateMinConfidence {
		return d, true
	}
	if d.State == state.UIPasswordManagerSheet && d.Confidence >= state.InvestigateMinConfidence {
		return d, true
	}
	if d.State == state.UIKeyboardSettings && d.Confidence >= state.InvestigateMinConfidence {
		return d, true
	}
	return d, false
}

func (e *Executor) recoveryStageRescan(invalidate func()) error {
	e.event("RECOVERY stage=1 action=rescan (fresh observe)")
	if invalidate != nil {
		invalidate()
	}
	observe, inv := e.cachedObserve()
	if invalidate == nil {
		invalidate = inv
	}
	snap0, pkg0, act0 := observe()
	det := state.NewDetector()
	before := det.Detect(snap0, pkg0, act0)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		_, _ = e.Session.Client.DumpUI(false)
		e.invalidateObserve(invalidate)
		snap, pkg, act := observe()
		after := det.Detect(snap, pkg, act)
		if after.State != before.State || after.Confidence != before.Confidence {
			return fmt.Errorf("rescan: deferred investigate")
		}
		time.Sleep(e.Session.PollInterval())
	}
	return fmt.Errorf("rescan: deferred investigate")
}

func (e *Executor) recoveryStageRefreshObserve(invalidate func()) error {
	e.event("RECOVERY stage=2 action=refresh_observe (cache bust)")
	if invalidate != nil {
		invalidate()
	}
	// One slow dump; invalidate again so the next observe() reads fresh hierarchy.
	_, _ = e.Session.Client.DumpUI(false)
	if invalidate != nil {
		invalidate()
	}
	time.Sleep(e.Session.PollInterval())
	return fmt.Errorf("refresh_observe: deferred investigate")
}

func (e *Executor) recoveryStageInspectForeground() error {
	e.event("RECOVERY stage=3 action=inspect_foreground (hint logging after investigate)")
	return fmt.Errorf("inspect_foreground: deferred investigate")
}

func (e *Executor) recoveryStageVisualDetect() error {
	e.event("RECOVERY stage=4 action=visual_detect (catalog scan after investigate)")
	return fmt.Errorf("visual_detect: deferred investigate")
}

func (e *Executor) scanKnownTextCatalogs(det *state.Detector, snap ui.Snapshot, pkg, act string) (state.Detection, bool) {
	catalogs := []struct {
		state state.UIState
		texts []string
	}{
		{state.UIPasswordManagerSheet, state.PasswordManagerTitleTexts},
		{state.UILocationServicesPrompt, state.LocationServicesIntroTexts},
		{state.UISaveLoginPrompt, state.SaveLoginLaterTexts},
		{state.UILogoutConfirmPrompt, state.LogoutConfirmTitleTexts},
		{state.UIContactFollowPrompt, state.ContactFollowSkipTexts},
		{state.UISavedProfileScreen, state.SavedProfileOtherTexts},
		{state.UIKeyboardSettings, state.KeyboardSettingsTitleTexts},
		{state.UILogin, state.LoginEmailFieldTexts},
		{state.UI2FACheckpoint, state.TwoFactorTitleTexts},
	}
	for _, c := range catalogs {
		if !e.Session.Resolver.TextExists(snap, c.texts) {
			continue
		}
		full := det.Detect(snap, pkg, act)
		if full.State == c.state && full.Confidence >= state.VerifyMinConfidence {
			return full, true
		}
	}
	return state.Detection{}, false
}

func (e *Executor) recoveryStageDiagnosticsStop(
	mem *state.DeviceMemory,
	inv state.Investigation,
	snap ui.Snapshot,
	pkg string,
) (state.Detection, error) {
	e.event("RECOVERY stage=5 action=diagnostics_stop kind=%s", inv.Detection.UnknownKind)
	screenshot, dump := e.captureRecoveryArtifacts("unknown", snap)
	obs := state.BuildObservationSnapshot(e.Session.Serial, mem, inv, screenshot, dump)
	e.event("RECOVERY observation serial=%s prev=%s last_action=%s screenshot=%s hierarchy=%s",
		obs.Serial, obs.PreviousState, obs.LastAction, obs.Screenshot, obs.Hierarchy)
	e.logRecoveryDiagnostic(snap, pkg, inv)
	mem.ResetUnknownCount()
	return inv.Detection, fmt.Errorf("recovery STOP: unknown UI after %d stages kind=%s artifacts=%s",
		recoveryMaxStages, inv.Detection.UnknownKind, screenshot)
}

func (e *Executor) captureRecoveryArtifacts(label string, snap ui.Snapshot) (screenshotPath, dumpPath string) {
	ts := time.Now().Unix()
	dir := filepath.Join(e.Session.ProjectRoot, "captures", e.Session.Serial)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.event("CAPTURE mkdir failed label=%s dir=%s err=%v", label, dir, err)
		return "", ""
	}
	screenshotPath = filepath.Join(dir, fmt.Sprintf("%s_%d.png", label, ts))
	dumpPath = filepath.Join(dir, fmt.Sprintf("%s_%d.xml", label, ts))
	if err := e.Session.Client.Screenshot(screenshotPath); err != nil {
		e.event("CAPTURE screenshot failed label=%s path=%s err=%v", label, screenshotPath, err)
		screenshotPath = ""
	} else {
		e.event("CAPTURE screenshot ok label=%s path=%s", label, screenshotPath)
	}
	if snap.XML == "" {
		snap = e.Session.ReadUI(true)
	}
	if err := os.WriteFile(dumpPath, []byte(snap.XML), 0o644); err != nil {
		e.event("CAPTURE hierarchy write failed label=%s path=%s err=%v elements=%d", label, dumpPath, err, len(snap.Elements))
		dumpPath = ""
	} else {
		e.event("CAPTURE hierarchy ok label=%s path=%s elements=%d", label, dumpPath, len(snap.Elements))
	}
	return screenshotPath, dumpPath
}
