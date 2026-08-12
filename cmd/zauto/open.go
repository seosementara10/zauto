package main

import "zauto/internal/panel"

func runOpenCLI(args []string) error {
	_ = args
	return panel.LaunchDesktopApp()
}
