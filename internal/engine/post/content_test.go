package post

import (
	"testing"

	"zauto/internal/store"
)

func TestPostContentCacheKeyPerCategory(t *testing.T) {
	personal := postContentCacheKey(store.PostTextCategoryPersonal)
	fanpage := postContentCacheKey(store.PostTextCategoryFanpage)
	if personal == fanpage {
		t.Fatal("cache keys must differ per category")
	}
}

func TestVerifyTextPrefix(t *testing.T) {
	if got := verifyTextPrefix("abc"); got != "abc" {
		t.Fatalf("short text unchanged, got %q", got)
	}
	long := "0123456789012345678901234567890123456789"
	if got := verifyTextPrefix(long); got != long[:30] {
		t.Fatalf("expected 30-char prefix, got %q", got)
	}
}
