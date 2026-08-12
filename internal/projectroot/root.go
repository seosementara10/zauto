package projectroot

import (
	"os"
	"path/filepath"
)

// Find returns the zauto project directory (cwd or executable dir with config/config.json).
func Find() string {
	if cwd, err := os.Getwd(); err == nil {
		if hasConfig(cwd) {
			return cwd
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if hasConfig(dir) {
			return dir
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

// ResolveConfig turns a -config flag value into an absolute path.
func ResolveConfig(flagPath string) string {
	if filepath.IsAbs(flagPath) {
		return flagPath
	}
	return filepath.Join(Find(), flagPath)
}

func hasConfig(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "config", "config.json"))
	return err == nil
}
