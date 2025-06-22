//go:build tool
package main

import (
	"path/filepath"
	"os"
	
	"entitydb/config"
	"entitydb/logger"
	"entitydb/storage/binary"
)

func main() {
	logger.Info("=== Dump Raw Relationship Data ===")
	
	// Load configuration using proper configuration system
	cfg := config.Load()
	
	// Allow override via command line argument
	if len(os.Args) > 1 {
		cfg.DataPath = os.Args[1]
		// Reconstruct database filename for the new data path
		cfg.DatabaseFilename = filepath.Join(cfg.DataPath, "entities.edb")
	}
	
	logger.Info("Using database file: %s", cfg.DatabaseFilename)
	
	reader, err := binary.NewReader(cfg.DatabaseFilename)
	if err != nil {
		logger.Error("Failed to create reader: %v", err)
		os.Exit(1)
	}
	defer reader.Close()
	
	// Find the specific relationship entity
	relID := "rel_9de03ca57e584fd7a658663cff67f297"
	logger.Info("Looking for relationship: %s", relID)
	
	entities, err := reader.GetAllEntities()
	if err != nil {
		logger.Error("Failed to get entities: %v", err)
		os.Exit(1)
	}
	
	found := false
	for _, entity := range entities {
		if entity.ID == relID {
			found = true
			logger.Info("Found entity ID: %s", entity.ID)
			logger.Info("Tags (%d):", len(entity.Tags))
			for i, tag := range entity.Tags {
				logger.Info("  [%d] %s", i, tag)
			}
			logger.Info("Content length: %d", len(entity.Content))
			if len(entity.Content) > 0 {
				logger.Info("Content (first 200 chars): %s", string(entity.Content[:min(200, len(entity.Content))]))
			}
			logger.Info("CreatedAt: %d", entity.CreatedAt)
			logger.Info("UpdatedAt: %d", entity.UpdatedAt)
			break
		}
	}
	
	if !found {
		logger.Error("Entity %s not found", relID)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}