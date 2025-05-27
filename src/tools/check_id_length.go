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
	
	// Look for the admin user
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to find admin users: %v", err)
	}
	
	if len(adminUsers) == 0 {
		fmt.Println("No admin users found")
		return
	}
	
	adminUser := adminUsers[0]
	fmt.Printf("Admin user ID: %s (length: %d)\n", adminUser.ID, len(adminUser.ID))
	fmt.Printf("ID bytes: %x\n", []byte(adminUser.ID))
	
	// Look for full ID in tags
	var fullID string
	for _, tag := range adminUser.Tags {
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "identity:uuid:") {
			fullID = strings.TrimPrefix(actualTag, "identity:uuid:")
			fmt.Printf("\nFull ID from tag: %s (length: %d)\n", fullID, len(fullID))
			break
		}
	}
	
	// Check GetRelationshipsBySource with truncated ID
	fmt.Printf("\n=== Testing GetRelationshipsBySource with truncated ID ===\n")
	relationships, err := repo.GetRelationshipsBySource(adminUser.ID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d relationships\n", len(relationships))
	}
	
	// Check GetRelationshipsBySource with full ID
	if fullID != "" && fullID != adminUser.ID {
		fmt.Printf("\n=== Testing GetRelationshipsBySource with full ID ===\n")
		relationships, err = repo.GetRelationshipsBySource(fullID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Found %d relationships\n", len(relationships))
		}
	}
}