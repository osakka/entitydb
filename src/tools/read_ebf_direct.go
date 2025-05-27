package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"log"
)

func main() {
	logger.Info("[main] === Direct EBF Reader ===")

	// Open the reader directly
	reader, err := binary.NewReader("var/entities.ebf")
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Try to read some entities directly
	count := 0
	maxShow := 10
	
	// Read the index to get entity IDs
	logger.Info("[main] Reading index...")
	
	// Try GetAllEntities
	entities, err := reader.GetAllEntities()
	if err != nil {
		logger.Error("[main] Failed to get all entities: %v", err)
	} else {
		logger.Info("[main] GetAllEntities returned %d entities", len(entities))
		for i, entity := range entities {
			if i < maxShow {
				logger.Info("[main] Entity %d: ID=%s, Tags=%d, Content=%d bytes", 
					i+1, entity.ID, len(entity.Tags), len(entity.Content))
				for j, tag := range entity.Tags {
					if j < 5 { // Show first 5 tags
						logger.Info("[main]   Tag[%d]: %s", j, tag)
					}
				}
			}
			count++
		}
	}
	
	logger.Info("[main] Total entities found: %d", count)
	
	// Also check specific entities we know should exist
	testIDs := []string{
		"user_3b0d8f209787b0baadfd00e555979f8",
		"rel_9de03ca57e584fd7a658663cff67f297",
		"cred_bc098257eb7fa98e885df720bbaa5f9a",
	}
	
	logger.Info("[main] \nChecking specific entities...")
	for _, id := range testIDs {
		entity, err := reader.GetEntity(id)
		if err != nil {
			logger.Error("[main] Failed to get entity %s: %v", id, err)
		} else if entity != nil {
			logger.Info("[main] Found entity %s: Tags=%d, Content=%d bytes", 
				id, len(entity.Tags), len(entity.Content))
			for i, tag := range entity.Tags {
				if i < 5 {
					logger.Info("[main]   Tag[%d]: %s", i, tag)
				}
			}
		} else {
			logger.Info("[main] Entity %s not found", id)
		}
	}
}