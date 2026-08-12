package flow

import (
	"zauto/internal/state"
	"zauto/internal/ui"
)

// Goal is satisfied when the screen is ready to proceed (e.g. login form or saved profile picker).
type Goal func(resolver *ui.Resolver, det state.Detection, snap ui.Snapshot) bool

// Hooks binds shared engine services into the ODAV loop (recovery, overlay dispatch, logging).
type Hooks struct {
	Event               func(msg string, args ...interface{})
	Resolver            func() *ui.Resolver
	RunRecovery         func(det *state.Detector, observe state.ObserveFn, mem *state.DeviceMemory, invalidate func()) (state.Detection, bool, error)
	TryResolved         func(d state.Detection) (state.Detection, bool)
	DispatchOverlay     func(det *state.Detector, observe state.ObserveFn, d state.Detection) error
	InvalidateObserve   func(invalidate func())
	SkipOnboarding      func() error
	CaptureFlowTimeout  func(label string, mem *state.DeviceMemory, det *state.Detector, observe state.ObserveFn)
	ErrUnhandledState   func(d state.Detection) error
}
