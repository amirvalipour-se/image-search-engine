package service

import (
	"errors"
	"fmt"

	"search-engine/internal/domain"
	"search-engine/internal/python"
)

var (
	ErrIndexImage   = errors.New("index image")
	ErrSaveMetadata = errors.New("save metadata")
)

type ProductInput struct {
	Name        string
	Category    string
	Description string
	Price       float64
	ImageURL    string
}

type SearchService struct {
	python   *python.Client
	metadata *domain.MetadataRepository
}

func NewSearchService(py *python.Client, meta *domain.MetadataRepository) *SearchService {
	return &SearchService{
		python:   py,
		metadata: meta,
	}
}

func (s *SearchService) SearchByImage(image []byte, k int) ([]domain.SearchResult, error) {
	pythonResp, err := s.python.Search(image, k)
	if err != nil {
		return nil, err
	}

	results := make([]domain.SearchResult, 0, len(pythonResp.Matches))
	for _, match := range pythonResp.Matches {
		product, ok := s.metadata.Get(match.ID)
		if !ok {
			continue
		}

		results = append(results, domain.SearchResult{
			ID:          product.ID,
			ProductName: product.ProductName,
			Category:    product.Category,
			Price:       product.Price,
			Image:       product.Image,
			Score:       match.Score,
		})
	}

	return results, nil
}

func (s *SearchService) CreateProduct(image []byte, input ProductInput) (domain.Product, error) {
	id, err := s.python.IndexImage(image)
	if err != nil {
		return domain.Product{}, fmt.Errorf("%w: %v", ErrIndexImage, err)
	}

	product := domain.Product{
		ID:          id,
		ProductName: input.Name,
		Category:    input.Category,
		Description: input.Description,
		Price:       input.Price,
		Image:       input.ImageURL,
	}

	if err := s.metadata.Put(product); err != nil {
		return domain.Product{}, fmt.Errorf("%w: %v", ErrSaveMetadata, err)
	}

	return product, nil
}
