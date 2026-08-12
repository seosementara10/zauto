//go:build windows

package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"zauto/internal/toolchain"
)

func launchScrcpyAt(scrcpy, scrcpyDir, serial string, hpNum int, tile WindowTile, opts Options) (*exec.Cmd, error) {
	args := scrcpyArgs(serial, hpNum, tile, opts)
	return launchScrcpyDirect(scrcpy, scrcpyDir, opts.ProjectRoot, args)
}

func launchScrcpyDirect(scrcpy, scrcpyDir, projectRoot string, args []string) (*exec.Cmd, error) {
	cmd := exec.Command(scrcpy, args...)
	cmd.Dir = scrcpyDir
	cmd.Env = toolchain.ScrcpyEnv(projectRoot, scrcpyDir)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	RegisterScrcpy(cmd)
	return cmd, nil
}

func scrcpyArgs(serial string, hpNum int, tile WindowTile, opts Options) []string {
	short := serial
	if len(short) > 8 {
		short = serial[len(short)-8:]
	}
	return []string{
		"-s", serial,
		"--max-size", fmt.Sprintf("%d", opts.MaxSize),
		"--video-buffer", "0",
		"--window-title", fmt.Sprintf("HP%d ...%s", hpNum, short),
		"--window-x", fmt.Sprintf("%d", tile.X),
		"--window-y", fmt.Sprintf("%d", tile.Y),
		"--window-width", fmt.Sprintf("%d", tile.W),
		"--window-height", fmt.Sprintf("%d", tile.H),
		"--window-borderless",
		"--stay-awake",
		"--no-audio",
	}
}

func scrcpySerialNeedle(serial string) string { return "-s " + serial }

func commandLineMatchesSerial(line, serial string) bool {
	needle := scrcpySerialNeedle(serial)
	idx := strings.Index(line, needle)
	if idx < 0 {
		return false
	}
	after := idx + len(needle)
	if after >= len(line) {
		return true
	}
	switch line[after] {
	case ' ', '\t', '"':
		return true
	default:
		return false
	}
}

func findScrcpyPIDs(serial string) []int {
	var pids []int
	seen := map[int]bool{}
	for _, pid := range scrcpyProcessPIDs() {
		line, err := processCommandLine(uint32(pid))
		if err != nil {
			continue
		}
		if !commandLineMatchesSerial(line, serial) {
			continue
		}
		if seen[pid] {
			continue
		}
		seen[pid] = true
		pids = append(pids, pid)
	}
	return pids
}

// ScrcpyRunningForSerial reports whether scrcpy is already running for a device.
func ScrcpyRunningForSerial(serial string) bool {
	if len(findScrcpyPIDs(serial)) > 0 {
		return true
	}
	return ScrcpyWindowRunning(serial)
}

// ScrcpyPIDForSerial returns one scrcpy PID for the device, if any.
func ScrcpyPIDForSerial(serial string) (int, bool) {
	pids := findScrcpyPIDs(serial)
	if len(pids) == 0 {
		return 0, false
	}
	return pids[0], true
}

// DedupeScrcpyForSerial closes extra scrcpy windows for one device, keeping one instance.
func DedupeScrcpyForSerial(serial string) {
	pids := findScrcpyPIDs(serial)
	if len(pids) <= 1 {
		return
	}
	for _, pid := range pids[1:] {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Kill()
		}
	}
}

// StopScrcpyForSerial kills every scrcpy process matching the device serial.
func StopScrcpyForSerial(serial string) {
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pids := findScrcpyPIDs(serial)
		if len(pids) == 0 {
			return
		}
		for _, pid := range pids {
			if proc, err := os.FindProcess(pid); err == nil {
				_ = proc.Kill()
			}
		}
		time.Sleep(80 * time.Millisecond)
	}
}

// KillScrcpyBySerial stops all scrcpy instances matched by device serial.
func KillScrcpyBySerial(serial string) { StopScrcpyForSerial(serial) }
