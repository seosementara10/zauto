package store

import (
	"context"
	"os"

	"zauto/internal/config"
)

const DefaultURL = "postgres://zauto:zauto_dev@127.0.0.1:5433/zauto?sslmode=disable"

// ResolveURL returns the PostgreSQL connection string from config or environment.
func ResolveURL(wf *config.Workflow) string {
	if wf != nil && wf.Database.URL != "" {
		return wf.Database.URL
	}
	if u := os.Getenv("ZAUTO_DATABASE_URL"); u != "" {
		return u
	}
	return DefaultURL
}

// OpenWorkflow connects using database.url from the workflow (or defaults).
func OpenWorkflow(ctx context.Context, wf *config.Workflow) (*Store, error) {
	return Open(ctx, ResolveURL(wf))
}
