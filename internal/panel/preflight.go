package panel

import (
	"context"
	"time"

	"zauto/internal/adb"
)

func (s *Server) preflightLocked() map[string]interface{} {
	var reasons []string

	enabledCount := 0
	for _, on := range s.enabled {
		if on {
			enabledCount++
		}
	}
	if enabledCount == 0 {
		reasons = append(reasons, "Aktifkan minimal 1 HP di menu Device")
	}

	if s.accountCount() == 0 {
		reasons = append(reasons, "Import akun di menu Akun")
	} else if s.Store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		rows, err := s.Store.ListAccountSummaries(ctx)
		cancel()
		if err == nil {
			assigned := 0
			ready := 0
			for _, a := range rows {
				if a.AssignedSerial != "" {
					assigned++
					if a.AutomationEnabled && a.AutomationFlow != "" {
						ready++
					}
				}
			}
			if assigned == 0 {
				reasons = append(reasons, "Assign akun ke HP di menu Akun")
			} else if ready == 0 {
				reasons = append(reasons, "Set skrip automation per akun di menu Akun")
			}
		}
	}

	all, err := adb.ListDevices()
	if err != nil {
		reasons = append(reasons, "ADB tidak merespon — cek USB/driver")
	} else if len(all) == 0 {
		reasons = append(reasons, "Tidak ada HP terhubung — colok USB")
	}

	canRun := len(reasons) == 0 && s.runStatus != "running" && s.runStatus != "paused"
	return map[string]interface{}{
		"can_run": canRun,
		"reasons": reasons,
	}
}
