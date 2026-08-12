package panel

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (s *Server) setDesktopCtx(ctx context.Context) {
	s.desktopMu.Lock()
	s.desktopCtx = ctx
	s.desktopMu.Unlock()
}

func (s *Server) desktopContext() context.Context {
	s.desktopMu.RLock()
	defer s.desktopMu.RUnlock()
	return s.desktopCtx
}

func (s *Server) handleWindowState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	ctx := s.desktopContext()
	if ctx == nil {
		writeJSON(w, map[string]interface{}{"desktop": false})
		return
	}
	writeJSON(w, map[string]interface{}{
		"desktop":    true,
		"maximized":  runtime.WindowIsMaximised(ctx),
		"fullscreen": runtime.WindowIsFullscreen(ctx),
	})
}

func (s *Server) handleWindowMaximize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	ctx := s.desktopContext()
	if ctx == nil {
		http.Error(w, "hanya tersedia di aplikasi desktop", http.StatusNotImplemented)
		return
	}
	if runtime.WindowIsMaximised(ctx) {
		runtime.WindowUnmaximise(ctx)
	} else {
		runtime.WindowMaximise(ctx)
	}
	writeJSON(w, map[string]bool{"maximized": runtime.WindowIsMaximised(ctx)})
}

func (s *Server) handleWindowFullscreen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	ctx := s.desktopContext()
	if ctx == nil {
		http.Error(w, "hanya tersedia di aplikasi desktop", http.StatusNotImplemented)
		return
	}
	if runtime.WindowIsFullscreen(ctx) {
		runtime.WindowUnfullscreen(ctx)
	} else {
		runtime.WindowFullscreen(ctx)
	}
	writeJSON(w, map[string]bool{"fullscreen": runtime.WindowIsFullscreen(ctx)})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
