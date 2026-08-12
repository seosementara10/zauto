package runtime

import (
	"context"
	"time"

	"zauto/internal/engine/flow"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// Exec is the automation surface used by login/, logout/, and overlay/ without importing engine.
type Exec interface {
	Sess() *Session
	Event(msg string, args ...interface{})
	Ctx() context.Context
	CachedObserve() (state.ObserveFn, func())
	ReadSnap(observe state.ObserveFn) ui.Snapshot
	InvalidateObserve(invalidate func())
	PollTapObserve(observe state.ObserveFn, invalidate func(), queries []ui.FindQuery, timeoutSec float64) error
	TapFirstQuery(queries []ui.FindQuery, timeoutSec float64) error
	TapFirstQueryObserve(observe state.ObserveFn, invalidate func(), queries []ui.FindQuery, timeoutSec float64) error
	WaitTapTexts(texts []string, prefer string, minY, maxY int, timeoutSec float64) error
	ActUntilNotState(det *state.Detector, observe state.ObserveFn, avoid state.UIState, timeout time.Duration, label string, act func() error) error
	WaitUntilNotState(det *state.Detector, observe state.ObserveFn, avoid state.UIState, timeout time.Duration, label string) error
	ActUntilState(det *state.Detector, observe state.ObserveFn, invalidate func(), target state.UIState, timeout time.Duration, label string, act func() error) error
	ActUntilPredicate(observe state.ObserveFn, invalidate func(), timeout time.Duration, label string, act func() error, untilFn func(ui.Snapshot) bool) error
	WaitUntilPredicate(observe state.ObserveFn, timeout time.Duration, label string, untilFn func(ui.Snapshot) bool) error
	FillLoginField(labels, resourceIDs []string, value string, editIndex int) error
	CaptureRecoveryArtifacts(label string, snap ui.Snapshot) (screenshotPath, dumpPath string)
	LogScreen(observe state.ObserveFn, where string, note ScreenNote)
	LogScreenIfStale(observe state.ObserveFn, where string, note ScreenNote, minInterval time.Duration)
	CaptureFailure(label, where, detail string, observe state.ObserveFn, note ScreenNote) (screenshotPath, dumpPath string)
	DeviceIndex() int
	ErrUnhandledState(d state.Detection) error
	RunStateLoop(ctx context.Context, det *state.Detector, observe state.ObserveFn, invalidate func(), mem *state.DeviceMemory, timeout time.Duration, fillCredentials func() error) error
	RunOverlayAwareFlow(ctx context.Context, det *state.Detector, observe state.ObserveFn, invalidate func(), mem *state.DeviceMemory, timeout time.Duration, watch []state.UIState, goal flow.Goal, onProgress func(state.Detection) error) error
}
