package api

import (
	"log"
	"net/http"

	"search-engine/internal/domain"
)

type searchPageData struct {
	K       int
	Results []domain.SearchResult
	Error   string
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := searchPageData{K: defaultSearchK}
	if r.Method == http.MethodPost {
		data = h.searchFromForm(r)
	}

	h.renderSearchPage(w, data)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	image, err := readImageFile(r)
	if err != nil {
		http.Error(w, "file field required", http.StatusBadRequest)
		return
	}

	results, err := h.search.SearchByImage(image, parseK(r, defaultSearchK))
	if err != nil {
		log.Printf("python search error: %v", err)
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, domain.SearchResponse{Results: results})
}

func (h *Handler) searchFromForm(r *http.Request) searchPageData {
	data := searchPageData{K: defaultSearchK}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		data.Error = "failed to parse form"
		return data
	}

	data.K = parseK(r, defaultSearchK)
	image, err := readImageFile(r)
	if err != nil {
		data.Error = "choose an image first"
		return data
	}

	results, err := h.search.SearchByImage(image, data.K)
	if err != nil {
		log.Printf("template search error: %v", err)
		data.Error = "search failed"
		return data
	}

	data.Results = results
	return data
}

func (h *Handler) renderSearchPage(w http.ResponseWriter, data searchPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.ExecuteTemplate(w, "search.html", data); err != nil {
		log.Printf("template render error: %v", err)
	}
}
