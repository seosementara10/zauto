package panel

import (
	"context"
	"fmt"
	"log"
	"time"

	"zauto/internal/adb"
	"zauto/internal/config"
	"zauto/internal/store"
	"zauto/internal/toolchain"
)

// DefaultPort is the headless panel HTTP port (`zauto serve`).
const DefaultPort = 8765

const (
	dbPrepareTimeout = 45 * time.Second
	dbRetryAttempts  = 30
	dbRetryDelay     = time.Second
)

// Options configures panel startup (Wails desktop or headless HTTP).
type Options struct {
	ProjectRoot string
	ConfigPath  string
	Port        int // headless only; 0 → DefaultPort
}

// Prepared holds a bootstrapped panel server.
type Prepared struct {
	Server *Server
}

// Close stops mirrors and releases panel resources.
func (p *Prepared) Close() {
	if p == nil || p.Server == nil {
		return
	}
	_ = p.Server.Close()
}

// Prepare bootstraps ADB, database, and the panel API handler (shared by Wails + serve).
func Prepare(ctx context.Context, opts Options) (*Prepared, error) {
	toolchain.ConfigurePanel(opts.ProjectRoot)

	if !adb.CheckAvailable() {
		return nil, fmt.Errorf("ADB tidak ditemukan — install Android Platform Tools")
	}

	wf, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	warmADBServer(wf)

	dbCtx, cancel := context.WithTimeout(ctx, dbPrepareTimeout)
	st, err := waitForDatabase(dbCtx, wf, dbRetryAttempts, dbRetryDelay)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	if err := bootstrapStore(context.Background(), st, opts.ProjectRoot); err != nil {
		st.Close()
		return nil, err
	}
	if counts, err := st.PostTextCounts(context.Background()); err == nil {
		log.Printf("Bootstrap: post texts personal=%d fanpage=%d group=%d",
			counts[store.PostTextCategoryPersonal],
			counts[store.PostTextCategoryFanpage],
			counts[store.PostTextCategoryGroup],
		)
	}

	port := opts.Port
	if port <= 0 {
		port = DefaultPort
	}

	srv := NewServer(opts.ProjectRoot, opts.ConfigPath, port)
	srv.Store = st
	srv.refreshDevicesWithRetry(4)
	srv.startUIHotReloadWatch()
	srv.broadcastState()
	return &Prepared{Server: srv}, nil
}

func warmADBServer(wf *config.Workflow) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := adb.StartServer(ctx); err != nil {
		log.Printf("panel: adb start-server: %v", err)
	}
	if wf != nil && wf.Emulator.AutoConnect {
		adb.ConnectLocalEmulators(adb.LDPlayerADBPorts(wf.Emulator.InstanceCount))
	}
}
