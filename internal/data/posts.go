package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PostEntry is one feed post: caption text and optional image filename (relative to images dir).
type PostEntry struct {
	Text      string
	ImageFile string
}

func LoadPosts(path string) ([]PostEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var posts []PostEntry
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		text := line
		image := ""
		if idx := strings.Index(line, "|"); idx >= 0 {
			text = strings.TrimSpace(line[:idx])
			image = strings.TrimSpace(line[idx+1:])
		}
		if text == "" && image == "" {
			return nil, fmt.Errorf("%s:%d empty post line", path, i+1)
		}
		posts = append(posts, PostEntry{Text: text, ImageFile: image})
	}
	if len(posts) == 0 {
		return nil, fmt.Errorf("no posts in %s", path)
	}
	return posts, nil
}

func GetPost(path string, index int) (PostEntry, error) {
	posts, err := LoadPosts(path)
	if err != nil {
		return PostEntry{}, err
	}
	if index < 0 || index >= len(posts) {
		return PostEntry{}, fmt.Errorf("post_index %d out of range (have %d)", index, len(posts))
	}
	return posts[index], nil
}

func ResolvePostImage(imagesDir, imageFile string) (string, error) {
	if strings.TrimSpace(imageFile) == "" {
		return "", nil
	}
	p := imageFile
	if !filepath.IsAbs(p) {
		p = filepath.Join(imagesDir, imageFile)
	}
	info, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("post image %q: %w", p, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("post image %q is a directory", p)
	}
	return p, nil
}
