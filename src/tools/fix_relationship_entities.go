package main

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"os"
	"strings"
)

func main() {
	logger.Info("=== Fix Relationship Entities ===")
	
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	// Find all relationship entities by ID pattern
	allEntities, err := repo.List()
	if err != nil {
		logger.Error("Failed to list entities: %v", err)
		os.Exit(1)
	}
	
	fixedCount := 0
	
	for _, entity := range allEntities {
		// Skip if not a relationship entity (by ID)
		if !strings.HasPrefix(entity.ID, "rel_") {
			continue
		}
		
		logger.Info("\nChecking relationship entity: %s", entity.ID)
		logger.Info("  Current tags: %v", entity.Tags)
		
		// Parse content to get relationship details
		var relData map[string]interface{}
		if len(entity.Content) > 0 {
			if err := json.Unmarshal(entity.Content, &relData); err == nil {
				logger.Info("  Content data: %v", relData)
				
				// Rebuild tags from content
				newTags := []string{"type:relationship"}
				
				if relType, ok := relData["relationship_type"].(string); ok && relType != "" {
					newTags = append(newTags, "_relationship:"+relType)
				}
				
				if sourceID, ok := relData["source_id"].(string); ok && sourceID != "" {
					newTags = append(newTags, "_source:"+sourceID)
				}
				
				if targetID, ok := relData["target_id"].(string); ok && targetID != "" {
					newTags = append(newTags, "_target:"+targetID)
				}
				
				// Add temporal timestamp to each tag
				timestamp := models.Now()
				timestampedTags := []string{}
				for _, tag := range newTags {
					timestampedTags = append(timestampedTags, models.FormatTemporalTag(tag))
				}
				
				// Update entity
				entity.Tags = timestampedTags
				entity.UpdatedAt = timestamp
				
				logger.Info("  New tags: %v", newTags)
				
				if err := repo.Update(entity); err != nil {
					logger.Error("  Failed to update entity: %v", err)
				} else {
					logger.Info("  Successfully fixed relationship entity")
					fixedCount++
				}
			} else {
				logger.Error("  Failed to parse content: %v", err)
			}
		} else {
			logger.Warn("  No content found for relationship entity")
		}
	}
	
	logger.Info("\nFixed %d relationship entities", fixedCount)
}