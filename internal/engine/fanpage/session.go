package fanpage

import (
	"fmt"
	"strings"

	"zauto/internal/engine/post"
	"zauto/internal/engine/runtime"
	"zauto/internal/store"
)

func fanpagePublishedKey(fbPageID string) string {
	return fmt.Sprintf("fanpage_published_%s", strings.TrimSpace(fbPageID))
}

func markFanpagePublished(e runtime.Exec, fbPageID string) {
	if fbPageID = strings.TrimSpace(fbPageID); fbPageID != "" {
		e.Sess().Runtime[fanpagePublishedKey(fbPageID)] = true
	}
	post.MarkPublished(e, store.PostTextCategoryFanpage)
}

func isFanpagePublished(e runtime.Exec, fbPageID string) bool {
	if v, _ := e.Sess().Runtime[fanpagePublishedKey(fbPageID)].(bool); v {
		return true
	}
	return false
}
