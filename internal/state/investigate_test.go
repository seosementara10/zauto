package state

import (
	"testing"

	"zauto/internal/ui"
)

func TestInvestigateProbeFindsPartialLogin(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Nomor ponsel atau email" bounds="[0,400][720,500]" class="android.widget.TextView"/>
		<node text="Kata sandi" bounds="[0,520][720,620]" class="android.widget.TextView"/>
	</hierarchy>`)
	inv := d.Investigate(snap, "com.facebook.katana", "")
	if inv.Method != "probe" && inv.Method != "primary" {
		t.Fatalf("method=%q want probe or primary", inv.Method)
	}
	if inv.Detection.State != UILogin {
		t.Fatalf("state=%q want login conf=%.2f", inv.Detection.State, inv.Detection.Confidence)
	}
}

func TestInvestigateUnresolvedEmpty(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy><node text="..." bounds="[0,0][10,10]"/></hierarchy>`)
	inv := d.Investigate(snap, "", "")
	if inv.Detection.State != UIUnknown {
		t.Fatalf("state=%q want unknown", inv.Detection.State)
	}
	if inv.Method != "unresolved" {
		t.Fatalf("method=%q want unresolved", inv.Method)
	}
}

func TestInvestigateSystemPermissionPackage(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Allow Facebook to send you notifications?" bounds="[0,400][720,500]" package="com.android.permissioncontroller" class="android.widget.TextView"/>
		<node text="Don't allow" clickable="true" bounds="[0,700][360,780]" class="android.widget.Button"/>
		<node text="Allow" clickable="true" bounds="[360,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	inv := d.Investigate(snap, "com.android.permissioncontroller", "")
	if inv.Detection.State != UIPermission {
		t.Fatalf("state=%q want permission method=%s conf=%.2f", inv.Detection.State, inv.Method, inv.Detection.Confidence)
	}
}
