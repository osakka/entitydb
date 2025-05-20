package main

import (
	"entitydb/storage/binary"
	"entitydb/logger"
	"fmt"
	"strings"
)

func main() {
	// Initialize logger
	logger.SetLogLevel("info")
	
	// Create repository
	repo, err := binary.NewEntityRepository("/opt/entitydb/var")
	if err != nil {
		logger.Fatalf("Failed to create repository: %v", err)
	}
	
	// List all user entities
	fmt.Println("Searching for user entities...")
	users, err := repo.ListByTag("type:user")
	if err != nil {
		logger.Fatalf("Failed to list users: %v", err)
	}
	
	fmt.Printf("Found %d user entities\n\n", len(users))
	
	// Check each user
	for i, user := range users {
		fmt.Printf("User %d (ID: %s):\n", i+1, user.ID)
		
		// Show all tags
		fmt.Println("  Tags:")
		for _, tag := range user.Tags {
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				fmt.Printf("    %s (timestamp: %s)\n", parts[1], parts[0])
			} else {
				fmt.Printf("    %s\n", tag)
			}
		}
		
		// Show content
		fmt.Println("  Content:")
		for _, content := range user.Content {
			if content.Type == "username" || content.Type == "password_hash" {
				if content.Type == "password_hash" {
					fmt.Printf("    %s: %s...\n", content.Type, content.Value[:10])
				} else {
					fmt.Printf("    %s: %s\n", content.Type, content.Value)
				}
			}
		}
		fmt.Println()
	}
	
	// Also try searching by admin tag
	fmt.Println("Searching for admin entities by username tag...")
	adminEntities, err := repo.ListByTag("id:username:admin")
	if err != nil {
		fmt.Printf("Error searching for admin tag: %v\n", err)
	} else {
		fmt.Printf("Found %d entities with id:username:admin tag\n", len(adminEntities))
	}
}