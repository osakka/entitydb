package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"entitydb/models"
	"entitydb/storage/binary"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <data_path>\n", os.Args[0])
		os.Exit(1)
	}

	dataPath := os.Args[1]
	os.Setenv("ENTITYDB_LOG_LEVEL", "ERROR")

	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Find all admin users
	userEntities, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to find admin users: %v", err)
	}
	
	fmt.Printf("Found %d admin users to consolidate\n", len(userEntities))
	
	// Delete all admin users and their relationships
	for _, user := range userEntities {
		fmt.Printf("Deleting admin user: %s\n", user.ID)
		
		// Delete relationships
		rels, _ := repo.GetRelationshipsBySource(user.ID)
		for _, rel := range rels {
			if relationship, ok := rel.(*models.EntityRelationship); ok {
				fmt.Printf("  Deleting relationship: %s\n", relationship.ID)
				repo.Delete(relationship.ID)
			}
		}
		
		// Delete user
		repo.Delete(user.ID)
	}
	
	// Delete all orphaned credentials
	credEntities, _ := repo.ListByTag("type:credential")
	for _, cred := range credEntities {
		fmt.Printf("Deleting credential: %s\n", cred.ID)
		repo.Delete(cred.ID)
	}
	
	// Create a single clean admin user
	adminID := fmt.Sprintf("user_%x", time.Now().UnixNano())
	adminUser := &models.Entity{
		ID: adminID,
		Tags: []string{
			"type:user",
			fmt.Sprintf("identity:username:admin"),
			fmt.Sprintf("identity:uuid:%s", adminID),
			"status:active",
			"profile:email:admin@entitydb.local",
			fmt.Sprintf("created:%d", time.Now().UnixNano()),
			"rbac:role:admin",
		},
	}
	
	if err := repo.Create(adminUser); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	
	fmt.Printf("\nCreated new admin user: %s\n", adminID)
	
	// Create credential with proper salt
	salt := fmt.Sprintf("%x", time.Now().UnixNano())
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
			fmt.Sprintf("user:%s", adminID),
			fmt.Sprintf("salt:%s", salt),
			fmt.Sprintf("created:%d", time.Now().UnixNano()),
		},
		Content: hash,
	}
	
	if err := repo.Create(credEntity); err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}
	
	fmt.Printf("Created credential: %s\n", credID)
	
	// Create has_credential relationship
	relID := fmt.Sprintf("rel_%x", time.Now().UnixNano())
	relationship := &models.EntityRelationship{
		ID:               relID,
		SourceID:         adminID,
		TargetID:         credID,
		Type:             "has_credential",
		RelationshipType: "has_credential",
		Properties:       map[string]string{"primary": "true"},
		CreatedAt:        time.Now().UnixNano(),
	}
	
	if err := repo.CreateRelationship(relationship); err != nil {
		log.Fatalf("Failed to create credential relationship: %v", err)
	}
	
	fmt.Printf("Created credential relationship: %s\n", relID)
	
	// Create member_of relationship to administrators group
	relID2 := fmt.Sprintf("rel_%x", time.Now().UnixNano())
	relationship2 := &models.EntityRelationship{
		ID:               relID2,
		SourceID:         adminID,
		TargetID:         "group_administrators",
		Type:             "member_of",
		RelationshipType: "member_of",
		Properties:       map[string]string{},
		CreatedAt:        time.Now().UnixNano(),
	}
	
	if err := repo.CreateRelationship(relationship2); err != nil {
		log.Fatalf("Failed to create group membership: %v", err)
	}
	
	fmt.Printf("Created group membership: %s\n", relID2)
	
	fmt.Printf("\n=== SUCCESS ===\n")
	fmt.Printf("Admin user consolidated!\n")
	fmt.Printf("Username: admin\n")
	fmt.Printf("Password: admin\n")
	fmt.Printf("User ID: %s\n", adminID)
}