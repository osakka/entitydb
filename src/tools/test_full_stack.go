package main

import (
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"os"
)

func main() {
	// Clean test directory
	testDir := "test_full_stack"
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Create the same repository stack as the main server
	fmt.Printf("Creating repository stack like main server...\n")
	
	// 1. Create TemporalRepository (which creates HighPerformanceRepository -> EntityRepository)
	temporalRepo, err := binary.NewTemporalRepository(testDir)
	if err != nil {
		fmt.Printf("ERROR: Failed to create temporal repository: %v\n", err)
		return
	}

	// 2. Wrap with CachedRepository
	finalRepo := binary.NewCachedRepository(temporalRepo, 300) // 5 minutes

	fmt.Printf("Repository stack created successfully\n")

	// Create a test entity with a known ID
	testEntity := &models.Entity{
		ID:   "perm_entity_view",
		Tags: []string{"type:permission", "resource:entity", "action:view"},
		Content: []byte("permission content"),
	}

	fmt.Printf("Creating entity with ID: %s\n", testEntity.ID)
	
	// Create the entity
	err = finalRepo.Create(testEntity)
	if err != nil {
		fmt.Printf("ERROR: Failed to create entity: %v\n", err)
		return
	}

	fmt.Printf("Entity created successfully\n")

	// Try to retrieve it immediately
	retrieved, err := finalRepo.GetByID(testEntity.ID)
	if err != nil {
		fmt.Printf("ERROR: Failed to retrieve entity: %v\n", err)
		return
	}

	if retrieved == nil {
		fmt.Printf("ERROR: Retrieved entity is nil\n")
		return
	}

	fmt.Printf("SUCCESS: Entity retrieved with ID: %s\n", retrieved.ID)
	fmt.Printf("Tags: %v\n", retrieved.Tags)
	fmt.Printf("Content: %s\n", string(retrieved.Content))
}