package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type MetadataRepository struct {
	path     string
	products map[int]Product
	mu       sync.RWMutex
}

func LoadMetadata(path string) (*MetadataRepository, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read metadata file: %w", err)
	}

	var products map[int]Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return &MetadataRepository{path: path, products: products}, nil
}

func (r *MetadataRepository) Get(id int) (Product, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	return p, ok
}

func (r *MetadataRepository) Put(product Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[product.ID] = product

	data, err := json.MarshalIndent(r.products, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(r.path, data, 0644); err != nil {
		return fmt.Errorf("write metadata file: %w", err)
	}

	return nil
}

func (r *MetadataRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.products)
}
