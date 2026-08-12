package panel

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"zauto/internal/config"
	"zauto/internal/store"
)

const defaultAccountsFile = "data/accounts.txt"

// bootstrapStore migrates schema and imports legacy accounts when the database is empty.
func bootstrapStore(ctx context.Context, st *store.Store, projectRoot string) error {
	if st == nil {
		return fmt.Errorf("database store is nil")
	}
	if err := st.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	n, err := st.CountAccounts(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	path := filepath.Join(projectRoot, defaultAccountsFile)
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	imported, err := st.ImportAccountsFile(ctx, path)
	if err != nil {
		return fmt.Errorf("import %s: %w", path, err)
	}
	if imported > 0 {
		log.Printf("Bootstrap: imported %d account(s) from %s", imported, path)
	}
	return nil
}

// waitForDatabase retries until PostgreSQL accepts connections.
func waitForDatabase(ctx context.Context, wf *config.Workflow, attempts int, delay time.Duration) (*store.Store, error) {
	if attempts <= 0 {
		attempts = 30
	}
	if delay <= 0 {
		delay = time.Second
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		dbCtx, cancel := context.WithTimeout(ctx, store.DefaultConnectTimeout())
		st, err := store.OpenWorkflow(dbCtx, wf)
		cancel()
		if err == nil {
			return st, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("database not ready after %d attempts: %w", attempts, lastErr)
}
