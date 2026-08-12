//go:build windows

package panel

import (
	"fmt"

	"golang.org/x/sys/windows"
)

const panelMutexName = "Global\\zauto-panel-single-instance"

func acquirePanelInstance() (alreadyRunning bool, release func(), err error) {
	name, err := windows.UTF16PtrFromString(panelMutexName)
	if err != nil {
		return false, nil, err
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		if errno, ok := err.(windows.Errno); ok && errno == windows.ERROR_ALREADY_EXISTS {
			return true, nil, nil
		}
		return false, nil, fmt.Errorf("panel mutex: %w", err)
	}
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(h)
		return true, nil, nil
	}
	return false, func() { _ = windows.CloseHandle(h) }, nil
}

// PanelInstanceActive reports whether a desktop panel instance holds the global mutex.
func PanelInstanceActive() bool {
	name, err := windows.UTF16PtrFromString(panelMutexName)
	if err != nil {
		return false
	}
	h, err := windows.OpenMutex(windows.SYNCHRONIZE, false, name)
	if err != nil {
		return false
	}
	_ = windows.CloseHandle(h)
	return true
}
