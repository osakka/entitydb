package main

import (
	"fmt"
	"log"
	"strings"

	"entitydb/config"
	"entitydb/storage/binary"
)

func main() {
	// Initialize configuration
	cfg := config.Load()
	
	// Create repository
	repo, err := binary.NewEntityRepositoryWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Define test entity types to clean up
	testTypes := []string{
		"type:concurrent_test",
		"type:persistence_test", 
		"type:related_test",
		"type:test",
	}

	totalDeleted := 0

	for _, testType := range testTypes {
		fmt.Printf("Cleaning up entities with tag: %s\n", testType)
		
		// Get entities with this test type
		entities, err := repo.ListByTag(testType)
		if err != nil {
			fmt.Printf("Error listing entities with tag %s: %v\n", testType, err)
			continue
		}

		fmt.Printf("Found %d entities to delete\n", len(entities))

		// Delete each entity
		for _, entity := range entities {
			// Skip if this is not actually a test entity (safety check)
			isTestEntity := false
			for _, tag := range entity.Tags {
				cleanTag := strings.TrimSpace(tag)
				if strings.HasPrefix(cleanTag, "type:") {
					entityType := strings.TrimPrefix(cleanTag, "type:")
					if strings.Contains(entityType, "test") || entityType == "test" {
						isTestEntity = true
						break
					}
				}
			}

			if !isTestEntity {
				fmt.Printf("Skipping %s - not a test entity\n", entity.ID)
				continue
			}

			fmt.Printf("Deleting test entity: %s\n", entity.ID)
			err := repo.Delete(entity.ID)
			if err != nil {
				fmt.Printf("Error deleting entity %s: %v\n", entity.ID, err)
			} else {
				totalDeleted++
			}
		}
	}

	fmt.Printf("\nCleanup complete! Deleted %d test entities.\n", totalDeleted)
	
	// Get final count by examining all entity types
	fmt.Println("\nCurrent entity breakdown:")
	entityTypes := map[string]int{}
	
	// Check various common entity type tags
	typePatterns := []string{"type:user", "type:session", "type:metric", "type:measurement", "type:metric_definition", "type:config"}
	totalRemaining := 0
	
	for _, pattern := range typePatterns {
		entities, err := repo.ListByTag(pattern)
		if err == nil && len(entities) > 0 {
			entityTypes[pattern] = len(entities)
			totalRemaining += len(entities)
			fmt.Printf("  %s: %d entities\n", pattern, len(entities))
		}
	}
	
	fmt.Printf("\nTotal remaining entities: %d\n", totalRemaining)
}