package check

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"zauto/internal/adb"
	"zauto/internal/monitor"
)

// Run prints environment validation (replaces legacy setup scripts).
func Run(projectRoot string) int {
	fmt.Println("=== zauto setup check ===")
	fmt.Println()

	failures := 0
	warn := func(msg string) { fmt.Printf("  WARN: %s\n", msg) }
	ok := func(msg string) { fmt.Printf("  OK: %s\n", msg) }
	fail := func(msg string) { fmt.Printf("  FAIL: %s\n", msg); failures++ }

	fmt.Println("[1] Go")
	if out, err := exec.Command("go", "version").CombinedOutput(); err == nil {
		ok(strings.TrimSpace(string(out)))
	} else {
		warn("Go not found — install from https://go.dev/dl/")
	}

	fmt.Println("[2] zauto binary")
	exePath := filepath.Join(projectRoot, "zauto.exe")
	if _, err := os.Stat(exePath); err == nil {
		ok("zauto.exe")
	} else {
		warn("zauto.exe not built — run: go build -o zauto.exe ./cmd/zauto")
	}

	fmt.Println("[3] ADB")
	if !adb.CheckAvailable() {
		fail("ADB not found — add Android Platform Tools to PATH")
	} else {
		out, _ := exec.Command("adb", "version").CombinedOutput()
		line := strings.TrimSpace(strings.Split(string(out), "\n")[0])
		ok(line)
	}

	fmt.Println("[4] Connected devices")
	devices, err := adb.ListDevices()
	if err != nil {
		fail(err.Error())
	} else if len(devices) == 0 {
		warn("No devices connected")
	} else {
		ok(fmt.Sprintf("%d device(s)", len(devices)))
		for _, serial := range devices {
			client := &adb.Client{Serial: serial, Timeout: 15 * time.Second}
			model := strings.TrimSpace(client.DeviceModel())
			res := client.DeviceResolution()
			fmt.Printf("    - %s  (%s, %s)\n", serial, model, res)
		}
	}

	fmt.Println("[5] Config")
	cfg := filepath.Join(projectRoot, "config", "config.json")
	if _, err := os.Stat(cfg); err != nil {
		fail("config/config.json missing")
	} else {
		ok("config/config.json")
	}

	fmt.Println("[6] scrcpy (monitor)")
	if _, err := monitor.FindScrcpy(projectRoot); err != nil {
		warn("scrcpy not found — download to tools/scrcpy/ for --monitor")
	} else {
		ok("scrcpy ready for --monitor")
	}

	if len(devices) > 0 && adb.CheckAvailable() {
		fmt.Println("[7] ADB tap permission")
		client := &adb.Client{Serial: devices[0], Timeout: 15 * time.Second}
		if _, err := client.Shell("input", "tap", "100", "100"); err != nil {
			if strings.Contains(err.Error(), "SecurityException") || strings.Contains(err.Error(), "INJECT_EVENTS") {
				fail("tap blocked — enable USB debugging (Security) on phone, keep screen unlocked")
			} else {
				warn(fmt.Sprintf("tap test: %v", err))
			}
		} else {
			ok("ADB tap allowed")
		}
	}

	fmt.Println()
	fmt.Println("=== Done ===")
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  .\\zauto --panel              # panel kontrol web")
	fmt.Println("  .\\zauto --max-devices 2     # jalankan login CLI")

	if failures > 0 {
		return 1
	}
	return 0
}
