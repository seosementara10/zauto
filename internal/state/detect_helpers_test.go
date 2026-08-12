package state

import (
	"testing"

	"zauto/internal/ui"
)

func TestLoginFormReadyFromEdits(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="61592118521974" enabled="true" bounds="[48,400][672,480]" class="android.widget.EditText" package="com.facebook.katana"/>
  <node text="" enabled="true" password="true" bounds="[48,520][672,600]" class="android.widget.EditText" package="com.facebook.katana"/>
</hierarchy>`)
	if !LoginFormReady(nil, snap) {
		t.Fatal("expected login form ready from edits")
	}
}

func TestInvestigateKeyboardSettingsParity(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Setelan" enabled="true" bounds="[0,200][720,280]" class="android.widget.TextView" package="com.google.android.inputmethod.latin"/>
  <node text="Preferensi" enabled="true" bounds="[0,320][720,400]" class="android.widget.TextView" package="com.google.android.inputmethod.latin"/>
</hierarchy>`)
	pkg := "com.google.android.inputmethod.latin"
	d := NewDetector().Detect(snap, pkg, "")
	inv := NewDetector().Investigate(snap, pkg, "")
	if d.State != UIKeyboardSettings {
		t.Fatalf("Detect state=%q", d.State)
	}
	if inv.Detection.State != UIKeyboardSettings && inv.Probes[UIKeyboardSettings] < InvestigateMinConfidence {
		t.Fatalf("Investigate probe=%.2f detection=%q", inv.Probes[UIKeyboardSettings], inv.Detection.State)
	}
}

func TestClassifyUnknownIMEAsOverlay(t *testing.T) {
	inv := Investigation{Probes: map[UIState]float64{UIKeyboardSettings: 0.9}}
	k := ClassifyUnknown(ui.Snapshot{}, "com.google.android.inputmethod.latin", inv)
	if k != UnknownKindOverlay {
		t.Fatalf("kind=%q want overlay", k)
	}
}
