package logout

import (
	"fmt"
	"time"

	"zauto/internal/config"
	"zauto/internal/engine/runtime"
	"zauto/internal/engine/flow"
	"zauto/internal/engine/overlay"
	"zauto/internal/state"
)

type phase int

const (
	phasePrecheck phase = iota
	phaseMenu
	phaseConfirm
)

type logoutFlow struct {
	action           config.Action
	phase            phase
	menuDone         bool
	confirmDone      bool
	keluarTapped     bool
	lastLogPhase     phase
	precheckDeadline time.Time
}

func (lf *logoutFlow) phaseName() string {
	switch lf.phase {
	case phasePrecheck:
		return "precheck"
	case phaseMenu:
		return "menu"
	case phaseConfirm:
		return "confirm"
	default:
		return "unknown"
	}
}

func (lf *logoutFlow) logPhase(e runtime.Exec, d state.Detection) {
	if lf.lastLogPhase == lf.phase {
		return
	}
	e.Event("LOGOUT phase=%s ui_state=%s confidence=%.0f%%", lf.phaseName(), d.State, d.Confidence*100)
	lf.lastLogPhase = lf.phase
}

func (lf *logoutFlow) progress(e runtime.Exec, observe state.ObserveFn, invalidate func()) func(state.Detection) error {
	return func(d state.Detection) error {
		lf.logPhase(e, d)
		switch lf.phase {
		case phasePrecheck:
			if d.State == state.UILoggedIn && d.Confidence >= state.VerifyMinConfidence {
				lf.phase = phaseMenu
				return nil
			}
			snap := e.ReadSnap(observe)
			feedHints := state.LoggedInFeedHints
			if e.Sess().Resolver.TextExists(snap, feedHints) {
				lf.phase = phaseMenu
				return nil
			}
			if time.Now().After(lf.precheckDeadline) {
				e.CaptureRecoveryArtifacts("logout_not_logged_in", snap)
				return fmt.Errorf("not on logged-in feed before logout (state=%s %.0f%%)", d.State, d.Confidence*100)
			}
			return nil

		case phaseMenu:
			if lf.menuDone {
				return nil
			}
			if state.IsOverlayState(d.State) && d.Confidence >= state.VerifyMinConfidence {
				return nil
			}
			snap := e.ReadSnap(observe)
			if lf.successGoal(e.Sess().Resolver, d, snap) {
				e.Event("VERIFY logout already on login screen (skip menu)")
				lf.menuDone = true
				lf.confirmDone = true
				return nil
			}
			onFeed := d.State == state.UILoggedIn && d.Confidence >= state.VerifyMinConfidence
			if !onFeed && !menuDrawerOpen(e, snap) {
				if !e.Sess().Resolver.TextExists(snap, state.LoggedInFeedHints) {
					return nil
				}
			}
			e.Event("ACT open menu for logout")
			menuTimeout := lf.action.ParamFloat("menu_timeout_sec", 10)
			if err := openMenu(e, observe, invalidate, menuTimeout); err != nil {
				return err
			}
			snap = e.ReadSnap(observe)
			if !e.Sess().Resolver.TextExists(snap, menuItemTexts) {
				if !scrollMenuDrawer(e, observe, invalidate, 6) {
					e.CaptureRecoveryArtifacts("logout_menu_scroll", snap)
					return fmt.Errorf("logout menu item not visible (menu drawer scroll failed; feed scroll avoided)")
				}
			}
			e.Event("ACT tap Keluar in menu")
			if err := tapMenuItem(e, observe, invalidate, lf.action.ParamFloat("verify_timeout_sec", 10)); err != nil {
				return fmt.Errorf("logout menu item: %w", err)
			}
			lf.keluarTapped = true
			e.Event("LOGOUT keluar tapped — expect save_login then logout_confirm overlays")
			e.InvalidateObserve(invalidate)
			lf.menuDone = true
			lf.phase = phaseConfirm
			lf.lastLogPhase = phase(-1)
			return nil

		case phaseConfirm:
			if lf.confirmDone {
				return nil
			}
			snap := e.ReadSnap(observe)
			if lf.successGoal(e.Sess().Resolver, d, snap) {
				switch d.State {
				case state.UISavedProfileScreen:
					e.Event("VERIFY logout reached saved profile picker state=%s confidence=%.0f%%", d.State, d.Confidence*100)
				default:
					e.Event("VERIFY logout reached login screen state=%s confidence=%.0f%%", d.State, d.Confidence*100)
				}
				lf.confirmDone = true
				return nil
			}
			if state.IsOverlayState(d.State) && d.Confidence >= state.VerifyMinConfidence {
				e.Event("LOGOUT confirm waiting overlay state=%s", d.State)
				return nil
			}
			if e.Sess().Resolver.TextExists(snap, state.LogoutConfirmTitleTexts) ||
				e.Sess().Resolver.TextExists(snap, state.LogoutConfirmButtonTexts) {
				e.Event("ACT tap logout confirm (fallback)")
				if err := overlay.TapLogoutConfirmButton(e, observe, invalidate, lf.action.ParamFloat("verify_timeout_sec", 12)); err != nil {
					return fmt.Errorf("logout confirm: %w", err)
				}
				e.InvalidateObserve(invalidate)
				return nil
			}
			return nil
		}
		return nil
	}
}

// Run executes the Facebook logout ODAV flow.
func Run(e runtime.Exec, action config.Action) error {
	ctx := e.Ctx()
	det := state.NewDetector()
	mem := e.Sess().Memory
	if mem == nil {
		mem = &state.DeviceMemory{}
		e.Sess().Memory = mem
	}
	mem.ResetFlow()

	observe, invalidate := e.CachedObserve()
	lf := &logoutFlow{
		action:           action,
		precheckDeadline: time.Now().Add(8 * time.Second),
	}

	menuSec := action.ParamFloat("menu_timeout_sec", 10)
	verifySec := action.ParamFloat("verify_timeout_sec", 25)
	timeout := time.Duration((menuSec + verifySec*2) * float64(time.Second))

	goal := flow.Goal(lf.successGoal)
	return e.RunOverlayAwareFlow(
		ctx, det, observe, invalidate, mem, timeout,
		state.LoginFlowWatchStates(),
		goal,
		lf.progress(e, observe, invalidate),
	)
}
