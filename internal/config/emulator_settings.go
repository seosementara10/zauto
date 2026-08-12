package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveEmulatorConfig merges emulator settings into config.json.
func SaveEmulatorConfig(path string, em EmulatorConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	em.normalize(raw)
	block, err := json.Marshal(em)
	if err != nil {
		return err
	}
	var emMap map[string]interface{}
	if err := json.Unmarshal(block, &emMap); err != nil {
		return err
	}
	raw["emulator"] = emMap
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
