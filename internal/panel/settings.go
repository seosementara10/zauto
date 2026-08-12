package panel

import (
	"encoding/json"
	"net/http"

	"zauto/internal/config"
)

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	settings := config.PanelSettingsFromWorkflow(wf)
	settings.Normalize()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

func (s *Server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body config.PanelSettings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := config.SavePanelSettings(s.ConfigPath, body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.invalidateWorkflowCache()
	s.broadcastState()

	s.mu.RLock()
	runActive := s.runStatus == "running" || s.runStatus == "paused"
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"status": "ok"}
	if runActive {
		resp["run_active"] = true
		resp["message"] = "Run sedang berjalan — perubahan berlaku untuk run berikutnya, bukan run yang sedang jalan."
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetSettings(w, r)
	case http.MethodPost:
		s.handleSaveSettings(w, r)
	default:
		http.Error(w, "GET or POST only", http.StatusMethodNotAllowed)
	}
}
