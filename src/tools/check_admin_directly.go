//go:build tool
package main

import (
	"fmt"
	"log"
	"os"

	"entitydb/logger"
	"entitydb/storage/binary"
)

func main() {
	// Enable only error logging to reduce noise
	if err := logger.SetLogLevel("ERROR"); err != nil {
		log.Fatalf("Failed to set log level: %v", err)
	}

	// Determine data path
	dataPath := "/opt/entitydb/var"
	if _, err := os.Stat(dataPath); err != nil {
		dataPath = "var"
	}

	// Initialize repository
	factory := &binary.RepositoryFactory{}
	repo, err := factory.CreateRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	// Check for admin with both old and new tag formats
	fmt.Println("=== Checking for admin users ===")
	
	// Try new format
	users1, _ := repo.ListByTag("identity:username:admin")
	fmt.Printf("Users with 'identity:username:admin': %d\n", len(users1))
	
	// Try old format
	users2, _ := repo.ListByTag("id:username:admin")
	fmt.Printf("Users with 'id:username:admin': %d\n", len(users2))
	
	// Try direct ID lookup
	fmt.Println("\n=== Checking specific user IDs ===")
	userIDs := []string{"user_admin", "user_3bab530eb56e122a050df463214872ad"}
	
	for _, id := range userIDs {
		user, err := repo.GetByID(id)
		if err != nil {
			fmt.Printf("User %s: NOT FOUND (%v)\n", id, err)
		} else {
			fmt.Printf("User %s: FOUND\n", id)
			fmt.Printf("  Tags: %v\n", user.GetTagsWithoutTimestamp())
		}
	}
	
	// Try to find any users
	fmt.Println("\n=== Looking for any users ===")
	allUsers, _ := repo.ListByTag("type:user")
	fmt.Printf("Total users found: %d\n", len(allUsers))
	for i, user := range allUsers {
		if i < 5 { // Show first 5
			fmt.Printf("  User: %s\n", user.ID)
			tags := user.GetTagsWithoutTimestamp()
			for _, tag := range tags {
				if len(tag) > 15 && (tag[:15] == "identity:userna" || tag[:11] == "id:username") {
					fmt.Printf("    Username tag: %s\n", tag)
				}
			}
		}
	}
}