// Package driver abstracts UI automation backends.
//
//   - ADB + uiautomator dump: lightweight, no Appium server
//   - Appium + UiAutomator2: stable find/click/fill for complex UI
package driver

import (
	"zauto/internal/adb"
	"zauto/internal/config"
)

// UI automates on-screen elements (find, tap, fill, wait).
type UI interface {
	Name() string
	TextExists(texts []string) (bool, error)
	WaitForTexts(texts []string, timeoutSec, pollSec float64) (bool, error)
	FindByTexts(texts []string, prefer string, minCenterY, maxCenterY int) (x, y int, ok bool)
	Tap(x, y int) error
	TapTexts(texts []string, prefer string, minCenterY, maxCenterY int) error
	FillField(fieldTexts []string, value string) error
	FillFieldByResource(resourceIDs []string, value string) error
	Swipe(x1, y1, x2, y2, durationMs int) error
	Close() error
}
// NewUI selects the automation backend from workflow config.
func NewUI(opts Options) (UI, error) {
	switch opts.Workflow.Automation.Driver {
	case "appium", "uiautomator2":
		return NewAppiumUI(opts)
	default:
		return NewAdbUI(opts)
	}
}

// Options passed when constructing a per-device UI driver.
type Options struct {
	Workflow    *config.Workflow
	Serial      string
	DeviceIndex int
	AppPackage  string
	AppActivity string
	ADBClient   *adb.Client
}
