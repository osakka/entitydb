package main

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"log"
)

func main() {
	logger.Info("[main] === Test GetRelationshipsBySource ===")

	// Open repository
	repo, err := binary.NewEntityRepository("var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Test user ID we know exists
	userID := "user_3b0d8f209787b0baadfd00e555979f8"
	logger.Info("[main] Testing GetRelationshipsBySource for user: %s", userID)

	// Call the method
	relationships, err := repo.GetRelationshipsBySource(userID)
	if err != nil {
		logger.Error("[main] GetRelationshipsBySource failed: %v", err)
		return
	}

	logger.Info("[main] GetRelationshipsBySource returned %d relationships", len(relationships))

	// Inspect what was returned
	for i, rel := range relationships {
		logger.Info("[main] Relationship %d type: %T", i, rel)
		
		// Try to cast to different types
		if entityRel, ok := rel.(*models.EntityRelationship); ok {
			logger.Info("[main]   Cast to *models.EntityRelationship successful")
			logger.Info("[main]   ID: %s", entityRel.ID)
			logger.Info("[main]   Type: %s", entityRel.Type)
			logger.Info("[main]   RelationshipType: %s", entityRel.RelationshipType)
			logger.Info("[main]   SourceID: %s", entityRel.SourceID)
			logger.Info("[main]   TargetID: %s", entityRel.TargetID)
		} else {
			logger.Info("[main]   Failed to cast to *models.EntityRelationship")
		}
		
		// Try map[string]interface{}
		if mapRel, ok := rel.(map[string]interface{}); ok {
			logger.Info("[main]   Cast to map[string]interface{} successful")
			for k, v := range mapRel {
				logger.Info("[main]   %s: %v", k, v)
			}
		}
		
		// Try other types
		logger.Info("[main]   Actual value: %+v", rel)
	}
	
	// Also test by directly calling ListByTag
	logger.Info("[main] \nDirect ListByTag test:")
	searchTag := "_source:" + userID
	entities, err := repo.ListByTag(searchTag)
	if err != nil {
		logger.Error("[main] ListByTag failed: %v", err)
	} else {
		logger.Info("[main] ListByTag returned %d entities for tag '%s'", len(entities), searchTag)
		for i, entity := range entities {
			logger.Info("[main] Entity %d: ID=%s, Tags=%d", i, entity.ID, len(entity.Tags))
		}
	}
}