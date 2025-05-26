package main

import (
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"os"
	"time"
)

func main() {
	// Clean test directory
	testDir := "test_security_init"
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Create the same repository stack as the main server
	fmt.Printf("Creating repository stack...\n")
	
	temporalRepo, err := binary.NewTemporalRepository(testDir)
	if err != nil {
		fmt.Printf("ERROR: Failed to create temporal repository: %v\n", err)
		return
	}

	finalRepo := binary.NewCachedRepository(temporalRepo, 300) // 5 minutes

	// Create security manager and initializer
	securityManager := models.NewSecurityManager(finalRepo)
	securityInit := models.NewSecurityInitializer(securityManager, finalRepo)

	fmt.Printf("Testing step-by-step security initialization...\n")

	// Test creating one permission
	fmt.Printf("Creating perm_entity_view...\n")
	permissionEntity := &models.Entity{
		ID: "perm_entity_view",
		Tags: []string{
			"type:" + models.EntityTypePermission,
			"resource:entity",
			"action:view",
			"scope:global",
			"created:" + models.NowString(),
		},
		Content:   nil,
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}

	if err := finalRepo.Create(permissionEntity); err != nil {
		fmt.Printf("ERROR: Failed to create permission: %v\n", err)
		return
	}

	fmt.Printf("Permission created successfully\n")

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Try to retrieve it
	fmt.Printf("Attempting to retrieve perm_entity_view...\n")
	retrieved, err := finalRepo.GetByID("perm_entity_view")
	if err != nil {
		fmt.Printf("ERROR: Failed to retrieve permission: %v\n", err)
		return
	}

	if retrieved == nil {
		fmt.Printf("ERROR: Retrieved permission is nil\n")
		return
	}

	fmt.Printf("SUCCESS: Permission retrieved with ID: %s\n", retrieved.ID)

	// Now test the full security initialization
	fmt.Printf("\nTesting full security initialization...\n")
	if err := securityInit.InitializeDefaultSecurityEntities(); err != nil {
		fmt.Printf("ERROR: Security initialization failed: %v\n", err)
		return
	}

	fmt.Printf("Security initialization completed\n")

	// Test retrieving a few key entities
	testEntities := []string{"perm_all", "role_admin", "group_administrators"}
	for _, entityID := range testEntities {
		fmt.Printf("Testing retrieval of %s...\n", entityID)
		entity, err := finalRepo.GetByID(entityID)
		if err != nil {
			fmt.Printf("  ERROR: Failed to retrieve %s: %v\n", entityID, err)
		} else if entity == nil {
			fmt.Printf("  ERROR: %s is nil\n", entityID)
		} else {
			fmt.Printf("  SUCCESS: %s retrieved\n", entityID)
		}
	}
}