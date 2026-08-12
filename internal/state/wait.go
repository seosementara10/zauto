package state

import (
	"context"
	"time"

	"zauto/internal/ui"
)

// ObserveFn returns a fresh UI snapshot plus optional package/activity sensors.
type ObserveFn func() (ui.Snapshot, string, string)

// WaitUntilState polls Observe until target state reaches minConfidence or timeout.
func (d *Detector) WaitUntilState(ctx context.Context, observe ObserveFn, target UIState, timeout time.Duration, minConfidence float64, poll time.Duration) (Detection, bool) {
	if poll <= 0 {
		poll = DefaultPollInterval
	}
	deadline := time.Now().Add(timeout)
	var last Detection
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return last, false
		default:
		}
		snap, pkg, act := observe()
		last = d.Detect(snap, pkg, act)
		if last.State == target && last.Confidence >= minConfidence {
			return last, true
		}
		time.Sleep(poll)
	}
	return last, false
}

// WaitUntilAny polls until one of the target states matches.
func (d *Detector) WaitUntilAny(ctx context.Context, observe ObserveFn, targets []UIState, timeout time.Duration, minConfidence float64, poll time.Duration) (Detection, bool) {
	if poll <= 0 {
		poll = DefaultPollInterval
	}
	want := map[UIState]bool{}
	for _, t := range targets {
		want[t] = true
	}
	deadline := time.Now().Add(timeout)
	var last Detection
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return last, false
		default:
		}
		snap, pkg, act := observe()
		last = d.Detect(snap, pkg, act)
		if want[last.State] && last.Confidence >= minConfidence {
			return last, true
		}
		time.Sleep(poll)
	}
	return last, false
}

// WaitUntilNotState polls until detection no longer matches avoid with minConfidence.
func (d *Detector) WaitUntilNotState(ctx context.Context, observe ObserveFn, avoid UIState, timeout time.Duration, minConfidence float64, poll time.Duration) (Detection, bool) {
	if poll <= 0 {
		poll = DefaultPollInterval
	}
	deadline := time.Now().Add(timeout)
	var last Detection
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return last, false
		default:
		}
		snap, pkg, act := observe()
		last = d.Detect(snap, pkg, act)
		if last.State != avoid || last.Confidence < minConfidence {
			return last, true
		}
		time.Sleep(poll)
	}
	return last, false
}
