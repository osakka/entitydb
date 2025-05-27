package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	
	"entitydb/storage/binary"
	"entitydb/models"
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
	
	// Find all admin users
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list admin users: %v", err)
	}
	
	fmt.Printf("Found %d admin users\n\n", len(adminUsers))
	
	// Check each admin user
	validAdmins := []struct {
		User       *models.Entity
		Credential *models.Entity
		Timestamp  time.Time
	}{}
	
	for i, user := range adminUsers {
		fmt.Printf("Admin user %d: %s\n", i+1, user.ID)
		
		// Get creation timestamp from first tag
		var timestamp time.Time
		if len(user.Tags) > 0 {
			parts := strings.SplitN(user.Tags[0], "|", 2)
			if len(parts) == 2 {
				if nanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					timestamp = time.Unix(0, nanos)
					fmt.Printf("  Created: %s\n", timestamp.Format(time.RFC3339))
				}
			}
		}
		
		// Check for credential relationships
		relationships, err := repo.GetRelationshipsBySource(user.ID)
		if err != nil {
			fmt.Printf("  Error getting relationships: %v\n", err)
			continue
		}
		
		fmt.Printf("  Relationships: %d\n", len(relationships))
		
		// Look for has_credential relationship
		var credentialID string
		for _, rel := range relationships {
			if relObj, ok := rel.(*models.EntityRelationship); ok {
				if relObj.RelationshipType == "has_credential" {
					credentialID = relObj.TargetID
					fmt.Printf("  Has credential: %s\n", credentialID)
					break
				}
			}
		}
		
		// If has credential, verify it exists
		if credentialID != "" {
			cred, err := repo.GetByID(credentialID)
			if err != nil {
				fmt.Printf("  ERROR: Credential not found: %v\n", err)
			} else {
				fmt.Printf("  ✓ Credential exists\n")
				validAdmins = append(validAdmins, struct {
					User       *models.Entity
					Credential *models.Entity
					Timestamp  time.Time
				}{user, cred, timestamp})
			}
		} else {
			fmt.Printf("  ✗ No credential relationship\n")
		}
		
		fmt.Println()
	}
	
	// Show summary
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total admin users: %d\n", len(adminUsers))
	fmt.Printf("Valid admin users (with credentials): %d\n", len(validAdmins))
	
	if len(validAdmins) > 0 {
		// Find most recent valid admin
		mostRecent := validAdmins[0]
		for _, admin := range validAdmins {
			if admin.Timestamp.After(mostRecent.Timestamp) {
				mostRecent = admin
			}
		}
		
		fmt.Printf("\nMost recent valid admin:\n")
		fmt.Printf("  User ID: %s\n", mostRecent.User.ID)
		fmt.Printf("  Created: %s\n", mostRecent.Timestamp.Format(time.RFC3339))
		fmt.Printf("  Credential ID: %s\n", mostRecent.Credential.ID)
	}
}