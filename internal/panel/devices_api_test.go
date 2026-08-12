package panel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Verifies /api/state exposes a devices array after ADB scan (needs adb + optional USB).
func TestHandleStateIncludesDevices(t *testing.T) {
	root := `C:\Users\WINDOWS 11\Desktop\z`
	cfg := root + `\config\config.json`
	srv := NewServer(root, cfg, 8765)
	if !srv.refreshDevices() {
		t.Skip("adb scan failed — skip live device test")
	}
	rec := httptest.NewRecorder()
	srv.handleState(rec, httptest.NewRequest(http.MethodGet, "/api/state", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	devs, ok := payload["devices"].([]interface{})
	if !ok {
		t.Fatalf("devices field missing or wrong type: %T", payload["devices"])
	}
	t.Logf("api/state devices: %d", len(devs))
	if payload["run_status"] == nil {
		t.Fatal("run_status missing from /api/state")
	}
}
