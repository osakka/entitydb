package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"entitydb/models"
	"entitydb/storage/binary"
	"golang.org/x/crypto/bcrypt"
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

	// First, delete all existing admin users except the first one
	users, err := repo.ListByTag("type:user")
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}

	adminUsers := []string{}
	for _, user := range users {
		for _, tag := range user.Tags {
			parts := strings.SplitN(tag, "|", 2)
			actualTag := tag
			if len(parts) == 2 {
				actualTag = parts[1]
			}
			
			if actualTag == "identity:username:admin" {
				adminUsers = append(adminUsers, user.ID)
				break
			}
		}
	}

	fmt.Printf("Found %d admin users\n", len(adminUsers))

	// Keep only the first admin user
	if len(adminUsers) > 1 {
		for i := 1; i < len(adminUsers); i++ {
			fmt.Printf("Deleting duplicate admin user: %s\n", adminUsers[i])
			if err := repo.Delete(adminUsers[i]); err != nil {
				fmt.Printf("Failed to delete %s: %v\n", adminUsers[i], err)
			}
		}
	}

	// Now ensure the first admin user has proper credentials
	if len(adminUsers) > 0 {
		adminID := adminUsers[0]
		fmt.Printf("\nFixing admin user: %s\n", adminID)

		// Check if credential exists
		credTag := fmt.Sprintf("credential:%s", adminID)
		creds, err := repo.ListByTag(credTag)
		if err != nil {
			log.Fatalf("Failed to check credentials: %v", err)
		}

		// Delete any existing credentials
		for _, cred := range creds {
			fmt.Printf("Deleting old credential: %s\n", cred.ID)
			if err := repo.Delete(cred.ID); err != nil {
				fmt.Printf("Failed to delete credential %s: %v\n", cred.ID, err)
			}
		}

		// Create new credential
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to generate password hash: %v", err)
		}

		credData := map[string]interface{}{
			"type":          "password",
			"password_hash": string(hash),
			"created_at":    time.Now().Format(time.RFC3339),
		}

		credContent, err := json.Marshal(credData)
		if err != nil {
			log.Fatalf("Failed to marshal credential data: %v", err)
		}

		// Generate a simple ID
		credID := fmt.Sprintf("cred_%x", time.Now().UnixNano())
		credEntity := &models.Entity{
			ID:      credID,
			Content: credContent,
			Tags: []string{
				"type:credential",
				fmt.Sprintf("credential:%s", adminID),
				"credential:type:password",
			},
		}

		if err := repo.Create(credEntity); err != nil {
			log.Fatalf("Failed to create credential: %v", err)
		}

		fmt.Printf("Created credential: %s\n", credID)

		// Get the admin user again
		users2, err := repo.ListByTag("type:user")
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}
		
		var adminUser *models.Entity
		for _, user := range users2 {
			if user.ID == adminID {
				adminUser = user
				break
			}
		}
		
		if adminUser == nil {
			log.Fatalf("Could not find admin user %s", adminID)
		}

		// Check if user has admin role
		hasAdminRole := false
		for _, tag := range adminUser.Tags {
			parts := strings.SplitN(tag, "|", 2)
			actualTag := tag
			if len(parts) == 2 {
				actualTag = parts[1]
			}
			if actualTag == "rbac:role:admin" {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			fmt.Printf("Adding admin role to user\n")
			adminUser.Tags = append(adminUser.Tags, "rbac:role:admin")
			if err := repo.Update(adminUser); err != nil {
				log.Fatalf("Failed to update admin user: %v", err)
			}
		}

		// Create relationship to administrators group
		relID := fmt.Sprintf("rel_%x", time.Now().UnixNano())
		relEntity := &models.Entity{
			ID: relID,
			Tags: []string{
				"type:relationship",
				fmt.Sprintf("relationship:from:%s", adminID),
				"relationship:to:group_administrators",
				"relationship:type:member_of",
				fmt.Sprintf("from:%s", adminID),
				"to:group_administrators",
				"relation:member_of",
			},
		}

		if err := repo.Create(relEntity); err != nil {
			fmt.Printf("Failed to create relationship (may already exist): %v\n", err)
		} else {
			fmt.Printf("Created relationship to administrators group: %s\n", relID)
		}

		fmt.Printf("\nAdmin user fixed successfully!\n")
		fmt.Printf("Username: admin\n")
		fmt.Printf("Password: admin\n")
		fmt.Printf("User ID: %s\n", adminID)
		fmt.Printf("Credential ID: %s\n", credID)
	} else {
		fmt.Printf("No admin users found to fix\n")
	}
}