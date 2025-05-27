package main

import (
	"entitydb/storage/binary"
	"fmt"
	"log"
	"strings"
)

func main() {
	// Open repository
	repo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()
	
	// List all entities
	entities, err := repo.List()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}
	
	fmt.Printf("Total entities: %d\n\n", len(entities))
	
	// Find user entities
	fmt.Println("=== USER ENTITIES ===")
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			// Remove timestamp prefix if present
			actualTag := tag
			if idx := strings.Index(tag, "|"); idx > 0 {
				actualTag = tag[idx+1:]
			}
			
			if actualTag == "type:user" {
				fmt.Printf("\nEntity ID: %s\n", entity.ID)
				fmt.Println("Tags:")
				for _, t := range entity.Tags {
					// Remove timestamp for display
					displayTag := t
					if idx := strings.Index(t, "|"); idx > 0 {
						displayTag = t[idx+1:]
					}
					fmt.Printf("  - %s\n", displayTag)
				}
				break
			}
		}
	}
	
	// Search specifically for identity:username:admin
	fmt.Println("\n=== SEARCHING FOR identity:username:admin ===")
	results, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
	} else {
		fmt.Printf("Found %d entities with tag 'identity:username:admin'\n", len(results))
		for _, entity := range results {
			fmt.Printf("  - %s\n", entity.ID)
		}
	}
}