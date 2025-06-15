//go:build tool
package main

import (
	"entitydb/storage/binary"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Parse command line arguments
	dataPath := flag.String("data", "/opt/entitydb/var", "Path to the data directory")
	entityIDs := flag.String("entities", "", "Comma-separated list of entity IDs to delete")
	flag.Parse()

	if *entityIDs == "" {
		fmt.Println("Error: --entities parameter is required")
		fmt.Println("Usage: delete_entities --entities=id1,id2,id3")
		os.Exit(1)
	}

	// Parse entity IDs
	ids := strings.Split(*entityIDs, ",")
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
	}

	fmt.Printf("Deleting %d entities from %s\n", len(ids), *dataPath)

	// Create binary repository
	repo, err := binary.NewEntityRepository(*dataPath)
	if err != nil {
		fmt.Printf("Error creating repository: %v\n", err)
		os.Exit(1)
	}

	// Delete each entity
	for _, entityID := range ids {
		fmt.Printf("Deleting entity: %s\n", entityID)
		
		// Check if entity exists first
		_, err := repo.GetByID(entityID)
		if err != nil {
			fmt.Printf("Warning: Entity %s not found: %v\n", entityID, err)
			continue
		}

		// Delete the entity
		err = repo.Delete(entityID)
		if err != nil {
			fmt.Printf("Error deleting entity %s: %v\n", entityID, err)
			continue
		}
		
		fmt.Printf("Successfully deleted entity: %s\n", entityID)
	}

	fmt.Println("Entity deletion completed")
}