package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func runComposeUp(args []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	composeFile := filepath.Join(root, "docker-compose.db.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return fmt.Errorf("docker-compose.db.yml not found — jalankan dari folder proyek zauto")
	}

	composeArgs := append([]string{"compose", "-f", composeFile, "up", "-d"}, args...)
	cmd := exec.Command("docker", composeArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = root
	fmt.Println("=== zauto compose up ===")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose: %w", err)
	}

	script := filepath.Join(root, "scripts", "start-panel.ps1")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("start-panel.ps1 not found")
	}

	time.Sleep(2 * time.Second)
	ps := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", script)
	ps.Stdout = os.Stdout
	ps.Stderr = os.Stderr
	ps.Dir = root
	if err := ps.Run(); err != nil {
		return fmt.Errorf("start panel: %w", err)
	}
	return nil
}

func runComposeCommand(ctx context.Context, subcmd string, args []string) error {
	switch subcmd {
	case "up":
		return runComposeUp(args)
	default:
		return fmt.Errorf("unknown compose subcommand: %s (supported: up)", subcmd)
	}
}
