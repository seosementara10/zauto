package post

import (
	"fmt"
	"path/filepath"
	"strings"

	"zauto/internal/config"
	"zauto/internal/data"
	"zauto/internal/engine/runtime"
	"zauto/internal/textutil"
)

// Content is resolved post payload for one run.
type Content struct {
	Text      string
	ImagePath string // local absolute path, empty for text-only
}

func ResolveContent(e runtime.Exec, action config.Action) (Content, error) {
	category := postCategoryForAction(action)
	if cached, ok := e.Sess().Runtime[postContentCacheKey(category)].(Content); ok && cached.Text != "" {
		return cached, nil
	}

	wf := e.Sess().Workflow
	if direct := strings.TrimSpace(action.ParamString("post_text", "")); direct != "" {
		img := strings.TrimSpace(action.ParamString("image_path", ""))
		if img != "" {
			img = e.Sess().ResolvePath(img)
		}
		c := Content{Text: textutil.SanitizeADBText(direct), ImagePath: img}
		e.Sess().Runtime[postContentCacheKey(category)] = c
		return c, nil
	}

	if c, err := resolveFromDB(e, action); err != nil {
		return Content{}, err
	} else if c.Text != "" {
		e.Sess().Runtime[postContentCacheKey(category)] = c
		return c, nil
	}

	postsFile := action.ParamString("posts_file", wf.RegString("posts_file", "data/posts.txt"))
	imagesDir := action.ParamString("images_dir", wf.RegString("images_dir", "data/post_images"))
	idx := action.ParamInt("post_index", wf.RegInt("post_index", 0))

	path := e.Sess().ResolvePath(postsFile)
	entry, err := data.GetPost(path, idx)
	if err != nil {
		return Content{}, err
	}
	imgPath := ""
	if entry.ImageFile != "" {
		imgPath, err = data.ResolvePostImage(e.Sess().ResolvePath(imagesDir), entry.ImageFile)
		if err != nil {
			return Content{}, err
		}
	}
	c := Content{Text: textutil.SanitizeADBText(entry.Text), ImagePath: imgPath}
	e.Sess().Runtime[postContentCacheKey(category)] = c
	return c, nil
}

func (c Content) Summary() string {
	if c.ImagePath != "" {
		return fmt.Sprintf("text=%q image=%s", truncate(c.Text, 40), filepath.Base(c.ImagePath))
	}
	return fmt.Sprintf("text=%q", truncate(c.Text, 60))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
