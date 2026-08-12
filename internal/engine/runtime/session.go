package runtime

import (
	"context"
	"path/filepath"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/driver"
	"zauto/internal/logging"
	"zauto/internal/state"
	"zauto/internal/store"
	"zauto/internal/ui"
)

// Session holds per-device context: ADB for device ops, UI driver for automation.
type Session struct {
	Client      *adb.Client
	UI          driver.UI
	Workflow    *config.Workflow
	Serial      string
	ProjectRoot string
	Resolver    *ui.Resolver
	Runtime     map[string]interface{}
	Memory      *state.DeviceMemory
	DevLog      *logging.DeviceLogger
	Ctx         context.Context
	Store       *store.Store
}

func NewSession(client *adb.Client, uiDriver driver.UI, wf *config.Workflow, projectRoot string) *Session {
	return &Session{
		Client:      client,
		UI:          uiDriver,
		Workflow:    wf,
		Serial:      client.Serial,
		ProjectRoot: projectRoot,
		Resolver:    ui.NewDefaultResolver(),
		Runtime:     map[string]interface{}{},
	}
}

func (s *Session) ReadUI(fast bool) ui.Snapshot {
	xml, _ := s.Client.DumpUI(fast)
	return ui.ParseHierarchy(xml)
}

func (s *Session) ScreenSize() (int, int) {
	return s.Client.ScreenSize()
}

func (s *Session) ScrollCoords(key string) config.ScrollCoords {
	w, h := s.ScreenSize()
	profiles := s.Workflow.ScreenProfiles
	if profiles != nil {
		if prof, ok := profiles[s.Workflow.ScreenProfile]; ok {
			if c, ok := prof[key]; ok {
				return c
			}
		}
		if c, ok := profiles["default"][key]; ok {
			return c
		}
	}
	if key == "scroll_down" {
		return config.ScrollCoords{X1: w / 2, Y1: h * 70 / 100, X2: w / 2, Y2: h * 25 / 100, DurationMs: 300}
	}
	return config.ScrollCoords{X1: w / 2, Y1: h * 25 / 100, X2: w / 2, Y2: h * 70 / 100, DurationMs: 300}
}

func (s *Session) ResolvePath(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(s.ProjectRoot, rel)
}

func (s *Session) WaitForTexts(texts []string, timeoutSec, pollSec float64) bool {
	ok, _ := s.UI.WaitForTexts(texts, timeoutSec, pollSec)
	return ok
}

func (s *Session) Locate(q ui.FindQuery, fast bool) *ui.Resolved {
	x, y, ok := s.UI.FindByTexts(q.Texts, q.Prefer, q.MinCenterY, q.MaxCenterY)
	if !ok {
		snap := s.ReadUI(fast)
		return s.Resolver.Find(snap, q)
	}
	_ = x
	_ = y
	return &ui.Resolved{
		Element: ui.Element{Text: q.Texts[0], Bounds: [4]int{x, y, x + 1, y + 1}, Clickable: true},
		Label:   q.Texts[0],
		Bounds:  [4]int{x, y, x + 1, y + 1},
	}
}

func (s *Session) TapTexts(texts []string, prefer string, minY, maxY int) error {
	return s.UI.TapTexts(texts, prefer, minY, maxY)
}

func (s *Session) FillField(fields []string, value string) error {
	return s.UI.FillField(fields, value)
}

func (s *Session) FillFieldByResource(resourceIDs []string, value string) error {
	return s.UI.FillFieldByResource(resourceIDs, value)
}

func (s *Session) Verify(spec *config.VerifySpec) bool {
	if spec == nil || len(spec.Texts) == 0 {
		return true
	}
	timeout := spec.TimeoutSec
	if timeout <= 0 {
		timeout = 15
	}
	pollSec := spec.PollSec
	if pollSec <= 0 {
		pollSec = s.PollInterval().Seconds()
	}
	ok, _ := s.UI.WaitForTexts(spec.Texts, timeout, pollSec)
	return ok
}

func (s *Session) Pause(seconds float64) {
	if seconds <= 0 {
		return
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}

// PollInterval returns UI observe poll tick from engine.poll_sec (0 → 400ms default).
func (s *Session) PollInterval() time.Duration {
	if s == nil || s.Workflow == nil {
		return state.DefaultPollInterval
	}
	return state.PollIntervalFromSec(s.Workflow.Engine.PollSec)
}

// SettleAfterAction applies configured post-action pacing when an explicit delay is intended.
func (s *Session) SettleAfterAction() {
	sec := s.Workflow.Engine.PostActionDelay
	if sec <= 0 {
		sec = s.Workflow.DelayBetweenActionsSec
	}
	if sec > 0 {
		time.Sleep(time.Duration(sec * float64(time.Second)))
	}
}
