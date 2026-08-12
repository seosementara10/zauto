package overlay

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestPostPromotePromptVisible(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Tingkatkan jangkauan Anda" enabled="true" bounds="[24,900][696,960]" class="android.view.ViewGroup"/>
  <node text="Promosikan postingan" clickable="true" enabled="true" bounds="[48,1100][672,1180]" class="android.widget.Button"/>
  <node text="Lain Kali" clickable="true" enabled="true" bounds="[48,1200][672,1280]" class="android.widget.Button"/>
  <node text="Jangan tampilkan pesan ini setelah postingan di masa mendatang." clickable="true" enabled="true" bounds="[48,1320][620,1400]" class="android.view.ViewGroup"/>
  <node text="" clickable="true" enabled="true" bounds="[620,1330][680,1390]" class="android.widget.CheckBox"/>
</hierarchy>`)
	if !state.IsState(snap, "com.facebook.katana", state.UIPostPromotePrompt) {
		t.Fatal("expected post promote prompt visible")
	}
}

func TestFindPostPromoteLaterButton(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Promosikan postingan" clickable="true" enabled="true" bounds="[48,1100][672,1180]" class="android.widget.Button"/>
  <node text="Lain Kali" clickable="true" enabled="true" bounds="[48,1200][672,1280]" class="android.widget.Button"/>
</hierarchy>`)
	hit := findPostPromoteLaterButton(snap, 1600*50/100)
	if hit == nil {
		t.Fatal("expected Lain Kali button")
	}
	if hit.Label != "Lain Kali" {
		t.Fatalf("got label %q", hit.Label)
	}
}

func TestDetectPostPromotePrompt(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Tingkatkan jangkauan Anda" enabled="true" bounds="[24,900][696,960]" class="android.view.ViewGroup"/>
  <node text="Promosikan postingan" clickable="true" enabled="true" bounds="[48,1100][672,1180]" class="android.widget.Button"/>
  <node text="Lain Kali" clickable="true" enabled="true" bounds="[48,1200][672,1280]" class="android.widget.Button"/>
</hierarchy>`)
	d := state.NewDetector().Detect(snap, "com.facebook.katana", "")
	if d.State != state.UIPostPromotePrompt {
		t.Fatalf("expected post_promote_prompt, got %s", d.State)
	}
}
