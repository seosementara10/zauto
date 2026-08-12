package engine

import (
	"fmt"
	"time"

	"zauto/internal/state"
	"zauto/internal/ui"
)

// actUntilNotState runs act then waits until avoid state clears (state-aware verify, not fixed sleep).
func (e *Executor) actUntilNotState(
	det *state.Detector,
	observe state.ObserveFn,
	avoid state.UIState,
	timeout time.Duration,
	label string,
	act func() error,
) error {
	if err := act(); err != nil {
		return err
	}
	_, ok := det.WaitUntilNotState(e.ctx(), observe, avoid, timeout, state.VerifyMinConfidence, e.Session.PollInterval())
	if ok {
		return nil
	}
	snap, pkg, actName := observe()
	d := det.Detect(snap, pkg, actName)
	if d.State == avoid && d.Confidence >= state.VerifyMinConfidence {
		return fmt.Errorf("%s: %s still visible (%.0f%%)", label, avoid, d.Confidence*100)
	}
	return nil
}

func (e *Executor) waitUntilNotState(det *state.Detector, observe state.ObserveFn, avoid state.UIState, timeout time.Duration, label string) error {
	return e.actUntilNotState(det, observe, avoid, timeout, label, func() error { return nil })
}

// actUntilState runs act then polls until target UIState is detected.
func (e *Executor) actUntilState(
	det *state.Detector,
	observe state.ObserveFn,
	invalidate func(),
	target state.UIState,
	timeout time.Duration,
	label string,
	act func() error,
) error {
	if err := act(); err != nil {
		return err
	}
	e.invalidateObserve(invalidate)
	_, ok := det.WaitUntilState(e.ctx(), observe, target, timeout, state.VerifyMinConfidence, e.Session.PollInterval())
	if ok {
		return nil
	}
	return fmt.Errorf("%s: %s not reached within timeout", label, target)
}

// actUntilPredicate runs act then polls until untilFn returns true on a fresh snapshot.
func (e *Executor) actUntilPredicate(
	observe state.ObserveFn,
	invalidate func(),
	timeout time.Duration,
	label string,
	act func() error,
	untilFn func(ui.Snapshot) bool,
) error {
	if err := act(); err != nil {
		return err
	}
	e.invalidateObserve(invalidate)
	if err := e.waitUntilPredicate(observe, timeout, label, untilFn); err != nil {
		return err
	}
	return nil
}

func (e *Executor) waitUntilPredicate(observe state.ObserveFn, timeout time.Duration, label string, untilFn func(ui.Snapshot) bool) error {
	if untilFn == nil {
		return nil
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := e.readSnap(observe)
		if untilFn(snap) {
			return nil
		}
		time.Sleep(e.Session.PollInterval())
	}
	return fmt.Errorf("%s: condition not met within timeout", label)
}
