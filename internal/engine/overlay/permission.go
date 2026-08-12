package overlay

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// PermissionPolicyRuntimeKey stores the active permission policy in session runtime.
const PermissionPolicyRuntimeKey = "permission_policy"

func permissionPolicy(e runtime.Exec) state.PermissionPolicy {
	if v, ok := e.Sess().Runtime[PermissionPolicyRuntimeKey].(string); ok {
		switch v {
		case "allow":
			return state.UniformPermissionPolicy(state.PermActionAllow)
		case "deny":
			return state.UniformPermissionPolicy(state.PermActionDeny)
		}
	}
	return state.DefaultPermissionPolicy()
}

// HandlePermissionDialog runs the permission overlay handler (used by handle_permission action).
func HandlePermissionDialog(e runtime.Exec, det *state.Detector, observe state.ObserveFn, invalidate func(), verifyTimeout time.Duration) error {
	return handlePermissionObserve(e, det, observe, invalidate, verifyTimeout)
}

func HandlePermission(e runtime.Exec, det *state.Detector, observe state.ObserveFn) error {
	return HandlePermissionDialog(e, det, observe, nil, 10*time.Second)
}

func handlePermissionObserve(
	e runtime.Exec,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	verifyTimeout time.Duration,
) error {
	snap, pkg, _ := observe()
	kind := state.IdentifyPermissionKind(e.Sess().Resolver, snap, pkg)
	action := permissionPolicy(e).ActionFor(kind)

	e.Event("DECIDE permission kind=%s action=%s (fail-closed)", kind, action)

	switch action {
	case state.PermActionDeny:
		return permissionDenyObserve(e, det, observe, invalidate, verifyTimeout)
	case state.PermActionAllow:
		return e.ActUntilNotState(det, observe, state.UIPermission, verifyTimeout, "permission allow", func() error {
			q := ui.FindQuery{Texts: state.PermissionAllowTexts, PreferClickable: true, Prefer: "first"}
			if err := e.PollTapObserve(observe, invalidate, []ui.FindQuery{q}, 5); err != nil {
				return fmt.Errorf("permission allow: %w", err)
			}
			return nil
		})
	default:
		return fmt.Errorf("permission fail-closed: kind=%s pkg=%s — manual review required", kind, pkg)
	}
}

func permissionDenyObserve(
	e runtime.Exec,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	verifyTimeout time.Duration,
) error {
	e.Event("ACT tap deny permission")
	queries := []ui.FindQuery{
		{Texts: []string{"Jangan izinkan", "JANGAN IZINKAN"}, PreferClickable: true, Prefer: "left"},
		{Texts: state.PermissionDenyTexts, PreferClickable: true, Prefer: "left"},
		{Texts: state.PermissionDenyTexts, PreferClickable: true, Prefer: "bottom"},
		{Texts: state.PermissionDenyTexts, PreferClickable: true, Prefer: "first"},
	}
	deadline := time.Now().Add(verifyTimeout)
	for time.Now().Before(deadline) {
		snap, pkg, actName := observe()
		d := det.Detect(snap, pkg, actName)
		if d.State != state.UIPermission || d.Confidence < state.VerifyMinConfidence {
			e.Event("VERIFY permission dismissed")
			return nil
		}
		for _, q := range queries {
			if err := e.PollTapObserve(observe, invalidate, []ui.FindQuery{q}, 2); err == nil {
				e.Event("ACT permission deny tapped")
				break
			}
		}
		time.Sleep(e.Sess().PollInterval())
	}
	snap, pkg, actName := observe()
	d := det.Detect(snap, pkg, actName)
	if d.State == state.UIPermission && d.Confidence >= state.VerifyMinConfidence {
		e.CaptureRecoveryArtifacts("permission_deny_failed", snap)
		return fmt.Errorf("permission deny: dialog still visible (%.0f%%)", d.Confidence*100)
	}
	return nil
}
