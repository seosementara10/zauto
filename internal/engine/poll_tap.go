package engine

import (
	"fmt"
	"time"

	"zauto/internal/state"
	"zauto/internal/ui"
)

// pollTap finds the first matching query and taps it until timeout (direct ReadUI).
func (e *Executor) pollTap(queries []ui.FindQuery, timeoutSec float64) error {
	return e.pollTapObserve(nil, nil, queries, timeoutSec)
}

// pollTapObserve uses cached observe when provided and invalidates after a successful tap.
func (e *Executor) pollTapObserve(
	observe state.ObserveFn,
	invalidate func(),
	queries []ui.FindQuery,
	timeoutSec float64,
) error {
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	var lastErr error
	for time.Now().Before(deadline) {
		snap := e.readSnap(observe)
		for _, q := range queries {
			if r := e.Session.Resolver.Find(snap, q); r != nil {
				x, y := r.Center()
				if err := e.Session.Client.Tap(x, y); err != nil {
					lastErr = err
					continue
				}
				e.invalidateObserve(invalidate)
				return nil
			}
		}
		time.Sleep(e.Session.PollInterval())
	}
	if lastErr != nil {
		return lastErr
	}
	if len(queries) == 1 && len(queries[0].Texts) > 0 {
		return fmt.Errorf("element not found: %v", queries[0].Texts)
	}
	return fmt.Errorf("element not found in queries")
}

func (e *Executor) readSnap(observe state.ObserveFn) ui.Snapshot {
	if observe != nil {
		snap, _, _ := observe()
		return snap
	}
	return e.Session.ReadUI(true)
}

func (e *Executor) waitTapTexts(texts []string, prefer string, minY, maxY int, timeoutSec float64) error {
	q := ui.FindQuery{
		Texts: texts, Prefer: prefer, MinCenterY: minY, MaxCenterY: maxY, PreferClickable: true,
	}
	return e.pollTap([]ui.FindQuery{q}, timeoutSec)
}

func (e *Executor) tapFirstQuery(queries []ui.FindQuery, timeoutSec float64) error {
	return e.pollTap(queries, timeoutSec)
}

func (e *Executor) tapFirstQueryObserve(
	observe state.ObserveFn,
	invalidate func(),
	queries []ui.FindQuery,
	timeoutSec float64,
) error {
	return e.pollTapObserve(observe, invalidate, queries, timeoutSec)
}
