package panel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/controller"
	"zauto/internal/monitor"
	"zauto/internal/runcontrol"
	"zauto/internal/store"
)

// Server is the panel API and UI state (served by Wails asset server or headless HTTP).
type Server struct {
	ProjectRoot string
	ConfigPath  string
	Port        int
	http        *http.Server

	mu          sync.RWMutex
	devices     []DeviceInfo
	enabled     map[string]bool
	activeTasks map[string]bool
	runStatus   string // idle, running, paused, done
	lastResults []resultEntry
	runCancel   context.CancelFunc

	adbLastError string

	mirrors         map[string]*exec.Cmd
	mirrorSlot      map[string]int
	mirrorLaunching map[string]bool
	mirrorMu        sync.Mutex
	mirrorSyncMu      sync.Mutex
	mirrorSyncPending bool
	mirrorSyncAgain   bool
	mirrorDebounce    *time.Timer
	stateRev          uint64
	shuttingDown      bool
	mirrorErrors      map[string]string
	watchStop         chan struct{}
	mirrorLaunchWg    sync.WaitGroup
	panel             panelBounds
	desktopMu         sync.RWMutex
	desktopCtx        context.Context

	uiRevMu       sync.RWMutex
	uiRev         uint64
	uiReloadMu    sync.Mutex
	uiReloadTimer *time.Timer

	emulatorPrepMu   sync.Mutex
	emulatorPrepared map[string]bool

	resCache resourceCache
	Store    *store.Store
	handler  http.Handler

	eventMu   sync.Mutex
	eventSubs map[chan []byte]struct{}
}

type DeviceInfo struct {
	Serial       string `json:"serial"`
	Model        string `json:"model"`
	Resolution   string `json:"resolution"`
	Connected    bool   `json:"connected"`
	Enabled      bool   `json:"enabled"`
	MirrorOpen   bool   `json:"mirror_open"`
	MirrorError  string `json:"mirror_error,omitempty"`
	Paused       bool   `json:"paused"`
	Status       string `json:"status"`
	Assigned     bool   `json:"assigned"`
	AccountName  string `json:"account_name,omitempty"`
	AccountLogin string `json:"account_login,omitempty"`
}

type resultEntry struct {
	Serial string `json:"serial"`
	Tasks  int    `json:"tasks"`
	Errors int    `json:"errors"`
}

type runRequest struct {
	Tasks      []string `json:"tasks"`
	MaxDevices int      `json:"max_devices"`
	Workers    int      `json:"workers"`
}

// NewServer creates the panel handler and API routes.
// For Wails, only Handler() is used. For headless mode, call ListenAndServe().
func NewServer(projectRoot, configPath string, port int) *Server {
	if port <= 0 {
		port = DefaultPort
	}
	s := &Server{
		ProjectRoot: projectRoot,
		ConfigPath:  configPath,
		Port:        port,
		enabled:     map[string]bool{},
		activeTasks: map[string]bool{
			"facebook_login":        false,
			"facebook_auto_post":    true,
			"facebook_fanpage_post": false,
		},
		runStatus:   "idle",
		eventSubs:   map[chan []byte]struct{}{},
		mirrors:         map[string]*exec.Cmd{},
		mirrorSlot:      map[string]int{},
		mirrorLaunching: map[string]bool{},
		mirrorErrors:    map[string]string{},
		watchStop:       make(chan struct{}),
		emulatorPrepared: map[string]bool{},
	}
	go s.startDeviceWatch()
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/assets/styles.css", s.handlePanelCSS)
	mux.HandleFunc("/assets/tailwind.css", s.handlePanelTailwind)
	mux.HandleFunc("/assets/app.js", s.handlePanelJS)
	mux.HandleFunc("/assets/pages/", s.handlePanelPage)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/accounts/import", s.handleAccountsImport)
	mux.HandleFunc("/api/accounts/import-preview", s.handleAccountsImportPreview)
	mux.HandleFunc("/api/accounts/create", s.handleAccountsCreate)
	mux.HandleFunc("/api/accounts/update", s.handleAccountsUpdate)
	mux.HandleFunc("/api/accounts/auto-assign", s.handleAccountsAutoAssign)
	mux.HandleFunc("/api/accounts/assign", s.handleAccountsAssign)
	mux.HandleFunc("/api/accounts/unassign", s.handleAccountsUnassign)
	mux.HandleFunc("/api/accounts/automation", s.handleAccountsAutomation)
	mux.HandleFunc("/api/flows", s.handleFlowsCatalog)
	mux.HandleFunc("/api/post-texts", s.handlePostTexts)
	mux.HandleFunc("/api/post-texts/delete", s.handlePostTextDelete)
	mux.HandleFunc("/api/post-texts/counts", s.handlePostTextsCounts)
	mux.HandleFunc("/api/devices/refresh", s.handleRefreshDevices)
	mux.HandleFunc("/api/devices/register", s.handleRegisterDevice)
	mux.HandleFunc("/api/devices/enable-all", s.handleEnableAll)
	mux.HandleFunc("/api/devices/disable-all", s.handleDisableAll)
	mux.HandleFunc("/api/devices/toggle", s.handleToggleDevice)
	mux.HandleFunc("/api/devices/mirror-retry", s.handleMirrorRetry)
	mux.HandleFunc("/api/devices/mirror-relayout", s.handleMirrorRelayout)
	mux.HandleFunc("/api/devices/pause", s.handlePauseDevice)
	mux.HandleFunc("/api/devices/resume", s.handleResumeDevice)
	mux.HandleFunc("/api/emulators/launch", s.handleEmulatorLaunch)
	mux.HandleFunc("/api/emulators/add", s.handleEmulatorAdd)
	mux.HandleFunc("/api/emulators/quit", s.handleEmulatorQuit)
	mux.HandleFunc("/api/window/state", s.handleWindowState)
	mux.HandleFunc("/api/window/maximize", s.handleWindowMaximize)
	mux.HandleFunc("/api/window/fullscreen", s.handleWindowFullscreen)
	mux.HandleFunc("/api/tasks/toggle", s.handleToggleTask)
	mux.HandleFunc("/api/run", s.handleRun)
	mux.HandleFunc("/api/run/ack", s.handleRunAck)
	mux.HandleFunc("/api/pause", s.handlePauseAll)
	mux.HandleFunc("/api/resume", s.handleResumeAll)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/logs/tail", s.handleLogsTail)
	s.handler = mux
	return s
}

// Handler returns the panel HTTP handler (for Wails asset server or tests).
func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) ListenAndServe() error {
	if s.http == nil {
		s.http = &http.Server{
			Addr:    fmt.Sprintf(":%d", s.Port),
			Handler: s.handler,
		}
	}
	return s.http.ListenAndServe()
}

// Close stops running automation, mirrors, and releases panel resources.
func (s *Server) Close() error {
	s.shutdownPanel()
	var err error
	if s.http != nil {
		err = s.http.Close()
		s.http = nil
	}
	if s.Store != nil {
		s.Store.Close()
		s.Store = nil
	}
	return err
}

func (s *Server) shutdownPanel() {
	s.mu.Lock()
	s.shuttingDown = true
	wasRunning := s.runStatus == "running" || s.runStatus == "paused"
	s.stopRunLocked()
	s.mu.Unlock()
	if wasRunning {
		log.Printf("panel: automation di-stop (panel ditutup)")
	}
	runcontrol.Default.Stop()
	runcontrol.Default.Reset()

	select {
	case <-s.watchStop:
	default:
		close(s.watchStop)
	}

	s.mirrorSyncMu.Lock()
	if s.mirrorDebounce != nil {
		s.mirrorDebounce.Stop()
		s.mirrorDebounce = nil
	}
	s.mirrorSyncMu.Unlock()

	s.mirrorLaunchWg.Wait()
	s.stopAllMirrors()
	log.Printf("panel: mirror ditutup")
}

func (s *Server) stopRunLocked() {
	if s.runCancel != nil {
		s.runCancel()
	}
	s.runStatus = "idle"
	s.runCancel = nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(s.panelHTML()))
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	s.writeState(w)
}

func (s *Server) writeState(w http.ResponseWriter) {
	payload := s.buildState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func (s *Server) handleRefreshDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.refreshDevices()
	s.broadcastState()
	s.writeState(w)
}

func (s *Server) handleEnableAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.refreshDevices()
	s.mu.Lock()
	s.initEnabledDevices(wf)
	serials := make([]string, 0, len(s.devices))
	for _, d := range s.devices {
		serials = append(serials, d.Serial)
	}
	s.bumpStateRevLocked()
	s.mu.Unlock()
	s.persistDevicesEnabled(r.Context(), serials)
	s.requestSyncMirrors()
	s.writeState(w)
}

func (s *Server) handleDisableAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	serials := make([]string, 0, len(s.enabled))
	for serial := range s.enabled {
		s.enabled[serial] = false
		serials = append(serials, serial)
	}
	s.syncEnabledFlagsLocked()
	s.bumpStateRevLocked()
	s.mu.Unlock()
	s.persistDevicesEnabled(r.Context(), serials)
	s.requestSyncMirrors()
	s.writeState(w)
}

// handleToggleDevice sets one device's enabled flag (explicit enabled=true/false from UI).
func (s *Server) handleToggleDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Serial  string `json:"serial"`
		Enabled *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Serial == "" {
		http.Error(w, "invalid serial", http.StatusBadRequest)
		return
	}
	enabled := false
	if body.Enabled != nil {
		enabled = *body.Enabled
	} else {
		s.mu.Lock()
		enabled = !s.enabled[body.Serial]
		s.mu.Unlock()
	}
	s.persistDeviceEnabled(r.Context(), body.Serial, enabled)
	s.mu.Lock()
	s.enabled[body.Serial] = enabled
	s.syncEnabledFlagsLocked()
	s.bumpStateRevLocked()
	s.mu.Unlock()
	s.requestSyncMirrors()
	s.writeState(w)
}

func (s *Server) handleMirrorRetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Serial string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Serial == "" {
		http.Error(w, "invalid serial", http.StatusBadRequest)
		return
	}
	s.mirrorMu.Lock()
	s.stopMirrorLocked(body.Serial)
	s.mirrorMu.Unlock()
	s.requestSyncMirrors()
	s.writeState(w)
}

func (s *Server) handleMirrorRelayout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mirrorMu.Lock()
	moved, alive, restarted := s.relayoutMirrorsLocked()
	s.mirrorMu.Unlock()

	if restarted > 0 {
		s.requestSyncMirrors()
	} else {
		s.syncMirrorOpenFlagsLocked()
		s.broadcastState()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"moved":     moved,
		"alive":     alive,
		"restarted": restarted,
		"start_x":   s.panel.mirrorStartX(),
	})
}

func (s *Server) handlePauseDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Serial string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Serial == "" {
		http.Error(w, "invalid serial", http.StatusBadRequest)
		return
	}
	runcontrol.Default.PauseDevice(body.Serial)
	s.syncDeviceRuntimeState()
	s.broadcastState()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleResumeDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Serial string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Serial == "" {
		http.Error(w, "invalid serial", http.StatusBadRequest)
		return
	}
	runcontrol.Default.ResumeDevice(body.Serial)
	s.syncDeviceRuntimeState()
	s.broadcastState()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleToggleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var body struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.activeTasks[body.Name] = !s.activeTasks[body.Name]
	s.mu.Unlock()
	s.broadcastState()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePauseAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	if s.runStatus != "running" {
		s.mu.Unlock()
		http.Error(w, "not running", http.StatusBadRequest)
		return
	}
	s.runStatus = "paused"
	s.mu.Unlock()
	runcontrol.Default.PauseAll()
	s.syncDeviceRuntimeState()
	s.broadcastState()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleResumeAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	if s.runStatus != "paused" {
		s.mu.Unlock()
		http.Error(w, "not paused", http.StatusBadRequest)
		return
	}
	s.runStatus = "running"
	s.mu.Unlock()
	runcontrol.Default.ResumeAll()
	s.syncDeviceRuntimeState()
	s.broadcastState()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	if s.runStatus != "running" && s.runStatus != "paused" {
		s.mu.Unlock()
		http.Error(w, "not running", http.StatusBadRequest)
		return
	}
	s.stopRunLocked()
	s.mu.Unlock()
	runcontrol.Default.Stop()
	runcontrol.Default.Reset()
	s.syncDeviceRuntimeState()
	s.broadcastState()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wf, err := s.loadWorkflow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	if s.runStatus == "running" || s.runStatus == "paused" {
		s.mu.Unlock()
		http.Error(w, "already running", http.StatusConflict)
		return
	}
	enabledSerials := s.enabledSerials()
	s.mu.Unlock()

	wf.Tasks = nil // tasks resolved per account in worker
	if req.MaxDevices > 0 {
		wf.MaxDevices = req.MaxDevices
	}
	if req.Workers > 0 {
		wf.ParallelWorkers = req.Workers
	}

	all, err := adb.ListDevices()
	if err != nil {
		http.Error(w, "ADB error — cek koneksi USB", http.StatusServiceUnavailable)
		return
	}
	serials := enabledSerialsFromList(all, enabledSerials)
	if len(serials) == 0 {
		http.Error(w, "no enabled devices", http.StatusBadRequest)
		return
	}
	if len(serials) > wf.MaxDevices {
		serials = serials[:wf.MaxDevices]
	}

	log.Printf("Panel run: %d HP aktif %v (max=%d workers=%d akun=%d)",
		len(serials), serials, wf.MaxDevices, wf.ParallelWorkers, s.accountCount())

	runcontrol.Default.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	runcontrol.Default.SetCancel(cancel)

	s.mu.Lock()
	if s.runStatus == "running" || s.runStatus == "paused" || s.shuttingDown {
		s.mu.Unlock()
		cancel()
		http.Error(w, "already running", http.StatusConflict)
		return
	}
	s.runStatus = "running"
	s.runCancel = cancel
	s.lastResults = nil
	s.mu.Unlock()
	s.setDeviceStatus(serials, "running")
	s.broadcastState()

	go func() {
		defer cancel()
		runID := time.Now().Format("20060102_150405")
		logDir := filepath.Join(s.ProjectRoot, "logs")
		ctrl := &controller.Controller{
			Workflow:    wf,
			ProjectRoot: s.ProjectRoot,
			LogDir:      logDir,
			RunID:       runID,
			Store:       s.Store,
		}
		results := ctrl.Run(ctx, serials)

		s.mu.Lock()
		if s.shuttingDown {
			s.mu.Unlock()
			return
		}
		s.runStatus = "done"
		s.runCancel = nil
		s.lastResults = nil
		for _, res := range results {
			st := "done"
			if len(res.Errors) > 0 {
				st = "error"
			}
			s.lastResults = append(s.lastResults, resultEntry{
				Serial: res.Serial, Tasks: res.TasksCompleted, Errors: len(res.Errors),
			})
			s.setDeviceStatusLocked([]string{res.Serial}, st)
		}
		s.mu.Unlock()
		s.refreshDevices()
		s.broadcastState()
		log.Printf("Panel run selesai: %d HP", len(results))
	}()

	w.WriteHeader(http.StatusAccepted)
	s.writeState(w)
}

func (s *Server) handleRunAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	if s.runStatus == "done" {
		s.runStatus = "idle"
	}
	s.mu.Unlock()
	s.writeState(w)
}

func (s *Server) refreshDevicesFrom(all []string) {
	connected := map[string]bool{}
	for _, serial := range all {
		connected[serial] = true
	}
	registered := s.registeredSerials()
	for serial := range registered {
		if !connected[serial] {
			all = append(all, serial)
		}
	}

	s.prepareConnectedEmulators(connected)

	persisted := s.loadPersistedEnabledFlags()

	s.mu.Lock()
	runSt := s.runStatus
	lastRes := append([]resultEntry(nil), s.lastResults...)
	for _, serial := range all {
		if _, had := s.enabled[serial]; had {
			continue // jangan timpa toggle in-memory dengan DB stale
		}
		if on, ok := persisted[serial]; ok {
			s.enabled[serial] = on
		} else {
			s.enabled[serial] = false
		}
	}
	s.mu.Unlock()

	assignments := map[string]store.DeviceAssignment{}
	if s.Store != nil {
		if m, err := s.Store.ListDeviceAssignments(context.Background()); err == nil {
			assignments = m
		}
	}

	infos := make([]DeviceInfo, 0, len(all))
	for _, serial := range all {
		isConnected := connected[serial]
		info := DeviceInfo{
			Serial:    serial,
			Connected: isConnected,
			Paused:    runcontrol.Default.IsPaused(serial),
			Status:    deviceStatusForRun(runSt, serial, lastRes),
		}
		if isConnected {
			client := &adb.Client{Serial: serial, Timeout: 10 * time.Second}
			info.Model = client.DeviceModel()
			info.Resolution = client.DeviceResolution()
		} else {
			if label := registered[serial]; label != "" {
				info.Model = label
			}
			info.Resolution = "—"
			info.Status = "offline"
		}
		if a, ok := assignments[serial]; ok {
			info.Assigned = true
			info.AccountName = a.Name
			info.AccountLogin = a.LoginID
		}
		infos = append(infos, info)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = infos
	s.syncEnabledFlagsLocked()
	for i := range s.devices {
		serial := s.devices[i].Serial
		if !s.enabled[serial] {
			s.devices[i].MirrorOpen = false
			s.devices[i].MirrorError = ""
			if monitor.ScrcpyRunningForSerial(serial) {
				monitor.StopScrcpyForSerial(serial)
			}
			continue
		}
		s.devices[i].MirrorOpen = monitor.ScrcpyRunningForSerial(serial)
	}
	s.bumpStateRevLocked()
}

func deviceStatusForRun(runSt, serial string, lastResults []resultEntry) string {
	st := "idle"
	switch runSt {
	case "running":
		if runcontrol.Default.IsPaused(serial) {
			st = "paused"
		} else {
			st = "running"
		}
	case "paused":
		st = "paused"
	case "done":
		for _, lr := range lastResults {
			if lr.Serial == serial {
				if lr.Errors > 0 {
					return "error"
				}
				return "done"
			}
		}
	}
	return st
}

func (s *Server) accountCount() int {
	if s.Store == nil {
		return 0
	}
	n, err := s.Store.CountAccounts(context.Background())
	if err != nil {
		return 0
	}
	return n
}

func (s *Server) assignedDeviceCount() int {
	if s.Store == nil {
		return 0
	}
	n, err := s.Store.CountDevicesWithAssignments(context.Background())
	if err != nil {
		return 0
	}
	return n
}

// effectiveRunLimits returns sensible max_devices/workers for the panel (not raw config 100).
// Uses the already-cached device list (caller must hold s.mu) instead of shelling out to adb.
func (s *Server) effectiveRunLimits(wf *config.Workflow) (maxDev, workers, assigned int) {
	n := len(s.devices)
	if n == 0 {
		n = 1
	}
	maxDev = n
	if wf.MaxDevices > 0 && wf.MaxDevices < maxDev {
		maxDev = wf.MaxDevices
	}
	assigned = s.assignedDeviceCount()
	if assigned > 0 && assigned < maxDev {
		maxDev = assigned
	}
	workers = maxDev
	if wf.ParallelWorkers > 0 && wf.ParallelWorkers < workers {
		workers = wf.ParallelWorkers
	}
	return maxDev, workers, assigned
}

// initEnabledDevices sets initial enabled flags from the cached device list. Caller must hold
// s.mu and should call refreshDevices beforehand so s.devices reflects what's connected now.
func (s *Server) initEnabledDevices(wf *config.Workflow) {
	if len(s.devices) == 0 {
		return
	}
	limit, _, _ := s.effectiveRunLimits(wf)
	for i, d := range s.devices {
		s.enabled[d.Serial] = i < limit
	}
	s.syncEnabledFlagsLocked()
	log.Printf("Panel: %d akun → %d HP aktif (dari %d terhubung)", s.accountCount(), limit, len(s.devices))
}

// syncEnabledFlagsLocked copies s.enabled into each cached DeviceInfo.Enabled so the API/UI
// reflect a toggle immediately, without re-querying adb for model/resolution. Caller must hold s.mu.
func (s *Server) bumpStateRevLocked() {
	s.stateRev++
}

func (s *Server) syncEnabledFlagsLocked() {
	for i := range s.devices {
		s.devices[i].Enabled = s.enabled[s.devices[i].Serial]
	}
}

func (s *Server) loadPersistedEnabledFlags() map[string]bool {
	if s.Store == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	m, err := s.Store.LoadDeviceMirrorEnabled(ctx)
	if err != nil {
		log.Printf("panel: load mirror_enabled: %v", err)
		return nil
	}
	return m
}

func (s *Server) persistDeviceEnabled(ctx context.Context, serial string, enabled bool) {
	if s.Store == nil || serial == "" {
		return
	}
	c, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := s.Store.SetDeviceMirrorEnabled(c, serial, enabled); err != nil {
		log.Printf("panel: persist mirror_enabled %s: %v", serial, err)
	}
}

func (s *Server) persistDevicesEnabled(ctx context.Context, serials []string) {
	if s.Store == nil || len(serials) == 0 {
		return
	}
	s.mu.RLock()
	on := make([]string, 0, len(serials))
	off := make([]string, 0, len(serials))
	for _, serial := range serials {
		if s.enabled[serial] {
			on = append(on, serial)
		} else {
			off = append(off, serial)
		}
	}
	s.mu.RUnlock()
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if len(on) > 0 {
		if err := s.Store.SetDevicesMirrorEnabled(c, on, true); err != nil {
			log.Printf("panel: persist mirror_enabled ON: %v", err)
		}
	}
	if len(off) > 0 {
		if err := s.Store.SetDevicesMirrorEnabled(c, off, false); err != nil {
			log.Printf("panel: persist mirror_enabled OFF: %v", err)
		}
	}
}

func (s *Server) enabledSerials() map[string]bool {
	out := map[string]bool{}
	for k, v := range s.enabled {
		if v {
			out[k] = true
		}
	}
	return out
}

func (s *Server) setDeviceStatus(serials []string, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setDeviceStatusLocked(serials, status)
}

func (s *Server) setDeviceStatusLocked(serials []string, status string) {
	set := map[string]bool{}
	for _, serial := range serials {
		set[serial] = true
	}
	for i := range s.devices {
		if set[s.devices[i].Serial] {
			s.devices[i].Status = status
		}
	}
}

// syncDeviceRuntimeState updates pause/status fields without re-querying adb.
func (s *Server) syncDeviceRuntimeState() {
	s.mu.Lock()
	defer s.mu.Unlock()
	runSt := s.runStatus
	lastRes := append([]resultEntry(nil), s.lastResults...)
	for i := range s.devices {
		serial := s.devices[i].Serial
		s.devices[i].Paused = runcontrol.Default.IsPaused(serial)
		s.devices[i].Status = deviceStatusForRun(runSt, serial, lastRes)
	}
}

func enabledSerialsFromList(all []string, enabled map[string]bool) []string {
	var serials []string
	for _, d := range all {
		if enabled[d] {
			serials = append(serials, d)
		}
	}
	return serials
}
