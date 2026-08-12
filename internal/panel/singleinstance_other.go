//go:build !windows

package panel

func acquirePanelInstance() (alreadyRunning bool, release func(), err error) {
	return false, func() {}, nil
}

// PanelInstanceActive is always false on non-Windows platforms.
func PanelInstanceActive() bool { return false }
