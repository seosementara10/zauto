package panel

import (
	"context"
	"time"

	"zauto/internal/adb"
	"zauto/internal/monitor"
)

func (s *Server) healthLocked() map[string]interface{} {
	issues := []string{}
	adbOK := false
	deviceCount := 0
	if all, err := adb.ListDevices(); err == nil {
		adbOK = true
		deviceCount = len(all)
		if deviceCount == 0 {
			issues = append(issues, "Tidak ada HP terhubung")
		}
	} else {
		issues = append(issues, "ADB tidak merespon")
	}

	dbOK := false
	if s.Store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := s.Store.Ping(ctx)
		cancel()
		if err == nil {
			dbOK = true
		} else {
			issues = append(issues, "Database tidak terhubung")
		}
	} else {
		issues = append(issues, "Database belum dikonfigurasi")
	}

	scrcpyCount := monitor.CountScrcpyProcesses()
	ok := adbOK && dbOK && len(issues) == 0
	return map[string]interface{}{
		"ok":            ok,
		"adb":           adbOK,
		"db":            dbOK,
		"device_count":  deviceCount,
		"scrcpy_count":  scrcpyCount,
		"issues":        issues,
	}
}

func (s *Server) checklistLocked() []map[string]interface{} {
	health := s.healthLocked()
	pf := s.preflightLocked()

	enabledCount := 0
	for _, d := range s.devices {
		if d.Enabled {
			enabledCount++
		}
	}
	accountCount := s.accountCount()
	assignedCount := s.assignedDeviceCount()
	automationReady := 0
	if s.Store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		rows, err := s.Store.ListAccountSummaries(ctx)
		cancel()
		if err == nil {
			for _, a := range rows {
				if a.AssignedSerial != "" && a.AutomationEnabled && a.AutomationFlow != "" {
					automationReady++
				}
			}
		}
	}

	items := []map[string]interface{}{
		{
			"id": "adb", "label": "ADB & HP terhubung", "page": "devices",
			"ok": health["adb"].(bool) && health["device_count"].(int) > 0,
		},
		{
			"id": "db", "label": "Database PostgreSQL", "page": "accounts",
			"ok": health["db"].(bool),
		},
		{
			"id": "devices", "label": "Minimal 1 HP aktif (mirror)", "page": "devices",
			"ok": enabledCount > 0,
		},
		{
			"id": "accounts", "label": "Akun di database", "page": "accounts",
			"ok": accountCount > 0,
		},
		{
			"id": "assign", "label": "HP sudah di-assign akun", "page": "accounts",
			"ok": assignedCount > 0,
		},
		{
			"id": "tasks", "label": "Skrip automation per akun", "page": "accounts",
			"ok": automationReady > 0,
		},
		{
			"id": "run", "label": "Siap jalankan automation", "page": "kontrol",
			"ok": pf["can_run"].(bool),
		},
	}
	return items
}
