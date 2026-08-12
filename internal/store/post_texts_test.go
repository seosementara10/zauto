package store

import "testing"

func TestNormalizePostTextCategory(t *testing.T) {
	cases := map[string]string{
		"personal": PostTextCategoryPersonal,
		"Fanpage":  PostTextCategoryFanpage,
		"GROUP":    PostTextCategoryGroup,
	}
	for in, want := range cases {
		got, err := normalizePostTextCategory(in)
		if err != nil || got != want {
			t.Fatalf("normalize(%q)=%q err=%v want %q", in, got, err, want)
		}
	}
	if _, err := normalizePostTextCategory("invalid"); err == nil {
		t.Fatal("expected error for invalid category")
	}
}
