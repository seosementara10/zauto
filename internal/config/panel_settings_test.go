package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSavePanelSettingsPreservesTasks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	seed := `{
  "max_devices": 2,
  "parallel_workers": 2,
  "tasks": [{"name": "facebook_login", "app": "facebook", "flow": "facebook_login_logout"}],
  "database": {"url": "postgres://x", "max_accounts_per_device": 50},
  "engine": {"retry": {"max_attempts": 2, "delay_sec": 0.5}}
}`
	if err := os.WriteFile(path, []byte(seed), 0644); err != nil {
		t.Fatal(err)
	}
	p := PanelSettings{MaxDevices: 3, ParallelWorkers: 3, DelayBetweenActionsSec: 0.1, AutomationDriver: "adb"}
	if err := SavePanelSettings(path, p); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["max_devices"].(float64) != 3 {
		t.Fatalf("max_devices=%v", raw["max_devices"])
	}
	tasks, ok := raw["tasks"].([]interface{})
	if !ok || len(tasks) != 1 {
		t.Fatalf("tasks preserved: %v", raw["tasks"])
	}
}
