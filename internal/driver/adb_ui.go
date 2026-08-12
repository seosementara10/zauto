package driver

import (
	"fmt"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/state"
	"zauto/internal/ui"
)

// AdbUI uses ADB shell uiautomator dump + local XML parsing (no Appium server).
type AdbUI struct {
	client   *adb.Client
	resolver *ui.Resolver
	poll     time.Duration
}

func NewAdbUI(opts Options) (*AdbUI, error) {
	client := opts.ADBClient
	if client == nil {
		client = &adb.Client{
			Serial:     opts.Serial,
			Timeout:    time.Duration(opts.Workflow.AdbTimeoutSec) * time.Second,
			Retries:    opts.Workflow.AdbRetries,
			DumpSettle: 30 * time.Millisecond,
		}
	}
	poll := state.DefaultPollInterval
	if opts.Workflow != nil {
		poll = state.PollIntervalFromSec(opts.Workflow.Engine.PollSec)
	}
	return &AdbUI{client: client, resolver: ui.NewDefaultResolver(), poll: poll}, nil
}

func (d *AdbUI) Name() string { return "adb-uiautomator-dump" }

func (d *AdbUI) read(fast bool) ui.Snapshot {
	xml, _ := d.client.DumpUI(fast)
	return ui.ParseHierarchy(xml)
}

func (d *AdbUI) TextExists(texts []string) (bool, error) {
	return d.resolver.TextExists(d.read(true), texts), nil
}

func (d *AdbUI) WaitForTexts(texts []string, timeoutSec, pollSec float64) (bool, error) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		if d.resolver.TextExists(d.read(true), texts) {
			return true, nil
		}
		if pollSec > 0 {
			time.Sleep(time.Duration(pollSec * float64(time.Second)))
		} else {
			time.Sleep(d.poll)
		}
	}
	return false, nil
}

func (d *AdbUI) FindByTexts(texts []string, prefer string, minCenterY, maxCenterY int) (int, int, bool) {
	r := d.resolver.Find(d.read(true), ui.FindQuery{
		Texts: texts, Prefer: prefer, PreferClickable: true,
		MinCenterY: minCenterY, MaxCenterY: maxCenterY,
	})
	if r == nil {
		return 0, 0, false
	}
	x, y := r.Center()
	return x, y, true
}

func (d *AdbUI) Tap(x, y int) error {
	return d.client.Tap(x, y)
}

func (d *AdbUI) TapTexts(texts []string, prefer string, minCenterY, maxCenterY int) error {
	x, y, ok := d.FindByTexts(texts, prefer, minCenterY, maxCenterY)
	if !ok {
		return fmt.Errorf("element not found: %v", texts)
	}
	return d.client.Tap(x, y)
}

func (d *AdbUI) FillField(fieldTexts []string, value string) error {
	snap := d.read(false)
	r := ui.FindInputField(d.resolver, snap, ui.FindQuery{Texts: fieldTexts})
	if r == nil {
		return fmt.Errorf("field not found: %v", fieldTexts)
	}
	return d.tapAndType(r, value)
}

func (d *AdbUI) FillFieldByResource(resourceIDs []string, value string) error {
	snap := d.read(false)
	r := ui.FindInputField(d.resolver, snap, ui.FindQuery{ResourceIDs: resourceIDs})
	if r == nil {
		return fmt.Errorf("field not found by resource-id: %v", resourceIDs)
	}
	return d.tapAndType(r, value)
}

func (d *AdbUI) Swipe(x1, y1, x2, y2, durationMs int) error {
	return d.client.Swipe(x1, y1, x2, y2, durationMs)
}

func (d *AdbUI) Close() error { return nil }

// ADB returns the underlying client for device-level ops (install, force-stop, etc.).
func (d *AdbUI) ADB() *adb.Client { return d.client }

// ClientForDevice builds an ADB client with optional user profile detection.
func ClientForDevice(wf *config.Workflow, serial string) *adb.Client {
	userID := wf.AdbUser
	if userID == nil && wf.AutoDetectUser {
		for _, pkg := range wf.Apps {
			userID = adb.FindUserForPackage(serial, pkg)
			if userID != nil {
				break
			}
		}
	}
	return &adb.Client{
		Serial:     serial,
		Timeout:    time.Duration(wf.AdbTimeoutSec) * time.Second,
		Retries:    wf.AdbRetries,
		UserID:     userID,
		DumpSettle: 30 * time.Millisecond,
	}
}
