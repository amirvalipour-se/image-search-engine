package api

import (
	"encoding/json"
	"net/http"

	"search-engine/internal/config"
	"search-engine/internal/domain"
	"search-engine/internal/python"
)

func NewRouter(cfg *config.Config, meta *domain.MetadataRepository, py *python.Client) http.Handler {
	h := NewHandler(py, meta)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("POST /", h.Home)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "mode": cfg.SearchMode})
	})
	mux.HandleFunc("POST /api/search", h.Search)
	mux.HandleFunc("POST /api/products", h.CreateProduct)

	return mux
}
