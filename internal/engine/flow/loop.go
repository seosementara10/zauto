package flow

import (
	"context"
	"fmt"
	"time"

	"zauto/internal/state"
)

const PollWindow = 2 * time.Second

// RunLoop executes OBSERVE → DETECT → RECOVERY → DISPATCH until success or timeout.
func RunLoop(
	ctx context.Context,
	h Hooks,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	mem *state.DeviceMemory,
	timeout time.Duration,
	pollTick time.Duration,
	spec Spec,
) error {
	if pollTick <= 0 {
		pollTick = state.DefaultPollInterval
	}
	deadline := time.Now().Add(timeout)
	var lastLogged state.UIState

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		poll := PollWindow
		if remaining < poll {
			poll = remaining
		}

		d, _ := det.WaitUntilAny(ctx, observe, spec.Watch, poll, state.VerifyMinConfidence, pollTick)
		mem.SetDetection(d)
		snap, _, _ := observe()

		if spec.Goal != nil && spec.Goal(h.Resolver(), d, snap) {
			if spec.LogGoalOnSuccess {
				h.Event("VERIFY flow goal satisfied state=%s confidence=%.0f%%", d.State, d.Confidence*100)
			}
			mem.ResetUnknownCount()
			return nil
		}

		if d.State != lastLogged {
			label := string(d.State)
			if d.UnknownKind != state.UnknownKindNone {
				label += "/" + string(d.UnknownKind)
			}
			h.Event("STATE %s confidence=%.0f%% evidence=%v", label, d.Confidence*100, d.Evidence)
			lastLogged = d.State
		}

		if d.State == state.UIError && d.Confidence >= state.VerifyMinConfidence {
			return fmt.Errorf("error screen detected")
		}
		if spec.ExitOnLoggedIn && d.State == state.UILoggedIn && d.Confidence >= state.VerifyMinConfidence {
			h.Event("VERIFY logged_in ok confidence=%.0f%%", d.Confidence*100)
			mem.TaskState = state.TaskSuccess
			mem.ClearAuthWait()
			return nil
		}

		if d.IsUncertain() || d.State == state.UIUnknown {
			h.Event("UNKNOWN — enter recovery engine")
			recovered, fatal, err := h.RunRecovery(det, observe, mem, invalidate)
			if fatal {
				return err
			}
			if err != nil {
				h.InvalidateObserve(invalidate)
				continue
			}
			h.InvalidateObserve(invalidate)
			d = recovered
			mem.SetDetection(d)
			snap, _, _ = observe()
			if spec.ExitOnLoggedIn && d.State == state.UILoggedIn && d.Confidence >= state.VerifyMinConfidence {
				mem.TaskState = state.TaskSuccess
				mem.ResetUnknownCount()
				return nil
			}
			if spec.Goal != nil && spec.Goal(h.Resolver(), d, snap) {
				mem.ResetUnknownCount()
				return nil
			}
		}

		if resolved, ok := h.TryResolved(d); ok && state.IsOverlayState(resolved.State) {
			if err := h.DispatchOverlay(det, observe, resolved); err != nil {
				return err
			}
			h.InvalidateObserve(invalidate)
			mem.ResetUnknownCount()
			continue
		}

		if spec.TrackAuthResult && mem.AuthPhase == state.AuthPhaseWaitResult {
			if d.State == state.UIError && d.Confidence >= state.VerifyMinConfidence {
				mem.ClearAuthWait()
				return fmt.Errorf("login error after credentials submitted")
			}
			if d.State == state.UILogin {
				if time.Since(mem.AuthFilledAt) > state.AuthResultTimeout {
					mem.ClearAuthWait()
					h.CaptureFlowTimeout("auth_timeout", mem, det, observe)
					return fmt.Errorf("auth timeout: no transition from login within %s", state.AuthResultTimeout)
				}
				continue
			}
			if d.State == state.UILoading {
				continue
			}
		}

		if spec.OnProgress != nil && d.Confidence >= state.VerifyMinConfidence {
			if err := spec.OnProgress(d); err != nil {
				return err
			}
			h.InvalidateObserve(invalidate)
			mem.ResetUnknownCount()
			continue
		}

		switch d.State {
		case state.UILoading:
			continue
		case state.UI2FACheckpoint:
			h.CaptureFlowTimeout("2fa_checkpoint", mem, det, observe)
			return fmt.Errorf("2FA/checkpoint detected — manual intervention required")
		case state.UIOnboarding:
			if !spec.HandleOnboarding {
				return h.ErrUnhandledState(d)
			}
			if err := h.SkipOnboarding(); err != nil {
				return err
			}
			h.InvalidateObserve(invalidate)
		case state.UILogin:
			if spec.FillCreds == nil {
				if d.Confidence >= state.VerifyMinConfidence {
					return h.ErrUnhandledState(d)
				}
				continue
			}
			if mem.AuthPhase == state.AuthPhaseWaitResult {
				continue
			}
			if !d.CanExecute() {
				return fmt.Errorf("decide: login screen not actionable (%.0f%%)", d.Confidence*100)
			}
			if err := spec.FillCreds(); err != nil {
				return err
			}
			h.InvalidateObserve(invalidate)
			mem.BeginAuthWait()
			h.Event("AUTH wait_auth_result started")
		default:
			if d.Confidence >= state.VerifyMinConfidence {
				return h.ErrUnhandledState(d)
			}
		}
	}
	h.CaptureFlowTimeout("flow_timeout", mem, det, observe)
	return fmt.Errorf("flow loop timeout")
}
