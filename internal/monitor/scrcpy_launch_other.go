//go:build !windows

package monitor

import (
	"fmt"
	"os/exec"

	"zauto/internal/executil"
	"zauto/internal/toolchain"
)

func launchScrcpyAt(scrcpy, scrcpyDir, serial string, hpNum int, tile WindowTile, opts Options) (*exec.Cmd, error) {
	args := scrcpyArgs(serial, hpNum, tile, opts)
	return launchScrcpyDirect(scrcpy, scrcpyDir, opts.ProjectRoot, args)
}

func launchScrcpyDirect(scrcpy, scrcpyDir, projectRoot string, args []string) (*exec.Cmd, error) {
	cmd := exec.Command(scrcpy, args...)
	cmd.Dir = scrcpyDir
	if env := toolchain.ScrcpyEnv(projectRoot, scrcpyDir); env != nil {
		cmd.Env = env
	}
	executil.HideWindow(cmd)
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

func KillScrcpyBySerial(string) {}

// ScrcpyRunningForSerial reports whether scrcpy is already running for a device.
func ScrcpyRunningForSerial(string) bool { return false }

// ScrcpyPIDForSerial returns one scrcpy PID for the device, if any.
func ScrcpyPIDForSerial(string) (int, bool) { return 0, false }

// DedupeScrcpyForSerial closes extra scrcpy windows for one device, keeping one instance.
func DedupeScrcpyForSerial(string) {}

// StopScrcpyForSerial kills every scrcpy process matching the device serial.
func StopScrcpyForSerial(string) {}
