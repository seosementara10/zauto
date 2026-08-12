package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zauto/internal/textutil"
)

const (
	PostTextCategoryPersonal = "personal"
	PostTextCategoryFanpage  = "fanpage"
	PostTextCategoryGroup    = "group"
)

type PostText struct {
	ID        int64     `json:"id"`
	Category  string    `json:"category"`
	Body      string    `json:"body"`
	ImageFile string    `json:"image_file,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func normalizePostTextCategory(category string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case PostTextCategoryPersonal, PostTextCategoryFanpage, PostTextCategoryGroup:
		return strings.ToLower(strings.TrimSpace(category)), nil
	default:
		return "", fmt.Errorf("invalid category %q (use personal, fanpage, or group)", category)
	}
}

func (s *Store) ListPostTexts(ctx context.Context, category string) ([]PostText, error) {
	cat, err := normalizePostTextCategory(category)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, category, body, image_file, created_at
		FROM post_texts
		WHERE category = $1 AND status = 'active'
		ORDER BY id DESC`, cat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PostText
	for rows.Next() {
		var t PostText
		if err := rows.Scan(&t.ID, &t.Category, &t.Body, &t.ImageFile, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) CreatePostText(ctx context.Context, category, body, imageFile string) (int64, error) {
	cat, err := normalizePostTextCategory(category)
	if err != nil {
		return 0, err
	}
	body = textutil.SanitizeADBText(body)
	if body == "" {
		return 0, fmt.Errorf("text body is empty after sanitizing (use ASCII only)")
	}
	var id int64
	err = s.pool.QueryRow(ctx, `
		INSERT INTO post_texts (category, body, image_file)
		VALUES ($1, $2, $3) RETURNING id`, cat, body, strings.TrimSpace(imageFile)).Scan(&id)
	return id, err
}

func (s *Store) DeletePostText(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid post text id")
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE post_texts SET status = 'inactive' WHERE id = $1 AND status = 'active'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("post text %d not found", id)
	}
	return nil
}

// PickRandomPostText returns one active text at random for the category.
func (s *Store) PickRandomPostText(ctx context.Context, category string) (PostText, error) {
	cat, err := normalizePostTextCategory(category)
	if err != nil {
		return PostText{}, err
	}
	var t PostText
	err = s.pool.QueryRow(ctx, `
		SELECT id, category, body, image_file, created_at
		FROM post_texts
		WHERE category = $1 AND status = 'active'
		ORDER BY random()
		LIMIT 1`, cat).Scan(&t.ID, &t.Category, &t.Body, &t.ImageFile, &t.CreatedAt)
	if err != nil {
		return PostText{}, fmt.Errorf("no post texts for category %q: add texts in panel Text menu", cat)
	}
	t.Body = textutil.SanitizeADBText(t.Body)
	return t, nil
}

func (s *Store) CountPostTexts(ctx context.Context, category string) (int, error) {
	cat, err := normalizePostTextCategory(category)
	if err != nil {
		return 0, err
	}
	var n int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM post_texts WHERE category = $1 AND status = 'active'`, cat).Scan(&n)
	return n, err
}
