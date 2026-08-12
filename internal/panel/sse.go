package panel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
)

const deviceWatchInterval = 5 * time.Second

func (s *Server) startDeviceWatch() {
	go s.deviceWatchLoop()
	go s.mirrorHealthLoop()
	go func() {
		time.Sleep(300 * time.Millisecond)
		if s.refreshDevices() {
			s.broadcastState()
		}
	}()
}

func (s *Server) deviceWatchLoop() {
	ticker := time.NewTicker(deviceWatchInterval)
	defer ticker.Stop()
	lastDeviceKey := ""
	lastResourceKey := ""
	s.scanDevicesLocked(&lastDeviceKey)
	for {
		select {
		case <-s.watchStop:
			return
		case <-ticker.C:
			deviceChanged := s.scanDevicesLocked(&lastDeviceKey)
			if deviceChanged {
				s.requestSyncMirrors()
			}
			changed := deviceChanged
			if s.resourcesChanged(lastResourceKey) {
				lastResourceKey = s.resourceWatchKey()
				changed = true
			}
			if changed {
				s.broadcastState()
			}
		}
	}
}

func (s *Server) scanDevicesLocked(lastDeviceKey *string) bool {
	s.connectConfiguredEmulators()
	all, err := adb.ListDevices()
	if err != nil {
		s.mu.Lock()
		s.adbLastError = err.Error()
		s.mu.Unlock()
		log.Printf("panel: scan USB gagal: %v", err)
		return false
	}
	s.mu.Lock()
	s.adbLastError = ""
	s.mu.Unlock()
	key := strings.Join(all, ",")
	if key == *lastDeviceKey {
		return false
	}
	*lastDeviceKey = key
	s.refreshDevicesFrom(all)
	return true
}

// connectConfiguredEmulators adb-connects LDPlayer ports from config (no-op when auto_connect is false).
func (s *Server) connectConfiguredEmulators() {
	wf, err := s.loadWorkflow()
	if err != nil || wf == nil || !wf.Emulator.AutoConnect {
		return
	}
	adb.ConnectLocalEmulators(adb.LDPlayerADBPorts(wf.Emulator.InstanceCount))
}

func (s *Server) mirrorHealthLoop() {
	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.watchStop:
			return
		case <-ticker.C:
			s.mu.RLock()
			devicesCopy := append([]DeviceInfo(nil), s.devices...)
			s.mu.RUnlock()

			s.mirrorMu.Lock()
			var needRetry bool
			for _, d := range devicesCopy {
				if !d.Enabled || d.MirrorOpen || !d.Connected {
					continue
				}
				if s.mirrorLaunching[d.Serial] {
					continue
				}
				needRetry = true
				break
			}
			s.mirrorMu.Unlock()
			if needRetry {
				log.Printf("panel: mirror belum terbuka — retry sync")
				s.requestSyncMirrors()
			}
		}
	}
}

func (s *Server) resourceWatchKey() string {
	key := s.ConfigPath
	if info, err := os.Stat(s.ConfigPath); err == nil {
		key += "|" + info.ModTime().String()
	}
	return key
}

func (s *Server) resourcesChanged(lastKey string) bool {
	return s.resourceWatchKey() != lastKey
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 8)
	s.subscribeEvents(ch)
	defer s.unsubscribeEvents(ch)

	if data, err := s.buildStateJSON(); err == nil {
		s.writeSSE(w, flusher, data)
	}

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-ch:
			s.writeSSE(w, flusher, data)
		case <-heartbeat.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) writeSSE(w http.ResponseWriter, flusher http.Flusher, data []byte) {
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func (s *Server) subscribeEvents(ch chan []byte) {
	s.eventMu.Lock()
	s.eventSubs[ch] = struct{}{}
	s.eventMu.Unlock()
}

func (s *Server) unsubscribeEvents(ch chan []byte) {
	s.eventMu.Lock()
	delete(s.eventSubs, ch)
	s.eventMu.Unlock()
}

func (s *Server) broadcastState() {
	data, err := s.buildStateJSON()
	if err != nil {
		return
	}
	s.eventMu.Lock()
	for ch := range s.eventSubs {
		select {
		case ch <- data:
		default:
		}
	}
	s.eventMu.Unlock()
}

func (s *Server) buildStateJSON() ([]byte, error) {
	payload := s.buildState()
	return json.Marshal(payload)
}

func (s *Server) buildState() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wf, _ := s.loadWorkflow()
	tasks := []map[string]interface{}{}
	maxDev, workers, accountCount, enabledCount, assignedCount, mirrorOpenCount := 2, 2, 0, 0, 0, 0
	for _, d := range s.devices {
		if d.Enabled {
			enabledCount++
		}
		if d.MirrorOpen {
			mirrorOpenCount++
		}
	}
	if wf != nil {
		maxDev, workers, assignedCount = s.effectiveRunLimits(wf)
		accountCount = s.accountCount()
		for _, t := range wf.Tasks {
			info := config.FlowInfoFor(t.Flow)
			steps := info.Steps
			if len(t.Actions) > 0 {
				steps = config.ActionLabels(t.Actions)
			}
			desc := t.Description
			if desc == "" {
				desc = info.Description
			}
			tasks = append(tasks, map[string]interface{}{
				"name":        t.Name,
				"app":         t.App,
				"flow":        t.Flow,
				"steps":       steps,
				"description": desc,
				"actions":     len(t.Actions),
				"active":      s.activeTasks[t.Name],
			})
		}
	}
	settings := config.PanelSettings{}
	if wf != nil {
		settings = config.PanelSettingsFromWorkflow(wf)
		settings.Normalize()
	}
	return map[string]interface{}{
		"state_rev":       s.stateRev,
		"ui_rev":          s.uiRevValue(),
		"panel_dev":       s.panelDevEnabled(),
		"run_status":      s.runStatus,
		"adb_error":       s.adbLastError,
		"max_devices":     maxDev,
		"workers":         workers,
		"account_count":   accountCount,
		"assigned_count":  assignedCount,
		"enabled_count":     enabledCount,
		"mirror_open_count": mirrorOpenCount,
		"devices":         s.devices,
		"tasks":           tasks,
		"last_results":    s.lastResults,
		"preflight":       s.preflightLocked(),
		"health":          s.healthLocked(),
		"checklist":       s.checklistLocked(),
		"accounts":        s.listAccountsLocked(),
		"settings":        settings,
		"ldplayer":        s.ldplayerSummary(),
		"panel_bounds":    s.panel.snapshot(),
	}
}
