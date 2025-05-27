package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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

	// Check if admin user already exists
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err == nil && len(adminUsers) > 0 {
		fmt.Printf("Admin user already exists: %s\n", adminUsers[0].ID)
		
		// Check if it has admin role
		hasRole := false
		for _, tag := range adminUsers[0].Tags {
			if tag == "rbac:role:admin" || strings.HasSuffix(tag, "|rbac:role:admin") {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			fmt.Printf("Adding admin role...\n")
			// Need to recreate entity with updated tags
			updatedEntity := &models.Entity{
				ID:      adminUsers[0].ID,
				Tags:    append(adminUsers[0].Tags, "rbac:role:admin"),
				Content: adminUsers[0].Content,
			}
			if err := repo.Update(updatedEntity); err != nil {
				fmt.Printf("Failed to add admin role: %v\n", err)
			}
		}
		
		// Try to authenticate
		user, err := securityManager.AuthenticateUser("admin", "admin")
		if err != nil {
			fmt.Printf("Authentication failed: %v\n", err)
			fmt.Printf("Admin user exists but cannot authenticate. Please check credentials.\n")
		} else {
			fmt.Printf("Authentication successful! Admin user is working.\n")
			fmt.Printf("Username: admin\n")
			fmt.Printf("Password: admin\n")
			fmt.Printf("User ID: %s\n", user.ID)
		}
		return
	}

	// Create admin user
	adminUser, err := securityManager.CreateUser("admin", "admin", "admin@entitydb.local")
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Created admin user: %s\n", adminUser.ID)

	// Add admin role
	userEntities, err := repo.ListByTag(fmt.Sprintf("identity:uuid:%s", adminUser.ID))
	if err != nil || len(userEntities) == 0 {
		log.Fatalf("Failed to find created user")
	}
	
	userEntity := userEntities[0]
	userEntity.Tags = append(userEntity.Tags, "rbac:role:admin")
	if err := repo.Update(userEntity); err != nil {
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
		fmt.Printf("Note: Group membership may already exist: %v\n", err)
	}

	fmt.Printf("\n=== Admin User Ready ===\n")
	fmt.Printf("Username: admin\n")
	fmt.Printf("Password: admin\n")
	fmt.Printf("ID: %s\n", adminUser.ID)
}