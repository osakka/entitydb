package main

import (
	"fmt"
	"log"
	"os"

	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <data_path>\n", os.Args[0])
		os.Exit(1)
	}

	dataPath := os.Args[1]

	// Open repository
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Create security manager
	securityManager := models.NewSecurityManager(repo)

	// Create admin user using the proper security manager method
	adminUser, err := securityManager.CreateUser("admin", "admin", "admin@entitydb.local")
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Created admin user: %s\n", adminUser.ID)

	// Add admin role by recreating entity with all tags
	adminEntity := &models.Entity{
		ID: adminUser.ID,
		Tags: append(adminUser.Entity.Tags, "rbac:role:admin"),
		Content: adminUser.Entity.Content,
	}
	if err := repo.Update(adminEntity); err != nil {
		log.Fatalf("Failed to add admin role: %v", err)
	}

	// Create relationship to administrators group
	relationship := &models.EntityRelationship{
		ID:               "rel_" + models.GenerateUUID(),
		SourceID:         adminUser.ID,
		TargetID:         "group_administrators",
		Type:             "member_of",
		RelationshipType: "member_of",
		Properties:       map[string]string{},
		CreatedAt:        models.Now(),
	}

	if err := repo.CreateRelationship(relationship); err != nil {
		// Ignore if already exists
		fmt.Printf("Note: Group membership may already exist\n")
	}

	fmt.Printf("\n=== Admin User Created Successfully ===\n")
	fmt.Printf("Username: admin\n")
	fmt.Printf("Password: admin\n")
	fmt.Printf("ID: %s\n", adminUser.ID)
}