package main

import "zauto/internal/panel"

func runDevCLI(args []string) error {
	_ = args
	return panel.LaunchDesktopAppDev()
}
