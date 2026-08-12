package main

import (
	"context"
	"fmt"
	"path/filepath"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/store"
)

func runDatabaseOps(ctx context.Context, wf *config.Workflow, migrate bool, importPath string, autoAssign bool) error {
	dbCtx, cancel := context.WithTimeout(ctx, store.DefaultConnectTimeout())
	st, err := store.OpenWorkflow(dbCtx, wf)
	cancel()
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}
	defer st.Close()

	if migrate {
		if err := st.Migrate(ctx); err != nil {
			return fmt.Errorf("db migrate: %w", err)
		}
		counts, err := st.PostTextCounts(ctx)
		if err != nil {
			return fmt.Errorf("db migrate counts: %w", err)
		}
		fmt.Println("Database schema applied.")
		fmt.Printf("Post texts: personal=%d fanpage=%d group=%d\n",
			counts[store.PostTextCategoryPersonal],
			counts[store.PostTextCategoryFanpage],
			counts[store.PostTextCategoryGroup],
		)
	}

	if importPath != "" {
		path := importPath
		if !filepath.IsAbs(path) {
			path = filepath.Join(wf.ProjectRoot, path)
		}
		n, err := st.ImportAccountsFile(ctx, path)
		if err != nil {
			return fmt.Errorf("db import: %w", err)
		}
		fmt.Printf("Imported %d account(s) from %s\n", n, path)
	}

	if autoAssign {
		serials, err := adb.ListDevices()
		if err != nil {
			return err
		}
		if len(serials) == 0 {
			return fmt.Errorf("db auto-assign: no devices connected")
		}
		n, err := st.AutoAssignDevices(ctx, serials, wf.Database.MaxAccountsPerDevice)
		if err != nil {
			return fmt.Errorf("db auto-assign: %w", err)
		}
		fmt.Printf("Auto-assigned %d device↔account slot(s)\n", n)
	}

	return nil
}
