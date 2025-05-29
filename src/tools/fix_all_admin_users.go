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
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "/opt/entitydb/var"
	}
	
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	fmt.Println("=== FIXING ALL ADMIN USERS ===")
	fmt.Println()
	
	// Find ALL users with username:admin
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list admin users: %v", err)
	}
	
	fmt.Printf("Found %d admin users total\n\n", len(adminUsers))
	
	// First, identify which user has the credential and make it the primary
	var primaryAdmin *models.Entity
	
	for _, user := range adminUsers {
		fmt.Printf("Checking user: %s\n", user.ID)
		
		// Check existing tags
		hasRBACAdmin := false
		hasActiveStatus := false
		
		for _, tag := range user.Tags {
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			if actualTag == "rbac:role:admin" {
				hasRBACAdmin = true
			}
			if actualTag == "status:active" {
				hasActiveStatus = true
			}
		}
		
		fmt.Printf("  Has rbac:role:admin: %v\n", hasRBACAdmin)
		fmt.Printf("  Has status:active: %v\n", hasActiveStatus)
		
		// Check for credential
		searchTag := "_source:" + user.ID
		entities, err := repo.ListByTag(searchTag)
		hasCredential := false
		if err == nil {
			for _, entity := range entities {
				for _, tag := range entity.Tags {
					actualTag := tag
					if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
						actualTag = parts[1]
					}
					if actualTag == "_relationship:has_credential" {
						hasCredential = true
						break
					}
				}
				if hasCredential {
					break
				}
			}
		}
		fmt.Printf("  Has credential: %v\n", hasCredential)
		
		// The primary admin should have all three
		if hasRBACAdmin && hasActiveStatus && hasCredential {
			primaryAdmin = user
			fmt.Printf("  *** This is the PRIMARY admin user ***\n")
		}
		
		fmt.Println()
	}
	
	// If we don't have a primary admin with all properties, fix the situation
	if primaryAdmin == nil {
		fmt.Println("No admin user has all required properties. Selecting one to fix...")
		
		// Prefer the one with rbac:role:admin
		for _, user := range adminUsers {
			hasRBACAdmin := false
			for _, tag := range user.Tags {
				actualTag := tag
				if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
					actualTag = parts[1]
				}
				if actualTag == "rbac:role:admin" {
					hasRBACAdmin = true
					break
				}
			}
			
			if hasRBACAdmin {
				primaryAdmin = user
				fmt.Printf("Selected user %s as primary (has rbac:role:admin)\n", user.ID)
				break
			}
		}
		
		// If still none, just take the first one
		if primaryAdmin == nil && len(adminUsers) > 0 {
			primaryAdmin = adminUsers[0]
			fmt.Printf("Selected user %s as primary (first found)\n", primaryAdmin.ID)
		}
	}
	
	if primaryAdmin == nil {
		log.Fatalf("No admin users found at all!")
	}
	
	// Now ensure the primary admin has all required properties
	fmt.Printf("\nEnsuring primary admin %s has all required properties...\n", primaryAdmin.ID)
	
	// 1. Ensure status:active
	hasActiveStatus := false
	for _, tag := range primaryAdmin.Tags {
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		if actualTag == "status:active" {
			hasActiveStatus = true
			break
		}
	}
	
	if !hasActiveStatus {
		fmt.Println("  Adding status:active...")
		primaryAdmin.Tags = append(primaryAdmin.Tags, "status:active")
		if err := repo.Update(primaryAdmin); err != nil {
			log.Printf("Failed to add status:active: %v", err)
		}
	}
	
	// 2. Ensure rbac:role:admin
	rbacManager := models.NewRBACTagManager(repo)
	if err := rbacManager.AssignRoleToUser(primaryAdmin.ID, "admin"); err != nil {
		log.Printf("Failed to ensure admin role: %v", err)
	} else {
		fmt.Println("  Ensured rbac:role:admin")
	}
	
	// 3. Ensure has_credential
	searchTag := "_source:" + primaryAdmin.ID
	entities, err := repo.ListByTag(searchTag)
	hasCredential := false
	if err == nil {
		for _, entity := range entities {
			for _, tag := range entity.Tags {
				actualTag := tag
				if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
					actualTag = parts[1]
				}
				if actualTag == "_relationship:has_credential" {
					hasCredential = true
					break
				}
			}
			if hasCredential {
				break
			}
		}
	}
	
	if !hasCredential {
		fmt.Println("  Creating credential...")
		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}
		
		credential := &models.Entity{
			ID:      models.GenerateUUID(),
			Tags:    []string{"type:credential", "credential:type:password"},
			Content: hashedPassword,
		}
		
		if err := repo.Create(credential); err != nil {
			log.Fatalf("Failed to create credential: %v", err)
		}
		
		relRepo := binary.NewRelationshipRepository(repo)
		relationship := &models.EntityRelationship{
			ID:               models.GenerateUUID(),
			SourceID:         primaryAdmin.ID,
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
		fmt.Printf("  Created credential and relationship\n")
	}
	
	// 4. Delete all other admin users to avoid conflicts
	fmt.Println("\nRemoving duplicate admin users...")
	for _, user := range adminUsers {
		if user.ID != primaryAdmin.ID {
			fmt.Printf("  Deleting duplicate admin user: %s\n", user.ID)
			if err := repo.Delete(user.ID); err != nil {
				log.Printf("    Failed to delete: %v", err)
			}
		}
	}
	
	fmt.Printf("\nPrimary admin user %s is now fully configured!\n", primaryAdmin.ID)
	fmt.Println("There should now be only ONE admin user with all required properties.")
}