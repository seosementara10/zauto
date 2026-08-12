package panel

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed web/*
var webFS embed.FS

func readWebFile(name string) string {
	return readWebFileEmbedded(name)
}

func panelHTML() string {
	return readWebFile("layout.html")
}

var panelPageNames = []string{
	"dashboard", "devices", "accounts", "skrip", "text", "settings", "kontrol", "log",
}

func panelPageHTML(name string) string {
	for _, n := range panelPageNames {
		if n == name {
			return readWebFile("pages/" + name + ".html")
		}
	}
	return ""
}

func panelCSS() string {
	return readWebFile("styles.css")
}

func panelJS() string {
	return readWebFile("app.js")
}

func (s *Server) handlePanelCSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(s.panelCSS()))
}

func (s *Server) handlePanelTailwind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(s.readWebAsset("tailwind.css")))
}

func (s *Server) handlePanelJS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(s.panelJS()))
}

func (s *Server) handlePanelPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/assets/pages/"), ".html")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	html := s.panelPageHTML(name)
	if html == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(html))
}
