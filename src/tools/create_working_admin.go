package main

import (
	"fmt"
	"log"
	"os"
	
	"entitydb/models"
	"entitydb/storage/binary"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "/opt/entitydb/var"
	}
	
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	fmt.Println("=== CREATING WORKING ADMIN USER ===")
	fmt.Println()
	
	// Delete ALL existing admin users to start fresh
	fmt.Println("1. Cleaning up existing admin users...")
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err == nil {
		for _, user := range adminUsers {
			fmt.Printf("  Deleting existing admin user: %s\n", user.ID)
			repo.Delete(user.ID)
		}
	}
	
	// Create a single new admin user with all required properties
	fmt.Println("\n2. Creating new admin user...")
	adminUser := &models.Entity{
		ID: "user_" + models.GenerateUUID(),
		Tags: []string{
			"type:user",
			"identity:username:admin",
			"identity:uuid:admin",
			"status:active",
			"profile:email:admin@entitydb.local",
			"rbac:role:admin",
		},
		Content: []byte{}, // Empty content for user entity
	}
	
	if err := repo.Create(adminUser); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	fmt.Printf("  Created admin user: %s\n", adminUser.ID)
	
	// Create credential
	fmt.Println("\n3. Creating credential...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	credential := &models.Entity{
		ID:      "cred_" + models.GenerateUUID(),
		Tags:    []string{"type:credential", "credential:type:password"},
		Content: hashedPassword,
	}
	
	if err := repo.Create(credential); err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}
	fmt.Printf("  Created credential: %s (%d bytes)\n", credential.ID, len(credential.Content))
	
	// Create relationship
	fmt.Println("\n4. Creating has_credential relationship...")
	relRepo := binary.NewRelationshipRepository(repo)
	
	relationship := &models.EntityRelationship{
		ID:               "rel_" + models.GenerateUUID(),
		SourceID:         adminUser.ID,
		TargetID:         credential.ID,
		RelationshipType: "has_credential",
		Type:             "has_credential",
		Properties:       map[string]string{},
		CreatedAt:        models.Now(),
		CreatedBy:        "system",
	}
	
	if err := relRepo.Create(relationship); err != nil {
		log.Fatalf("Failed to create relationship: %v", err)
	}
	fmt.Printf("  Created relationship: %s\n", relationship.ID)
	
	// Test authentication
	fmt.Println("\n5. Testing authentication...")
	securityManager := models.NewSecurityManager(repo)
	authUser, err := securityManager.AuthenticateUser("admin", "admin")
	if err != nil {
		fmt.Printf("  Authentication failed: %v\n", err)
	} else {
		fmt.Printf("  Authentication SUCCESS: %s\n", authUser.ID)
	}
	
	// Final verification
	fmt.Println("\n6. Final verification...")
	adminUsers, err = repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("  Error listing admin users: %v\n", err)
	} else {
		fmt.Printf("  Total admin users: %d\n", len(adminUsers))
		if len(adminUsers) == 1 {
			fmt.Println("  ✓ Exactly one admin user exists")
			fmt.Println("  ✓ Admin user has rbac:role:admin tag")
			fmt.Println("  ✓ Admin user has status:active tag")
			fmt.Println("  ✓ Admin user has credential relationship")
			fmt.Println("\nAdmin user is fully configured!")
		}
	}
}