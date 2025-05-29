package main

import (
	"fmt"
	"os"
	"strings"
	
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"golang.org/x/crypto/bcrypt"
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
		logger.Fatal("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	logger.Info("=== TESTING RBAC AUTHENTICATION ===")
	
	// Create security manager
	securityManager := models.NewSecurityManager(repo)
	
	// Test authentication
	logger.Info("Testing authentication for admin/admin...")
	userEntity, err := securityManager.AuthenticateUser("admin", "admin")
	if err != nil {
		logger.Error("Authentication failed: %v", err)
		
		// Let's debug what's happening
		logger.Info("Debugging authentication process...")
		
		// Find admin users
		adminUsers, err := repo.ListByTag("identity:username:admin")
		if err != nil {
			logger.Fatal("Failed to list admin users: %v", err)
		}
		
		logger.Info("Found %d admin users", len(adminUsers))
		
		for i, user := range adminUsers {
			fmt.Printf("\n--- Admin User %d ---\n", i+1)
			fmt.Printf("ID: %s\n", user.ID)
			fmt.Printf("Tags:\n")
			for _, tag := range user.Tags {
				actualTag := tag
				if parts := strings.SplitN(tag, " < /dev/null | ", 2); len(parts) == 2 {
					actualTag = parts[1]
				}
				fmt.Printf("  - %s\n", actualTag)
			}
			
			// Check for has_credential relationships stored as entities
			credRels, err := repo.ListByTag("rel:has_credential:source:" + user.ID)
			if err == nil && len(credRels) > 0 {
				fmt.Printf("Found %d credential relationships\n", len(credRels))
				for _, rel := range credRels {
					fmt.Printf("  Relationship entity: %s\n", rel.ID)
					
					// Try to parse the relationship
					if len(rel.Content) > 0 {
						relObj := &models.EntityRelationship{}
						if err := relObj.UnmarshalBinary(rel.Content); err == nil {
							fmt.Printf("    Target credential: %s\n", relObj.TargetID)
							
							// Get the credential
							cred, err := repo.GetByID(relObj.TargetID)
							if err == nil {
								fmt.Printf("    Credential found, testing password...\n")
								err = bcrypt.CompareHashAndPassword(cred.Content, []byte("admin"))
								fmt.Printf("    Password 'admin' valid: %v\n", err == nil)
							}
						}
					}
				}
			}
		}
	} else {
		fmt.Printf("Authentication successful!\n")
		fmt.Printf("User ID: %s\n", userEntity.ID)
		fmt.Printf("Username: %s\n", userEntity.Username)
		fmt.Printf("Email: %s\n", userEntity.Email)
		fmt.Printf("Roles: %v\n", userEntity.Roles)
	}
}
