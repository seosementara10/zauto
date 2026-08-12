package main

import "zauto/internal/panel"

func runReloadCLI(args []string) error {
	_ = args
	return panel.ReloadDesktopApp()
}
