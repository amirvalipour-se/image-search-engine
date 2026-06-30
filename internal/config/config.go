package config

import "os"

type Config struct {
	ServerPort     string
	MetadataPath   string
	PythonServiceURL string
	SearchMode     string
}

func Load() *Config {
	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MetadataPath:   getEnv("METADATA_PATH", "index/metadata.json"),
		PythonServiceURL: getEnv("PYTHON_SERVICE_URL", "http://localhost:8000"),
		SearchMode:     getEnv("SEARCH_MODE", "python"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
