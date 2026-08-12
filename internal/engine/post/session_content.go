package post

import (
	"fmt"

	"zauto/internal/engine/runtime"
)

func postContentCacheKey(category string) string {
	return fmt.Sprintf("post_content_%s", category)
}

func postPublishedKey(category string) string {
	return fmt.Sprintf("post_published_%s", category)
}

func markPublished(e runtime.Exec, category string) {
	e.Sess().Runtime[postPublishedKey(category)] = true
}

func isPublished(e runtime.Exec, category string) bool {
	v, _ := e.Sess().Runtime[postPublishedKey(category)].(bool)
	return v
}
