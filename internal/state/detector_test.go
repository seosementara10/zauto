package state

import (
	"testing"

	"zauto/internal/ui"
)

func TestDetectLoginScreen(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Nomor ponsel atau email" bounds="[0,400][720,500]" class="android.widget.TextView"/>
		<node text="Kata sandi" bounds="[0,520][720,620]" class="android.widget.TextView"/>
		<node text="Login" clickable="true" bounds="[0,700][720,780]" class="android.widget.Button"/>
		<node text="Lupa kata sandi?" bounds="[0,800][720,850]" class="android.widget.TextView"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UILogin {
		t.Fatalf("state=%q want login conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.9 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectSavedProfileScreen(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Sherlly Angelica" bounds="[0,400][720,480]" class="android.widget.TextView"/>
		<node text="Lanjut" clickable="true" bounds="[40,600][680,680]" class="android.widget.Button"/>
		<node text="Gunakan profil lain" clickable="true" bounds="[0,700][720,760]" class="android.widget.TextView"/>
		<node text="Buat akun baru" clickable="true" bounds="[40,900][680,980]" class="android.widget.Button"/>
		<node content-desc="More options" clickable="true" bounds="[620,200][700,280]" class="android.widget.ImageButton"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UISavedProfileScreen {
		t.Fatalf("state=%q want saved_profile_screen conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
}

func TestDetectContactSkipConfirm(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Yakin tidak ingin mengunggah kontak Anda?" bounds="[0,400][720,500]" class="android.widget.TextView"/>
		<node text="Temukan teman lebih cepat dengan mengunggah kontak Anda." bounds="[0,520][720,620]" class="android.widget.TextView"/>
		<node text="Lewati" clickable="true" bounds="[400,800][520,880]" class="android.widget.Button"/>
		<node text="Unggah kontak" clickable="true" bounds="[540,800][700,880]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UIContactFollowPrompt {
		t.Fatalf("state=%q want contact_follow_prompt conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectContactFollowPrompt(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Izinkan akses kontak menemukan orang untuk diikuti" bounds="[0,200][720,400]" class="android.widget.TextView"/>
		<node text="Lewati" clickable="true" bounds="[580,180][700,240]" class="android.widget.TextView"/>
		<node text="Lanjutkan" clickable="true" bounds="[0,900][720,980]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UIContactFollowPrompt {
		t.Fatalf("state=%q want contact_follow_prompt conf=%.2f", got.State, got.Confidence)
	}
}

func TestDetectSaveLoginPrompt(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Simpan info login Anda?" bounds="[0,200][720,280]" class="android.widget.TextView"/>
		<node text="Simpan" clickable="true" bounds="[0,900][720,980]" class="android.widget.Button"/>
		<node text="LAIN KALI" clickable="true" bounds="[0,1000][360,1080]" class="android.widget.Button"/>
		<node text="SIMPAN" clickable="true" bounds="[360,1000][720,1080]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UISaveLoginPrompt {
		t.Fatalf("state=%q want save_login_prompt conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectLogoutConfirmPrompt(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Apa yang Anda pikirkan?" bounds="[0,400][720,460]" class="android.widget.TextView"/>
		<node text="Logout dari akun Anda?" bounds="[80,500][640,580]" class="android.widget.TextView"/>
		<node text="BATALKAN" clickable="true" bounds="[320,620][480,680]" class="android.widget.Button"/>
		<node text="LOGOUT" clickable="true" bounds="[500,620][660,680]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UILogoutConfirmPrompt {
		t.Fatalf("state=%q want logout_confirm_prompt conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectLocationServicesPrompt(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Untuk menggunakan Layanan Lokasi, izinkan Facebook untuk mengakses lokasi Anda" bounds="[0,200][720,320]" class="android.widget.TextView"/>
		<node text="Lewati" clickable="true" bounds="[580,80][700,140]" class="android.widget.TextView"/>
		<node text="Lanjutkan" clickable="true" bounds="[40,1400][680,1480]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UILocationServicesPrompt {
		t.Fatalf("state=%q want location_services_prompt conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
}

func TestDetectPasswordManagerSheet(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Pengelola Sandi Google" bounds="[0,500][720,560]" class="android.widget.TextView"/>
		<node text="Simpan sandi untuk login ke Facebook?" bounds="[0,600][720,680]" class="android.widget.TextView"/>
		<node text="Lanjutkan" clickable="true" bounds="[400,900][680,980]" class="android.widget.Button"/>
		<node content-desc="Close" clickable="true" bounds="[640,480][720,560]" class="android.widget.ImageButton"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.google.android.gms", "")
	if got.State != UIPasswordManagerSheet {
		t.Fatalf("state=%q want password_manager_sheet conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectPriorityOverlayOverFeed(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="What's on your mind?" bounds="[0,200][720,280]" class="android.widget.TextView"/>
		<node text="Izinkan Facebook mengirim notifikasi kepada Anda?" bounds="[0,400][720,500]" class="android.widget.TextView"/>
		<node text="Izinkan" clickable="true" bounds="[0,600][720,680]" class="android.widget.Button"/>
		<node text="Jangan izinkan" clickable="true" bounds="[0,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UIPermission {
		t.Fatalf("state=%q want permission (overlay priority) conf=%.2f", got.State, got.Confidence)
	}
}

func TestDetectNotificationPermission(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Izinkan Facebook mengirim notifikasi kepada Anda?" bounds="[0,400][720,500]" class="android.widget.TextView"/>
		<node text="Izinkan" clickable="true" bounds="[0,600][720,680]" class="android.widget.Button"/>
		<node text="Jangan izinkan" clickable="true" bounds="[0,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.android.permissioncontroller", "")
	if got.State != UIPermission {
		t.Fatalf("state=%q want permission conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectLocaleSetupError(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="We're having trouble setting up Facebook in Indonesia right now. You can try again, or use Facebook in English (US) for now." bounds="[0,300][720,500]" class="android.widget.TextView"/>
		<node text="Try again" clickable="true" bounds="[0,800][720,880]" class="android.widget.Button"/>
		<node text="Continue in English (US)" clickable="true" bounds="[0,900][720,980]" class="android.widget.Button"/>
	</hierarchy>`)
	got := d.Detect(snap, "com.facebook.katana", "")
	if got.State != UILocaleSetupError {
		t.Fatalf("state=%q want locale_setup_error conf=%.2f evidence=%v", got.State, got.Confidence, got.Evidence)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence too low: %.2f", got.Confidence)
	}
}

func TestDetectUncertainWithOnlyLoginWord(t *testing.T) {
	d := NewDetector()
	snap := ui.ParseHierarchy(`<hierarchy><node text="Login" bounds="[0,0][10,10]"/></hierarchy>`)
	got := d.Detect(snap, "", "")
	if got.State != UIUnknown {
		t.Fatalf("state=%q want unknown for weak evidence", got.State)
	}
}
