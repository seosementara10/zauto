package panel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RunDesktop starts the panel as a native Wails desktop app (WebView2 on Windows).
func RunDesktop(opts Options) error {
	already, release, err := acquirePanelInstance()
	if err != nil {
		return err
	}
	if already {
		log.Printf("panel: instance sudah berjalan")
		FocusPanelWindow()
		return nil
	}
	defer release()

	prep, err := Prepare(context.Background(), opts)
	if err != nil {
		return err
	}
	srv := prep.Server
	width, height, posX, posY := desktopWindowBounds()
	log.Printf("panel: siap — jendela %dx%d @ (%d,%d)", width, height, posX, posY)

	err = wails.Run(&options.App{
		Title:            "zauto Panel",
		Width:            width,
		Height:           height,
		MinWidth:         WindowMinWidth,
		MinHeight:        520,
		DisableResize:    false,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Handler: srv.Handler(),
		},
		BackgroundColour: &options.RGBA{R: 240, G: 244, B: 248, A: 255},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		OnStartup: func(ctx context.Context) {
			srv.setDesktopCtx(ctx)
			runtime.WindowSetPosition(ctx, posX, posY)
			runtime.WindowSetSize(ctx, width, height)
			srv.panel.set(posX, posY, width, height)
			go trackPanelWindow(ctx, srv)
			go func() {
				srv.refreshDevices()
				srv.broadcastState()
			}()
		},
		OnShutdown: func(ctx context.Context) {
			prep.Close()
		},
	})
	if err != nil {
		prep.Close()
		return fmt.Errorf("wails: %w", err)
	}
	return nil
}

func trackPanelWindow(ctx context.Context, srv *Server) {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			x, y := runtime.WindowGetPosition(ctx)
			w, h := runtime.WindowGetSize(ctx)
			if w > 0 {
				prev := srv.panel.mirrorStartX()
				srv.panel.set(x, y, w, h)
				if srv.panel.mirrorStartX() != prev {
					srv.mirrorMu.Lock()
					_, _, restarted := srv.relayoutMirrorsLocked()
					srv.mirrorMu.Unlock()
					if restarted > 0 {
						srv.requestSyncMirrors()
					} else {
						srv.broadcastState()
					}
				}
			}
		}
	}
}
