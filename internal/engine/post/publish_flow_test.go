package post

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestFindExactLabelDoesNotMatchPostinganBaruAsPost(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Postingan baru" content-desc="Postingan baru" clickable="false" enabled="true" bounds="[245,105][475,145]" class="android.view.ViewGroup"/>
  <node text="Berikutnya" content-desc="Berikutnya" clickable="true" enabled="true" bounds="[503,1400][696,1472]" class="android.widget.Button"/>
</hierarchy>`)
	if hit := findExactLabel(snap, []string{"Post", "POST"}, 0, 2000); hit != nil {
		t.Fatalf("exact match should not hit Postingan baru, got %q", hit.Label)
	}
	hit := findExactLabel(snap, state.PostComposerNextTexts, 1300, 1600)
	if hit == nil || hit.Label != "Berikutnya" {
		t.Fatalf("expected Berikutnya, got %v", hit)
	}
}

func TestFindComposerNextButtonHeaderFanpage(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Postingan baru" enabled="true" bounds="[245,105][475,145]" class="android.view.ViewGroup"/>
  <node text="Berikutnya" content-desc="Berikutnya" clickable="true" enabled="true" bounds="[580,95][680,145]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findComposerNextButton(snap, 720, 1600)
	if hit == nil || hit.Label != "Berikutnya" {
		t.Fatalf("expected header Berikutnya, got %v", hit)
	}
	_, y := hit.Center()
	if y > 1600*25/100 {
		t.Fatalf("expected header Berikutnya, y=%d", y)
	}
}

func TestFindPublishButtonBerbagiSettings(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Pengaturan postingan" enabled="true" bounds="[96,80][400,168]" class="android.view.ViewGroup"/>
  <node text="BERBAGI" content-desc="BERBAGI" clickable="true" enabled="true" bounds="[580,95][680,145]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findPublishButton(snap, 720, 1600)
	if hit == nil || hit.Label != "BERBAGI" {
		t.Fatalf("expected BERBAGI, got %v", hit)
	}
}

func TestComposerNeedsNextHeaderFanpage(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Postingan baru" enabled="true" bounds="[245,105][475,145]" class="android.view.ViewGroup"/>
  <node text="Berikutnya" content-desc="Berikutnya" clickable="true" enabled="true" bounds="[580,95][680,145]" class="android.widget.Button"/>
</hierarchy>`)
	acts := scanComposerActions(snap, 720, 1600)
	if !composerNeedsNextBeforePublish(acts, 1600) {
		t.Fatal("fanpage composer should tap Berikutnya before publish")
	}
}

func TestFindComposerNextButtonBottom(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node content-desc="Berikutnya" clickable="true" enabled="true" bounds="[503,1400][696,1472]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findComposerNextButton(snap, 720, 1600)
	if hit == nil || hit.Label != "Berikutnya" {
		t.Fatalf("expected Berikutnya, got %v", hit)
	}
}

func TestFindPublishButtonTopRight(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Buat postingan" content-desc="Buat postingan" clickable="false" enabled="true" bounds="[200,100][400,140]" class="android.view.ViewGroup"/>
  <node text="Posting" content-desc="Posting" clickable="true" enabled="true" bounds="[580,95][680,145]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findPublishButton(snap, 720, 1600)
	if hit == nil || hit.Label != "Posting" {
		t.Fatalf("expected Posting button, got %v", hit)
	}
	x, y := hit.Center()
	if x < 500 || y > 250 {
		t.Fatalf("expected top-right tap, got (%d,%d)", x, y)
	}
}

func TestFindBottomPublishButtonAfterBerikutnya(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Postingan baru" content-desc="Postingan baru" clickable="false" enabled="true" bounds="[245,105][475,145]" class="android.view.ViewGroup"/>
  <node content-desc="Posting" clickable="true" enabled="true" bounds="[24,1413][696,1504]" class="android.widget.Button">
    <node text="Posting" content-desc="Posting" clickable="false" enabled="true" bounds="[301,1441][418,1481]" class="android.view.ViewGroup"/>
  </node>
</hierarchy>`)
	hit := findPublishButton(snap, 720, 1600)
	if hit == nil || hit.Label != "Posting" {
		t.Fatalf("expected bottom Posting button, got %v", hit)
	}
	_, y := hit.Center()
	if y < 1300 {
		t.Fatalf("expected bottom tap, got y=%d", y)
	}
}

func TestIsFinalPublishLabelExact(t *testing.T) {
	if !isFinalPublishLabel("Posting") || !isFinalPublishLabel("Kirim") || !isFinalPublishLabel("BERBAGI") {
		t.Fatal("expected publish labels")
	}
	if isFinalPublishLabel("Postingan baru") || isFinalPublishLabel("Bagikan juga ke") {
		t.Fatal("should not match title or share menu row")
	}
}

func TestFindKeyboardDoneButtonTopRight(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Buat postingan" content-desc="Buat postingan" clickable="false" enabled="true" bounds="[200,100][400,140]" class="android.view.ViewGroup"/>
  <node content-desc="Selesai" clickable="true" enabled="true" bounds="[585,97][720,151]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findKeyboardDoneButton(snap, 720, 1600)
	if hit == nil || hit.Label != "Selesai" {
		t.Fatalf("expected Selesai keyboard done, got %v", hit)
	}
	if isFinalPublishLabel("Selesai") {
		t.Fatal("Selesai must not be treated as publish label")
	}
}

func TestDetectComposerVariant(t *testing.T) {
	resolver := ui.NewDefaultResolver()

	directSnap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Buat postingan" content-desc="Buat postingan" clickable="false" enabled="true" bounds="[200,100][400,140]" class="android.view.ViewGroup"/>
  <node text="Posting" content-desc="Posting" clickable="true" enabled="true" bounds="[580,95][680,145]" class="android.widget.Button"/>
</hierarchy>`)
	if got := detectComposerVariant(resolver, directSnap, 720, 1600); got != ComposerDirectPublish {
		t.Fatalf("expected DIRECT, got %v", got)
	}

	nextSnap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Postingan baru" content-desc="Postingan baru" clickable="false" enabled="true" bounds="[245,105][475,145]" class="android.view.ViewGroup"/>
  <node text="Berikutnya" content-desc="Berikutnya" clickable="true" enabled="true" bounds="[503,1400][696,1472]" class="android.widget.Button"/>
</hierarchy>`)
	if got := detectComposerVariant(resolver, nextSnap, 720, 1600); got != ComposerNextThenPublish {
		t.Fatalf("expected NEXT_THEN_PUBLISH, got %v", got)
	}
}

func TestFindExactLabelTutupTopLeft(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node content-desc="Tutup" clickable="true" enabled="true" bounds="[20,88][92,160]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findExactLabel(snap, []string{"Tutup", "Close"}, 0, 288, 180)
	if hit == nil || hit.Label != "Tutup" {
		t.Fatalf("expected Tutup, got %v", hit)
	}
}
