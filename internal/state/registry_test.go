package state

import "testing"

func TestRegistrySavedProfileInLoginFlow(t *testing.T) {
	def, ok := DefaultRegistry.Def(UISavedProfileScreen)
	if !ok {
		t.Fatal("saved_profile_screen missing from registry")
	}
	if !def.LoginFlow {
		t.Fatal("saved_profile must be in login flow watch")
	}
	if !def.PostResetOK {
		t.Fatal("saved_profile must be valid after reset")
	}
	if !def.IsOverlay || !def.CanBlock {
		t.Fatal("saved_profile must be blocking overlay")
	}
}

func TestPostResetValidStatesIncludesSavedProfile(t *testing.T) {
	valid := PostResetValidStates()
	found := false
	for _, s := range valid {
		if s == UISavedProfileScreen {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PostResetValidStates missing saved_profile_screen: %v", valid)
	}
}

func TestLoginWatchIncludesAllBlockingOverlays(t *testing.T) {
	watch := LoginFlowWatchStates()
	watchSet := map[UIState]bool{}
	for _, s := range watch {
		watchSet[s] = true
	}
	for _, s := range DefaultRegistry.BlockingOverlays() {
		if !watchSet[s] {
			t.Fatalf("LoginFlowWatchStates missing overlay %s", s)
		}
	}
}
