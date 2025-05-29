package main

import (
	"encoding/json"
	"fmt"
	"entitydb/models"
	"entitydb/storage/binary"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("EntityDB Admin Authentication Diagnosis")
	fmt.Println("=====================================")
	
	// Open repository
	repo, err := binary.NewEntityRepository("/opt/entitydb/var")
	if err != nil {
		fmt.Printf("Failed to open repository: %v\n", err)
		return
	}
	defer repo.Close()
	
	// Find admin users
	users, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Failed to find admin users: %v\n", err)
		return
	}
	
	fmt.Printf("\nFound %d admin users:\n", len(users))
	
	for i, user := range users {
		fmt.Printf("\n--- Admin User %d ---\n", i+1)
		fmt.Printf("ID: %s\n", user.ID)
		
		// Check tags
		cleanTags := user.GetTagsWithoutTimestamp()
		fmt.Printf("Tags: %v\n", cleanTags)
		
		// Check for rbac:role:admin
		hasAdminRole := false
		for _, tag := range cleanTags {
			if tag == "rbac:role:admin" {
				hasAdminRole = true
				break
			}
		}
		fmt.Printf("Has rbac:role:admin: %v\n", hasAdminRole)
		
		// Find credential
		rels, err := repo.GetRelationshipsBySource(user.ID)
		if err != nil {
			fmt.Printf("Failed to get relationships: %v\n", err)
			continue
		}
		
		var credentialID string
		for _, rel := range rels {
			if relationship, ok := rel.(*models.EntityRelationship); ok {
				if relationship.Type == "has_credential" {
					credentialID = relationship.TargetID
					break
				}
			}
		}
		
		if credentialID == "" {
			fmt.Println("No credential relationship found!")
			continue
		}
		
		fmt.Printf("Credential ID: %s\n", credentialID)
		
		// Get credential entity
		cred, err := repo.GetByID(credentialID)
		if err != nil || cred == nil {
			fmt.Printf("Failed to get credential: %v\n", err)
			continue
		}
		
		// Extract salt from tags
		var salt string
		credTags := cred.GetTagsWithoutTimestamp()
		for _, tag := range credTags {
			if len(tag) > 5 && tag[:5] == "salt:" {
				salt = tag[5:]
				break
			}
		}
		
		fmt.Printf("Salt: %s\n", salt)
		fmt.Printf("Hash length: %d bytes\n", len(cred.Content))
		
		// Test password verification
		testPassword := "admin"
		testWithSalt := testPassword + salt
		
		err = bcrypt.CompareHashAndPassword(cred.Content, []byte(testWithSalt))
		if err == nil {
			fmt.Println("✓ Password 'admin' with salt verifies correctly")
		} else {
			fmt.Printf("✗ Password verification failed: %v\n", err)
			
			// Try without salt
			err = bcrypt.CompareHashAndPassword(cred.Content, []byte(testPassword))
			if err == nil {
				fmt.Println("✓ Password 'admin' WITHOUT salt verifies correctly")
			} else {
				fmt.Println("✗ Password verification failed both with and without salt")
			}
		}
		
		// Check user content
		if user.Content != nil && len(user.Content) > 0 {
			fmt.Printf("User has content (%d bytes)\n", len(user.Content))
			
			// Try to parse as JSON
			var content map[string]interface{}
			if err := json.Unmarshal(user.Content, &content); err == nil {
				fmt.Printf("Content: %+v\n", content)
			}
		}
	}
	
	fmt.Println("\n=====================================")
}