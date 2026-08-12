package panel

import (
	"fmt"
	"os/exec"
	"time"

	"zauto/internal/projectroot"
)

// LaunchDesktopApp starts zautopanel.exe (must be built with Wails tags).
// If dev flag file exists, enables hot reload. Running instance is focused unless dev restart is needed.
func LaunchDesktopApp() error {
	root := projectroot.Find()
	dev := PanelDevEnabled(root)
	if PanelInstanceActive() {
		if dev {
			fmt.Println("zauto Panel DEV sudah berjalan — simpan file UI untuk auto reload (atau F5).")
			fmt.Println("Rebuild setelah edit Go: zauto reload")
		} else if FocusPanelWindow() {
			fmt.Println("zauto Panel sudah berjalan — jendela difokuskan.")
			fmt.Println("Tip: zauto dev = auto reload UI · zauto reload = rebuild + restart")
		} else {
			fmt.Println("zauto Panel sudah berjalan — cek taskbar.")
		}
		if FocusPanelWindow() {
			return nil
		}
	}
	return startDesktopPanelProcess()
}

func waitForPanelReady(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if PanelInstanceActive() && FocusPanelWindow() {
			return
		}
		time.Sleep(400 * time.Millisecond)
	}
}

func stopDesktopPanel() {
	if !PanelInstanceActive() {
		return
	}
	_ = exec.Command("taskkill", "/F", "/IM", "zautopanel.exe").Run()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !PanelInstanceActive() {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("panel: menunggu instance lama berhenti…")
}
