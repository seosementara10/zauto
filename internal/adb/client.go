package adb

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zauto/internal/executil"
)

var keycodes = map[string]int{
	"HOME": 3, "BACK": 4, "RECENT": 187, "WAKEUP": 224, "POWER": 26, "ENTER": 66,
	"DEL": 67, "MOVE_END": 123,
}

var foregroundPkgRe = regexp.MustCompile(`(?:mCurrentFocus|mFocusedApp)=Window\{[^ ]+ [^ ]+ ([^/\s}]+)/`)

type Client struct {
	Serial     string
	Timeout    time.Duration
	Retries    int
	UserID     *int
	DumpSettle time.Duration
}

func adbCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, adbPath(), args...)
	executil.HideWindow(cmd)
	return cmd
}

func adbPath() string {
	if p := strings.TrimSpace(os.Getenv("ZAUTO_REAL_ADB")); p != "" {
		return p
	}
	if p, err := exec.LookPath("adb"); err == nil {
		return p
	}
	return "adb"
}

// StartServer ensures the adb daemon is running (silent on Windows).
func StartServer(ctx context.Context) error {
	return adbCommand(ctx, "start-server").Run()
}

func CheckAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return adbCommand(ctx, "version").Run() == nil
}

func ListDevices() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	out, err := adbCommand(ctx, "devices").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("adb devices: %w", err)
	}
	var devices []string
	for _, line := range strings.Split(string(out), "\n")[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "device" {
			devices = append(devices, parts[0])
		}
	}
	return DedupeDevices(devices), nil
}

func FindUserForPackage(serial, pkg string) *int {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	out, _ := adbCommand(ctx, "-s", serial, "shell", "pm", "list", "users").CombinedOutput()
	re := regexp.MustCompile(`UserInfo\{(\d+):`)
	matches := re.FindAllStringSubmatch(string(out), -1)
	userIDs := []string{"0"}
	for _, m := range matches {
		if len(m) > 1 {
			userIDs = append(userIDs, m[1])
		}
	}
	for _, uid := range userIDs {
		pathOut, _ := adbCommand(ctx, "-s", serial, "shell", "pm", "path", "--user", uid, pkg).CombinedOutput()
		if strings.Contains(string(pathOut), "package:") {
			id, _ := strconv.Atoi(uid)
			return &id
		}
	}
	return nil
}

func (c *Client) Shell(args ...string) (string, error) {
	cmdArgs := append([]string{"-s", c.Serial, "shell"}, args...)
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		cmd := adbCommand(ctx, cmdArgs...)
		out, err := cmd.CombinedOutput()
		cancel()
		if err == nil {
			return string(out), nil
		}
		lastErr = fmt.Errorf("[%s] adb shell %v: %w — %s", c.Serial, args, err, strings.TrimSpace(string(out)))
		if attempt < c.Retries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	return "", lastErr
}

func (c *Client) userArgs() []string {
	if c.UserID != nil {
		return []string{"--user", strconv.Itoa(*c.UserID)}
	}
	return nil
}

func (c *Client) WakeScreen() {
	_ = c.KeyEvent("WAKEUP")
}

func (c *Client) Tap(x, y int) error {
	_, err := c.Shell("input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	return err
}

func (c *Client) Swipe(x1, y1, x2, y2, durationMs int) error {
	_, err := c.Shell("input", "swipe",
		strconv.Itoa(x1), strconv.Itoa(y1), strconv.Itoa(x2), strconv.Itoa(y2), strconv.Itoa(durationMs))
	return err
}

func (c *Client) KeyEvent(key string) error {
	code, ok := keycodes[strings.ToUpper(key)]
	if !ok {
		return fmt.Errorf("unknown key: %s", key)
	}
	_, err := c.Shell("input", "keyevent", strconv.Itoa(code))
	return err
}

// ClearTextBackspaces moves the cursor to the end and sends DEL key events.
func (c *Client) ClearTextBackspaces(count int) error {
	if count <= 0 {
		return nil
	}
	_ = c.KeyEvent("MOVE_END")
	for i := 0; i < count; i++ {
		if err := c.KeyEvent("DEL"); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) InputText(text string) error {
	if text == "" {
		return nil
	}
	if bulk, ok := adbBulkInputArg(text); ok {
		if _, err := c.Shell("input", "text", bulk); err == nil {
			return nil
		}
	}
	for _, ch := range text {
		switch ch {
		case ' ':
			if _, err := c.Shell("input", "text", "%s"); err != nil {
				return err
			}
		case '@':
			if _, err := c.Shell("input", "keyevent", "77"); err != nil {
				return err
			}
		case '*':
			if _, err := c.Shell("input", "keyevent", "17"); err != nil {
				return err
			}
		default:
			s := string(ch)
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || strings.ContainsRune("._-+", ch) {
				if _, err := c.Shell("input", "text", s); err != nil {
					return err
				}
			} else {
				if _, err := c.Shell("input", "text", s); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// adbBulkInputArg returns a single adb "input text" argument when the whole string is safe.
func adbBulkInputArg(text string) (string, bool) {
	if strings.ContainsAny(text, "&|;<>(){}[]$`\\\"'") {
		return "", false
	}
	var b strings.Builder
	for _, ch := range text {
		switch ch {
		case ' ':
			b.WriteString("%s")
		case '%':
			return "", false
		default:
			if ch < 32 || ch > 126 {
				return "", false
			}
			b.WriteRune(ch)
		}
	}
	return b.String(), true
}

func (c *Client) ForceStop(pkg string) error {
	args := append([]string{"am", "force-stop"}, c.userArgs()...)
	args = append(args, pkg)
	_, err := c.Shell(args...)
	return err
}

// ClearPackageData runs `pm clear` — wipes app data (not just cache). Returns error unless output indicates success.
func (c *Client) ClearPackageData(pkg string) error {
	args := append([]string{"pm", "clear"}, c.userArgs()...)
	args = append(args, pkg)
	out, err := c.Shell(args...)
	if err != nil {
		return err
	}
	if !pmClearSucceeded(out) {
		return fmt.Errorf("pm clear %s: %s", pkg, strings.TrimSpace(out))
	}
	return nil
}

func pmClearSucceeded(out string) bool {
	s := strings.ToLower(strings.TrimSpace(out))
	return strings.Contains(s, "success")
}

func (c *Client) OpenApp(pkg, activity string) error {
	user := c.userArgs()
	if activity != "" {
		component := pkg + "/" + activity
		args := append([]string{"am", "start"}, user...)
		args = append(args, "-n", component)
		_, err := c.Shell(args...)
		return err
	}
	args := append([]string{"cmd", "package", "resolve-activity", "--brief"}, user...)
	args = append(args, "-a", "android.intent.action.MAIN", "-c", "android.intent.category.LAUNCHER", pkg)
	out, _ := c.Shell(args...)
	lines := strings.Fields(strings.TrimSpace(out))
	if len(lines) > 0 {
		last := lines[len(lines)-1]
		if strings.Contains(last, "/") && !strings.Contains(last, "No activity") {
			startArgs := append([]string{"am", "start"}, user...)
			startArgs = append(startArgs, "-n", last)
			_, err := c.Shell(startArgs...)
			return err
		}
	}
	component := pkg + "/.MainActivity"
	startArgs := append([]string{"am", "start"}, user...)
	startArgs = append(startArgs, "-n", component)
	_, err := c.Shell(startArgs...)
	return err
}

func (c *Client) IsPackageInstalled(pkg string) bool {
	args := append([]string{"pm", "list", "packages"}, c.userArgs()...)
	args = append(args, pkg)
	out, err := c.Shell(args...)
	return err == nil && strings.Contains(out, "package:"+pkg)
}

// InstallAPK installs an APK on the device (-r replace existing).
func (c *Client) InstallAPK(apkPath string) error {
	if apkPath == "" {
		return fmt.Errorf("empty apk path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	args := []string{"-s", c.Serial, "install", "-r", apkPath}
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		cmd := adbCommand(ctx, args...)
		out, err := cmd.CombinedOutput()
		if err == nil && strings.Contains(string(out), "Success") {
			return nil
		}
		lastErr = fmt.Errorf("[%s] adb install: %w — %s", c.Serial, err, strings.TrimSpace(string(out)))
		if attempt < c.Retries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	return lastErr
}

// UninstallPackage removes a package from the device.
func (c *Client) UninstallPackage(pkg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	cmd := adbCommand(ctx, "-s", c.Serial, "uninstall", pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[%s] adb uninstall %s: %w — %s", c.Serial, pkg, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// DeviceModel returns ro.product.model from getprop.
func (c *Client) DeviceModel() string {
	out, _ := c.Shell("getprop", "ro.product.model")
	return strings.TrimSpace(out)
}

// DeviceResolution returns wm size as "WxH".
func (c *Client) DeviceResolution() string {
	w, h := c.ScreenSize()
	return fmt.Sprintf("%dx%d", w, h)
}

func (c *Client) ScreenSize() (int, int) {
	out, err := c.Shell("wm", "size")
	if err != nil {
		return 720, 1600
	}
	re := regexp.MustCompile(`(\d+)x(\d+)`)
	if m := re.FindStringSubmatch(out); len(m) == 3 {
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		if w > 0 && h > 0 {
			return w, h
		}
	}
	return 720, 1600
}

// ForegroundPackage returns the package name in the current focus window.
func (c *Client) ForegroundPackage() string {
	out, err := c.Shell("dumpsys", "window", "windows")
	if err != nil {
		return ""
	}
	re := foregroundPkgRe
	if m := re.FindStringSubmatch(out); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func (c *Client) DumpUI(fast bool) (string, error) {
	remote := "/sdcard/window_dump.xml"
	settle := c.DumpSettle
	if fast {
		settle = 20 * time.Millisecond
	}
	if settle <= 0 {
		settle = 30 * time.Millisecond
	}
	var last string
	for i := 0; i < 2; i++ {
		_, _ = c.Shell("uiautomator", "dump", "--compressed", remote)
		time.Sleep(settle)
		out, _ := c.Shell("cat", remote)
		last = out
		if strings.Contains(out, "<hierarchy") && strings.Contains(out, "node") {
			return out, nil
		}
		if fast {
			break
		}
	}
	return last, nil
}

func (c *Client) Screenshot(path string) error {
	data, err := c.ScreenshotPNG()
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

// ScreenshotPNG captures the device screen as PNG bytes.
func (c *Client) ScreenshotPNG() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	cmd := adbCommand(ctx, "-s", c.Serial, "exec-out", "screencap", "-p")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("[%s] empty screenshot", c.Serial)
	}
	return buf.Bytes(), nil
}

// PushFile copies a local file to the device via adb push.
func (c *Client) PushFile(local, remote string) error {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout*2)
		cmd := adbCommand(ctx, "-s", c.Serial, "push", local, remote)
		out, err := cmd.CombinedOutput()
		cancel()
		if err == nil {
			return nil
		}
		lastErr = fmt.Errorf("[%s] adb push %s -> %s: %w — %s", c.Serial, local, remote, err, strings.TrimSpace(string(out)))
		if attempt < c.Retries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	return lastErr
}

// ScanMediaFile notifies the media scanner about a new file on device storage.
func (c *Client) ScanMediaFile(remotePath string) error {
	uri := "file://" + strings.TrimPrefix(remotePath, "file://")
	_, err := c.Shell("am", "broadcast", "-a", "android.intent.action.MEDIA_SCANNER_SCAN_FILE", "-d", uri)
	return err
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
