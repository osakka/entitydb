package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"os"
)

func main() {
	logger.Info("=== Dump Raw Relationship Data ===")
	
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	// Use NewReader directly to bypass any caching/indexing
	reader, err := binary.NewReader(dataDir + "/entities.ebf")
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