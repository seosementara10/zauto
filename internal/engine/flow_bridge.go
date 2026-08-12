package engine

import (
	"context"
	"time"

	"zauto/internal/engine/flow"
	"zauto/internal/engine/overlay"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func (e *Executor) flowHooks() flow.Hooks {
	return flow.Hooks{
		Event:    e.Event,
		Resolver: func() *ui.Resolver { return e.Session.Resolver },
		RunRecovery: func(det *state.Detector, observe state.ObserveFn, mem *state.DeviceMemory, invalidate func()) (state.Detection, bool, error) {
			return e.RunRecoveryEngine(det, observe, mem, invalidate)
		},
		TryResolved: e.TryResolvedDetection,
		DispatchOverlay: func(det *state.Detector, observe state.ObserveFn, d state.Detection) error {
			return overlay.Dispatch(e, det, observe, d)
		},
		InvalidateObserve:  e.InvalidateObserve,
		SkipOnboarding:     e.skipOnboarding,
		CaptureFlowTimeout: e.CaptureFlowTimeout,
		ErrUnhandledState:  e.ErrUnhandledState,
	}
}

func (e *Executor) skipOnboarding() error {
	e.Event("DECIDE skip onboarding → login")
	texts := append(append([]string(nil), state.OnboardingHaveAccountTexts...), state.LoginButtonTexts...)
	return e.WaitTapTexts(texts, "first", 0, 0, 8)
}

func (e *Executor) runFlowLoop(
	ctx context.Context,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	mem *state.DeviceMemory,
	timeout time.Duration,
	spec flow.Spec,
) error {
	return flow.RunLoop(ctx, e.flowHooks(), det, observe, invalidate, mem, timeout, e.Session.PollInterval(), spec)
}

// RunStateLoop runs the login ODAV loop until logged_in or timeout.
func (e *Executor) RunStateLoop(
	ctx context.Context,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	mem *state.DeviceMemory,
	timeout time.Duration,
	fillCredentials func() error,
) error {
	return e.runStateLoop(ctx, det, observe, invalidate, mem, timeout, fillCredentials)
}

func (e *Executor) runStateLoop(
	ctx context.Context,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	mem *state.DeviceMemory,
	timeout time.Duration,
	fillCredentials func() error,
) error {
	return e.runFlowLoop(ctx, det, observe, invalidate, mem, timeout, flow.LoginSpec(fillCredentials))
}

// RunOverlayAwareFlow runs overlay dispatch + optional progress until goal is met.
func (e *Executor) RunOverlayAwareFlow(
	ctx context.Context,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	mem *state.DeviceMemory,
	timeout time.Duration,
	watch []state.UIState,
	goal flow.Goal,
	onProgress func(state.Detection) error,
) error {
	return e.runOverlayAwareFlow(ctx, det, observe, invalidate, mem, timeout, watch, goal, onProgress)
}

func (e *Executor) runOverlayAwareFlow(
	ctx context.Context,
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	mem *state.DeviceMemory,
	timeout time.Duration,
	watch []state.UIState,
	goal flow.Goal,
	onProgress func(state.Detection) error,
) error {
	spec := flow.OverlaySpec(goal, onProgress)
	spec.Watch = watch
	return e.runFlowLoop(ctx, det, observe, invalidate, mem, timeout, spec)
}
