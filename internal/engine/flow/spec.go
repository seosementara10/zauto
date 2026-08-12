package flow

import "zauto/internal/state"

// Spec configures one ODAV loop run. Login and logout each build their own spec.
type Spec struct {
	Watch            []state.UIState
	Goal             Goal
	OnProgress       func(state.Detection) error
	FillCreds        func() error
	ExitOnLoggedIn   bool
	HandleOnboarding bool
	TrackAuthResult  bool
	LogGoalOnSuccess bool
}

func LoginSpec(fillCreds func() error) Spec {
	return Spec{
		Watch:            state.LoginFlowWatchStates(),
		FillCreds:        fillCreds,
		ExitOnLoggedIn:   true,
		HandleOnboarding: true,
		TrackAuthResult:  true,
	}
}

func OverlaySpec(goal Goal, onProgress func(state.Detection) error) Spec {
	return Spec{
		Watch:            state.LoginFlowWatchStates(),
		Goal:             goal,
		OnProgress:       onProgress,
		LogGoalOnSuccess: true,
	}
}
