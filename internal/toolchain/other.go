//go:build !windows

package toolchain

// ConfigurePanel is a no-op on non-Windows platforms.
func ConfigurePanel(projectRoot string) {}

// ScrcpyEnv returns the current environment unchanged.
func ScrcpyEnv(projectRoot string, scrcpyDir string) []string { return nil }

// WithScrcpyEnv runs fn without modifying the environment.
func WithScrcpyEnv(projectRoot string, fn func() error) error { return fn() }
