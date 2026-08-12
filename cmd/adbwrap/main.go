package main

import (
	"os"
	"os/exec"

	"zauto/internal/executil"
)

func main() {
	real := resolveRealAdb()
	cmd := exec.Command(real, os.Args[1:]...)
	// Inherit stdin/stdout/stderr from parent (scrcpy reads adb output via pipe).
	executil.HideWindow(cmd)
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}

func resolveRealAdb() string {
	if v := os.Getenv("ZAUTO_REAL_ADB"); v != "" {
		return v
	}
	if p, err := exec.LookPath("adb"); err == nil {
		return p
	}
	return "adb"
}
