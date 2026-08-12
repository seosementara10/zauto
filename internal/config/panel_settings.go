package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PanelSettings are user-editable options exposed in zauto Panel (persisted to config.json).
type PanelSettings struct {
	MaxDevices             int     `json:"max_devices"`
	ParallelWorkers        int     `json:"parallel_workers"`
	LoopCount              int     `json:"loop_count"`
	DelayBetweenLoopsSec   float64 `json:"delay_between_loops_sec"`
	DelayBetweenActionsSec float64 `json:"delay_between_actions_sec"`
	DelayAfterForceStopSec float64 `json:"delay_after_force_stop_sec"`
	PostActionDelaySec     float64 `json:"post_action_delay_sec"`
	PollSec                float64 `json:"poll_sec"`
	AppLaunchWaitSec       float64 `json:"app_launch_wait_sec"`
	RetryMaxAttempts       int     `json:"retry_max_attempts"`
	RetryDelaySec          float64 `json:"retry_delay_sec"`
	AdbTimeoutSec          int     `json:"adb_timeout_sec"`
	AdbRetries             int     `json:"adb_retries"`
	AutomationDriver       string  `json:"automation_driver"`
	WakeScreenBeforeTask   bool    `json:"wake_screen_before_task"`
	ForceStopBeforeOpen    bool    `json:"force_stop_before_open"`
	ClearAppBeforeOpen     bool    `json:"clear_app_before_open"`
	ForceStopAfterTask     bool    `json:"force_stop_after_task"`
	ScreenshotOnError      bool    `json:"screenshot_on_error"`
	MirrorOnRun            bool    `json:"mirror_on_run"`
}

func PanelSettingsFromWorkflow(wf *Workflow) PanelSettings {
	if wf == nil {
		return PanelSettings{MaxDevices: 1, ParallelWorkers: 1, LoopCount: 1, AutomationDriver: "adb"}
	}
	return PanelSettings{
		MaxDevices:             wf.MaxDevices,
		ParallelWorkers:        wf.ParallelWorkers,
		LoopCount:              wf.LoopCount,
		DelayBetweenLoopsSec:   wf.DelayBetweenLoopsSec,
		DelayBetweenActionsSec: wf.DelayBetweenActionsSec,
		DelayAfterForceStopSec: wf.DelayAfterForceStopSec,
		PostActionDelaySec:     wf.Engine.PostActionDelay,
		PollSec:                wf.Engine.PollSec,
		AppLaunchWaitSec:       wf.Engine.AppLaunchWaitSec,
		RetryMaxAttempts:       wf.Engine.RetryMaxAttempts,
		RetryDelaySec:          wf.Engine.RetryDelaySec,
		AdbTimeoutSec:          wf.AdbTimeoutSec,
		AdbRetries:             wf.AdbRetries,
		AutomationDriver:       wf.Automation.Driver,
		WakeScreenBeforeTask:   wf.WakeScreenBeforeTask,
		ForceStopBeforeOpen:    wf.ForceStopBeforeOpen,
		ClearAppBeforeOpen:     wf.ClearAppBeforeOpen,
		ForceStopAfterTask:     wf.ForceStopAfterTask,
		ScreenshotOnError:      wf.ScreenshotOnError,
		MirrorOnRun:            wf.MirrorOnRun,
	}
}

func (p *PanelSettings) Normalize() {
	if p.MaxDevices <= 0 {
		p.MaxDevices = 1
	}
	if p.ParallelWorkers <= 0 {
		p.ParallelWorkers = p.MaxDevices
	}
	if p.LoopCount <= 0 {
		p.LoopCount = 1
	}
	if p.RetryMaxAttempts <= 0 {
		p.RetryMaxAttempts = 2
	}
	if p.AdbTimeoutSec <= 0 {
		p.AdbTimeoutSec = 30
	}
	if p.AdbRetries <= 0 {
		p.AdbRetries = 2
	}
	if p.AutomationDriver != "appium" {
		p.AutomationDriver = "adb"
	}
	if p.DelayBetweenActionsSec < 0 {
		p.DelayBetweenActionsSec = 0
	}
	if p.PostActionDelaySec < 0 {
		p.PostActionDelaySec = 0
	}
	if p.RetryDelaySec < 0 {
		p.RetryDelaySec = 0
	}
}

func (p PanelSettings) Validate() error {
	p.Normalize()
	if p.MaxDevices > 100 {
		return fmt.Errorf("max_devices max 100")
	}
	if p.ParallelWorkers > 100 {
		return fmt.Errorf("parallel_workers max 100")
	}
	return nil
}

func (p PanelSettings) applyToMap(raw map[string]interface{}) {
	raw["max_devices"] = p.MaxDevices
	raw["parallel_workers"] = p.ParallelWorkers
	raw["loop_count"] = p.LoopCount
	raw["delay_between_loops_sec"] = p.DelayBetweenLoopsSec
	raw["delay_between_actions_sec"] = p.DelayBetweenActionsSec
	raw["delay_after_force_stop_sec"] = p.DelayAfterForceStopSec
	raw["wake_screen_before_task"] = p.WakeScreenBeforeTask
	raw["force_stop_before_open"] = p.ForceStopBeforeOpen
	raw["clear_app_before_open"] = p.ClearAppBeforeOpen
	raw["force_stop_after_task"] = p.ForceStopAfterTask
	raw["screenshot_on_error"] = p.ScreenshotOnError
	raw["mirror_on_run"] = p.MirrorOnRun
	raw["adb_timeout_sec"] = p.AdbTimeoutSec
	raw["adb_retries"] = p.AdbRetries

	engine, _ := raw["engine"].(map[string]interface{})
	if engine == nil {
		engine = map[string]interface{}{}
		raw["engine"] = engine
	}
	engine["post_action_delay_sec"] = p.PostActionDelaySec
	engine["poll_sec"] = p.PollSec
	engine["app_launch_wait_sec"] = p.AppLaunchWaitSec
	retry, _ := engine["retry"].(map[string]interface{})
	if retry == nil {
		retry = map[string]interface{}{}
		engine["retry"] = retry
	}
	retry["max_attempts"] = p.RetryMaxAttempts
	retry["delay_sec"] = p.RetryDelaySec

	automation, _ := raw["automation"].(map[string]interface{})
	if automation == nil {
		automation = map[string]interface{}{}
		raw["automation"] = automation
	}
	automation["driver"] = p.AutomationDriver
}

// SavePanelSettings merges settings into config.json (preserves tasks, database, etc.).
func SavePanelSettings(path string, p PanelSettings) error {
	if err := p.Validate(); err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.applyToMap(raw)
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadPanelSettings reads config.json and returns panel-editable fields.
func LoadPanelSettings(path string) (PanelSettings, error) {
	wf, err := Load(path)
	if err != nil {
		return PanelSettings{}, err
	}
	s := PanelSettingsFromWorkflow(wf)
	s.Normalize()
	return s, nil
}

// ConfigDir returns the directory containing the config file.
func ConfigDir(path string) string {
	return filepath.Dir(path)
}
