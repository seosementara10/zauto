package ui

import "testing"

func TestSelectBestLeftPicksLainKali(t *testing.T) {
	candidates := []Resolved{
		{Element: Element{Bounds: [4]int{360, 1000, 720, 1080}, Clickable: true}, Label: "SIMPAN"},
		{Element: Element{Bounds: [4]int{0, 1000, 360, 1080}, Clickable: true}, Label: "LAIN KALI"},
	}
	got := selectBest(candidates, "left")
	if got == nil || got.Label != "LAIN KALI" {
		t.Fatalf("selectBest(left) = %v want LAIN KALI", got)
	}
}

func TestFindSaveLoginLaterPrefersLeft(t *testing.T) {
	r := NewResolver(70)
	snap := ParseHierarchy(`<hierarchy>
		<node text="LAIN KALI" clickable="true" bounds="[40,1000][320,1080]" class="android.widget.Button"/>
		<node text="SIMPAN" clickable="true" bounds="[400,1000][680,1080]" class="android.widget.Button"/>
	</hierarchy>`)
	q := FindQuery{Texts: []string{"LAIN KALI", "Lain Kali"}, PreferClickable: true, Prefer: "left"}
	got := r.Find(snap, q)
	if got == nil || got.Label != "LAIN KALI" {
		t.Fatalf("Find(left) = %v want LAIN KALI", got)
	}
}
