package uiutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// OpenAppWindow opens url as a standalone browser app window (monitor dashboard only).
func OpenAppWindow(url string) {
	OpenAppWindowAt(url, 1400, 900, -1, -1)
}

// OpenAppWindowAt opens a sized app window via Edge/Chrome --app mode.
func OpenAppWindowAt(url string, width, height, x, y int) {
	if runtime.GOOS != "windows" {
		_ = exec.Command("xdg-open", url).Start()
		return
	}
	profileDir := filepath.Join(os.TempDir(), "zauto-monitor-profile")
	for _, bin := range windowsBrowsers() {
		if bin == "" {
			continue
		}
		args := []string{
			"--app=" + url,
			"--user-data-dir=" + profileDir,
			fmt.Sprintf("--window-size=%d,%d", width, height),
			"--no-first-run",
			"--no-default-browser-check",
		}
		if x >= 0 && y >= 0 {
			args = append(args, fmt.Sprintf("--window-position=%d,%d", x, y))
		}
		if err := exec.Command(bin, args...).Start(); err == nil {
			return
		}
	}
	_ = exec.Command("cmd", "/c", "start", "", url).Start()
}

func windowsBrowsers() []string {
	var list []string
	for _, p := range []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
	} {
		if _, err := os.Stat(p); err == nil {
			list = append(list, p)
		}
	}
	return list
}
