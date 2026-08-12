package state

import (
	"testing"

	"zauto/internal/ui"
)

func TestDetectKeyboardSettingsRule(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Setelan" enabled="true" bounds="[0,200][720,280]" class="android.widget.TextView" package="com.google.android.inputmethod.latin"/>
  <node text="Preferensi" enabled="true" bounds="[0,320][720,400]" class="android.widget.TextView" package="com.google.android.inputmethod.latin"/>
</hierarchy>`)
	got := NewDetector().Detect(snap, "com.google.android.inputmethod.latin", "")
	if got.State != UIKeyboardSettings {
		t.Fatalf("state=%q want keyboard_settings score=%.0f evidence=%v", got.State, got.Score, got.Evidence)
	}
}

func TestRegistryKeyboardSettingsIsOverlay(t *testing.T) {
	def, ok := DefaultRegistry.Def(UIKeyboardSettings)
	if !ok {
		t.Fatal("keyboard_settings missing from registry")
	}
	if !def.IsOverlay || !def.LoginFlow {
		t.Fatalf("keyboard_settings must be login overlay: %+v", def)
	}
}
