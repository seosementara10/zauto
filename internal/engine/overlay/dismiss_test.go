package overlay

import (
	"testing"

	"zauto/internal/ui"
)

func TestPasswordManagerCloseQueriesMatchCloseButton(t *testing.T) {
	resolver := ui.NewDefaultResolver()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Pengelola Sandi Google" bounds="[0,500][720,560]" class="android.widget.TextView"/>
		<node content-desc="Close" clickable="true" bounds="[640,480][720,560]" class="android.widget.ImageButton"/>
	</hierarchy>`)
	r := findPasswordManagerClose(resolver, snap, 720, 1600)
	if r == nil {
		t.Fatal("expected close X to be found")
	}
	cx, _ := r.Center()
	if cx < 600 {
		t.Fatalf("expected top-right close button, center x=%d", cx)
	}
}

func TestPasswordManagerCloseFallbackPoint(t *testing.T) {
	resolver := ui.NewDefaultResolver()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Pengelola Sandi Google" bounds="[48,520][380,580]" class="android.widget.TextView"/>
		<node text="Lanjutkan" clickable="true" bounds="[520,900][680,980]" class="android.widget.Button"/>
	</hierarchy>`)
	x, y, ok := passwordManagerCloseFallbackPoint(resolver, snap, 720, 1600)
	if !ok {
		t.Fatal("expected fallback point")
	}
	if x < 640 || y < 540 || y > 560 {
		t.Fatalf("unexpected fallback (%d,%d)", x, y)
	}
}

func TestPasswordManagerCloseGeometricImageView(t *testing.T) {
	resolver := ui.NewDefaultResolver()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Pengelola Sandi Google" bounds="[48,520][380,580]" class="android.widget.TextView"/>
		<node text="Simpan sandi untuk login ke Facebook?" bounds="[48,600][680,680]" class="android.widget.TextView"/>
		<node clickable="true" bounds="[648,508][708,568]" class="android.widget.ImageView"/>
		<node text="Lanjutkan" clickable="true" bounds="[520,900][680,980]" class="android.widget.Button"/>
	</hierarchy>`)
	r := findPasswordManagerClose(resolver, snap, 720, 1600)
	if r == nil {
		t.Fatal("expected geometric close match")
	}
	cx, cy := r.Center()
	if cx < 620 || cy < 480 || cy > 600 {
		t.Fatalf("unexpected close position (%d,%d)", cx, cy)
	}
}
