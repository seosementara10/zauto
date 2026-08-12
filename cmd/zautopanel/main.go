package main

import (
	"log"
	"os"
	"path/filepath"

	"zauto/internal/panel"
	"zauto/internal/projectroot"
)

func main() {
	hideConsoleWindow()
	root := projectroot.Find()
	setupFileLog(root)

	cfg := projectroot.ResolveConfig("config/config.json")
	log.Printf("zauto Panel (Wails) — %s", cfg)
	if err := panel.RunDesktop(panel.Options{
		ProjectRoot: root,
		ConfigPath:  cfg,
	}); err != nil {
		log.Printf("panel error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func setupFileLog(root string) {
	logDir := filepath.Join(root, "logs")
	_ = os.MkdirAll(logDir, 0755)
	f, err := os.OpenFile(filepath.Join(logDir, "panel-desktop.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	log.SetOutput(f)
}
