package state

import (
	"testing"

	"zauto/internal/ui"
)

func TestPermissionPolicyFailClosed(t *testing.T) {
	p := DefaultPermissionPolicy()
	if p.ActionFor(PermKindNotification) != PermActionDeny {
		t.Fatal("notification should deny")
	}
	if p.ActionFor(PermKindContacts) != PermActionDeny {
		t.Fatal("contacts should deny")
	}
	if p.ActionFor(PermKindUnknown) != PermActionReview {
		t.Fatal("unknown permission should review/stop")
	}
	if p.ActionFor(PermKindLocation) != PermActionDeny {
		t.Fatal("location should deny")
	}
}

func TestIdentifyPermissionNotification(t *testing.T) {
	r := ui.NewResolver(70)
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Allow Facebook to send you notifications?" bounds="[0,400][720,500]" class="android.widget.TextView"/>
		<node text="Don't allow" clickable="true" bounds="[0,700][360,780]" class="android.widget.Button"/>
	</hierarchy>`)
	if got := IdentifyPermissionKind(r, snap, "com.android.permissioncontroller"); got != PermKindNotification {
		t.Fatalf("kind=%q want notification", got)
	}
}

func TestIdentifyPermissionContacts(t *testing.T) {
	r := ui.NewResolver(70)
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Allow Facebook to access your contacts?" bounds="[0,400][720,500]" class="android.widget.TextView"/>
	</hierarchy>`)
	if got := IdentifyPermissionKind(r, snap, "com.facebook.katana"); got != PermKindContacts {
		t.Fatalf("kind=%q want contacts", got)
	}
}

func TestClassifyUnknownPermission(t *testing.T) {
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Allow" package="com.android.permissioncontroller" bounds="[0,700][360,780]" class="android.widget.Button"/>
	</hierarchy>`)
	inv := Investigation{Probes: map[UIState]float64{}}
	kind := ClassifyUnknown(snap, "com.android.permissioncontroller", inv)
	if kind != UnknownKindPermission {
		t.Fatalf("kind=%q want unknown_permission", kind)
	}
}

func TestClassifyUnknownLoading(t *testing.T) {
	inv := Investigation{Probes: map[UIState]float64{UILoading: 0.6}}
	kind := ClassifyUnknown(ui.Snapshot{}, "com.facebook.katana", inv)
	if kind != UnknownKindLoading {
		t.Fatalf("kind=%q want unknown_loading", kind)
	}
}
