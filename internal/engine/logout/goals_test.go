package logout

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestSuccessGoalRequiresKeluarTapped(t *testing.T) {
	lf := &logoutFlow{}
	resolver := ui.NewResolver(70)
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Nomor ponsel atau email" bounds="[0,400][720,500]" class="android.widget.EditText"/>
		<node text="Kata sandi" bounds="[0,520][720,620]" class="android.widget.EditText"/>
		<node text="Masuk" clickable="true" bounds="[0,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	d := state.Detection{State: state.UILogin, Confidence: 1}
	if lf.successGoal(resolver, d, snap) {
		t.Fatal("expected false before Keluar is tapped")
	}
	lf.keluarTapped = true
	if !lf.successGoal(resolver, d, snap) {
		t.Fatal("expected true after Keluar tapped and login form visible")
	}
}

func TestSuccessGoalAcceptsSavedProfilePicker(t *testing.T) {
	lf := &logoutFlow{keluarTapped: true}
	resolver := ui.NewResolver(70)
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Nurhayati Guguk" bounds="[0,300][720,360]" class="android.widget.TextView"/>
		<node text="Lanjut" clickable="true" bounds="[40,500][680,580]" class="android.widget.Button"/>
		<node text="Gunakan profil lain" clickable="true" bounds="[0,620][720,680]" class="android.widget.TextView"/>
		<node text="Buat akun baru" clickable="true" bounds="[40,900][680,980]" class="android.widget.Button"/>
	</hierarchy>`)
	d := state.Detection{State: state.UISavedProfileScreen, Confidence: 1}
	if !lf.successGoal(resolver, d, snap) {
		t.Fatal("expected saved profile picker to count as logout success")
	}
}

func TestSuccessGoalRejectsLoggedInFeed(t *testing.T) {
	lf := &logoutFlow{keluarTapped: true}
	resolver := ui.NewResolver(70)
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Apa yang Anda pikirkan?" bounds="[0,400][720,460]" class="android.widget.TextView"/>
		<node text="Menu" clickable="true" bounds="[620,100][700,180]" class="android.widget.Button"/>
	</hierarchy>`)
	d := state.Detection{State: state.UILoggedIn, Confidence: 1}
	if lf.successGoal(resolver, d, snap) {
		t.Fatal("expected false while still on logged-in feed")
	}
}

func TestTapLogoutConfirmPrefersRightButton(t *testing.T) {
	resolver := ui.NewResolver(70)
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="BATALKAN" clickable="true" bounds="[320,620][480,680]" class="android.widget.Button"/>
		<node text="LOGOUT" clickable="true" bounds="[500,620][660,680]" class="android.widget.Button"/>
	</hierarchy>`)
	q := ui.FindQuery{Texts: []string{"LOGOUT", "Log out"}, PreferClickable: true, Prefer: "right"}
	el := resolver.Find(snap, q)
	if el == nil {
		t.Fatal("expected LOGOUT button")
	}
	if el.Label != "LOGOUT" {
		t.Fatalf("label=%q want LOGOUT", el.Label)
	}
	cx, _ := el.Center()
	if cx < 500 {
		t.Fatalf("expected right-side LOGOUT, center x=%d", cx)
	}
}
