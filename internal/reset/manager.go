package reset

import (
	"context"
	"fmt"
	"time"

	"zauto/internal/adb"
	"zauto/internal/logging"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// Input for ResetAndLaunch.
type Input struct {
	Client        *adb.Client
	Package       string
	Activity      string
	Log           *logging.DeviceLogger
	VerifyTimeout time.Duration
	Poll          time.Duration
}

// Manager resets app data via pm clear and verifies initial UI state after launch.
type Manager struct {
	det *state.Detector
}

func NewManager() *Manager {
	return &Manager{det: state.NewDetector()}
}

func ObserveFromClient(c *adb.Client) state.ObserveFn {
	return func() (ui.Snapshot, string, string) {
		xml, _ := c.DumpUI(true)
		return ui.ParseHierarchy(xml), c.ForegroundPackage(), ""
	}
}

// WaitInitialUIState polls until a valid post-launch state appears (no fixed sleep).
// When allowLoggedIn is true (resume without pm clear), an existing session on the feed is accepted.
func WaitInitialUIState(ctx context.Context, client *adb.Client, timeout time.Duration, poll time.Duration, allowLoggedIn bool) (state.Detection, error) {
	if timeout <= 0 {
		timeout = 25 * time.Second
	}
	if poll <= 0 {
		poll = state.DefaultPollInterval
	}
	det := state.NewDetector()
	observe := ObserveFromClient(client)
	return verifyInitialState(ctx, det, observe, timeout, poll, allowLoggedIn)
}

func verifyInitialState(ctx context.Context, det *state.Detector, observe state.ObserveFn, timeout, poll time.Duration, allowLoggedIn bool) (state.Detection, error) {
	targets := state.PostResetValidStates()
	if allowLoggedIn {
		targets = append(append([]state.UIState(nil), targets...), state.UILoggedIn)
	}
	d, ok := det.WaitUntilAny(ctx, observe, targets, timeout, state.VerifyMinConfidence, poll)
	if !ok {
		snap, pkg, act := observe()
		d = det.Detect(snap, pkg, act)
	}
	if d.IsUncertain() && allowLoggedIn {
		snap, pkg, act := observe()
		if feedLoggedInVisible(snap) {
			return state.Detection{
				State:      state.UILoggedIn,
				Score:      75,
				Confidence: 0.75,
				Evidence:   []string{"feed_hints (resume)"},
				Package:    pkg,
				Activity:   act,
				At:         time.Now(),
			}, nil
		}
	}
	if d.IsUncertain() {
		return d, fmt.Errorf("initial state uncertain: %s (%.0f%%)", d.State, d.Confidence*100)
	}
	if d.State == state.UILoggedIn {
		if allowLoggedIn {
			return d, nil
		}
		return d, fmt.Errorf("still logged_in (%.0f%%) — session not cleared", d.Confidence*100)
	}
	if !state.IsPostResetValidState(d.State) {
		return d, fmt.Errorf("unexpected initial state %s (%.0f%%)", d.State, d.Confidence*100)
	}
	return d, nil
}

func feedLoggedInVisible(snap ui.Snapshot) bool {
	return ui.NewDefaultResolver().TextExists(snap, state.LoggedInFeedHints)
}

// ResetAndLaunch: force-stop → pm clear → launch → wait initial UI state.
func (m *Manager) ResetAndLaunch(ctx context.Context, in Input) (state.Detection, error) {
	if in.Client == nil {
		return state.Detection{}, fmt.Errorf("reset: nil adb client")
	}
	if in.Package == "" {
		return state.Detection{}, fmt.Errorf("reset: empty package")
	}
	if in.VerifyTimeout <= 0 {
		in.VerifyTimeout = 25 * time.Second
	}

	log := in.Log
	step := func(msg string, args ...interface{}) {
		if log != nil {
			log.Info("RESET "+msg, args...)
		}
	}

	if !in.Client.IsPackageInstalled(in.Package) {
		return state.Detection{}, fmt.Errorf("reset: package not installed: %s", in.Package)
	}
	step("CHECK_DEVICE ok pkg=%s", in.Package)

	step("FORCE_STOP %s", in.Package)
	if err := in.Client.ForceStop(in.Package); err != nil {
		return state.Detection{}, fmt.Errorf("reset force-stop: %w", err)
	}

	step("PM_CLEAR %s", in.Package)
	if err := in.Client.ClearPackageData(in.Package); err != nil {
		return state.Detection{}, fmt.Errorf("reset pm clear: %w", err)
	}
	step("PM_CLEAR ok")

	step("LAUNCH %s", in.Package)
	if err := in.Client.OpenApp(in.Package, in.Activity); err != nil {
		return state.Detection{}, fmt.Errorf("reset launch: %w", err)
	}

	observe := ObserveFromClient(in.Client)
	poll := in.Poll
	if poll <= 0 {
		poll = state.DefaultPollInterval
	}
	d, err := verifyInitialState(ctx, m.det, observe, in.VerifyTimeout, poll, false)
	if err != nil {
		return d, err
	}

	if state.IsOverlayState(d.State) {
		step("INITIAL_STATE %s confidence=%.0f%% (overlay — handoff to ODAV)", d.State, d.Confidence*100)
	} else {
		step("INITIAL_STATE %s confidence=%.0f%%", d.State, d.Confidence*100)
	}
	return d, nil
}
