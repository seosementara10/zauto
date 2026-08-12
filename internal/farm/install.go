package farm

import (
	"fmt"

	"zauto/internal/monitor"
)

// Install verifies scrcpy is present.
func Install(projectRoot string) error {
	fmt.Println("=== Farm Install ===")
	if _, err := monitor.FindScrcpy(projectRoot); err != nil {
		return err
	}
	fmt.Println("✓ scrcpy siap")
	return nil
}

// PrintSTFGuide shows STF setup for scale 10+ devices on WSL2.
func PrintSTFGuide() {
	fmt.Print(`
=== STF (DeviceFarmer) — scale 10–500+ HP ===

STF adalah standar lab profesional (browser dashboard + WebSocket stream).
Di Windows: jalankan provider di WSL2 + usbipd-win untuk USB passthrough.

Langkah ringkas:
  1. Install WSL2 Ubuntu + Docker Desktop
  2. winget install dorssel.usbipd
  3. usbipd bind --busid <ID>   (per HP, PowerShell Admin)
  4. usbipd attach --wsl --busid <ID>
  5. Di WSL: docker compose -f farm/stf/docker-compose.yml up -d
  6. Buka http://localhost:7100

File: farm/stf/docker-compose.yml (template)
`)
}
