package overlay

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestDetectKeyboardSettings(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Setelan" enabled="true" bounds="[0,200][720,280]" class="android.widget.TextView" package="com.google.android.inputmethod.latin"/>
  <node text="Bahasa" enabled="true" bounds="[0,320][720,400]" class="android.widget.TextView" package="com.google.android.inputmethod.latin"/>
  <node content-desc="Kembali ke atas" clickable="true" enabled="true" bounds="[0,88][112,200]" class="android.widget.ImageButton" package="com.google.android.inputmethod.latin"/>
</hierarchy>`)
	if !state.IsState(snap, "com.google.android.inputmethod.latin", state.UIKeyboardSettings) {
		t.Fatal("expected keyboard_settings via detector")
	}
}

func TestDetectKeyboardSettingsIgnoresFacebookSetelan(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Ubah Pengaturan" enabled="true" clickable="true" bounds="[32,692][283,732]" class="android.widget.Button" package="com.facebook.katana"/>
</hierarchy>`)
	if state.IsState(snap, "com.facebook.katana", state.UIKeyboardSettings) {
		t.Fatal("Tri banner Ubah Pengaturan must not classify as keyboard settings")
	}
}
