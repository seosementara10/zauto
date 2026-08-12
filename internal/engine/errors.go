package engine

import (
	"zauto/internal/engine/overlay"
	"zauto/internal/state"
)

func errUnhandledState(d state.Detection) error {
	return &unhandledStateError{state: d.State, confidence: d.Confidence}
}

type unhandledStateError struct {
	state      state.UIState
	confidence float64
}

func (e *unhandledStateError) Error() string {
	return "engine has no handler for detected state=" + string(e.state)
}

func (e *Executor) withPermissionPolicy(policy string, fn func() error) error {
	e.Session.Runtime[overlay.PermissionPolicyRuntimeKey] = policy
	defer delete(e.Session.Runtime, overlay.PermissionPolicyRuntimeKey)
	return fn()
}
