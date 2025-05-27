package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"log"
)

func main() {
	logger.Info("[main] === Debug GetByID ===")

	// Open repository
	repo, err := binary.NewEntityRepository("var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Test GetByID for the relationship we know exists
	relID := "rel_9de03ca57e584fd7a658663cff67f297"
	logger.Info("[main] Testing GetByID for: %s", relID)

	entity, err := repo.GetByID(relID)
	if err != nil {
		logger.Error("[main] GetByID failed: %v", err)
		return
	}

	logger.Info("[main] GetByID returned entity with ID: %s", entity.ID)
	logger.Info("[main] Entity has %d tags:", len(entity.Tags))
	for i, tag := range entity.Tags {
		logger.Info("[main]   Tag[%d]: %s", i, tag)
	}
	logger.Info("[main] Entity content: %d bytes", len(entity.Content))

	// Now test GetRelationshipByID
	logger.Info("[main] \nTesting GetRelationshipByID...")
	rel, err := repo.GetRelationshipByID(relID)
	if err != nil {
		logger.Error("[main] GetRelationshipByID failed: %v", err)
	} else {
		logger.Info("[main] GetRelationshipByID succeeded: %+v", rel)
	}

	// Finally test GetRelationshipsBySource
	userID := "user_3b0d8f209787b0baadfd00e555979f8"
	logger.Info("[main] \nTesting GetRelationshipsBySource for user: %s", userID)
	relationships, err := repo.GetRelationshipsBySource(userID)
	if err != nil {
		logger.Error("[main] GetRelationshipsBySource failed: %v", err)
	} else {
		logger.Info("[main] GetRelationshipsBySource returned %d relationships", len(relationships))
		for i, rel := range relationships {
			logger.Info("[main]   Relationship %d: %+v", i, rel)
		}
	}
}