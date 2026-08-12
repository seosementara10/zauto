package state

var freshUIStates = []UIState{UILogin, UIOnboarding, UILoading, UIPermission}

// FreshStates returns core app states right after pm clear (login path).
func FreshStates() []UIState {
	return append([]UIState(nil), freshUIStates...)
}

// PostResetValidStates includes fresh states plus overlays valid after pm clear + launch.
func PostResetValidStates() []UIState {
	return DefaultRegistry.PostResetValidStates()
}

// IsFreshState reports core app UI states (login/onboarding/loading/permission).
func IsFreshState(s UIState) bool {
	for _, t := range freshUIStates {
		if s == t {
			return true
		}
	}
	return false
}

// IsPostResetValidState reports any acceptable UI after pm clear succeeded.
func IsPostResetValidState(s UIState) bool {
	return IsFreshState(s) || IsOverlayState(s)
}
