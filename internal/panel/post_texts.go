package panel

import (
	"encoding/json"
	"net/http"
	"strings"

	"zauto/internal/store"
)

func (s *Server) handlePostTexts(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.handlePostTextsList(w, r)
	case http.MethodPost:
		s.handlePostTextsCreate(w, r)
	default:
		http.Error(w, "GET or POST only", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePostTextsList(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category == "" {
		category = store.PostTextCategoryPersonal
	}
	rows, err := s.Store.ListPostTexts(r.Context(), category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if rows == nil {
		rows = []store.PostText{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"category": category,
		"items":    rows,
		"count":    len(rows),
	})
}

func (s *Server) handlePostTextsCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Category  string `json:"category"`
		Body      string `json:"body"`
		ImageFile string `json:"image_file"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := s.Store.CreatePostText(r.Context(), body.Category, body.Body, body.ImageFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "ok": true})
}

func (s *Server) handlePostTextDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := s.Store.DeletePostText(r.Context(), body.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

func (s *Server) handlePostTextsCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	if s.Store == nil {
		http.Error(w, "database not connected", http.StatusServiceUnavailable)
		return
	}
	counts := map[string]int{}
	for _, cat := range []string{store.PostTextCategoryPersonal, store.PostTextCategoryFanpage, store.PostTextCategoryGroup} {
		n, err := s.Store.CountPostTexts(r.Context(), cat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		counts[cat] = n
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(counts)
}
