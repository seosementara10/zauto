package overlay

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

type pollDismissOpts struct {
	avoid       state.UIState
	timeout     time.Duration
	act         func() error
	doneEvent   string
	missEvent   string
	successEvent string
	errLabel    string
}

func pollUntilDismissed(e runtime.Exec, det *state.Detector, observe state.ObserveFn, opts pollDismissOpts) error {
	deadline := time.Now().Add(opts.timeout)
	for time.Now().Before(deadline) {
		snap, pkg, actName := observe()
		d := det.Detect(snap, pkg, actName)
		if d.State != opts.avoid || d.Confidence < state.VerifyMinConfidence {
			e.Event(opts.doneEvent)
			return nil
		}
		if err := opts.act(); err != nil {
			e.Event(opts.missEvent, err)
		} else if opts.successEvent != "" {
			e.Event(opts.successEvent)
		}
		time.Sleep(e.Sess().PollInterval())
	}
	return fmt.Errorf("%s: not dismissed within timeout", opts.errLabel)
}
