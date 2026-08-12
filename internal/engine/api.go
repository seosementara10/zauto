package engine

import (
	"context"
	"time"

	"zauto/internal/state"
	"zauto/internal/ui"
)

// API exports executor capabilities to login/, logout/, and overlay/ subpackages.

func (e *Executor) Sess() *Session { return e.Session }

func (e *Executor) Event(msg string, args ...interface{}) { e.event(msg, args...) }

func (e *Executor) Ctx() context.Context { return e.ctx() }

func (e *Executor) CachedObserve() (state.ObserveFn, func()) { return e.cachedObserve() }

func (e *Executor) ReadSnap(observe state.ObserveFn) ui.Snapshot { return e.readSnap(observe) }

func (e *Executor) InvalidateObserve(invalidate func()) { e.invalidateObserve(invalidate) }

func (e *Executor) PollTapObserve(observe state.ObserveFn, invalidate func(), queries []ui.FindQuery, timeoutSec float64) error {
	return e.pollTapObserve(observe, invalidate, queries, timeoutSec)
}

func (e *Executor) TapFirstQuery(queries []ui.FindQuery, timeoutSec float64) error {
	return e.tapFirstQuery(queries, timeoutSec)
}

func (e *Executor) TapFirstQueryObserve(observe state.ObserveFn, invalidate func(), queries []ui.FindQuery, timeoutSec float64) error {
	return e.tapFirstQueryObserve(observe, invalidate, queries, timeoutSec)
}

func (e *Executor) WaitTapTexts(texts []string, prefer string, minY, maxY int, timeoutSec float64) error {
	return e.waitTapTexts(texts, prefer, minY, maxY, timeoutSec)
}

func (e *Executor) ActUntilNotState(det *state.Detector, observe state.ObserveFn, avoid state.UIState, timeout time.Duration, label string, act func() error) error {
	return e.actUntilNotState(det, observe, avoid, timeout, label, act)
}

func (e *Executor) WaitUntilNotState(det *state.Detector, observe state.ObserveFn, avoid state.UIState, timeout time.Duration, label string) error {
	return e.waitUntilNotState(det, observe, avoid, timeout, label)
}

func (e *Executor) ActUntilState(det *state.Detector, observe state.ObserveFn, invalidate func(), target state.UIState, timeout time.Duration, label string, act func() error) error {
	return e.actUntilState(det, observe, invalidate, target, timeout, label, act)
}

func (e *Executor) ActUntilPredicate(observe state.ObserveFn, invalidate func(), timeout time.Duration, label string, act func() error, untilFn func(ui.Snapshot) bool) error {
	return e.actUntilPredicate(observe, invalidate, timeout, label, act, untilFn)
}

func (e *Executor) WaitUntilPredicate(observe state.ObserveFn, timeout time.Duration, label string, untilFn func(ui.Snapshot) bool) error {
	return e.waitUntilPredicate(observe, timeout, label, untilFn)
}

func (e *Executor) FillLoginField(labels, resourceIDs []string, value string, editIndex int) error {
	return e.fillLoginField(labels, resourceIDs, value, editIndex)
}

func (e *Executor) CaptureRecoveryArtifacts(label string, snap ui.Snapshot) (screenshotPath, dumpPath string) {
	return e.captureRecoveryArtifacts(label, snap)
}

// LogScreen, LogScreenIfStale, CaptureFailure are on Executor in diagnostics.go.

func (e *Executor) RunRecoveryEngine(det *state.Detector, observe state.ObserveFn, mem *state.DeviceMemory, invalidate func()) (state.Detection, bool, error) {
	return e.runRecoveryEngine(det, observe, mem, invalidate)
}

func (e *Executor) TryResolvedDetection(d state.Detection) (state.Detection, bool) {
	return e.tryResolvedDetection(d)
}

func (e *Executor) CaptureFlowTimeout(label string, mem *state.DeviceMemory, det *state.Detector, observe state.ObserveFn) {
	e.captureFlowTimeout(label, mem, det, observe)
}

func (e *Executor) LogRecoveryDiagnostic(snap ui.Snapshot, pkg string, inv state.Investigation) {
	e.logRecoveryDiagnostic(snap, pkg, inv)
}

func (e *Executor) DeviceIndex() int { return e.deviceIndex() }

func (e *Executor) ErrUnhandledState(d state.Detection) error { return errUnhandledState(d) }
