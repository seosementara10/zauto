package adb

import (
	"fmt"
	"strconv"
	"strings"
)

// emulatorPeerSerial returns the alternate adb serial for the same local emulator.
// emulator-5554 and 127.0.0.1:5555 refer to one LDPlayer/Android Emulator instance.
func emulatorPeerSerial(serial string) (peer string, ok bool) {
	serial = strings.TrimSpace(serial)
	if strings.HasPrefix(serial, "emulator-") {
		portStr := strings.TrimPrefix(serial, "emulator-")
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 0 {
			return "", false
		}
		return fmt.Sprintf("127.0.0.1:%d", port+1), true
	}
	if host, portStr, found := strings.Cut(serial, ":"); found && (host == "127.0.0.1" || host == "localhost") {
		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 {
			return "", false
		}
		return fmt.Sprintf("emulator-%d", port-1), true
	}
	return "", false
}

// DedupeDevices removes duplicate adb serials that point to the same local emulator.
// When both emulator-5554 and 127.0.0.1:5555 are present, the TCP serial is kept.
func DedupeDevices(devices []string) []string {
	if len(devices) < 2 {
		return devices
	}
	set := make(map[string]bool, len(devices))
	for _, serial := range devices {
		set[serial] = true
	}
	drop := make(map[string]bool)
	for serial := range set {
		peer, ok := emulatorPeerSerial(serial)
		if !ok || !set[peer] {
			continue
		}
		if strings.HasPrefix(serial, "emulator-") {
			drop[serial] = true
		}
	}
	if len(drop) == 0 {
		return devices
	}
	out := make([]string, 0, len(devices)-len(drop))
	seen := make(map[string]bool, len(devices))
	for _, serial := range devices {
		if drop[serial] || seen[serial] {
			continue
		}
		seen[serial] = true
		out = append(out, serial)
	}
	return out
}
