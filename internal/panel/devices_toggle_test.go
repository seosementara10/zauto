package panel

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToggleDevicePersistsAcrossRefresh(t *testing.T) {
	srv := NewServer(t.TempDir(), "", 8765)
	serial := "test-serial-abc123"

	srv.mu.Lock()
	srv.devices = []DeviceInfo{{Serial: serial, Connected: true, Status: "idle"}}
	srv.enabled[serial] = false
	srv.mu.Unlock()

	body, _ := json.Marshal(map[string]interface{}{"serial": serial, "enabled": true})
	rec := httptest.NewRecorder()
	srv.handleToggleDevice(rec, httptest.NewRequest(http.MethodPost, "/api/devices/toggle", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("toggle status=%d body=%s", rec.Code, rec.Body.String())
	}

	srv.mu.RLock()
	on := srv.enabled[serial]
	srv.mu.RUnlock()
	if !on {
		t.Fatal("enabled map not updated after toggle")
	}

	srv.refreshDevicesFrom([]string{serial})

	srv.mu.RLock()
	on = srv.enabled[serial]
	devEnabled := false
	for _, d := range srv.devices {
		if d.Serial == serial {
			devEnabled = d.Enabled
			break
		}
	}
	srv.mu.RUnlock()
	if !on || !devEnabled {
		t.Fatalf("toggle lost after refreshDevicesFrom: map=%v device.Enabled=%v", on, devEnabled)
	}
}

func TestDisableAllClearsEnabled(t *testing.T) {
	srv := NewServer(t.TempDir(), "", 8765)
	serial := "test-serial-xyz"

	srv.mu.Lock()
	srv.devices = []DeviceInfo{{Serial: serial, Connected: true}}
	srv.enabled[serial] = true
	srv.mu.Unlock()

	rec := httptest.NewRecorder()
	srv.handleDisableAll(rec, httptest.NewRequest(http.MethodPost, "/api/devices/disable-all", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("disable-all status=%d", rec.Code)
	}

	srv.mu.RLock()
	on := srv.enabled[serial]
	srv.mu.RUnlock()
	if on {
		t.Fatal("expected disabled after disable-all")
	}
}
