package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"zauto/internal/adb"
	"zauto/internal/check"
	"zauto/internal/config"
	"zauto/internal/controller"
	"zauto/internal/farm"
	"zauto/internal/monitor"
	"zauto/internal/panel"
	"zauto/internal/projectroot"
	"zauto/internal/store"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			if err := panel.RunServeCLI(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "compose":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: zauto compose up [-d]")
				os.Exit(1)
			}
			if err := runComposeCommand(context.Background(), os.Args[2], os.Args[3:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "open":
			if err := runOpenCLI(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "dev":
			if err := runDevCLI(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "reload":
			if err := runReloadCLI(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}

	configPath := flag.String("config", "config/config.json", "Path to config JSON")
	dryRun := flag.Bool("dry-run", false, "List plan without executing")
	listDevices := flag.Bool("list-devices", false, "List connected devices")
	monitorScreens := flag.Bool("monitor", false, "Monitor all devices (unified browser dashboard)")
	monitorWindows := flag.Bool("monitor-windows", false, "Monitor via separate scrcpy windows")
	monitorPort := flag.Int("monitor-port", 8765, "Dashboard port for --monitor")
	panelMode := flag.Bool("panel", false, "Open zauto panel (toggle HP + run scripts)")
	farmMode := flag.Bool("farm", false, "Same as --panel")
	farmInstall := flag.Bool("farm-install", false, "Verify farm tools (scrcpy)")
	farmSTF := flag.Bool("farm-stf", false, "Show STF setup guide (scale 10+ devices)")
	noMirror := flag.Bool("no-mirror", false, "Skip auto mirror when running automation")
	noClear := flag.Bool("no-clear", false, "Keep Facebook session (no pm clear); resume from logged-in feed")
	checkSetup := flag.Bool("check", false, "Validate environment (ADB, devices, config)")
	maxDevices := flag.Int("max-devices", 0, "Override max_devices")
	workers := flag.Int("workers", 0, "Override parallel_workers")
	dbMigrate := flag.Bool("db-migrate", false, "Apply PostgreSQL schema (database.url)")
	dbImport := flag.String("db-import", "", "Import legacy accounts txt into PostgreSQL")
	dbAutoAssign := flag.Bool("db-auto-assign", false, "Assign account N to connected device N (slot 1)")
	flag.Parse()

	if !adb.CheckAvailable() {
		fmt.Fprintln(os.Stderr, "ADB not found. Install Android Platform Tools and add to PATH.")
		os.Exit(1)
	}

	root := projectroot.Find()
	cfg := projectroot.ResolveConfig(*configPath)

	if *dbMigrate || *dbImport != "" || *dbAutoAssign {
		wf, err := config.Load(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Config error:", err)
			os.Exit(1)
		}
		if err := runDatabaseOps(context.Background(), wf, *dbMigrate, *dbImport, *dbAutoAssign); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !*farmMode && !*panelMode {
			return
		}
	}

	if *farmInstall {
		if err := farm.Install(root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Siap. Jalankan: .\\up.ps1")
		return
	}

	if *farmSTF {
		farm.PrintSTFGuide()
		return
	}

	if *farmMode || *panelMode {
		if err := panel.LaunchDesktopApp(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if *monitorScreens || *monitorWindows {
		if *monitorWindows {
			if err := monitor.Run(monitor.DefaultOptions(root)); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			if err := monitor.RunDashboard(*monitorPort); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return
	}

	if *checkSetup {
		os.Exit(check.Run(root))
	}

	if *listDevices {
		devices, err := controller.ListConnectedDevices()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(devices) == 0 {
			fmt.Println("No devices connected.")
			return
		}
		fmt.Println("Connected devices:")
		for _, d := range devices {
			fmt.Printf("  %s\n", d)
		}
		return
	}

	wf, err := config.Load(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Config error:", err)
		os.Exit(1)
	}

	if *maxDevices > 0 {
		wf.MaxDevices = *maxDevices
	}
	if *workers > 0 {
		wf.ParallelWorkers = *workers
	}
	if *noClear {
		wf.ClearAppBeforeOpen = false
	}

	projectRoot := wf.ProjectRoot
	runID := time.Now().Format("20060102_150405")
	logDir := filepath.Join(projectRoot, "logs")

	log.Printf("Config: %s", cfg)
	log.Printf("Mode: %s", wf.Mode)

	devices, err := adb.ListDevices()
	if err != nil {
		log.Fatal(err)
	}
	serials, err := controller.FilterDevices(devices, wf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using %d device(s)", len(serials))

	controller.WarnMissingApps(wf, serials)
	controller.PrintPlan(wf, serials)

	if *dryRun {
		fmt.Println("\nDry run — no actions executed.")
		for _, s := range serials {
			fmt.Printf("  Would run on: %s\n", s)
		}
		for _, t := range wf.Tasks {
			fmt.Printf("  Task: %s (app=%s, actions=%d)\n", t.Name, t.App, len(t.Actions))
		}
		if len(devices) > len(serials) {
			fmt.Printf("\n  Mirror: %d HP (semua terhubung), automation: %d HP\n", len(devices), len(serials))
		}
		return
	}

	var mirrorProcs []*exec.Cmd
	var dashboard *monitor.Dashboard
	if wf.MirrorOnRun && !*noMirror {
		if wf.MirrorMode == "windows" {
			mirrorProcs, err = monitor.Start(devices, monitor.DefaultOptions(projectRoot), false)
			if err != nil {
				log.Printf("WARN: mirror windows: %v", err)
			} else {
				log.Printf("Mirror scrcpy: %d jendela", len(mirrorProcs))
			}
			defer monitor.Stop(mirrorProcs)
		} else {
			port := wf.MirrorPort
			if *monitorPort != 8765 {
				port = *monitorPort
			}
			dashboard, err = monitor.StartDashboard(port)
			if err != nil {
				log.Printf("WARN: dashboard: %v", err)
			}
			if dashboard != nil {
				defer dashboard.Stop()
			}
		}
	}

	ctrl := &controller.Controller{
		Workflow:    wf,
		ProjectRoot: projectRoot,
		LogDir:      logDir,
		RunID:       runID,
	}
	dbCtx, dbCancel := context.WithTimeout(context.Background(), store.DefaultConnectTimeout())
	st, err := store.OpenWorkflow(dbCtx, wf)
	dbCancel()
	if err != nil {
		log.Fatal("database: ", err)
	}
	defer st.Close()
	ctrl.Store = st

	results := ctrl.Run(context.Background(), serials)
	controller.PrintSummary(results)

	for _, res := range results {
		if len(res.Errors) > 0 {
			os.Exit(1)
		}
	}
}
