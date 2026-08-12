package textutil

import "testing"

func TestSanitizeADBTextEmDashAndEllipsis(t *testing.T) {
	in := "Hello! Spread kindness today — dunia butuh lebih banyak ke…"
	want := "Hello! Spread kindness today - dunia butuh lebih banyak ke..."
	got := SanitizeADBText(in)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSanitizeADBTextDropsNonASCII(t *testing.T) {
	got := SanitizeADBText("café résumé")
	if got != "caf rsum" {
		t.Fatalf("got %q", got)
	}
}
