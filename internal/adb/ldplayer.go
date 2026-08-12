package adb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"zauto/internal/executil"
)

const defaultLDPlayerInstall = `C:\LDPlayer\LDPlayer14`

// LDPlayerInstance describes one LDPlayer multi-instance slot.
type LDPlayerInstance struct {
	Index     int    `json:"index"`
	ADBPort   int    `json:"adb_port"`
	ADBSerial string `json:"adb_serial"`
	Running   bool   `json:"running"`
	Connected bool   `json:"connected"`
}

// LDPlayerManager controls LDPlayer via ldconsole.exe (dnconsole requires admin from GUI apps).
type LDPlayerManager struct {
	InstallPath string
}

func NewLDPlayerManager(installPath string) *LDPlayerManager {
	p := strings.TrimSpace(installPath)
	if p == "" {
		p = defaultLDPlayerInstall
	}
	return &LDPlayerManager{InstallPath: p}
}

func (m *LDPlayerManager) consolePath() string {
	ld := filepath.Join(m.InstallPath, "ldconsole.exe")
	if _, err := os.Stat(ld); err == nil {
		return ld
	}
	return filepath.Join(m.InstallPath, "dnconsole.exe")
}

func (m *LDPlayerManager) available() bool {
	_, err := os.Stat(m.consolePath())
	return err == nil
}

func (m *LDPlayerManager) run(ctx context.Context, args ...string) (string, error) {
	if !m.available() {
		return "", fmt.Errorf("ldconsole tidak ditemukan di %s", m.InstallPath)
	}
	cmd := executil.CommandContext(ctx, m.consolePath(), args...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return text, fmt.Errorf("%w: %s", err, text)
		}
		return text, err
	}
	return text, nil
}

// ListInstances scans leidian*.config and checks running/ADB state.
func (m *LDPlayerManager) ListInstances(connectedSerials map[string]bool) ([]LDPlayerInstance, error) {
	if !m.available() {
		return nil, fmt.Errorf("LDPlayer tidak terinstall di %s", m.InstallPath)
	}
	pattern := filepath.Join(m.InstallPath, "vms", "config", "leidian*.config")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var indexes []int
	for _, f := range files {
		base := filepath.Base(f)
		if base == "leidians.config" {
			continue
		}
		num := strings.TrimSuffix(strings.TrimPrefix(base, "leidian"), ".config")
		idx, err := strconv.Atoi(num)
		if err != nil {
			continue
		}
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out := make([]LDPlayerInstance, 0, len(indexes))
	for _, idx := range indexes {
		port := LDPlayerBaseADBPort + idx*2
		serial := fmt.Sprintf("127.0.0.1:%d", port)
		inst := LDPlayerInstance{
			Index:     idx,
			ADBPort:   port,
			ADBSerial: serial,
			Connected: connectedSerials[serial],
		}
		runningOut, runErr := m.run(ctx, "isrunning", "--index", strconv.Itoa(idx))
		if runErr == nil {
			v := strings.TrimSpace(runningOut)
			inst.Running = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "running")
		}
		out = append(out, inst)
	}
	return out, nil
}

// Launch starts one LDPlayer instance by index.
func (m *LDPlayerManager) Launch(index int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	_, err := m.run(ctx, "launch", "--index", strconv.Itoa(index))
	return err
}

// LaunchAll starts every configured instance (or indices 0..count-1 when no configs are listed).
func (m *LDPlayerManager) LaunchAll(count int) error {
	instances, err := m.ListInstances(nil)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return fmt.Errorf("tidak ada instance LDPlayer")
	}
	var targets []int
	for _, inst := range instances {
		if count > 0 && inst.Index >= count {
			continue
		}
		if inst.Running {
			continue
		}
		targets = append(targets, inst.Index)
	}
	if len(targets) == 0 {
		return nil
	}
	var errs []string
	for _, idx := range targets {
		if err := m.Launch(idx); err != nil {
			errs = append(errs, fmt.Sprintf("index %d: %v", idx, err))
		}
		time.Sleep(1500 * time.Millisecond)
	}
	if len(errs) > 0 {
		return fmt.Errorf("launch sebagian gagal: %s", strings.Join(errs, "; "))
	}
	return nil
}

// AddInstance clones instance 0 or creates a fresh instance.
func (m *LDPlayerManager) AddInstance(name string, cloneFrom int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	before, _ := m.ListInstances(nil)
	beforeSet := map[int]bool{}
	for _, inst := range before {
		beforeSet[inst.Index] = true
	}
	if name == "" {
		name = fmt.Sprintf("zauto-%d", time.Now().Unix()%100000)
	}
	if _, err := m.run(ctx, "copy", "--name", name, "--from", strconv.Itoa(cloneFrom)); err != nil {
		if _, err2 := m.run(ctx, "add", "--name", name); err2 != nil {
			return -1, fmt.Errorf("copy: %v; add: %w", err, err2)
		}
	}
	instances, err := m.ListInstances(nil)
	if err != nil {
		return -1, err
	}
	for _, inst := range instances {
		if !beforeSet[inst.Index] {
			_, _ = m.run(ctx, "modify", "--index", strconv.Itoa(inst.Index), "--adb", "1", "--memory", "2048", "--cpu", "2")
			return inst.Index, nil
		}
	}
	return -1, fmt.Errorf("instance baru tidak terdeteksi")
}

// Quit stops one instance.
func (m *LDPlayerManager) Quit(index int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	_, err := m.run(ctx, "quit", "--index", strconv.Itoa(index))
	return err
}
