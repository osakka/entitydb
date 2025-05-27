package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"entitydb/storage/binary"
)

func main() {
	// Get data path from command line
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <data_path>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s ../var\n", os.Args[0])
		os.Exit(1)
	}

	dataPath := os.Args[1]

	// Suppress logs
	os.Setenv("ENTITYDB_LOG_LEVEL", "ERROR")

	// Open repository
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Search for all users
	users, err := repo.ListByTag("type:user")
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}

	fmt.Printf("Found %d users\n\n", len(users))

	// Check each user
	adminFound := false
	for _, user := range users {
		// Look for username
		var username string
		for _, tag := range user.Tags {
			// Handle temporal tags
			parts := strings.SplitN(tag, "|", 2)
			actualTag := tag
			if len(parts) == 2 {
				actualTag = parts[1]
			}
			
			if strings.HasPrefix(actualTag, "identity:username:") {
				username = strings.TrimPrefix(actualTag, "identity:username:")
			}
		}
		
		if username == "admin" {
			adminFound = true
			fmt.Printf("=== ADMIN USER FOUND ===\n")
			fmt.Printf("User ID: %s\n", user.ID)
			fmt.Printf("Username: %s\n", username)
			
			// Check credentials
			credTag := fmt.Sprintf("credential:%s", user.ID)
			creds, err := repo.ListByTag(credTag)
			if err != nil {
				fmt.Printf("Error checking credentials: %v\n", err)
			} else {
				fmt.Printf("Found %d credential entities\n", len(creds))
				for _, cred := range creds {
					var credData map[string]interface{}
					if err := json.Unmarshal(cred.Content, &credData); err == nil {
						fmt.Printf("  Credential ID: %s\n", cred.ID)
						fmt.Printf("  Credential type: %v\n", credData["type"])
						if hash, ok := credData["password_hash"].(string); ok {
							fmt.Printf("  Has password hash: %s... (length: %d)\n", hash[:20], len(hash))
						}
					}
				}
			}
			
			// Show roles
			var roles []string
			for _, tag := range user.Tags {
				parts := strings.SplitN(tag, "|", 2)
				actualTag := tag
				if len(parts) == 2 {
					actualTag = parts[1]
				}
				
				if strings.HasPrefix(actualTag, "rbac:role:") {
					roles = append(roles, strings.TrimPrefix(actualTag, "rbac:role:"))
				}
			}
			fmt.Printf("Roles: %v\n", roles)
			
			// Show all tags
			fmt.Printf("\nAll tags for admin user:\n")
			for _, tag := range user.Tags {
				fmt.Printf("  %s\n", tag)
			}
			
			fmt.Printf("=======================\n\n")
		}
	}
	
	if !adminFound {
		fmt.Printf("NO ADMIN USER FOUND!\n")
		fmt.Printf("\nAll users found:\n")
		for _, user := range users {
			var username string
			for _, tag := range user.Tags {
				parts := strings.SplitN(tag, "|", 2)
				actualTag := tag
				if len(parts) == 2 {
					actualTag = parts[1]
				}
				
				if strings.HasPrefix(actualTag, "identity:username:") {
					username = strings.TrimPrefix(actualTag, "identity:username:")
				}
			}
			fmt.Printf("  ID: %s, Username: %s\n", user.ID, username)
		}
	}
}