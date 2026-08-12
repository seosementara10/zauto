package post

import (
	"context"
	"strings"

	"zauto/internal/config"
	"zauto/internal/data"
	"zauto/internal/engine/runtime"
	"zauto/internal/store"
)

func postCategoryForAction(action config.Action) string {
	if cat := strings.TrimSpace(action.ParamString("post_category", "")); cat != "" {
		return cat
	}
	switch action.Type {
	case "facebook_fanpage_post":
		return store.PostTextCategoryFanpage
	default:
		return store.PostTextCategoryPersonal
	}
}

func resolveFromDB(e runtime.Exec, action config.Action) (Content, error) {
	st := e.Sess().Store
	if st == nil {
		return Content{}, nil
	}
	wf := e.Sess().Workflow
	postSource := action.ParamString("post_source", wf.RegString("post_source", "db"))
	if postSource != "db" {
		return Content{}, nil
	}
	category := postCategoryForAction(action)
	entry, err := st.PickRandomPostText(context.Background(), category)
	if err != nil {
		return Content{}, err
	}
	e.Event("POST picked text id=%d category=%s", entry.ID, category)
	imagesDir := action.ParamString("images_dir", wf.RegString("images_dir", "data/post_images"))
	imgPath := ""
	if entry.ImageFile != "" {
		imgPath, err = data.ResolvePostImage(e.Sess().ResolvePath(imagesDir), entry.ImageFile)
		if err != nil {
			return Content{}, err
		}
	}
	return Content{Text: entry.Body, ImagePath: imgPath}, nil
}
