package api

import (
	"embed"
	"html/template"

	"search-engine/internal/domain"
	"search-engine/internal/python"
	"search-engine/internal/service"
)

//go:embed templates/*.html
var templatesFS embed.FS

const defaultSearchK = 500

type Handler struct {
	search   *service.SearchService
	template *template.Template
}

func NewHandler(py *python.Client, meta *domain.MetadataRepository) *Handler {
	return &Handler{
		search:   service.NewSearchService(py, meta),
		template: template.Must(template.ParseFS(templatesFS, "templates/*.html")),
	}
}
