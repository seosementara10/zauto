package panel

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/store"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	health := s.healthLocked()
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if ok, _ := health["ok"].(bool); !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(health)
}

type accountsImportRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (s *Server) handleAccountsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		ProfileID string `json:"profile_id"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := s.Store.CreateAccount(r.Context(), body.Name, body.Email, body.ProfileID, body.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.broadcastState()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

func (s *Server) handleAccountsImportPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req accountsImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	content := req.Content
	if content == "" && req.Path != "" {
		path := req.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(s.ProjectRoot, path)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		content = string(raw)
	}
	if content == "" {
		http.Error(w, "content or path required", http.StatusBadRequest)
		return
	}
	rows, err := store.PreviewImportContent(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	valid := 0
	for _, row := range rows {
		if row.Valid {
			valid++
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"rows":  rows,
		"total": len(rows),
		"valid": valid,
	})
}

func (s *Server) handleAccountsImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	path := defaultAccountsFile
	content := ""
	if r.Body != nil {
		var req accountsImportRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Content != "" {
			content = req.Content
		} else if req.Path != "" {
			path = req.Path
		}
	}
	var n int
	var err error
	if content != "" {
		n, err = s.Store.ImportAccountsContent(r.Context(), content)
	} else {
		if !filepath.IsAbs(path) {
			path = filepath.Join(s.ProjectRoot, path)
		}
		n, err = s.Store.ImportAccountsFile(r.Context(), path)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.broadcastState()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": n,
		"path":     path,
	})
}

func (s *Server) handleAccountsAutoAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serials, err := adb.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(serials) == 0 {
		http.Error(w, "tidak ada HP terhubung — colok USB lalu refresh", http.StatusBadRequest)
		return
	}
	n, err := s.Store.AutoAssignDevices(r.Context(), serials, wf.Database.MaxAccountsPerDevice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.refreshDevicesFrom(serials)
	s.broadcastState()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"assigned": n,
		"devices":  len(serials),
	})
}

func (s *Server) handleAccountsAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Serial    string `json:"serial"`
		AccountID int64  `json:"account_id"`
		SlotNo    int    `json:"slot_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Serial == "" || body.AccountID <= 0 {
		http.Error(w, "invalid serial or account_id", http.StatusBadRequest)
		return
	}
	if body.SlotNo <= 0 {
		body.SlotNo = 1
	}
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maxAcc := 50
	if wf != nil && wf.Database.MaxAccountsPerDevice > 0 {
		maxAcc = wf.Database.MaxAccountsPerDevice
	}
	if _, err := s.Store.UpsertDevice(r.Context(), body.Serial, body.Serial, 0, maxAcc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.Store.AssignAccount(r.Context(), body.Serial, body.AccountID, body.SlotNo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if all, err := adb.ListDevices(); err == nil {
		s.refreshDevicesFrom(all)
	}
	s.broadcastState()
	s.writeState(w)
}

func (s *Server) handleAccountsUnassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Serial    string `json:"serial"`
		AccountID int64  `json:"account_id"`
		SlotNo    int    `json:"slot_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var err error
	if body.AccountID > 0 {
		err = s.Store.UnassignAccount(r.Context(), body.AccountID)
	} else if body.Serial != "" {
		err = s.Store.UnassignDevice(r.Context(), body.Serial, body.SlotNo)
	} else {
		http.Error(w, "serial or account_id required", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if all, err := adb.ListDevices(); err == nil {
		s.refreshDevicesFrom(all)
	}
	s.broadcastState()
	s.writeState(w)
}

func (s *Server) handleAccountsAutomation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		AccountID int64    `json:"account_id"`
		Flow      string   `json:"flow"`
		Steps     []string `json:"steps"`
		Enabled   *bool    `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccountID <= 0 {
		http.Error(w, "invalid account_id", http.StatusBadRequest)
		return
	}
	flow := body.Flow
	params := map[string]interface{}{}
	if len(body.Steps) > 0 {
		steps, err := config.NormalizePipelineSteps(body.Steps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		flow = "facebook_pipeline"
		params = config.DefaultPipelineParams(steps)
	} else {
		if flow == "" {
			flow = "facebook_login_logout"
		}
		for _, opt := range config.AvailableFlows() {
			if opt.Flow == flow {
				for k, v := range opt.Params {
					params[k] = v
				}
				break
			}
		}
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	if err := s.Store.SetAccountAutomation(r.Context(), body.AccountID, flow, params, enabled); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.broadcastState()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

func (s *Server) handleFlowsCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	opts := config.AvailableFlows()
	out := make([]map[string]interface{}, 0, len(opts)+1)
	out = append(out, map[string]interface{}{
		"flow":        "facebook_pipeline",
		"label":       "Pipeline kustom (centang per akun)",
		"steps":       config.PipelineStepLabels(config.AllPipelineSteps),
		"pipeline":    config.AllPipelineSteps,
		"description": "Centang langkah di halaman Akun — urutan: PM Clear → Login → Beranda → Fanpage → Logout",
	})
	for _, o := range opts {
		info := config.FlowInfoFor(o.Flow)
		out = append(out, map[string]interface{}{
			"flow":        o.Flow,
			"label":       o.Label,
			"steps":       info.Steps,
			"description": info.Description,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) listAccountsLocked() []map[string]interface{} {
	if s.Store == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.Store.ListAccountSummaries(ctx)
	if err != nil {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, a := range rows {
		steps := a.AutomationSteps
		if len(steps) == 0 {
			steps = config.AccountPipelineSteps(a.AutomationFlow, a.AutomationParams)
		}
		out = append(out, map[string]interface{}{
			"id":                 a.ID,
			"name":               a.Name,
			"login_id":           a.LoginID,
			"assigned_serial":    a.AssignedSerial,
			"slot_no":            a.SlotNo,
			"automation_flow":    a.AutomationFlow,
			"automation_enabled": a.AutomationEnabled,
			"automation_steps":   config.PipelineStepLabels(steps),
			"pipeline_steps":     steps,
			"fanpage_count":      a.FanpageCount,
			"fanpages":           fanpagesJSON(a.Fanpages),
		})
	}
	return out
}

func fanpagesJSON(fps []store.Fanpage) []map[string]string {
	if len(fps) == 0 {
		return nil
	}
	out := make([]map[string]string, 0, len(fps))
	for _, fp := range fps {
		out = append(out, map[string]string{
			"name":       fp.Name,
			"fb_page_id": fp.FBPageID,
		})
	}
	return out
}
