package main

import (
	"flag"
	"fmt"

	"zauto/internal/panel"
)

func runDevCLI(args []string) error {
	fs := flag.NewFlagSet("dev", flag.ExitOnError)
	browser := fs.Bool("browser", false, "Panel di browser (HTTP :8765, hot reload, DevTools F12)")
	port := fs.Int("port", panel.DefaultPort, "Port HTTP saat --browser")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *browser {
		serveArgs := []string{fmt.Sprintf("-port=%d", *port), "-open=true"}
		return panel.ServeBrowserDev(serveArgs)
	}
	return panel.LaunchDesktopAppDev()
}
