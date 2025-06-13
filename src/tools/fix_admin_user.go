package main

import (
	"crypto/rand"
	"encoding/hex"
	"entitydb/config"
	"flag"
	"fmt"
	"log"

	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"golang.org/x/crypto/bcrypt"
)

func generateSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func main() {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	logger.SetLogLevel("INFO")
	
	factory := &binary.RepositoryFactory{}
	repo, err := factory.CreateRepository(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	// Look for existing admin user
	adminUsers, _ := repo.ListByTag("identity:username:admin")
	
	var adminID string
	if len(adminUsers) > 0 {
		adminID = adminUsers[0].ID
		fmt.Printf("Found existing admin user: %s\n", adminID)
		
		// Check for existing credential relationship
		relationships, _ := repo.GetRelationshipsBySource(adminID)
		for _, rel := range relationships {
			if entityRel, ok := rel.(*models.EntityRelationship); ok {
				if entityRel.Type == models.RelationshipHasCredential || entityRel.RelationshipType == models.RelationshipHasCredential {
					fmt.Printf("Found existing credential relationship: %s -> %s\n", entityRel.SourceID, entityRel.TargetID)
					fmt.Println("Admin user already has credentials.")
					return
				}
			}
		}
	} else {
		// Create admin user
		adminID = "user_admin_" + models.GenerateUUID()
		userEntity := &models.Entity{
			ID: adminID,
			Tags: []string{
				"type:user",
				"dataset:_system",
				"identity:username:admin",
				"identity:uuid:" + adminID,
				"status:active",
				"profile:email:admin@entitydb.local",
				"created:" + models.NowString(),
				"rbac:role:admin",
			},
			Content:   []byte{},
			CreatedAt: models.Now(),
			UpdatedAt: models.Now(),
		}
		
		if err := repo.Create(userEntity); err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		fmt.Printf("Created admin user: %s\n", adminID)
	}
	
	// Create credential
	credentialID := "cred_admin_" + models.GenerateUUID()
	salt := generateSalt()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"+salt), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	credentialEntity := &models.Entity{
		ID: credentialID,
		Tags: []string{
			"type:credential",
			"dataset:_system",
			"algorithm:bcrypt",
			"user:" + adminID,
			"salt:" + salt,
			"created:" + models.NowString(),
		},
		Content:   hashedPassword,
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}
	
	if err := repo.Create(credentialEntity); err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}
	fmt.Printf("Created credential: %s\n", credentialID)
	
	// Create relationship
	relationship := &models.EntityRelationship{
		ID:               "rel_admin_cred_" + models.GenerateUUID(),
		SourceID:         adminID,
		TargetID:         credentialID,
		Type:             models.RelationshipHasCredential,
		RelationshipType: models.RelationshipHasCredential,
		Properties:       map[string]string{"primary": "true"},
		CreatedAt:        models.Now(),
	}
	
	if err := repo.CreateRelationship(relationship); err != nil {
		log.Fatalf("Failed to create relationship: %v", err)
	}
	fmt.Printf("Created credential relationship: %s\n", relationship.ID)
	
	fmt.Println("\nAdmin user setup complete!")
}
