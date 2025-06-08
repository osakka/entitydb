package main

import (
	"fmt"
	"log"
	"entitydb/models"
	"entitydb/storage/binary"
	"strings"
)

func main() {
	// Initialize storage
	repo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Find admin user
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list admin users: %v", err)
	}

	fmt.Printf("Found %d admin users\n", len(adminUsers))
	
	if len(adminUsers) == 0 {
		fmt.Println("No admin user found!")
		return
	}

	adminUser := adminUsers[0]
	fmt.Printf("\nAdmin User ID: %s\n", adminUser.ID)
	fmt.Println("Tags:")
	for _, tag := range adminUser.GetTagsWithoutTimestamp() {
		fmt.Printf("  - %s\n", tag)
		if strings.HasPrefix(tag, "relationship:has_credential:") {
			credID := strings.TrimPrefix(tag, "relationship:has_credential:")
			fmt.Printf("    -> Found credential relationship to: %s\n", credID)
			
			// Try to fetch the credential
			cred, err := repo.GetByID(credID)
			if err != nil {
				fmt.Printf("    -> ERROR fetching credential: %v\n", err)
			} else {
				fmt.Printf("    -> Credential found! Content length: %d bytes\n", len(cred.Content))
			}
		}
	}

	// Look for any credential entities
	fmt.Println("\nSearching for credential entities...")
	
	// List all entities and filter
	allEntities, err := repo.ListByTag("type:credential")
	if err != nil {
		// Try a different approach - list all and filter
		fmt.Println("Direct tag search failed, checking all entities...")
		allEntities = []*models.Entity{}
		
		// Get some entities to check
		for _, tag := range []string{"dataspace:_system", "algorithm:bcrypt"} {
			entities, err := repo.ListByTag(tag)
			if err == nil {
				for _, e := range entities {
					tags := e.GetTagsWithoutTimestamp()
					for _, t := range tags {
						if t == "type:credential" {
							allEntities = append(allEntities, e)
							break
						}
					}
				}
			}
		}
	}
	
	fmt.Printf("Found %d credential entities\n", len(allEntities))
	for _, cred := range allEntities {
		fmt.Printf("\nCredential ID: %s\n", cred.ID)
		fmt.Println("Tags:")
		for _, tag := range cred.GetTagsWithoutTimestamp() {
			fmt.Printf("  - %s\n", tag)
		}
		fmt.Printf("Content length: %d bytes\n", len(cred.Content))
	}
}