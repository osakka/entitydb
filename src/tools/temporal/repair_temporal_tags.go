//go:build tool
package main

import (
	"entitydb/storage/binary"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// Parse command line arguments
	dataPath := flag.String("data", "/opt/entitydb/var", "Path to the data directory")
	flag.Parse()

	// Create binary repository
	repo, err := binary.NewEntityRepository(*dataPath)
	if err != nil {
		fmt.Printf("Error creating repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting temporal tag fix...")
	start := time.Now()

	// Get all tags
	fmt.Println("Getting all entities...")
	entities, err := repo.List()
	if err != nil {
		fmt.Printf("Error getting entities: %v\n", err)
		os.Exit(1)
	}

	// Count stats
	tagCount := 0
	plainTagCount := 0

	// Rebuild tagIndex
	tagIndex := make(map[string][]string)

	// First pass - collect all tags
	fmt.Printf("Processing %d entities...\n", len(entities))
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			// Add the full tag
			tagIndex[tag] = append(tagIndex[tag], entity.ID)
			tagCount++

			// Extract and add the plain tag
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					plainTag := parts[1]
					tagIndex[plainTag] = append(tagIndex[plainTag], entity.ID)
					plainTagCount++
				}
			}
		}
	}

	// Set the tag index
	// Since we can't directly replace the tag index, we'll add all entities again
	// This is inefficient but doesn't require changing the repository code
	// In a real fix, we'd add a method to replace the tag index directly
	for _, entity := range entities {
		// First remove the entity from all indexes
		err := repo.Delete(entity.ID)
		if err != nil {
			fmt.Printf("Error deleting entity %s: %v\n", entity.ID, err)
			continue
		}

		// Then add it back
		err = repo.Create(entity)
		if err != nil {
			fmt.Printf("Error recreating entity %s: %v\n", entity.ID, err)
			continue
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Temporal tag fix completed in %v\n", elapsed)
	fmt.Printf("Processed %d tags and indexed %d plain tags\n", tagCount, plainTagCount)
}