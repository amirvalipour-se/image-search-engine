package api

import (
	"errors"
	"log"
	"net/http"

	"search-engine/internal/service"
)

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
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

	product, err := h.search.CreateProduct(image, service.ProductInput{
		Name:        r.FormValue("productName"),
		Category:    r.FormValue("category"),
		Description: r.FormValue("description"),
		Price:       parsePrice(r.FormValue("price")),
		ImageURL:    r.FormValue("image"),
	})
	if err != nil {
		log.Printf("product create error: %v", err)
		if errors.Is(err, service.ErrIndexImage) {
			http.Error(w, "failed to index image", http.StatusInternalServerError)
			return
		}
		http.Error(w, "failed to save product", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, product)
}
