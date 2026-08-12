//go:build !windows

package monitor

import (
	"os/exec"
	"syscall"
)

// ProcessAlive reports whether cmd's process is still running.
func ProcessAlive(cmd *exec.Cmd) bool {
	if cmd == nil || cmd.Process == nil {
		return false
	}
	return cmd.Process.Signal(syscall.Signal(0)) == nil
}
