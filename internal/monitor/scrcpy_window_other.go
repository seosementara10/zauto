//go:build !windows

package monitor

func ScrcpyWindowRunning(string) bool { return false }
