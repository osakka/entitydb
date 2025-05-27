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
	
	// List all entities
	entities, err := repo.List()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}
	
	fmt.Printf("Total entities: %d\n\n", len(entities))
	
	// Look for user entities
	fmt.Printf("=== User Entities ===\n")
	userCount := 0
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			
			if actualTag == "type:user" {
				userCount++
				fmt.Printf("\nUser Entity: %s\n", entity.ID)
				fmt.Printf("Tags:\n")
				for _, t := range entity.Tags {
					at := t
					if p := strings.SplitN(t, "|", 2); len(p) == 2 {
						at = p[1]
					}
					fmt.Printf("  - %s\n", at)
				}
				break
			}
		}
	}
	fmt.Printf("\nTotal users: %d\n", userCount)
	
	// Look for credential entities
	fmt.Printf("\n=== Credential Entities ===\n")
	credCount := 0
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			
			if actualTag == "type:credential" {
				credCount++
				fmt.Printf("\nCredential Entity: %s\n", entity.ID)
				fmt.Printf("Tags:\n")
				for _, t := range entity.Tags {
					at := t
					if p := strings.SplitN(t, "|", 2); len(p) == 2 {
						at = p[1]
					}
					fmt.Printf("  - %s\n", at)
				}
				break
			}
		}
	}
	fmt.Printf("\nTotal credentials: %d\n", credCount)
	
	// Check for identity:username:admin tag
	fmt.Printf("\n=== Checking for identity:username:admin ===\n")
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d entities with tag 'identity:username:admin'\n", len(adminUsers))
		for _, u := range adminUsers {
			fmt.Printf("  - %s\n", u.ID)
		}
	}
}