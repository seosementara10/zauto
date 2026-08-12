package engine

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"zauto/internal/config"
	"zauto/internal/data"
	"zauto/internal/engine/fanpage"
	"zauto/internal/engine/login"
	"zauto/internal/engine/logout"
	"zauto/internal/engine/overlay"
	"zauto/internal/engine/post"
	"zauto/internal/state"
	"zauto/internal/ui"
)

type Executor struct {
	Session *Session
}

func (e *Executor) RunTask(task config.Task) error {
	for _, action := range task.Actions {
		if err := e.runAction(action); err != nil {
			if action.Optional {
				log.Printf("[%s] optional action skipped: %s — %v", e.Session.Serial, action.Type, err)
				continue
			}
			return fmt.Errorf("%s: %w", action.Type, err)
		}
	}
	return nil
}

func (e *Executor) runAction(action config.Action) error {
	switch strings.ToLower(action.Type) {
	case "wait":
		sec := action.ParamFloat("seconds", 1)
		time.Sleep(time.Duration(sec * float64(time.Second)))
		return nil
	case "wait_for_text":
		texts := config.ActionTexts(action)
		pollSec := action.ParamFloat("poll_sec", 0)
		if pollSec <= 0 {
			pollSec = e.Session.PollInterval().Seconds()
		}
		if !e.Session.WaitForTexts(texts, action.ParamFloat("timeout_sec", 30), pollSec) {
			return fmt.Errorf("wait_for_text timeout: %v", texts)
		}
		return nil
	case "tap_text":
		return e.tapText(action)
	case "tap":
		x, y := action.ParamInt("x", 0), action.ParamInt("y", 0)
		return e.Session.Client.Tap(x, y)
	case "tap_create_account":
		return e.tapCreateAccount(action)
	case "fill_name":
		return e.fillName(action)
	case "fill_email":
		return e.fillEmail(action)
	case "set_birthday":
		return e.setBirthday(action)
	case "set_gender":
		return e.setGender(action)
	case "facebook_login":
		return e.facebookLogin(action)
	case "facebook_logout":
		return e.facebookLogout(action)
	case "facebook_auto_post":
		return e.facebookAutoPost(action)
	case "facebook_fanpage_post":
		return e.facebookFanpagePost(action)
	case "force_stop_app":
		return e.forceStopApp(action)
	case "handle_permission":
		return e.handlePermission(action)
	case "scroll_down":
		c := e.Session.ScrollCoords("scroll_down")
		for i := 0; i < action.ParamInt("times", 1); i++ {
			_ = e.Session.Client.Swipe(c.X1, c.Y1, c.X2, c.Y2, c.DurationMs)
		}
		return nil
	case "scroll_up":
		c := e.Session.ScrollCoords("scroll_up")
		for i := 0; i < action.ParamInt("times", 1); i++ {
			_ = e.Session.Client.Swipe(c.X1, c.Y1, c.X2, c.Y2, c.DurationMs)
		}
		return nil
	case "press":
		return e.Session.Client.KeyEvent(action.ParamString("key", "HOME"))
	case "screenshot":
		name := action.ParamString("filename", fmt.Sprintf("%s_%d.png", e.Session.Serial, time.Now().Unix()))
		return e.Session.Client.Screenshot(filepath.Join("screenshots", name))
	case "dump_ui":
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *Executor) tapText(action config.Action) error {
	texts := config.ActionTexts(action)
	timeout := action.ParamFloat("timeout_sec", 25)
	if action.Optional {
		timeout = action.ParamFloat("timeout_sec", 0.6)
	}
	prefer := action.ParamString("prefer", "first")
	maxY := action.ParamInt("max_center_y", 0)
	minY := action.ParamInt("min_center_y", 0)

	if err := e.waitTapTexts(texts, prefer, minY, maxY, timeout); err != nil {
		if action.Optional {
			return nil
		}
		return err
	}
	if actionVerify(action) != nil && !action.ParamBool("skip_verify", false) {
		if !e.Session.Verify(actionVerify(action)) {
			return fmt.Errorf("verification failed after tap_text")
		}
	}
	return nil
}

func (e *Executor) tapCreateAccount(action config.Action) error {
	serial := e.Session.Serial
	create := []string{"Create new account", "Create New Account", "Buat Akun Baru", "Buat akun baru"}
	join := []string{"Join Facebook", "Gabung Facebook", "Create an account to connect"}
	nameForm := []string{"What's your name?", "First name", "Last name", "SKIP", "Skip"}

	if err := e.waitTapTexts(create, "bottom", 0, 0, action.ParamFloat("timeout_sec", 20)); err != nil {
		return fmt.Errorf("Create account button not found")
	}
	log.Printf("[%s] tap_create_account: bottom tap", serial)

	if !e.Session.WaitForTexts(join, 5, 0) {
		return fmt.Errorf("Join Facebook screen not reached")
	}
	if err := e.waitTapTexts(create, "first", 0, 2000, 8); err != nil {
		return fmt.Errorf("second Create account not found")
	}
	log.Printf("[%s] tap_create_account: upper tap OK", serial)

	if !e.Session.WaitForTexts(nameForm, 5, 0) {
		return fmt.Errorf("Name screen not reached")
	}
	return nil
}

func (e *Executor) fillName(action config.Action) error {
	wf := e.Session.Workflow
	path := e.Session.ResolvePath(action.ParamString("names_file", wf.RegString("names_file", "data/names.txt")))
	idx := action.ParamInt("name_index", wf.RegInt("name_index", 0))
	entry, err := data.GetName(path, idx)
	if err != nil {
		return err
	}
	e.Session.Runtime["gender"] = entry.Gender
	form := []string{"What's your name?", "First name", "Last name"}
	if !e.Session.WaitForTexts(form, action.ParamFloat("form_timeout_sec", 5), 0) {
		return fmt.Errorf("name form not found")
	}
	if err := e.Session.FillField([]string{"Last name", "Last name,"}, entry.Last); err != nil {
		return err
	}
	if err := e.Session.FillField([]string{"First name", "First name,"}, entry.First); err != nil {
		return err
	}
	_ = e.Session.Client.KeyEvent("BACK")
	return e.waitTapTexts([]string{"Next", "Berikutnya"}, "first", 0, 0, 8)
}

func (e *Executor) fillEmail(action config.Action) error {
	wf := e.Session.Workflow
	path := e.Session.ResolvePath(action.ParamString("emails_file", wf.RegString("emails_file", "data/email.txt")))
	idx := action.ParamInt("email_index", wf.RegInt("email_index", 0))
	email, err := data.GetEmail(path, idx)
	if err != nil {
		return err
	}
	form := []string{"What's your email?", "Enter the email where you can be contacted", "Sign up with mobile number"}
	fields := []string{"Email,", "Email", "Email address", "Alamat email"}
	if !e.Session.WaitForTexts(form, action.ParamFloat("form_timeout_sec", 8), 0) {
		return fmt.Errorf("email form not found")
	}
	if err := e.Session.FillField(fields, email); err != nil {
		return err
	}
	_ = e.Session.Client.KeyEvent("BACK")
	log.Printf("[%s] fill_email: %s", e.Session.Serial, email)
	return e.waitTapTexts([]string{"Next", "Berikutnya"}, "first", 0, 0, 8)
}

func (e *Executor) setBirthday(action config.Action) error {
	serial := e.Session.Serial
	minAge := action.ParamInt("min_age", 21)
	maxAge := action.ParamInt("max_age", 49)
	yearSwipes := action.ParamInt("year_swipes", 0)
	if yearSwipes <= 0 {
		yearSwipes = (minAge + maxAge) / 2 / 5
		if yearSwipes < 3 {
			yearSwipes = 3
		}
	}
	monthSwipes := action.ParamInt("month_swipes", 2)
	daySwipes := action.ParamInt("day_swipes", 2)

	if !e.Session.WaitForTexts([]string{"What's your birthday?", "Birthday", "Set date"}, 5, 0) {
		field := e.Session.Locate(ui.FindQuery{
			Texts: []string{"Birthday", "Set date", "DD", "MM", "YYYY"}, PreferClickable: true, MinCenterY: 500,
		}, true)
		if field != nil {
			x, y := field.Center()
			_ = e.Session.Client.Tap(x, y)
			time.Sleep(200 * time.Millisecond)
		}
	}

	cols := findPickerColumns(e.Session.ReadUI(true))
	if len(cols) < 3 {
		return fmt.Errorf("date picker columns not found")
	}
	log.Printf("[%s] set_birthday rough swipes year=%d month=%d day=%d", serial, yearSwipes, monthSwipes, daySwipes)
	for i := 0; i < yearSwipes; i++ {
		hardSwipe(e.Session, cols[0], true)
	}
	for i := 0; i < monthSwipes; i++ {
		hardSwipe(e.Session, cols[1], rand.Intn(2) == 0)
	}
	for i := 0; i < daySwipes; i++ {
		hardSwipe(e.Session, cols[2], rand.Intn(2) == 0)
	}

	setBtn := e.Session.Locate(ui.FindQuery{Texts: []string{"Set"}, ResourceIDs: []string{"android:id/button1"}}, true)
	if setBtn != nil {
		x, y := setBtn.Center()
		_ = e.Session.Client.Tap(x, y)
	}
	snap := e.Session.ReadUI(true)
	next := e.Session.Resolver.Find(snap, ui.FindQuery{Texts: []string{"Next", "Berikutnya"}, PreferClickable: true})
	if next != nil {
		x, y := next.Center()
		_ = e.Session.Client.Tap(x, y)
	}
	return nil
}

func (e *Executor) setGender(action config.Action) error {
	gender, _ := e.Session.Runtime["gender"].(string)
	if gender == "" {
		wf := e.Session.Workflow
		path := e.Session.ResolvePath(wf.RegString("names_file", "data/names.txt"))
		idx := wf.RegInt("name_index", 0)
		entry, err := data.GetName(path, idx)
		if err != nil {
			return err
		}
		gender = entry.Gender
	}
	var labels []string
	if gender == "male" {
		labels = []string{"Male", "Laki-laki", "Pria"}
	} else {
		labels = []string{"Female", "Perempuan", "Wanita"}
	}
	if !e.Session.WaitForTexts(labels, 4, 0) {
		return fmt.Errorf("gender screen not found")
	}
	if err := e.waitTapTexts(labels, "first", 0, 0, 5); err != nil {
		return fmt.Errorf("gender button not found")
	}
	if err := e.waitTapTexts([]string{"Next", "Berikutnya"}, "first", 0, 0, 8); err != nil {
		return fmt.Errorf("Next after gender not found")
	}
	return nil
}

func (e *Executor) facebookLogin(action config.Action) error {
	return login.Run(e, action)
}

func (e *Executor) fillLoginField(labels, resourceIDs []string, value string, editIndex int) error {
	resolver := e.Session.Resolver
	snap := e.Session.ReadUI(true)

	if editIndex == 0 {
		if r := ui.FindLoginEmailField(resolver, snap, labels, resourceIDs); r != nil && ui.FieldHasExpectedValue(r.Element, value) {
			return nil
		}
	} else if editIndex == 1 {
		if typed, _ := e.Session.Runtime["login_password_typed"].(bool); typed {
			return nil
		}
	}

	if err := e.Session.FillFieldByResource(resourceIDs, value); err == nil {
		if editIndex == 1 && !overlay.ShouldMarkPasswordTyped(e) {
			return fmt.Errorf("password fill blocked by keyboard overlay")
		}
		e.markLoginPasswordTyped(editIndex)
		return nil
	}
	if err := e.Session.FillField(labels, value); err == nil {
		if editIndex == 1 && !overlay.ShouldMarkPasswordTyped(e) {
			return fmt.Errorf("password fill blocked by keyboard overlay")
		}
		e.markLoginPasswordTyped(editIndex)
		return nil
	}
	if err := e.tapLoginFieldAndType(resolver, labels, resourceIDs, value, editIndex); err == nil {
		return nil
	}
	return e.fillLoginEditByIndex(labels, resourceIDs, editIndex, value)
}

func (e *Executor) markLoginPasswordTyped(editIndex int) {
	if editIndex == 1 {
		e.Session.Runtime["login_password_typed"] = true
	}
}

func (e *Executor) tapLoginFieldAndType(resolver *ui.Resolver, labels, resourceIDs []string, value string, editIndex int) error {
	snap := e.Session.ReadUI(false)
	var field *ui.Resolved
	if editIndex == 0 {
		field = ui.FindLoginEmailField(resolver, snap, labels, resourceIDs)
	} else {
		field = ui.FindLoginPasswordField(resolver, snap, labels, resourceIDs)
	}
	if field == nil {
		return fmt.Errorf("login field not found")
	}
	if editIndex == 0 && ui.FieldHasExpectedValue(field.Element, value) {
		return nil
	}
	x, y := field.EditTapPoint()
	if err := e.Session.Client.Tap(x, y); err != nil {
		return err
	}
	time.Sleep(e.Session.PollInterval())
	overlay.DismissKeyboardSettingsIfPresent(e)
	if overlay.KeyboardSettingsOpen(e) {
		return fmt.Errorf("login field tap opened keyboard settings")
	}
	if err := e.clearLoginFieldIfNeeded(field.Element, value); err != nil {
		return err
	}
	if err := e.Session.Client.InputText(value); err != nil {
		return err
	}
	if editIndex == 1 {
		if !overlay.ShouldMarkPasswordTyped(e) {
			return fmt.Errorf("password not entered — keyboard settings or IME foreground")
		}
	}
	e.markLoginPasswordTyped(editIndex)
	return nil
}

func (e *Executor) clearLoginFieldIfNeeded(field ui.Element, value string) error {
	if field.Password {
		return nil
	}
	if !ui.FieldNeedsClear(field, value) {
		return nil
	}
	got := strings.TrimSpace(field.Text)
	if got == "" {
		got = strings.TrimSpace(field.ContentDesc)
	}
	if got == "" {
		return nil
	}
	return e.Session.Client.ClearTextBackspaces(len(got) + 8)
}

func (e *Executor) fillLoginEditByIndex(labels, resourceIDs []string, index int, value string) error {
	snap := e.Session.ReadUI(true)
	resolver := e.Session.Resolver
	var field *ui.Resolved
	if index == 0 {
		field = ui.FindLoginEmailField(resolver, snap, labels, resourceIDs)
	} else if index == 1 {
		field = ui.FindLoginPasswordField(resolver, snap, labels, resourceIDs)
	}
	if field == nil {
		edits := ui.LoginFormEdits(snap)
		if index >= len(edits) {
			return fmt.Errorf("login edit index %d not found (%d fields)", index, len(edits))
		}
		elem := edits[index]
		field = &ui.Resolved{Element: elem, Bounds: elem.Bounds}
	}

	if index == 0 {
		if ui.FieldHasExpectedValue(field.Element, value) {
			return nil
		}
	} else if index == 1 {
		if typed, _ := e.Session.Runtime["login_password_typed"].(bool); typed {
			return nil
		}
	}

	x, y := field.EditTapPoint()
	if err := e.Session.Client.Tap(x, y); err != nil {
		return err
	}
	time.Sleep(e.Session.PollInterval())
	overlay.DismissKeyboardSettingsIfPresent(e)
	if overlay.KeyboardSettingsOpen(e) {
		return fmt.Errorf("login field tap opened keyboard settings")
	}
	if err := e.clearLoginFieldIfNeeded(field.Element, value); err != nil {
		return err
	}
	if err := e.Session.Client.InputText(value); err != nil {
		return err
	}
	if index == 1 {
		if !overlay.ShouldMarkPasswordTyped(e) {
			return fmt.Errorf("password not entered — keyboard settings or IME foreground")
		}
	}
	e.markLoginPasswordTyped(index)
	return nil
}

func (e *Executor) facebookLogout(action config.Action) error {
	serial := e.Session.Serial
	log.Printf("[%s] facebook_logout: start", serial)
	e.Event("LOGOUT start")

	e.Session.Pause(action.ParamFloat("settle_sec", 1.5))

	if err := logout.Run(e, action); err != nil {
		e.Event("LOGOUT failed: %v", err)
		return err
	}
	e.Event("LOGOUT ok — verified logged out")
	log.Printf("[%s] facebook_logout: ok", serial)
	return nil
}

func (e *Executor) facebookAutoPost(action config.Action) error {
	e.Event("POST flow start personal")
	if err := post.Run(e, action); err != nil {
		e.Event("POST failed: %v", err)
		return err
	}
	return nil
}

func (e *Executor) facebookFanpagePost(action config.Action) error {
	e.Event("POST flow start fanpage")
	if err := fanpage.Run(e, action); err != nil {
		e.Event("POST fanpage failed: %v", err)
		return err
	}
	return nil
}

func (e *Executor) forceStopApp(action config.Action) error {
	appKey := action.ParamString("app", "facebook")
	pkg, ok := e.Session.Workflow.Apps[appKey]
	if !ok || pkg == "" {
		return fmt.Errorf("unknown app key: %s", appKey)
	}
	log.Printf("[%s] force_stop_app: %s", e.Session.Serial, pkg)
	return e.Session.Client.ForceStop(pkg)
}

func (e *Executor) handlePermission(action config.Action) error {
	det := state.NewDetector()
	observe, invalidate := e.cachedObserve()
	snap, pkg, _ := observe()
	d := det.Detect(snap, pkg, "")
	if d.State != state.UIPermission || d.Confidence < state.InvestigateMinConfidence {
		if action.Optional {
			return nil
		}
		return fmt.Errorf("permission dialog not visible")
	}

	policy := action.ParamString("policy", "deny")
	var err error
	err = e.withPermissionPolicy(policy, func() error {
		return overlay.HandlePermissionDialog(e, det, observe, invalidate, 4*time.Second)
	})
	e.invalidateObserve(invalidate)
	if err != nil && action.Optional {
		return nil
	}
	return err
}

type pickerCol struct {
	center [2]int
	span   int
}

func findPickerColumns(snap ui.Snapshot) []pickerCol {
	var nums []pickerCol
	for _, elem := range snap.Elements {
		t := strings.TrimSpace(elem.Text)
		if len(t) >= 2 && len(t) <= 4 {
			allNum := true
			for _, c := range t {
				if c < '0' || c > '9' {
					allNum = false
					break
				}
			}
			if allNum || isMonth(t) {
				cx, cy := elem.Center()
				h := elem.Height()
				if h <= 0 {
					h = 80
				}
				nums = append(nums, pickerCol{center: [2]int{cx, cy}, span: h / 2})
			}
		}
	}
	if len(nums) >= 3 {
		return nums[:3]
	}
	return nums
}

func isMonth(s string) bool {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	lower := strings.ToLower(s)
	for _, m := range months {
		if strings.HasPrefix(strings.ToLower(m), lower) || strings.Contains(lower, strings.ToLower(m)) {
			return true
		}
	}
	return false
}

func hardSwipe(s *Session, col pickerCol, decrease bool) {
	cx, cy := col.center[0], col.center[1]
	span := col.span
	if span < 40 {
		span = 60
	}
	if decrease {
		_ = s.Client.Swipe(cx, cy-span, cx, cy+span, 100)
	} else {
		_ = s.Client.Swipe(cx, cy+span, cx, cy-span, 100)
	}
}

func actionVerify(action config.Action) *config.VerifySpec {
	if action.Verify != nil {
		return action.Verify
	}
	v, ok := action.Extra["verify"].(map[string]interface{})
	if !ok {
		return nil
	}
	spec := &config.VerifySpec{}
	if texts, ok := v["texts"].([]interface{}); ok {
		for _, t := range texts {
			if s, ok := t.(string); ok {
				spec.Texts = append(spec.Texts, s)
			}
		}
	}
	if f, ok := v["timeout_sec"].(float64); ok {
		spec.TimeoutSec = f
	}
	if f, ok := v["poll_sec"].(float64); ok {
		spec.PollSec = f
	}
	if len(spec.Texts) == 0 {
		return nil
	}
	return spec
}
