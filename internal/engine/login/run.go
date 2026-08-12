package login

import (
	"time"

	"zauto/internal/config"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
)

// Run executes the Facebook login ODAV flow.
func Run(e runtime.Exec, action config.Action) error {
	det := state.NewDetector()
	mem := e.Sess().Memory
	if mem == nil {
		mem = &state.DeviceMemory{TaskState: state.TaskRunning, UIState: state.UIUnknown}
		e.Sess().Memory = mem
	} else {
		mem.TaskState = state.TaskRunning
		if mem.LastDetection.State != "" && mem.LastDetection.State != state.UIUnknown {
			e.Event("HANDOFF initial_state=%s confidence=%.0f%%", mem.LastDetection.State, mem.LastDetection.Confidence*100)
		}
	}

	if mem.LastDetection.State == state.UILoggedIn && mem.LastDetection.Confidence >= state.VerifyMinConfidence {
		e.Event("SKIP login — already logged_in confidence=%.0f%%", mem.LastDetection.Confidence*100)
		mem.TaskState = state.TaskSuccess
		return nil
	}

	mem.ResetFlow()

	observe, invalidate := e.CachedObserve()

	e.Event("OBSERVE start package=%s", e.Sess().Client.ForegroundPackage())

	fill := func() error {
		e.Event("ACT fill credentials")
		err := FillFields(e, action)
		if err == nil {
			invalidate()
		}
		return err
	}

	return e.RunStateLoop(
		e.Ctx(),
		det, observe, invalidate, mem,
		90*time.Second,
		fill,
	)
}
