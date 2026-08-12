package panel

import (
	"os"
	"path/filepath"
)

const panelDevFlagFile = ".panel-dev"

// PanelDevEnabled reports whether UI should load from disk with hot reload.
func PanelDevEnabled(projectRoot string) bool {
	if DevMode() {
		return true
	}
	if projectRoot == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(projectRoot, panelDevFlagFile))
	return err == nil
}

// EnablePanelDevFlag turns on persistent dev mode for this project directory.
func EnablePanelDevFlag(projectRoot string) error {
	if projectRoot == "" {
		return nil
	}
	return os.WriteFile(filepath.Join(projectRoot, panelDevFlagFile), []byte("1\n"), 0644)
}

func (s *Server) panelDevEnabled() bool {
	return PanelDevEnabled(s.ProjectRoot)
}
