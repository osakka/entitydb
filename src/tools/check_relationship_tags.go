package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"os"
	"strings"
)

func main() {
	logger.Info("=== Check Relationship Tags ===")
	
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	// Check specific relationship
	relID := "rel_9de03ca57e584fd7a658663cff67f297"
	logger.Info("Checking tags for relationship: %s", relID)
	
	entity, err := repo.GetByID(relID)
	if err != nil {
		logger.Error("Failed to get entity: %v", err)
		os.Exit(1)
	}
	
	logger.Info("Entity ID: %s", entity.ID)
	logger.Info("Tags (%d):", len(entity.Tags))
	for _, tag := range entity.Tags {
		logger.Info("  - %s", tag)
	}
	
	// Check if it has type:relationship tag
	hasTypeRelationship := false
	for _, tag := range entity.Tags {
		if strings.Contains(tag, "type:relationship") {
			hasTypeRelationship = true
			break
		}
	}
	
	logger.Info("Has type:relationship tag: %v", hasTypeRelationship)
}