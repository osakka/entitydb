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
	logger.Info("=== Fix Relationship IDs Tool ===")
	
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	// Find all entities
	allEntities, err := repo.List()
	if err != nil {
		logger.Error("Failed to list entities: %v", err)
		os.Exit(1)
	}
	
	// Build a map of user IDs for quick lookup
	userIDs := make(map[string]bool)
	for _, entity := range allEntities {
		if strings.HasPrefix(entity.ID, "user_") {
			userIDs[entity.ID] = true
			logger.Info("Found user: %s", entity.ID)
		}
	}
	
	// Find relationships that need fixing
	needsFixing := []*models.Entity{}
	
	for _, entity := range allEntities {
		// Check if this is a relationship with has_credential
		isHasCredential := false
		var sourceID string
		
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "_relationship:has_credential") {
				isHasCredential = true
			}
			if strings.Contains(tag, "_source:") {
				parts := strings.SplitN(tag, "_source:", 2)
				if len(parts) == 2 {
					sourceID = parts[1]
				}
			}
		}
		
		if isHasCredential && sourceID != "" {
			// Check if source ID is too long (has extra char)
			possibleUserID := sourceID[:len(sourceID)-1]
			if userIDs[possibleUserID] {
				logger.Info("Found relationship %s with incorrect source ID: %s (should be %s)", 
					entity.ID, sourceID, possibleUserID)
				needsFixing = append(needsFixing, entity)
			}
		}
	}
	
	if len(needsFixing) == 0 {
		logger.Info("No relationships need fixing")
		return
	}
	
	logger.Info("\nFound %d relationships that need fixing", len(needsFixing))
	
	// Fix each relationship
	for _, entity := range needsFixing {
		logger.Info("\nFixing relationship: %s", entity.ID)
		
		// Create new tags with corrected source ID
		newTags := []string{}
		var oldSourceID, newSourceID string
		
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "_source:") {
				parts := strings.SplitN(tag, "_source:", 2)
				if len(parts) == 2 {
					oldSourceID = parts[1]
					newSourceID = oldSourceID[:len(oldSourceID)-1]
					// Add the corrected tag
					newTags = append(newTags, "_source:"+newSourceID)
					logger.Info("  Changing _source:%s to _source:%s", oldSourceID, newSourceID)
				}
			} else {
				// Keep other tags as is
				newTags = append(newTags, tag)
			}
		}
		
		// Also fix the content if it contains the source ID
		if len(entity.Content) > 0 {
			var content map[string]interface{}
			if err := json.Unmarshal(entity.Content, &content); err == nil {
				if sourceIDInContent, ok := content["source_id"].(string); ok && sourceIDInContent == oldSourceID {
					content["source_id"] = newSourceID
					newContent, _ := json.Marshal(content)
					entity.Content = newContent
					logger.Info("  Also fixed source_id in content")
				}
			}
		}
		
		// Update the entity with corrected tags
		entity.Tags = newTags
		entity.UpdatedAt = models.Now()
		
		if err := repo.Update(entity); err != nil {
			logger.Error("  Failed to update entity: %v", err)
		} else {
			logger.Info("  Successfully updated relationship")
		}
	}
	
	logger.Info("\nFix complete!")
}