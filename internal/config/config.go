package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ScrollCoords struct {
	X1          int `json:"x1"`
	Y1          int `json:"y1"`
	X2          int `json:"x2"`
	Y2          int `json:"y2"`
	DurationMs  int `json:"duration_ms"`
}

type VerifySpec struct {
	Texts      []string `json:"texts"`
	TimeoutSec float64  `json:"timeout_sec"`
	PollSec    float64  `json:"poll_sec"`
}

type Action struct {
	Type     string                 `json:"type"`
	Optional bool                   `json:"optional"`
	Texts    []string               `json:"texts"`
	Verify   *VerifySpec            `json:"verify"`
	Extra    map[string]interface{} `json:"-"`
	raw      map[string]interface{}
}

func (a *Action) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &a.raw); err != nil {
		return err
	}
	a.Type, _ = a.raw["type"].(string)
	if v, ok := a.raw["optional"].(bool); ok {
		a.Optional = v
	}
	if texts, ok := a.raw["texts"].([]interface{}); ok {
		for _, t := range texts {
			if s, ok := t.(string); ok {
				a.Texts = append(a.Texts, s)
			}
		}
	}
	if v, ok := a.raw["verify"].(map[string]interface{}); ok {
		a.Verify = &VerifySpec{}
		if texts, ok := v["texts"].([]interface{}); ok {
			for _, t := range texts {
				if s, ok := t.(string); ok {
					a.Verify.Texts = append(a.Verify.Texts, s)
				}
			}
		}
		if f, ok := v["timeout_sec"].(float64); ok {
			a.Verify.TimeoutSec = f
		}
		if f, ok := v["poll_sec"].(float64); ok {
			a.Verify.PollSec = f
		}
	}
	a.Extra = a.raw
	return nil
}

func (a *Action) ParamString(key, def string) string {
	if v, ok := a.Extra[key].(string); ok {
		return v
	}
	return def
}

func (a *Action) ParamFloat(key string, def float64) float64 {
	if v, ok := a.Extra[key].(float64); ok {
		return v
	}
	return def
}

func (a *Action) ParamInt(key string, def int) int {
	if v, ok := a.Extra[key].(float64); ok {
		return int(v)
	}
	return def
}

func (a *Action) ParamBool(key string, def bool) bool {
	if v, ok := a.Extra[key].(bool); ok {
		return v
	}
	return def
}

type Task struct {
	Name        string                 `json:"name"`
	App         string                 `json:"app"`
	Flow        string                 `json:"flow"`
	Description string                 `json:"description,omitempty"`
	Params      map[string]interface{} `json:"params"`
	Actions     []Action               `json:"actions"`
}

type EngineConfig struct {
	RetryMaxAttempts int     `json:"-"`
	RetryDelaySec    float64 `json:"-"`
	AppLaunchWaitSec float64 `json:"app_launch_wait_sec"`
	PostActionDelay  float64 `json:"post_action_delay_sec"`
	PollSec          float64 `json:"poll_sec"`
}

// AutomationConfig selects UI backend: "adb" (dump XML) or "appium" (UiAutomator2).
type AutomationConfig struct {
	Driver string       `json:"driver"`
	Appium AppiumConfig `json:"appium"`
}

type AppiumConfig struct {
	ServerURL            string `json:"server_url"`
	AutomationName       string `json:"automation_name"`
	PortBase             int    `json:"port_base"`
	PortStride           int    `json:"port_stride"`
	PortPerDevice        bool   `json:"port_per_device"`
	NoReset              bool   `json:"no_reset"`
	AutoGrantPermissions bool   `json:"auto_grant_permissions"`
}

// EmulatorConfig controls automatic LDPlayer / local emulator ADB connection from panel.
type EmulatorConfig struct {
	AutoConnect   bool   `json:"auto_connect"`
	InstanceCount int    `json:"instance_count"`
	InstallPath   string `json:"install_path"`
}

func (e *EmulatorConfig) normalize(raw map[string]interface{}) {
	if e.InstanceCount <= 0 {
		e.InstanceCount = 10
	}
	if strings.TrimSpace(e.InstallPath) == "" {
		e.InstallPath = `C:\LDPlayer\LDPlayer14`
	}
	if em, ok := raw["emulator"].(map[string]interface{}); ok {
		if _, set := em["auto_connect"]; !set {
			e.AutoConnect = true
		}
	} else {
		e.AutoConnect = true
	}
}

// DatabaseConfig holds PostgreSQL settings for accounts, devices, and assignments.
type DatabaseConfig struct {
	URL                  string `json:"url"`
	MaxAccountsPerDevice int    `json:"max_accounts_per_device"`
}

type Workflow struct {
	Mode                     string                            `json:"mode"`
	MaxDevices               int                               `json:"max_devices"`
	ParallelWorkers          int                               `json:"parallel_workers"`
	LoopCount                int                               `json:"loop_count"`
	DelayBetweenLoopsSec     float64                           `json:"delay_between_loops_sec"`
	DelayBetweenActionsSec   float64                           `json:"delay_between_actions_sec"`
	AutoDetectUser           bool                              `json:"auto_detect_user"`
	AdbUser                  *int                              `json:"adb_user"`
	WakeScreenBeforeTask       bool                              `json:"wake_screen_before_task"`
	ForceStopBeforeOpen        bool                              `json:"force_stop_before_open"`
	ClearAppBeforeOpen         bool                              `json:"clear_app_before_open"`
	ForceStopAfterTask         bool                              `json:"force_stop_after_task"`
	DelayAfterForceStopSec     float64                           `json:"delay_after_force_stop_sec"`
	ScreenshotOnError          bool                              `json:"screenshot_on_error"`
	MirrorOnRun                bool                              `json:"mirror_on_run"`
	MirrorMode                 string                            `json:"mirror_mode"`
	MirrorPort                 int                               `json:"mirror_port"`
	AdbTimeoutSec              int                               `json:"adb_timeout_sec"`
	AdbRetries                 int                               `json:"adb_retries"`
	DeviceFilter               []string                          `json:"device_filter"`
	ScreenProfile              string                            `json:"screen_profile"`
	ScreenProfiles             map[string]map[string]ScrollCoords `json:"screen_profiles"`
	Apps                       map[string]string                 `json:"apps"`
	AppActivities              map[string]string                 `json:"app_activities"`
	Registration               map[string]interface{}            `json:"registration"`
	Database                   DatabaseConfig                    `json:"database"`
	Engine                     EngineConfig                      `json:"engine"`
	Automation                 AutomationConfig                  `json:"automation"`
	Emulator                   EmulatorConfig                    `json:"emulator"`
	Tasks                      []Task                            `json:"tasks"`
	ProjectRoot                string                            `json:"-"`
}

func Load(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, err
	}
	wf.ProjectRoot = filepath.Dir(filepath.Dir(path))

	if wf.MaxDevices <= 0 {
		wf.MaxDevices = 1
	}
	if wf.ParallelWorkers <= 0 {
		wf.ParallelWorkers = wf.MaxDevices
	}
	if wf.LoopCount <= 0 {
		wf.LoopCount = 1
	}
	if wf.AdbTimeoutSec <= 0 {
		wf.AdbTimeoutSec = 30
	}
	if wf.AdbRetries <= 0 {
		wf.AdbRetries = 2
	}
	if wf.ScreenProfile == "" {
		wf.ScreenProfile = "default"
	}
	if wf.WakeScreenBeforeTask == false && raw["wake_screen_before_task"] == nil {
		wf.WakeScreenBeforeTask = true
	}
	if raw["force_stop_before_open"] == nil {
		wf.ForceStopBeforeOpen = true
	}
	if raw["clear_app_before_open"] == nil {
		wf.ClearAppBeforeOpen = false
	}
	if raw["force_stop_after_task"] == nil {
		wf.ForceStopAfterTask = true
	}
	if raw["mirror_on_run"] == nil {
		wf.MirrorOnRun = true
	}
	if wf.MirrorMode == "" {
		wf.MirrorMode = "dashboard"
	}
	if wf.MirrorPort <= 0 {
		wf.MirrorPort = 8765
	}
	wf.Emulator.normalize(raw)

	if retry, ok := raw["engine"].(map[string]interface{}); ok {
		if r, ok := retry["retry"].(map[string]interface{}); ok {
			if v, ok := r["max_attempts"].(float64); ok {
				wf.Engine.RetryMaxAttempts = int(v)
			}
			if v, ok := r["delay_sec"].(float64); ok {
				wf.Engine.RetryDelaySec = v
			}
		}
	}
	if wf.Engine.RetryMaxAttempts <= 0 {
		wf.Engine.RetryMaxAttempts = 2
	}
	if wf.Automation.Driver == "" {
		wf.Automation.Driver = "adb"
	}
	if wf.Automation.Appium.AutomationName == "" {
		wf.Automation.Appium.AutomationName = "UiAutomator2"
	}
	if wf.Automation.Appium.PortBase <= 0 {
		wf.Automation.Appium.PortBase = 4723
	}
	if wf.Automation.Appium.PortStride <= 0 {
		wf.Automation.Appium.PortStride = 1
	}
	if wf.Database.MaxAccountsPerDevice <= 0 {
		wf.Database.MaxAccountsPerDevice = 50
	}

	for i := range wf.Tasks {
		if wf.Tasks[i].Flow != "" && len(wf.Tasks[i].Actions) == 0 {
			actions, err := ExpandFlow(wf.Tasks[i].Flow, wf.Tasks[i].Params)
			if err != nil {
				return nil, err
			}
			wf.Tasks[i].Actions = actions
		}
	}
	return &wf, nil
}

func (wf *Workflow) RegString(key, def string) string {
	if wf.Registration == nil {
		return def
	}
	if v, ok := wf.Registration[key].(string); ok {
		return v
	}
	return def
}

func (wf *Workflow) RegInt(key string, def int) int {
	if wf.Registration == nil {
		return def
	}
	if v, ok := wf.Registration[key].(float64); ok {
		return int(v)
	}
	return def
}

func (wf *Workflow) RegBool(key string, def bool) bool {
	if wf.Registration == nil {
		return def
	}
	if v, ok := wf.Registration[key].(bool); ok {
		return v
	}
	return def
}

func ExpandFlow(name string, params map[string]interface{}) ([]Action, error) {
	switch name {
	case "facebook_lite_register":
		return FacebookLiteRegisterFlow(params), nil
	case "facebook_login":
		return FacebookLoginFlow(params), nil
	case "facebook_logout":
		return FacebookLogoutFlow(params), nil
	case "facebook_login_logout":
		return FacebookLoginLogoutFlow(params), nil
	case "facebook_auto_post":
		return FacebookAutoPostFlow(params), nil
	case "facebook_login_auto_post":
		return FacebookLoginAutoPostFlow(params), nil
	case "facebook_fanpage_post":
		return FacebookFanpagePostFlow(params), nil
	case "facebook_login_fanpage_post":
		return FacebookLoginFanpagePostFlow(params), nil
	case "facebook_login_auto_post_logout":
		return FacebookLoginAutoPostLogoutFlow(params), nil
	case "facebook_login_fanpage_post_logout":
		return FacebookLoginFanpagePostLogoutFlow(params), nil
	case "facebook_pipeline":
		steps := AccountPipelineSteps(name, params)
		return ExpandPipelineSteps(steps, params)
	default:
		return nil, fmt.Errorf("unknown flow: %s", name)
	}
}
