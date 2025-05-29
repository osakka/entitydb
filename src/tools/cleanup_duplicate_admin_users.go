package main

import (
	"fmt"
	"log"
	"os"

	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	fmt.Println("=== Cleaning up duplicate admin users ===")
	
	// Initialize repository
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "../var"
	}

	factory := &binary.RepositoryFactory{}
	repo, err := factory.CreateRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	// Get all admin users
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list admin users: %v", err)
	}

	fmt.Printf("Found %d admin users\n", len(adminUsers))

	if len(adminUsers) <= 1 {
		fmt.Println("No duplicate users to clean up.")
		return
	}

	// Find the oldest user (by creation timestamp)
	var oldestUser *models.Entity
	var oldestTimestamp int64 = 9223372036854775807 // Max int64 value

	for _, user := range adminUsers {
		if user.CreatedAt < oldestTimestamp {
			oldestTimestamp = user.CreatedAt
			oldestUser = user
		}
	}

	fmt.Printf("Keeping oldest admin user: %s (created: %d)\n", oldestUser.ID, oldestUser.CreatedAt)

	// Delete all other admin users and their relationships
	deletedCount := 0
	for _, user := range adminUsers {
		if user.ID != oldestUser.ID {
			fmt.Printf("Deleting duplicate admin user: %s (created: %d)\n", user.ID, user.CreatedAt)
			
			// Get and delete relationships for this user
			relationships, relErr := repo.GetRelationshipsBySource(user.ID)
			if relErr == nil {
				for _, rel := range relationships {
					if entityRel, ok := rel.(*models.EntityRelationship); ok {
						fmt.Printf("  Deleting relationship: %s\n", entityRel.ID)
						repo.DeleteRelationship(entityRel.ID)
					}
				}
			}
			
			// Delete the user entity
			if err := repo.Delete(user.ID); err != nil {
				fmt.Printf("  ERROR deleting user %s: %v\n", user.ID, err)
			} else {
				fmt.Printf("  Successfully deleted user %s\n", user.ID)
				deletedCount++
			}
		}
	}

	fmt.Printf("\nCleanup complete: deleted %d duplicate admin users\n", deletedCount)
	fmt.Printf("Remaining admin user: %s\n", oldestUser.ID)
}