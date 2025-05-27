package main

import (
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

	// Find admin user
	userEntities, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to find admin user: %v", err)
	}
	
	if len(userEntities) == 0 {
		log.Fatalf("No admin user found")
	}
	
	adminUser := userEntities[0]
	fmt.Printf("Found admin user: %s\n", adminUser.ID)
	
	// Check for existing credential relationships
	relationships, err := repo.GetRelationshipsBySource(adminUser.ID)
	if err != nil {
		fmt.Printf("Error getting relationships: %v\n", err)
	}
	
	// Check if credential relationship exists
	hasCredentialRel := false
	for _, rel := range relationships {
		if relationship, ok := rel.(*models.EntityRelationship); ok {
			if relationship.Type == "has_credential" {
				hasCredentialRel = true
				fmt.Printf("Found existing credential relationship: %s -> %s\n", relationship.SourceID, relationship.TargetID)
			}
		}
	}
	
	if !hasCredentialRel {
		// Create new credential entity with salt
		salt := generateSalt()
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"+salt), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to generate password hash: %v", err)
		}
		
		credID := fmt.Sprintf("cred_%x", time.Now().UnixNano())
		credEntity := &models.Entity{
			ID: credID,
			Tags: []string{
				"type:credential",
				"algorithm:bcrypt",
				fmt.Sprintf("user:%s", adminUser.ID),
				fmt.Sprintf("salt:%s", salt),
				fmt.Sprintf("created:%d", time.Now().UnixNano()),
			},
			Content: hash,
		}
		
		if err := repo.Create(credEntity); err != nil {
			log.Fatalf("Failed to create credential: %v", err)
		}
		
		fmt.Printf("Created credential entity: %s\n", credID)
		
		// Create relationship
		relID := fmt.Sprintf("rel_%x", time.Now().UnixNano())
		relationship := &models.EntityRelationship{
			ID:         relID,
			SourceID:   adminUser.ID,
			TargetID:   credID,
			Type:       "has_credential",
			RelationshipType: "has_credential",
			Properties: map[string]string{"primary": "true"},
			CreatedAt:  time.Now().UnixNano(),
		}
		
		if err := repo.CreateRelationship(relationship); err != nil {
			log.Fatalf("Failed to create relationship: %v", err)
		}
		
		fmt.Printf("Created credential relationship: %s\n", relID)
	}
	
	// Ensure admin has role
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
		adminUser.Tags = append(adminUser.Tags, "rbac:role:admin")
		if err := repo.Update(adminUser); err != nil {
			log.Fatalf("Failed to update admin user: %v", err)
		}
		fmt.Printf("Added admin role to user\n")
	}
	
	// Check for administrators group membership
	hasMembership := false
	for _, rel := range relationships {
		if relationship, ok := rel.(*models.EntityRelationship); ok {
			if relationship.Type == "member_of" && relationship.TargetID == "group_administrators" {
				hasMembership = true
				break
			}
		}
	}
	
	if !hasMembership {
		relID := fmt.Sprintf("rel_%x", time.Now().UnixNano()) 
		relationship := &models.EntityRelationship{
			ID:         relID,
			SourceID:   adminUser.ID,
			TargetID:   "group_administrators",
			Type:       "member_of",
			RelationshipType: "member_of",
			Properties: map[string]string{},
			CreatedAt:  time.Now().UnixNano(),
		}
		
		if err := repo.CreateRelationship(relationship); err != nil {
			fmt.Printf("Failed to create group membership (may already exist): %v\n", err)
		} else {
			fmt.Printf("Created group membership relationship: %s\n", relID)
		}
	}
	
	fmt.Printf("\nAdmin user setup complete!\n")
	fmt.Printf("Username: admin\n")
	fmt.Printf("Password: admin\n")
}

func generateSalt() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}