package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPostsTextAndImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "posts.txt")
	if err := os.WriteFile(path, []byte("Hello feed\nCaption|pic.jpg\n"), 0644); err != nil {
		t.Fatal(err)
	}
	posts, err := LoadPosts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("len=%d want 2", len(posts))
	}
	if posts[0].Text != "Hello feed" || posts[0].ImageFile != "" {
		t.Fatalf("post0=%+v", posts[0])
	}
	if posts[1].ImageFile != "pic.jpg" {
		t.Fatalf("post1=%+v", posts[1])
	}
}

func TestResolvePostImage(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(img, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ResolvePostImage(dir, "a.jpg")
	if err != nil || got != img {
		t.Fatalf("got=%q err=%v", got, err)
	}
}
