package state

import "time"

// UIState is the classified screen state for one device (independent per HP).
type UIState string

const (
	UIUnknown                  UIState = "unknown"
	UILoading                  UIState = "loading"
	UILogin                    UIState = "login"
	UIOnboarding               UIState = "onboarding"
	UILoggedIn                 UIState = "logged_in"
	UISaveLoginPrompt          UIState = "save_login_prompt"
	UILoginAccountFinderPrompt UIState = "login_account_finder_prompt"
	UILogoutConfirmPrompt      UIState = "logout_confirm_prompt"
	UIContactFollowPrompt      UIState = "contact_follow_prompt"
	UILocationServicesPrompt   UIState = "location_services_prompt"
	UISavedProfileScreen       UIState = "saved_profile_screen"
	UIPasswordManagerSheet     UIState = "password_manager_sheet"
	UILocaleSetupError         UIState = "locale_setup_error"
	UIPermission               UIState = "permission"
	UIFanpageHomeIntro         UIState = "fanpage_home_intro"
	UIPostPromotePrompt        UIState = "post_promote_prompt"
	UIKeyboardSettings         UIState = "keyboard_settings"
	UI2FACheckpoint            UIState = "2fa_checkpoint"
	UIError                    UIState = "error"
)

// UnknownKind classifies an uncertain screen for recovery routing (sub-state of UIUnknown).
type UnknownKind string

const (
	UnknownKindNone         UnknownKind = ""
	UnknownKindLoading      UnknownKind = "unknown_loading"
	UnknownKindOverlay      UnknownKind = "unknown_overlay"
	UnknownKindPermission   UnknownKind = "unknown_permission"
	UnknownKindAppScreen    UnknownKind = "unknown_app_screen"
	UnknownKindSystemWindow UnknownKind = "unknown_system_window"
	UnknownKindError        UnknownKind = "unknown_error"
)

// AuthPhase tracks post-credential submission (distinct from UIUnknown).
type AuthPhase string

const (
	AuthPhaseNone       AuthPhase = ""
	AuthPhaseWaitResult AuthPhase = "wait_auth_result"
)

const AuthResultTimeout = 25 * time.Second

// TaskState tracks automation job progress (separate from UIState).
type TaskState string

const (
	TaskIdle     TaskState = "idle"
	TaskRunning  TaskState = "running"
	TaskWaiting  TaskState = "waiting"
	TaskSuccess  TaskState = "success"
	TaskFailed   TaskState = "failed"
	TaskRetrying TaskState = "retrying"
)

// Detection is the result of Observe + Detect with evidence scoring.
type Detection struct {
	State       UIState     `json:"state"`
	Confidence  float64     `json:"confidence"`
	Score       float64     `json:"score"`
	Evidence    []string    `json:"evidence"`
	UnknownKind UnknownKind `json:"unknown_kind,omitempty"`
	Package     string      `json:"package,omitempty"`
	Activity    string      `json:"activity,omitempty"`
	At          time.Time   `json:"at"`
}

func (d Detection) CanExecute() bool {
	return d.Confidence >= DefaultDetectionPolicy.ExecuteThreshold
}
func (d Detection) IsUncertain() bool {
	return d.Confidence < DefaultDetectionPolicy.VerifyThreshold
}

// DeviceMemory holds per-device context for Decide/Recovery.
type DeviceMemory struct {
	UIState       UIState     `json:"ui_state"`
	PreviousUI    UIState     `json:"previous_ui_state"`
	TaskState     TaskState   `json:"task_state"`
	LastAction    string      `json:"last_action"`
	RetryCount    int         `json:"retry_count"`
	UnknownCount  int         `json:"unknown_count"`
	UnknownKind   UnknownKind `json:"unknown_kind,omitempty"`
	AuthPhase     AuthPhase   `json:"auth_phase,omitempty"`
	AuthFilledAt  time.Time   `json:"auth_filled_at,omitempty"`
	LastDetection Detection   `json:"last_detection"`
}

func (m *DeviceMemory) SetDetection(d Detection) {
	if m.UIState != d.State && m.UIState != "" {
		m.PreviousUI = m.UIState
	}
	m.UIState = d.State
	if d.UnknownKind != UnknownKindNone {
		m.UnknownKind = d.UnknownKind
	}
	m.LastDetection = d
}

func (m *DeviceMemory) ResetUnknownCount() { m.UnknownCount = 0 }

func (m *DeviceMemory) BeginAuthWait() {
	m.AuthPhase = AuthPhaseWaitResult
	m.AuthFilledAt = time.Now()
}

func (m *DeviceMemory) ClearAuthWait() {
	m.AuthPhase = AuthPhaseNone
	m.AuthFilledAt = time.Time{}
}

// ResetFlow clears transient login/recovery counters at the start of a new flow attempt.
func (m *DeviceMemory) ResetFlow() {
	m.ResetUnknownCount()
	m.ClearAuthWait()
}
