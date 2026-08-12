package reset

import (
	"testing"

	"zauto/internal/state"
)

func TestFreshStateRecognition(t *testing.T) {
	for _, s := range state.FreshStates() {
		if !state.IsFreshState(s) {
			t.Fatalf("%s should be fresh", s)
		}
		if !state.IsPostResetValidState(s) {
			t.Fatalf("%s should be post-reset valid", s)
		}
	}
	if state.IsFreshState(state.UILoggedIn) {
		t.Fatal("logged_in is not fresh")
	}
}

func TestPostResetOverlayStatesValid(t *testing.T) {
	for _, s := range []state.UIState{
		state.UIPasswordManagerSheet, state.UISaveLoginPrompt, state.UILocaleSetupError,
	} {
		if !state.IsPostResetValidState(s) {
			t.Fatalf("%s should be valid after reset", s)
		}
		if state.IsFreshState(s) {
			t.Fatalf("%s overlay should not be core fresh", s)
		}
	}
}
