package state

// StateDefinition describes one classified UI state and how flows should treat it.
// Handlers are bound in internal/engine — metadata lives here so every flow shares one catalog.
type StateDefinition struct {
	State       UIState
	Priority    int
	IsOverlay   bool
	CanBlock    bool
	PostResetOK bool
	LoginFlow   bool
}

// Registry is the central catalog of Facebook UI states.
type Registry struct {
	defs map[UIState]StateDefinition
}

// DefaultRegistry is the Facebook katana state catalog.
var DefaultRegistry = newFacebookRegistry()

func newFacebookRegistry() *Registry {
	defs := []StateDefinition{
		{State: UIPermission, Priority: 100, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UIPasswordManagerSheet, Priority: 90, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UIKeyboardSettings, Priority: 88, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UILocaleSetupError, Priority: 85, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UISaveLoginPrompt, Priority: 80, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UILogoutConfirmPrompt, Priority: 79, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UIContactFollowPrompt, Priority: 78, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UILocationServicesPrompt, Priority: 77, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UISavedProfileScreen, Priority: 76, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UILoginAccountFinderPrompt, Priority: 75, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UIFanpageHomeIntro, Priority: 74, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UIPostPromotePrompt, Priority: 73, IsOverlay: true, CanBlock: true, PostResetOK: true, LoginFlow: true},
		{State: UIError, Priority: 70, CanBlock: true, LoginFlow: true},
		{State: UI2FACheckpoint, Priority: 65, CanBlock: true, LoginFlow: true},
		{State: UILoading, Priority: 50, LoginFlow: true, PostResetOK: true},
		{State: UIOnboarding, Priority: 40, CanBlock: true, LoginFlow: true, PostResetOK: true},
		{State: UILogin, Priority: 30, CanBlock: true, LoginFlow: true, PostResetOK: true},
		{State: UILoggedIn, Priority: 10, LoginFlow: true},
		{State: UIUnknown, Priority: 0},
	}
	m := make(map[UIState]StateDefinition, len(defs))
	for _, d := range defs {
		m[d.State] = d
	}
	return &Registry{defs: m}
}

func (r *Registry) Def(s UIState) (StateDefinition, bool) {
	d, ok := r.defs[s]
	return d, ok
}

func (r *Registry) Priority(s UIState) int {
	if d, ok := r.defs[s]; ok {
		return d.Priority
	}
	return 0
}

func (r *Registry) IsOverlay(s UIState) bool {
	if d, ok := r.defs[s]; ok {
		return d.IsOverlay
	}
	return false
}

func (r *Registry) All() []StateDefinition {
	out := make([]StateDefinition, 0, len(r.defs))
	for _, d := range r.defs {
		out = append(out, d)
	}
	return out
}

func (r *Registry) LoginWatchStates() []UIState {
	return r.statesWhere(func(d StateDefinition) bool { return d.LoginFlow })
}

func (r *Registry) PostResetValidStates() []UIState {
	return r.statesWhere(func(d StateDefinition) bool { return d.PostResetOK })
}

func (r *Registry) BlockingOverlays() []UIState {
	return r.statesWhere(func(d StateDefinition) bool { return d.IsOverlay && d.CanBlock })
}

func (r *Registry) statesWhere(pred func(StateDefinition) bool) []UIState {
	var out []UIState
	for _, d := range r.All() {
		if pred(d) {
			out = append(out, d.State)
		}
	}
	return out
}
