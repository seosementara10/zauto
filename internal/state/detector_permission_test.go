package state

import (
	"testing"

	"zauto/internal/ui"
)

func TestPermissionRequiresAllowDenyPair(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Manage your privacy and access settings" bounds="[0,400][720,500]" class="android.widget.TextView"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State == UIPermission {
		t.Fatalf("vague access text should not classify as permission: state=%q conf=%.2f", got.State, got.Confidence)
	}
}

func TestPermissionHighConfidenceWithSystemWindow(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Allow Facebook to send you notifications?" bounds="[0,400][720,500]" package="com.android.permissioncontroller" class="android.widget.TextView"/>
		<node text="Don't allow" clickable="true" bounds="[0,700][360,780]" class="android.widget.Button"/>
		<node text="Allow" clickable="true" bounds="[360,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.android.permissioncontroller", "")
	if got.State != UIPermission {
		t.Fatalf("state=%q want permission conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < VerifyMinConfidence {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}
