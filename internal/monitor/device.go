package monitor

import (
	"os/exec"
)

// StartOneAt opens one scrcpy window at an exact tile position.
func StartOneAt(serial string, hpNum int, tile WindowTile, opts Options) (*exec.Cmd, error) {
	scrcpy, err := FindScrcpy(opts.ProjectRoot)
	if err != nil {
		return nil, err
	}
	return StartOneAtWith(scrcpy, ScrcpyDir(scrcpy), serial, hpNum, tile, opts)
}

// StartOneAtWith starts scrcpy using a pre-resolved binary path (avoids repeated lookups).
func StartOneAtWith(scrcpy, scrcpyDir, serial string, hpNum int, tile WindowTile, opts Options) (*exec.Cmd, error) {
	return launchScrcpyAt(scrcpy, scrcpyDir, serial, hpNum, tile, opts)
}
