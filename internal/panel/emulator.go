package panel

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
)

func (s *Server) ldplayerSummary() map[string]interface{} {
	wf, err := s.loadWorkflow()
	if err != nil || wf == nil {
		return map[string]interface{}{"available": false}
	}
	mgr := adb.NewLDPlayerManager(wf.Emulator.InstallPath)
	if _, err := mgr.ListInstances(nil); err != nil {
		return map[string]interface{}{
			"available":    false,
			"install_path": wf.Emulator.InstallPath,
			"error":        err.Error(),
		}
	}
	connected := map[string]bool{}
	if all, err := adb.ListDevices(); err == nil {
		for _, serial := range all {
			connected[serial] = true
		}
	}
	instances, err := mgr.ListInstances(connected)
	if err != nil {
		return map[string]interface{}{
			"available":    false,
			"install_path": wf.Emulator.InstallPath,
			"error":        err.Error(),
		}
	}
	return map[string]interface{}{
		"available":      true,
		"install_path":   wf.Emulator.InstallPath,
		"auto_connect":   wf.Emulator.AutoConnect,
		"instance_count": wf.Emulator.InstanceCount,
		"instances":      instances,
	}
}

func (s *Server) handleEmulatorLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Index *int `json:"index"`
		All   bool `json:"all"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mgr := adb.NewLDPlayerManager(wf.Emulator.InstallPath)
	if body.All {
		err = mgr.LaunchAll(wf.Emulator.InstanceCount)
	} else {
		idx := 0
		if body.Index != nil {
			idx = *body.Index
		}
		err = mgr.Launch(idx)
	}
	if err != nil {
		log.Printf("panel: ldplayer launch: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	time.Sleep(2 * time.Second)
	s.connectConfiguredEmulators()
	s.refreshDevicesWithRetry(8)
	s.broadcastState()
	s.writeState(w)
}

func (s *Server) handleEmulatorAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name      string `json:"name"`
		CloneFrom *int   `json:"clone_from"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cloneFrom := 0
	if body.CloneFrom != nil {
		cloneFrom = *body.CloneFrom
	}
	mgr := adb.NewLDPlayerManager(wf.Emulator.InstallPath)
	newIndex, err := mgr.AddInstance(body.Name, cloneFrom)
	if err != nil {
		log.Printf("panel: ldplayer add: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	em := wf.Emulator
	if newIndex+1 > em.InstanceCount {
		em.InstanceCount = newIndex + 1
		if err := config.SaveEmulatorConfig(s.ConfigPath, em); err != nil {
			log.Printf("panel: save emulator config: %v", err)
		} else {
			s.invalidateWorkflowCache()
		}
	}
	s.connectConfiguredEmulators()
	s.refreshDevicesWithRetry(4)
	s.broadcastState()
	s.writeState(w)
}

func (s *Server) handleEmulatorQuit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Index int `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mgr := adb.NewLDPlayerManager(wf.Emulator.InstallPath)
	if err := mgr.Quit(body.Index); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	s.refreshDevices()
	s.broadcastState()
	s.writeState(w)
}
