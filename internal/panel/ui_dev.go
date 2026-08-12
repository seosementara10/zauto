package panel

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DevMode is true when ZAUTO_PANEL_DEV=1 (serve UI from disk + hot reload).
func DevMode() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("ZAUTO_PANEL_DEV")))
	return v == "1" || v == "true" || v == "yes"
}

func readWebFileEmbedded(name string) string {
	b, err := webFS.ReadFile("web/" + name)
	if err != nil {
		return ""
	}
	return string(b)
}

func (s *Server) readWebAsset(name string) string {
	if s.panelDevEnabled() && s.ProjectRoot != "" {
		path := filepath.Join(s.ProjectRoot, "internal", "panel", "web", filepath.FromSlash(name))
		if b, err := os.ReadFile(path); err == nil {
			return string(b)
		}
	}
	return readWebFileEmbedded(name)
}

func (s *Server) panelHTML() string  { return s.readWebAsset("layout.html") }
func (s *Server) panelCSS() string  { return s.readWebAsset("styles.css") }
func (s *Server) panelJS() string   { return s.readWebAsset("app.js") }

func (s *Server) panelPageHTML(name string) string {
	for _, n := range panelPageNames {
		if n == name {
			return s.readWebAsset("pages/" + name + ".html")
		}
	}
	return ""
}

func webAssetsSignature(projectRoot string) string {
	webDir := filepath.Join(projectRoot, "internal", "panel", "web")
	var b strings.Builder
	_ = filepath.WalkDir(webDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(webDir, path)
		if err != nil {
			return nil
		}
		b.WriteString(filepath.ToSlash(rel))
		b.WriteByte('|')
		b.WriteString(info.ModTime().UTC().Format(time.RFC3339Nano))
		b.WriteByte(';')
		return nil
	})
	return b.String()
}

func (s *Server) startUIHotReloadWatch() {
	if !s.panelDevEnabled() || s.ProjectRoot == "" {
		return
	}
	go func() {
		var lastSig string
		ticker := time.NewTicker(1200 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			sig := webAssetsSignature(s.ProjectRoot)
			if lastSig == "" {
				lastSig = sig
				continue
			}
			if sig != lastSig {
				lastSig = sig
				s.scheduleUIReload()
			}
		}
	}()
	log.Printf("panel dev: hot reload UI aktif — edit internal/panel/web lalu simpan")
}

func (s *Server) scheduleUIReload() {
	s.uiReloadMu.Lock()
	if s.uiReloadTimer != nil {
		s.uiReloadTimer.Stop()
	}
	s.uiReloadTimer = time.AfterFunc(450*time.Millisecond, func() {
		s.bumpUIRev()
	})
	s.uiReloadMu.Unlock()
}

func (s *Server) bumpUIRev() {
	s.uiRevMu.Lock()
	s.uiRev++
	rev := s.uiRev
	s.uiRevMu.Unlock()
	log.Printf("panel dev: reload UI (rev %d)", rev)
	s.broadcastState()
}

func (s *Server) uiRevValue() uint64 {
	s.uiRevMu.RLock()
	defer s.uiRevMu.RUnlock()
	return s.uiRev
}
