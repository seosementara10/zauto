package state

import "testing"

func TestPostResetValidStatesIncludesOverlays(t *testing.T) {
	valid := PostResetValidStates()
	want := map[UIState]bool{
		UILogin: true, UIOnboarding: true, UILoading: true, UIPermission: true,
		UIPasswordManagerSheet: true, UISaveLoginPrompt: true, UILocaleSetupError: true,
		UIContactFollowPrompt: true, UISavedProfileScreen: true,
	}
	for s := range want {
		if !containsState(valid, s) {
			t.Fatalf("PostResetValidStates missing %s: %v", s, valid)
		}
	}
}

func containsState(list []UIState, s UIState) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
