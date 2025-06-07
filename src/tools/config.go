// Package tools provides common configuration for EntityDB tools
package main

import (
	"flag"
	"os"
	"path/filepath"
)

// ToolConfig holds configuration for tools
type ToolConfig struct {
	DataPath    string
	APIEndpoint string
	Debug       bool
}

// GetToolConfig returns configuration for tools with proper defaults
func GetToolConfig() *ToolConfig {
	// Determine EntityDB root directory
	execPath, _ := os.Executable()
	entityDBRoot := filepath.Join(filepath.Dir(execPath), "..", "..")
	
	// Use environment variables with relative path defaults
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = filepath.Join(entityDBRoot, "var")
	}
	
	apiEndpoint := os.Getenv("ENTITYDB_API_ENDPOINT")
	if apiEndpoint == "" {
		// Check if SSL is enabled
		if os.Getenv("ENTITYDB_USE_SSL") == "true" {
			port := os.Getenv("ENTITYDB_SSL_PORT")
			if port == "" {
				port = "8085"
			}
			apiEndpoint = "https://localhost:" + port
		} else {
			port := os.Getenv("ENTITYDB_PORT")
			if port == "" {
				port = "8085"
			}
			apiEndpoint = "http://localhost:" + port
		}
	}
	
	return &ToolConfig{
		DataPath:    dataPath,
		APIEndpoint: apiEndpoint,
		Debug:       os.Getenv("ENTITYDB_DEBUG") == "true",
	}
}

// RegisterToolFlags registers common tool flags
func RegisterToolFlags(cfg *ToolConfig) {
	flag.StringVar(&cfg.DataPath, "data-path", cfg.DataPath, "Path to EntityDB data directory")
	flag.StringVar(&cfg.APIEndpoint, "api-endpoint", cfg.APIEndpoint, "EntityDB API endpoint URL")
	flag.BoolVar(&cfg.Debug, "debug", cfg.Debug, "Enable debug logging")
}