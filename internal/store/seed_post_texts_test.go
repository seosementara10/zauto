package store

import (
	"testing"
)

func TestDefaultPostTextsCount(t *testing.T) {
	for cat, lines := range defaultPostTexts {
		if len(lines) != 20 {
			t.Fatalf("category %q has %d lines, want 20", cat, len(lines))
		}
	}
}

func TestDefaultPostTextsCategories(t *testing.T) {
	want := []string{PostTextCategoryPersonal, PostTextCategoryFanpage, PostTextCategoryGroup}
	for _, cat := range want {
		if _, ok := defaultPostTexts[cat]; !ok {
			t.Fatalf("missing default texts for %q", cat)
		}
	}
}
