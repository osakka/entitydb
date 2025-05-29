package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	
	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "/opt/entitydb/var"
	}
	
	// Create repository
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	fmt.Println("=== ALL RELATIONSHIPS ===")
	
	// Get all entities
	entities, err := repo.ListAll()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}
	
	for _, entity := range entities {
		// Check if this is a relationship entity
		for _, tag := range entity.Tags {
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			if strings.HasPrefix(actualTag, "type:relationship") {
				fmt.Printf("\nRelationship Entity: %s\n", entity.ID)
				
				// Try to parse as EntityRelationship
				if len(entity.Content) > 0 {
					rel := &models.EntityRelationship{}
					if err := rel.UnmarshalBinary(entity.Content); err == nil {
						fmt.Printf("  Type: %s\n", rel.RelationshipType)
						fmt.Printf("  Source: %s\n", rel.SourceID)
						fmt.Printf("  Target: %s\n", rel.TargetID)
						
						if rel.RelationshipType == "has_credential" {
							fmt.Println("  *** FOUND has_credential relationship! ***")
						}
					}
				}
				break
			}
		}
	}
}