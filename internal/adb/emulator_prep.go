package adb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// LDPlayerBaseADBPort is the ADB port for LDPlayer instance index 0.
const LDPlayerBaseADBPort = 5555

// LDPlayerADBPorts returns ADB ports for LDPlayer instances 0..count-1 (5555, 5557, ...).
func LDPlayerADBPorts(count int) []int {
	if count <= 0 {
		count = 10
	}
	ports := make([]int, count)
	for i := 0; i < count; i++ {
		ports[i] = LDPlayerBaseADBPort + i*2
	}
	return ports
}

// EmulatorPrepResult summarizes emulator hardening applied via adb shell.
type EmulatorPrepResult struct {
	Serial   string
	Applied  []string
	Warnings []string
	Errors   []string
}

// ConnectLocalEmulators tries adb connect on LDPlayer-style local emulator ports.
func ConnectLocalEmulators(ports []int) {
	if len(ports) == 0 {
		ports = LDPlayerADBPorts(10)
	}
	timeout := time.Duration(len(ports)*2+5) * time.Second
	if timeout < 20*time.Second {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for _, port := range ports {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		cmd := adbCommand(ctx, "connect", addr)
		out, err := cmd.CombinedOutput()
		text := strings.TrimSpace(string(out))
		if err != nil {
			log.Printf("adb: connect %s: %v (%s)", addr, err, text)
			continue
		}
		if strings.Contains(text, "connected") {
			log.Printf("adb: %s", text)
		}
	}
}

// IsEmulatorSerial reports whether the adb serial looks like a local emulator.
func IsEmulatorSerial(serial string) bool {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return false
	}
	if strings.HasPrefix(serial, "emulator-") {
		return true
	}
	return strings.HasPrefix(serial, "127.0.0.1:") || strings.HasPrefix(serial, "localhost:")
}

// IsEmulator reports whether the device exposes common emulator boot properties.
func (c *Client) IsEmulator() bool {
	if IsEmulatorSerial(c.Serial) {
		return true
	}
	switch strings.TrimSpace(c.GetProp("ro.boot.qemu")) {
	case "1", "true":
		return true
	}
	switch strings.TrimSpace(c.GetProp("ro.kernel.qemu")) {
	case "1", "true":
		return true
	}
	hw := strings.ToLower(strings.TrimSpace(c.GetProp("ro.hardware")))
	return hw == "goldfish" || hw == "ranchu" || hw == "vbox86" || hw == "vbox"
}

// GetProp reads a single Android system property.
func (c *Client) GetProp(name string) string {
	out, err := c.Shell("getprop", name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// PrepareEmulator applies best-effort emulator property tweaks on connect.
// Play Integrity / SafetyNet cannot be passed programmatically without manual Magisk setup.
func PrepareEmulator(c *Client) EmulatorPrepResult {
	res := EmulatorPrepResult{Serial: c.Serial}

	for _, item := range []struct {
		key string
		val string
	}{
		{"ro.boot.qemu", "0"},
		{"ro.kernel.qemu", "0"},
		{"ro.kernel.android.qemud", ""},
	} {
		if _, err := c.Shell("setprop", item.key, item.val); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("setprop %s: %v", item.key, err))
			continue
		}
		got := c.GetProp(item.key)
		if item.val == "" && got != "" {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s still %q after clear", item.key, got))
			continue
		}
		if item.val != "" && got != item.val {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s is %q (wanted %q)", item.key, got, item.val))
			continue
		}
		if item.val == "" {
			res.Applied = append(res.Applied, item.key+"=cleared")
		} else {
			res.Applied = append(res.Applied, item.key+"="+item.val)
		}
	}

	if route, err := c.Shell("ip", "-4", "route", "show", "dev", "wlan0"); err == nil {
		route = strings.TrimSpace(route)
		if strings.Contains(route, "172.16.") || strings.Contains(route, "10.0.2.") {
			res.Warnings = append(res.Warnings,
				"internal NAT IP on wlan0 — cannot hide from apps via adb; use physical device for production")
		}
	}

	if abi := c.GetProp("ro.product.cpu.abi"); strings.EqualFold(abi, "x86_64") || strings.EqualFold(abi, "x86") {
		res.Warnings = append(res.Warnings,
			"CPU ABI "+abi+" — apps can detect x86 emulator even when ro.boot.qemu is hidden")
	}

	res.Warnings = append(res.Warnings, playIntegrityWarning(c))

	return res
}

func playIntegrityWarning(c *Client) string {
	// gservices / GMS must be present for real attestation; emulators usually fail regardless of setprop.
	if out, err := c.Shell("pm", "path", "com.google.android.gms"); err != nil || strings.TrimSpace(out) == "" {
		return "Play Integrity: Google Play services not found — attestation unavailable"
	}
	return "Play Integrity / SafetyNet: cannot auto-pass on emulator (needs manual Magisk + integrity module); use physical device for Facebook production"
}

// LogEmulatorPrep writes prep results to the standard logger.
func LogEmulatorPrep(res EmulatorPrepResult) {
	if len(res.Applied) > 0 {
		log.Printf("emulator prep [%s]: applied %s", res.Serial, strings.Join(res.Applied, ", "))
	}
	for _, w := range res.Warnings {
		log.Printf("emulator prep [%s]: warn: %s", res.Serial, w)
	}
	for _, e := range res.Errors {
		log.Printf("emulator prep [%s]: error: %s", res.Serial, e)
	}
}
