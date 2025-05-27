package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	
	"entitydb/models"
	"entitydb/storage/binary"
	
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <data_path>\n", os.Args[0])
		os.Exit(1)
	}
	
	dataPath := os.Args[1]
	
	// Create repository
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	// Find admin user
	fmt.Println("=== Looking for admin user ===")
	userEntities, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}
	
	fmt.Printf("Found %d users with username:admin\n", len(userEntities))
	
	for _, user := range userEntities {
		fmt.Printf("\nUser ID: %s\n", user.ID)
		fmt.Println("Tags:")
		for _, tag := range user.GetTagsWithoutTimestamp() {
			fmt.Printf("  - %s\n", tag)
		}
		
		// Get relationships
		relationships, err := repo.GetRelationshipsBySource(user.ID)
		if err != nil {
			fmt.Printf("  Error getting relationships: %v\n", err)
			continue
		}
		
		fmt.Printf("  Relationships: %d\n", len(relationships))
		
		// Look for credential relationships
		for _, rel := range relationships {
			if relationship, ok := rel.(*models.EntityRelationship); ok {
				fmt.Printf("    - %s -> %s (Type: %s, RelationshipType: %s)\n", 
					relationship.SourceID, relationship.TargetID, 
					relationship.Type, relationship.RelationshipType)
				
				// If it's a credential relationship, check the credential
				if relationship.Type == "has_credential" || relationship.RelationshipType == "has_credential" {
					cred, err := repo.GetByID(relationship.TargetID)
					if err != nil {
						fmt.Printf("      Error getting credential: %v\n", err)
						continue
					}
					
					fmt.Printf("      Credential ID: %s\n", cred.ID)
					fmt.Printf("      Credential Tags:\n")
					for _, tag := range cred.GetTagsWithoutTimestamp() {
						fmt.Printf("        - %s\n", tag)
					}
					
					// Extract salt
					var salt string
					for _, tag := range cred.GetTagsWithoutTimestamp() {
						if strings.HasPrefix(tag, "salt:") {
							salt = strings.TrimPrefix(tag, "salt:")
							break
						}
					}
					
					// Test password
					testPassword := "admin"
					err = bcrypt.CompareHashAndPassword(cred.Content, []byte(testPassword+salt))
					if err == nil {
						fmt.Printf("      ✓ Password '%s' is VALID\n", testPassword)
					} else {
						fmt.Printf("      ✗ Password '%s' is INVALID: %v\n", testPassword, err)
						
						// Try without salt
						err = bcrypt.CompareHashAndPassword(cred.Content, []byte(testPassword))
						if err == nil {
							fmt.Printf("      ✓ Password '%s' is VALID (without salt)\n", testPassword)
						}
					}
					
					fmt.Printf("      Credential content length: %d bytes\n", len(cred.Content))
				}
			}
		}
	}
	
	// Also check all credentials
	fmt.Println("\n=== All Credentials ===")
	credEntities, err := repo.ListByTag("type:credential")
	if err == nil {
		fmt.Printf("Found %d total credentials\n", len(credEntities))
		for _, cred := range credEntities {
			fmt.Printf("  - %s\n", cred.ID)
			for _, tag := range cred.GetTagsWithoutTimestamp() {
				if strings.HasPrefix(tag, "user:") {
					fmt.Printf("    For user: %s\n", strings.TrimPrefix(tag, "user:"))
				}
			}
		}
	}
}