package driver

import (
	"fmt"
	"strings"
	"time"

	"zauto/internal/adb"
	"zauto/internal/appium"
)

// AppiumUI drives UI through Appium server with UiAutomator2 backend.
type AppiumUI struct {
	adb    *adb.Client
	appium *appium.Client
}

func NewAppiumUI(opts Options) (*AppiumUI, error) {
	wf := opts.Workflow
	ap := wf.Automation.Appium
	serverURL := ap.ServerURL
	if ap.PortPerDevice {
		serverURL = appium.ServerURL(ap.ServerURL, opts.DeviceIndex, ap.PortBase, ap.PortStride)
	}
	client := appium.NewClient(appium.SessionOptions{
		ServerURL:      serverURL,
		Serial:         opts.Serial,
		AutomationName: ap.AutomationName,
		AppPackage:     opts.AppPackage,
		AppActivity:    opts.AppActivity,
		NoReset:        ap.NoReset,
		AutoGrantPerms: ap.AutoGrantPermissions,
	})
	if err := client.CreateSession(appium.SessionOptions{
		ServerURL:      serverURL,
		Serial:         opts.Serial,
		AutomationName: ap.AutomationName,
		AppPackage:     opts.AppPackage,
		AppActivity:    opts.AppActivity,
		NoReset:        ap.NoReset,
		AutoGrantPerms: ap.AutoGrantPermissions,
	}); err != nil {
		return nil, err
	}
	return &AppiumUI{
		adb:    ClientForDevice(wf, opts.Serial),
		appium: client,
	}, nil
}

func (d *AppiumUI) Name() string { return "appium-uiautomator2" }

func (d *AppiumUI) TextExists(texts []string) (bool, error) {
	src, err := d.appium.PageSource()
	if err != nil {
		return false, err
	}
	lower := strings.ToLower(src)
	for _, t := range texts {
		if t != "" && strings.Contains(lower, strings.ToLower(t)) {
			return true, nil
		}
	}
	return false, nil
}

func (d *AppiumUI) WaitForTexts(texts []string, timeoutSec, pollSec float64) (bool, error) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	if pollSec <= 0 {
		pollSec = 0.5
	}
	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for time.Now().Before(deadline) {
		ok, err := d.TextExists(texts)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		time.Sleep(time.Duration(pollSec * float64(time.Second)))
	}
	return false, nil
}

func (d *AppiumUI) FindByTexts(texts []string, prefer string, minCenterY, maxCenterY int) (int, int, bool) {
	_, err := d.appium.FindByTexts(texts)
	if err != nil {
		return 0, 0, false
	}
	// Appium clicks by element id; coordinates unused
	return 0, 0, true
}

func (d *AppiumUI) Tap(x, y int) error {
	if x == 0 && y == 0 {
		return fmt.Errorf("use TapTexts with appium driver")
	}
	return d.appium.Tap(x, y)
}

func (d *AppiumUI) TapTexts(texts []string, prefer string, minCenterY, maxCenterY int) error {
	id, err := d.appium.FindByTexts(texts)
	if err != nil {
		return err
	}
	return d.appium.Click(id)
}

func (d *AppiumUI) FillField(fieldTexts []string, value string) error {
	id, err := d.appium.FindByTexts(fieldTexts)
	if err != nil {
		return err
	}
	if err := d.appium.Click(id); err != nil {
		return err
	}
	return d.appium.SendKeys(id, value)
}

func (d *AppiumUI) FillFieldByResource(resourceIDs []string, value string) error {
	return fmt.Errorf("field not found by resource-id: %v", resourceIDs)
}

func (d *AppiumUI) Swipe(x1, y1, x2, y2, durationMs int) error {
	return d.appium.Swipe(x1, y1, x2, y2, durationMs)
}

func (d *AppiumUI) Close() error {
	return d.appium.DeleteSession()
}

func (d *AppiumUI) ADB() *adb.Client { return d.adb }
