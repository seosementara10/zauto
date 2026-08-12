//go:build windows

package toolchain

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const AdbWrapExe = "adbwrap.exe"

// ConfigurePanel records the real adb path for in-process zauto calls (never adbwrap).
func ConfigurePanel(projectRoot string) {
	if real := resolveRealADB(projectRoot); real != "" {
		_ = os.Setenv("ZAUTO_REAL_ADB", real)
	}
}

// resolveRealADB locates the platform adb binary (not adbwrap).
// Prefers adb.exe bundled with tools/scrcpy (matches scrcpy) over PATH adb.
func resolveRealADB(projectRoot string) string {
	if p := strings.TrimSpace(os.Getenv("ZAUTO_REAL_ADB")); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if bundled := findBundledADB(projectRoot); bundled != "" {
		return bundled
	}
	if p, err := exec.LookPath("adb"); err == nil {
		return p
	}
	return ""
}

func findBundledADB(projectRoot string) string {
	var found string
	toolsDir := filepath.Join(projectRoot, "tools", "scrcpy")
	_ = filepath.WalkDir(toolsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "adb.exe") {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// ScrcpyEnv returns environment for scrcpy. Uses adb.exe bundled next to scrcpy (not adbwrap —
// scrcpy must read adb stdout via pipe; adbwrap breaks that when launched as scrcpy child).
func ScrcpyEnv(projectRoot string, scrcpyDir string) []string {
	env := os.Environ()
	real := filepath.Join(scrcpyDir, "adb.exe")
	if _, err := os.Stat(real); err != nil {
		real = resolveRealADB(projectRoot)
	}
	if real == "" {
		return env
	}
	env = setEnvVar(env, "ADB", real)
	env = setEnvVar(env, "ZAUTO_REAL_ADB", real)
	return env
}

// WithScrcpyEnv temporarily sets process env for scrcpy-noconsole.vbs launches.
func WithScrcpyEnv(projectRoot string, fn func() error) error {
	wrap := filepath.Join(projectRoot, AdbWrapExe)
	if _, err := os.Stat(wrap); err != nil {
		return fn()
	}
	real := resolveRealADB(projectRoot)
	restoreADB := patchEnv("ADB", wrap)
	restoreReal := func() {}
	if real != "" {
		restoreReal = patchEnv("ZAUTO_REAL_ADB", real)
	}
	defer func() {
		restoreReal()
		restoreADB()
	}()
	return fn()
}

func setEnvVar(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return append(out, prefix+value)
}

func patchEnv(key, value string) func() {
	old, ok := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	return func() {
		if ok {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}
