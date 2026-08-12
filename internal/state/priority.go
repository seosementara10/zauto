package state

// Priority returns overlay/screen priority (higher wins on tie).
func Priority(s UIState) int {
	return DefaultRegistry.Priority(s)
}

// LoginFlowWatchStates is the full set of states the login ODAV loop may observe.
func LoginFlowWatchStates() []UIState {
	return DefaultRegistry.LoginWatchStates()
}

// IsOverlayState reports transient dialogs/sheets that block the target screen.
func IsOverlayState(s UIState) bool {
	return DefaultRegistry.IsOverlay(s)
}
