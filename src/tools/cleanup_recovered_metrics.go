package main

import (
	"entitydb/storage/binary"
	"log"
	"os"
	"strings"
)

func main() {
	// logger.InitLogger("INFO", false)
	
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "/opt/entitydb/var"
	}
	
	log.Printf("Loading repository from %s...", dataPath)
	
	// Create repository factory
	factory := &binary.RepositoryFactory{}
	
	// Create temporal repository
	repo, err := factory.CreateRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	
	// Get all metric entities
	entities, err := repo.ListByTag("type:metric")
	if err != nil {
		log.Fatalf("Failed to list metric entities: %v", err)
	}
	
	deletedCount := 0
	
	// Find and delete recovered metric entities
	for _, entity := range entities {
		// Check if it's a metric entity
		if !strings.HasPrefix(entity.ID, "metric_") {
			continue
		}
		
		// Check if it has recovery tags
		isRecovered := false
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if tag == "status:recovered" || tag == "recovery:partial" || tag == "recovery:placeholder" {
				isRecovered = true
				break
			}
		}
		
		if isRecovered {
			log.Printf("Deleting recovered metric: %s", entity.ID)
			
			// Delete the entity
			if err := repo.Delete(entity.ID); err != nil {
				log.Printf("Failed to delete %s: %v", entity.ID, err)
			} else {
				deletedCount++
			}
		}
	}
	
	log.Printf("Deleted %d recovered metric entities", deletedCount)
}