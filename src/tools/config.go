// Package tools provides common configuration for EntityDB tools
package main

import (
	"entitydb/config"
	"flag"
	"fmt"
	"log"
)

// ToolConfig holds configuration for tools - now wraps the main Config
type ToolConfig struct {
	*config.Config
	APIEndpoint string
	Debug       bool
}

// GetToolConfig returns configuration for tools using the centralized config system
func GetToolConfig() *ToolConfig {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	// Build API endpoint from configuration
	var apiEndpoint string
	if cfg.UseSSL {
		apiEndpoint = fmt.Sprintf("https://localhost:%d", cfg.SSLPort)
	} else {
		apiEndpoint = fmt.Sprintf("http://localhost:%d", cfg.Port)
	}
	
	return &ToolConfig{
		Config:      cfg,
		APIEndpoint: apiEndpoint,
		Debug:       cfg.DevMode,
	}
}

// RegisterToolFlags registers common tool flags (now delegates to ConfigManager)
func RegisterToolFlags(cfg *ToolConfig) {
	// Tool-specific flags can be added here if needed
	// The core configuration flags are handled by ConfigManager.RegisterFlags()
}