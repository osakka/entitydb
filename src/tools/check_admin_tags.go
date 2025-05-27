package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	
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
	
	// Look for all user entities
	fmt.Printf("=== Searching for User Entities ===\n\n")
	
	userEntities, err := repo.ListByTag("type:user")
	if err != nil {
		log.Printf("Error listing by tag 'type:user': %v", err)
	} else {
		fmt.Printf("Found %d entities with tag 'type:user'\n", len(userEntities))
		
		for _, user := range userEntities {
			fmt.Printf("\nUser Entity: %s\n", user.ID)
			fmt.Printf("Tags:\n")
			for _, tag := range user.Tags {
				// Show both temporal and actual tag
				actualTag := tag
				if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
					actualTag = parts[1]
					fmt.Printf("  - %s (temporal: %s)\n", actualTag, tag)
				} else {
					fmt.Printf("  - %s\n", tag)
				}
			}
			
			// Check if this is the admin user
			for _, tag := range user.Tags {
				actualTag := tag
				if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
					actualTag = parts[1]
				}
				
				if actualTag == "identity:username:admin" {
					fmt.Printf("\n*** THIS IS THE ADMIN USER ***\n")
					
					// Check relationships
					rels, err := repo.GetRelationshipsBySource(user.ID)
					if err != nil {
						fmt.Printf("Error getting relationships: %v\n", err)
					} else {
						fmt.Printf("Relationships from this user: %d\n", len(rels))
					}
				}
			}
		}
	}
	
	// Also check what tags exist with identity:username:
	fmt.Printf("\n\n=== Checking identity:username: tags ===\n")
	identityTags, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Error listing by tag 'identity:username:admin': %v\n", err)
	} else {
		fmt.Printf("Found %d entities with tag 'identity:username:admin'\n", len(identityTags))
	}
}