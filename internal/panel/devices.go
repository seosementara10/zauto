package panel

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"zauto/internal/adb"
)

func (s *Server) refreshDevicesWithRetry(n int) {
	for i := 0; i < n; i++ {
		if s.refreshDevices() {
			return
		}
		if i+1 < n {
			time.Sleep(800 * time.Millisecond)
		}
	}
}

// refreshDevices scans USB via ADB. Returns true when scan succeeded.
func (s *Server) refreshDevices() bool {
	s.connectConfiguredEmulators()
	all, err := adb.ListDevices()
	s.mu.Lock()
	if err != nil {
		s.adbLastError = err.Error()
		s.mu.Unlock()
		log.Printf("panel: adb devices gagal: %v", err)
		return false
	}
	s.adbLastError = ""
	s.mu.Unlock()
	s.refreshDevicesFrom(all)
	return true
}

func (s *Server) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Serial string `json:"serial"`
		Label  string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Serial == "" {
		http.Error(w, "serial wajib diisi", http.StatusBadRequest)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	label := body.Label
	if label == "" {
		label = body.Serial
	}
	wf, _ := s.loadWorkflow()
	maxAcc := 50
	if wf != nil && wf.Database.MaxAccountsPerDevice > 0 {
		maxAcc = wf.Database.MaxAccountsPerDevice
	}
	if _, err := s.Store.UpsertDevice(r.Context(), body.Serial, label, 0, maxAcc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.refreshDevices()
	s.broadcastState()
	s.writeState(w)
}

func (s *Server) registeredSerials() map[string]string {
	out := map[string]string{}
	if s.Store == nil {
		return out
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.Store.ListRegisteredDevices(ctx)
	if err != nil {
		return out
	}
	for _, d := range rows {
		out[d.Serial] = d.Label
	}
	return out
}

// prepareConnectedEmulators runs one-time adb hardening when an emulator appears.
func (s *Server) prepareConnectedEmulators(connected map[string]bool) {
	s.emulatorPrepMu.Lock()
	defer s.emulatorPrepMu.Unlock()

	for serial := range connected {
		if !s.emulatorPrepared[serial] {
			client := &adb.Client{Serial: serial, Timeout: 15 * time.Second, Retries: 1}
			if !client.IsEmulator() {
				continue
			}
			res := adb.PrepareEmulator(client)
			adb.LogEmulatorPrep(res)
			s.emulatorPrepared[serial] = true
		}
	}
	for serial := range s.emulatorPrepared {
		if !connected[serial] {
			delete(s.emulatorPrepared, serial)
		}
	}
}
