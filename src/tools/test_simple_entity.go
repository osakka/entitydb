package main

import (
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"os"
)

func main() {
	// Clean test directory
	testDir := "test_entity_simple"
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Create repository
	repo, err := binary.NewEntityRepository(testDir)
	if err != nil {
		fmt.Printf("ERROR: Failed to create repository: %v\n", err)
		return
	}

	// Create a simple test entity
	testEntity := &models.Entity{
		ID:   "test_entity_simple",
		Tags: []string{"type:test", "name:simple"},
		Content: []byte("test content"),
	}

	fmt.Printf("Creating entity with ID: %s\n", testEntity.ID)
	
	// Create the entity
	err = repo.Create(testEntity)
	if err != nil {
		fmt.Printf("ERROR: Failed to create entity: %v\n", err)
		return
	}

	fmt.Printf("Entity created successfully\n")

	// Try to retrieve it immediately
	retrieved, err := repo.GetByID(testEntity.ID)
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