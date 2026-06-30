package main

import (
	"log"
	"net/http"

	"search-engine/internal/api"
	"search-engine/internal/config"
	"search-engine/internal/domain"
	"search-engine/internal/python"
)

func main() {
	cfg := config.Load()

	log.Printf("Loading metadata from %s...", cfg.MetadataPath)
	meta, err := domain.LoadMetadata(cfg.MetadataPath)
	if err != nil {
		log.Fatalf("Failed to load metadata: %v", err)
	}
	log.Printf("Loaded %d products", meta.Count())

	pythonClient := python.NewClient(cfg.PythonServiceURL)
	router := api.NewRouter(cfg, meta, pythonClient)

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
