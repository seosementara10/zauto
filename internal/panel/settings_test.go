package panel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	seed := `{
  "max_devices": 2,
  "parallel_workers": 2,
  "loop_count": 1,
  "delay_between_actions_sec": 0.5,
  "engine": {"post_action_delay_sec": 0.5, "poll_sec": 0, "retry": {"max_attempts": 2, "delay_sec": 0.5}},
  "automation": {"driver": "adb"},
  "tasks": [{"name": "facebook_login", "app": "facebook", "flow": "facebook_login_logout"}],
  "database": {"url": "postgres://x", "max_accounts_per_device": 50},
  "apps": {"facebook": "com.facebook.katana"}
}`
	if err := os.WriteFile(cfgPath, []byte(seed), 0644); err != nil {
		t.Fatal(err)
	}
	srv := NewServer(dir, cfgPath, 8765)

	getReq := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getRec := httptest.NewRecorder()
	srv.handleSettings(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", getRec.Code, getRec.Body.String())
	}

	postBody := `{"max_devices":4,"parallel_workers":3,"delay_between_actions_sec":0.1,"automation_driver":"adb","retry_max_attempts":2}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(postBody))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	srv.handleSettings(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("POST status=%d body=%s", postRec.Code, postRec.Body.String())
	}

	data, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(data), `"max_devices": 4`) {
		t.Fatalf("config not updated: %s", data)
	}

	stateRec := httptest.NewRecorder()
	srv.handleState(stateRec, httptest.NewRequest(http.MethodGet, "/api/state", nil))
	var payload map[string]interface{}
	if err := json.Unmarshal(stateRec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	settings, ok := payload["settings"].(map[string]interface{})
	if !ok || settings["max_devices"].(float64) != 4 {
		t.Fatalf("state.settings=%v", payload["settings"])
	}
}
