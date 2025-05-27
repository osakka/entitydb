package main

import (
	"fmt"
	"log"
	"os"
	
	"entitydb/storage/binary"
)

func main() {
	dataDir := "./var"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	fmt.Printf("Fixing broken index in %s...\n", dataDir)
	
	// Create repository
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	
	// Force reindex
	fmt.Println("Forcing complete reindex...")
	if err := repo.ReindexTags(); err != nil {
		log.Fatalf("Failed to reindex tags: %v", err)
	}
	
	fmt.Println("Reindex complete. Verifying index health...")
	
	// Verify index health
	if err := repo.VerifyIndexHealth(); err != nil {
		fmt.Printf("Warning: Index health check failed: %v\n", err)
	} else {
		fmt.Println("Index health check passed!")
	}
	
	// Test query for admin users
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Error querying admin users: %v\n", err)
	} else {
		fmt.Printf("Found %d admin users\n", len(adminUsers))
		for _, user := range adminUsers {
			fmt.Printf("  - %s\n", user.ID)
		}
	}
	
	// Test query for relationships
	relationships, err := repo.ListByTag("type:relationship")
	if err != nil {
		fmt.Printf("Error querying relationships: %v\n", err)
	} else {
		fmt.Printf("Found %d relationships\n", len(relationships))
	}
	
	// Test query for has_credential relationships  
	hasCredRels, err := repo.ListByTag("_relationship:has_credential")
	if err != nil {
		fmt.Printf("Error querying has_credential relationships: %v\n", err)
	} else {
		fmt.Printf("Found %d has_credential relationships\n", len(hasCredRels))
		for _, rel := range hasCredRels {
			fmt.Printf("  - %s\n", rel.ID)
		}
	}
}