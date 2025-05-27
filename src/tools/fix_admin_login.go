package main

import (
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"log"
)

func main() {
	// Open repository factory
	factory := &binary.RepositoryFactory{}
	repo, err := factory.CreateRepository("../var")
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	
	fmt.Println("Checking for existing admin user...")
	
	// Get the binary repository directly
	var binaryRepo *binary.EntityRepository
	switch r := repo.(type) {
	case *binary.CachedRepository:
		if hr, ok := r.GetUnderlying().(*binary.HighPerformanceRepository); ok {
			binaryRepo = hr.GetBaseRepository()
		} else if br, ok := r.GetUnderlying().(*binary.EntityRepository); ok {
			binaryRepo = br
		}
	case *binary.HighPerformanceRepository:
		binaryRepo = r.GetBaseRepository()
	case *binary.EntityRepository:
		binaryRepo = r
	}
	
	if binaryRepo == nil {
		log.Fatal("Cannot access binary repository")
	}
	
	// List all entities to find admin user
	entities, err := binaryRepo.List()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}
	
	var adminUser *models.Entity
	var adminCred *models.Entity
	
	// Find admin user and credential
	for _, entity := range entities {
		tags := entity.GetTagsWithoutTimestamp()
		
		// Check if it's a user entity
		isUser := false
		isAdmin := false
		for _, tag := range tags {
			if tag == "type:user" {
				isUser = true
			}
			if tag == "identity:username:admin" {
				isAdmin = true
			}
		}
		
		if isUser && isAdmin {
			adminUser = entity
			fmt.Printf("Found admin user: %s\n", entity.ID)
		}
		
		// Check if it's a credential for admin
		if adminUser != nil {
			for _, tag := range tags {
				if tag == "type:credential" {
					// Check if it references our admin user
					for _, t := range tags {
						if t == "user:"+adminUser.ID {
							adminCred = entity
							fmt.Printf("Found admin credential: %s\n", entity.ID)
						}
					}
				}
			}
		}
	}
	
	// Find the credential relationship
	for _, entity := range entities {
		tags := entity.GetTagsWithoutTimestamp()
		isRelationship := false
		correctSource := false
		correctTarget := false
		
		for _, tag := range tags {
			if tag == "_relationship" {
				isRelationship = true
			}
			if adminUser != nil && tag == "_source:"+adminUser.ID {
				correctSource = true
			}
			if adminCred != nil && tag == "_target:"+adminCred.ID {
				correctTarget = true
			}
		}
		
		if isRelationship && correctSource && correctTarget {
			fmt.Printf("Found credential relationship: %s\n", entity.ID)
		}
	}
	
	if adminUser == nil {
		fmt.Println("No admin user found!")
		return
	}
	
	if adminCred == nil {
		fmt.Println("No admin credential found!")
		return
	}
	
	// Now we need to create a proper relationship entity that the SecurityManager can find
	fmt.Println("\nCreating proper relationship for SecurityManager...")
	
	// Create a proper relationship entity with the expected tags
	relID := fmt.Sprintf("rel_%s_has_credential_%s", adminUser.ID, adminCred.ID)
	properRelationship := &models.Entity{
		ID: relID,
		Tags: []string{
			"type:relationship",
			"source_id:" + adminUser.ID,
			"target_id:" + adminCred.ID,
			"relationship_type:has_credential",
			"created_by:fix_script",
		},
		Content: []byte(fmt.Sprintf(`{"relationship_type":"has_credential","source_id":"%s","target_id":"%s"}`, adminUser.ID, adminCred.ID)),
	}
	
	// Check if it already exists
	existing, err := binaryRepo.GetByID(relID)
	if err == nil && existing != nil {
		fmt.Println("Proper relationship already exists")
	} else {
		// Create it
		if err := binaryRepo.Create(properRelationship); err != nil {
			log.Fatalf("Failed to create relationship: %v", err)
		}
		fmt.Println("Created proper relationship entity")
	}
	
	// Test authentication
	fmt.Println("\nTesting authentication...")
	securityManager := models.NewSecurityManager(repo)
	
	authUser, err := securityManager.AuthenticateUser("admin", "admin")
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		
		// Let's also try with the init password
		fmt.Println("\nTrying with default init password...")
		authUser, err = securityManager.AuthenticateUser("admin", "admin123")
		if err != nil {
			fmt.Printf("Authentication with admin123 also failed: %v\n", err)
		} else {
			fmt.Println("Authentication successful with admin123!")
			fmt.Printf("User ID: %s\n", authUser.ID)
		}
	} else {
		fmt.Println("Authentication successful with admin!")
		fmt.Printf("User ID: %s\n", authUser.ID)
	}
	
	// The tag index will be saved when the repository closes
	fmt.Println("\nTag index will be updated on close...")
}