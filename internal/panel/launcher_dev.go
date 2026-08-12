package panel

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"zauto/internal/projectroot"
)

// LaunchDesktopAppDev enables persistent dev mode, rebuilds if needed, and starts the panel.
func LaunchDesktopAppDev() error {
	root := projectroot.Find()
	if err := EnablePanelDevFlag(root); err != nil {
		return err
	}
	_ = os.Setenv("ZAUTO_PANEL_DEV", "1")
	if PanelInstanceActive() {
		fmt.Println("panel dev: restart…")
		stopDesktopPanel()
		time.Sleep(600 * time.Millisecond)
	}
	return startDesktopPanelProcess()
}

// ReloadDesktopApp rebuilds zautopanel.exe and restarts the panel (no build.ps1 needed).
func ReloadDesktopApp() error {
	root := projectroot.Find()
	_ = EnablePanelDevFlag(root)
	_ = os.Setenv("ZAUTO_PANEL_DEV", "1")
	fmt.Println("Rebuild zautopanel.exe…")
	build := exec.Command("go", "build", "-tags", "desktop,production", "-ldflags", "-w -s", "-o", "zautopanel.exe", "./cmd/zautopanel")
	build.Dir = root
	build.Env = append(os.Environ(), "CGO_ENABLED=1")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build zautopanel: %w", err)
	}
	cliBuild := exec.Command("go", "build", "-o", "zauto.exe", "./cmd/zauto")
	cliBuild.Dir = root
	cliBuild.Stdout = os.Stdout
	cliBuild.Stderr = os.Stderr
	_ = cliBuild.Run()

	if PanelInstanceActive() {
		stopDesktopPanel()
		time.Sleep(600 * time.Millisecond)
	}
	fmt.Println("Mode DEV aktif — UI auto reload saat file disimpan.")
	return startDesktopPanelProcess()
}

func startDesktopPanelProcess() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(self)
	panelExe := filepath.Join(dir, "zautopanel.exe")
	if _, err := os.Stat(panelExe); err != nil {
		return fmt.Errorf("zautopanel.exe tidak ditemukan — jalankan: zauto reload")
	}
	cmd := exec.Command(panelExe)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}
	root := projectroot.Find()
	if PanelDevEnabled(root) {
		fmt.Println("zauto Panel DEV — simpan file di internal/panel/web → auto reload (~1–2 detik)")
		fmt.Println("Badge DEV di footer = mode aktif. F5 = refresh manual.")
	} else {
		fmt.Println("zauto Panel dibuka — setelah edit UI jalankan: zauto reload")
	}
	waitForPanelReady(20 * time.Second)
	if FocusPanelWindow() {
		fmt.Println("zauto Panel siap.")
	}
	return nil
}
