package worker

import (
	"context"
	"fmt"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/driver"
	"zauto/internal/engine"
	"zauto/internal/engine/runtime"
	"zauto/internal/logging"
	"zauto/internal/pool"
	"zauto/internal/reset"
	"zauto/internal/runcontrol"
	"zauto/internal/state"
	"zauto/internal/store"
)

// Result of one device run.
type Result struct {
	Serial         string
	LoopsCompleted int
	TasksCompleted int
	Errors         []string
}

// RunDevice executes all loops/tasks on one device with retry policy.
func RunDevice(ctx context.Context, wf *config.Workflow, dev *pool.Device, projectRoot, logDir, runID string, st *store.Store) Result {
	result := Result{Serial: dev.Serial}
	lg := logging.ForDevice(logDir, dev.Serial, runID)
	defer lg.Close()

	client := driver.ClientForDevice(wf, dev.Serial)

	if client.IsEmulator() {
		res := adb.PrepareEmulator(client)
		adb.LogEmulatorPrep(res)
	}

	ui, err := driver.NewUI(driver.Options{
		Workflow:    wf,
		Serial:      dev.Serial,
		DeviceIndex: dev.Index,
		ADBClient:   client,
	})
	if err != nil {
		msg := fmt.Sprintf("UI driver init: %v", err)
		result.Errors = append(result.Errors, msg)
		lg.Error(msg)
		return result
	}
	defer ui.Close()

	session := engine.NewSession(client, ui, wf, projectRoot)
	session.Ctx = ctx
	session.Store = st
	session.Runtime["device_index"] = dev.Index
	session.DevLog = lg
	if err := bindAccount(ctx, st, session, dev, wf.Database.MaxAccountsPerDevice); err != nil {
		msg := fmt.Sprintf("account bind: %v", err)
		result.Errors = append(result.Errors, msg)
		lg.Error(msg)
		return result
	}
	accountTasks, err := resolveAccountTasks(ctx, st, session, wf)
	if err != nil {
		msg := fmt.Sprintf("account tasks: %v", err)
		result.Errors = append(result.Errors, msg)
		lg.Error(msg)
		return result
	}
	flow, _ := session.Runtime["automation_flow"].(string)
	lg.Info("Starting — driver=%s loops=%d account_flow=%s tasks=%d", ui.Name(), wf.LoopCount, flow, len(accountTasks))
	exec := &engine.Executor{Session: session}

	for loop := 1; loop <= wf.LoopCount; loop++ {
		if err := runcontrol.Default.WaitIfPaused(ctx, dev.Serial); err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result
		}
		lg.Info("=== Loop %d/%d ===", loop, wf.LoopCount)
		if wf.WakeScreenBeforeTask {
			client.WakeScreen()
		}
		for _, task := range accountTasks {
			if err := runcontrol.Default.WaitIfPaused(ctx, dev.Serial); err != nil {
				result.Errors = append(result.Errors, err.Error())
				return result
			}
			if err := prepareApp(ctx, client, wf, task, lg, session); err != nil {
				msg := fmt.Sprintf("Loop %d, task '%s': %v", loop, task.Name, err)
				result.Errors = append(result.Errors, msg)
				lg.Error(msg)
				continue
			}
			lg.Info("Task: %s", task.Name)
			var dbRunID int64
			if st != nil {
				if accID, ok := session.Runtime["account_id"].(int64); ok && accID > 0 {
					dbRunID, _ = st.StartRun(ctx, dev.Serial, accID, task.Name)
				}
			}
			taskErr := runWithRetry(exec, task, wf, lg)
			if st != nil && dbRunID > 0 {
				status, errMsg := "ok", ""
				if taskErr != nil {
					status, errMsg = "error", taskErr.Error()
				}
				_ = st.FinishRun(ctx, dbRunID, status, errMsg)
			}
			if taskErr != nil {
				msg := fmt.Sprintf("Loop %d, task '%s': %v", loop, task.Name, taskErr)
				result.Errors = append(result.Errors, msg)
				lg.Error(msg)
				if wf.ScreenshotOnError {
					observe, _ := exec.CachedObserve()
					shot, dump := exec.CaptureFailure(
						fmt.Sprintf("task_err_%s", task.Name),
						task.Name,
						taskErr.Error(),
						observe,
						runtime.ScreenNote{Detail: taskErr.Error()},
					)
					if shot != "" || dump != "" {
						lg.Warn("CAPTURE task error screenshot=%s hierarchy=%s", shot, dump)
					} else {
						lg.Warn("CAPTURE task error failed — no artifacts (see EVENT CAPTURE lines)")
					}
				}
				continue
			}
			result.TasksCompleted++
			lg.Info("Task completed: %s", task.Name)
			if err := teardownApp(client, wf, task, lg); err != nil {
				msg := fmt.Sprintf("Loop %d, task '%s' teardown: %v", loop, task.Name, err)
				result.Errors = append(result.Errors, msg)
				lg.Warn(msg)
			}
		}
		result.LoopsCompleted++
		if loop < wf.LoopCount && wf.DelayBetweenLoopsSec > 0 {
			time.Sleep(time.Duration(wf.DelayBetweenLoopsSec * float64(time.Second)))
		}
	}
	lg.Info("Finished — loops=%d tasks_ok=%d errors=%d", result.LoopsCompleted, result.TasksCompleted, len(result.Errors))
	return result
}

func runWithRetry(exec *engine.Executor, task config.Task, wf *config.Workflow, lg *logging.DeviceLogger) error {
	maxAttempts := wf.Engine.RetryMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	delay := wf.Engine.RetryDelaySec
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = exec.RunTask(task)
		if err == nil {
			return nil
		}
		if attempt < maxAttempts-1 {
			lg.Warn("RETRY %d/%d: %v", attempt+1, maxAttempts, err)
			if delay > 0 {
				time.Sleep(time.Duration(delay * float64(time.Second)))
			}
		}
	}
	return err
}

func prepareApp(ctx context.Context, client *adb.Client, wf *config.Workflow, task config.Task, lg *logging.DeviceLogger, session *engine.Session) error {
	if task.App == "" {
		return nil
	}
	pkg, ok := wf.Apps[task.App]
	if !ok {
		return fmt.Errorf("unknown app key: %s", task.App)
	}
	activity := wf.AppActivities[task.App]

	if shouldPmClear(wf, task) {
		lg.Info("Reset app %s (%s) — pm clear + detect initial state", task.App, pkg)
		d, err := reset.NewManager().ResetAndLaunch(ctx, reset.Input{
			Client:     client,
			Package:    pkg,
			Activity:   activity,
			Log:        lg,
			Poll:       state.PollIntervalFromSec(wf.Engine.PollSec),
		})
		if err != nil {
			return err
		}
		seedInitialState(session, d)
		return nil
	}

	if !client.IsPackageInstalled(pkg) {
		return fmt.Errorf("app not installed: %s (%s)", task.App, pkg)
	}
	if wf.ForceStopBeforeOpen {
		lg.Info("Force stop %s (%s)", task.App, pkg)
		_ = client.ForceStop(pkg)
		time.Sleep(time.Duration(wf.DelayAfterForceStopSec * float64(time.Second)))
	}
	lg.Info("Open %s (%s) — keep session (no pm clear)", task.App, pkg)
	if err := client.OpenApp(pkg, activity); err != nil {
		return err
	}
	d, err := reset.WaitInitialUIState(ctx, client, 25*time.Second, state.PollIntervalFromSec(wf.Engine.PollSec), true)
	if err != nil {
		return err
	}
	if d.State == state.UILoggedIn {
		lg.Info("INITIAL_STATE logged_in confidence=%.0f%% (resume — login will be skipped)", d.Confidence*100)
	} else {
		lg.Info("INITIAL_STATE %s confidence=%.0f%%", d.State, d.Confidence*100)
	}
	seedInitialState(session, d)
	return nil
}

// shouldPmClear decides whether to wipe app data before open.
// Per-account pipeline pm_clear step takes precedence; global clear_app_before_open applies only
// when the task has no custom pipeline steps (CLI/config tasks without DB account).
func shouldPmClear(wf *config.Workflow, task config.Task) bool {
	if config.PipelineIncludesPmClear(task.Flow, task.Params) {
		return true
	}
	if config.ParamsHasSteps(task.Params) {
		return false
	}
	return wf.ClearAppBeforeOpen
}

func seedInitialState(session *engine.Session, d state.Detection) {
	if session == nil {
		return
	}
	if session.Memory == nil {
		session.Memory = &state.DeviceMemory{}
	}
	session.Memory.SetDetection(d)
}

func teardownApp(client *adb.Client, wf *config.Workflow, task config.Task, lg *logging.DeviceLogger) error {
	if !wf.ForceStopAfterTask || task.App == "" {
		return nil
	}
	pkg, ok := wf.Apps[task.App]
	if !ok || pkg == "" {
		return nil
	}
	lg.Info("Close app %s (%s)", task.App, pkg)
	if err := client.ForceStop(pkg); err != nil {
		return err
	}
	if wf.DelayAfterForceStopSec > 0 {
		time.Sleep(time.Duration(wf.DelayAfterForceStopSec * float64(time.Second)))
	}
	return nil
}

func bindAccount(ctx context.Context, st *store.Store, session *engine.Session, dev *pool.Device, maxAccounts int) error {
	if st == nil {
		return fmt.Errorf("database not configured — set database.url in config")
	}
	if maxAccounts <= 0 {
		maxAccounts = 50
	}
	if _, err := st.UpsertDevice(ctx, dev.Serial, dev.Serial, dev.Index, maxAccounts); err != nil {
		return err
	}
	acc, slotID, err := st.GetNextAccountForDevice(ctx, dev.Serial)
	if err != nil {
		return err
	}
	session.Runtime["account_id"] = acc.ID
	session.Runtime["slot_id"] = slotID
	session.Runtime["login_id"] = acc.LoginID()
	session.Runtime["account_name"] = acc.Name
	session.Runtime["fanpages"] = acc.Fanpages
	session.Runtime["automation_flow"] = acc.AutomationFlow
	return st.TouchAssignment(ctx, slotID)
}

func resolveAccountTasks(ctx context.Context, st *store.Store, session *engine.Session, wf *config.Workflow) ([]config.Task, error) {
	if st == nil {
		return wf.Tasks, nil
	}
	accID, ok := session.Runtime["account_id"].(int64)
	if !ok || accID <= 0 {
		return nil, fmt.Errorf("no account bound to device")
	}
	acc, err := st.GetAccountByID(ctx, accID)
	if err != nil {
		return nil, err
	}
	return TasksForAccount(acc)
}
