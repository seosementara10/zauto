//go:build windows

package monitor

import (
	"os/exec"
	"syscall"
)

const (
	stillActive                      = 259 // STILL_ACTIVE
	processQueryLimitedInformation   = 0x1000
)

// ProcessAlive reports whether cmd's process is still running.
func ProcessAlive(cmd *exec.Cmd) bool {
	if cmd == nil || cmd.Process == nil {
		return false
	}
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(cmd.Process.Pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)
	var code uint32
	if err := syscall.GetExitCodeProcess(handle, &code); err != nil {
		return false
	}
	return code == stillActive
}
