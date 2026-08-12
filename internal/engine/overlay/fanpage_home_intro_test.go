package overlay

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestFindFanpageHomeIntroCloseTopLeft(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="" content-desc="Close" clickable="true" enabled="true" bounds="[20,88][92,160]" class="android.widget.Button"/>
  <node text="Lewati" clickable="true" enabled="true" bounds="[600,88][700,160]" class="android.widget.Button"/>
  <node text="Memperkenalkan Beranda Khusus untuk Nurhayati Fans" enabled="true" bounds="[24,500][696,600]" class="android.view.ViewGroup"/>
  <node text="Berinteraksi sebagai Halaman Anda dalam ruang khusus yang terpisah dari profil Anda." enabled="true" bounds="[24,620][696,720]" class="android.view.ViewGroup"/>
</hierarchy>`)
	resolver := ui.NewDefaultResolver()
	hit := findFanpageHomeIntroClose(resolver, snap, 720, 1600)
	if hit == nil {
		t.Fatal("expected close X in top-left")
	}
	x, y := hit.Center()
	if x > 720*28/100 || y > 1600*20/100 {
		t.Fatalf("close X should be in header band, got (%d,%d)", x, y)
	}
}

func TestFanpageHomeIntroVisible(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Memperkenalkan Beranda Khusus untuk Nurhayati Fans" enabled="true" bounds="[24,500][696,600]" class="android.view.ViewGroup"/>
  <node text="Berinteraksi sebagai Halaman Anda dalam ruang khusus yang terpisah dari profil Anda." enabled="true" bounds="[24,620][696,720]" class="android.view.ViewGroup"/>
</hierarchy>`)
	if !state.IsState(snap, "com.facebook.katana", state.UIFanpageHomeIntro) {
		t.Fatal("expected fanpage home intro visible")
	}
}

func TestDetectFanpageHomeIntro(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Lewati" enabled="true" bounds="[600,88][700,160]" class="android.widget.Button"/>
  <node text="Memperkenalkan Beranda Khusus untuk Nurhayati Fans" enabled="true" bounds="[24,500][696,600]" class="android.view.ViewGroup"/>
  <node text="Berinteraksi sebagai Halaman Anda dalam ruang khusus yang terpisah dari profil Anda." enabled="true" bounds="[24,620][696,720]" class="android.view.ViewGroup"/>
</hierarchy>`)
	d := state.NewDetector().Detect(snap, "com.facebook.katana", "")
	if d.State != state.UIFanpageHomeIntro {
		t.Fatalf("expected fanpage_home_intro, got %s", d.State)
	}
}
