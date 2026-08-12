package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/driver"
	"zauto/internal/pool"
	"zauto/internal/scheduler"
	"zauto/internal/store"
	"zauto/internal/worker"
)

// Controller orchestrates device discovery, pool, scheduler, and workers.
type Controller struct {
	Workflow    *config.Workflow
	ProjectRoot string
	LogDir      string
	RunID       string
	Store       *store.Store
}

// Run executes the workflow on all devices with bounded concurrency.
func (c *Controller) Run(ctx context.Context, serials []string) []worker.Result {
	if err := EnsureLogDir(c.LogDir); err != nil {
		log.Printf("WARN: log dir: %v", err)
	}
	p := pool.New(serials)
	sched := &scheduler.Scheduler{
		Workflow:    c.Workflow,
		Pool:        p,
		Workers:     c.Workflow.ParallelWorkers,
		ProjectRoot: c.ProjectRoot,
		LogDir:      c.LogDir,
		RunID:       c.RunID,
		Store:       c.Store,
	}
	return sched.Run(ctx)
}

// FilterDevices applies device_filter and max_devices from workflow config.
func FilterDevices(all []string, wf *config.Workflow) ([]string, error) {
	if len(all) == 0 {
		return nil, fmt.Errorf("no Android devices found — connect phones and run: adb devices")
	}
	filtered := all
	if len(wf.DeviceFilter) > 0 {
		allowed := map[string]bool{}
		for _, s := range wf.DeviceFilter {
			allowed[s] = true
		}
		filtered = nil
		for _, s := range all {
			if allowed[s] {
				filtered = append(filtered, s)
			}
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no devices matched device_filter")
	}
	if len(filtered) > wf.MaxDevices {
		filtered = filtered[:wf.MaxDevices]
	}
	return filtered, nil
}

// EnsureLogDir creates the logs directory if missing.
func EnsureLogDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// WarnMissingApps logs a warning for each device missing a configured app package.
func WarnMissingApps(wf *config.Workflow, serials []string) {
	for _, serial := range serials {
		client := driver.ClientForDevice(wf, serial)
		for name, pkg := range wf.Apps {
			if !client.IsPackageInstalled(pkg) {
				log.Printf("WARN [%s] Missing app: %s (%s)", serial, name, pkg)
			}
		}
	}
}

// PrintPlan shows run summary before execution.
func PrintPlan(wf *config.Workflow, serials []string) {
	driverName := wf.Automation.Driver
	if driverName == "" {
		driverName = "adb"
	}
	fmt.Printf("\nDevices : %d\n", len(serials))
	fmt.Printf("Tasks   : %d per loop\n", len(wf.Tasks))
	fmt.Printf("Loops   : %d\n", wf.LoopCount)
	fmt.Printf("Workers : %d\n", wf.ParallelWorkers)
	fmt.Printf("Driver  : %s\n", driverName)
}

// PrintSummary prints final results across devices.
func PrintSummary(results []worker.Result) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	totalErr := 0
	for _, r := range results {
		status := "OK"
		if len(r.Errors) > 0 {
			status = fmt.Sprintf("%d error(s)", len(r.Errors))
			totalErr += len(r.Errors)
		}
		fmt.Printf("  %s: loops=%d, tasks=%d, %s\n", r.Serial, r.LoopsCompleted, r.TasksCompleted, status)
	}
	fmt.Println(strings.Repeat("=", 60))
	if totalErr > 0 {
		fmt.Printf("Completed with %d total error(s).\n", totalErr)
	} else {
		fmt.Println("All devices completed successfully.")
	}
}

// ListConnectedDevices wraps adb.ListDevices for CLI use.
func ListConnectedDevices() ([]string, error) {
	return adb.ListDevices()
}
