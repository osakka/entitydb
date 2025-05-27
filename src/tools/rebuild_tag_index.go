package main

import (
	"entitydb/storage/binary"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rebuild_tag_index <data_path>")
		fmt.Println("Example: rebuild_tag_index ../var")
		os.Exit(1)
	}
	
	dataPath := os.Args[1]
	
	// Delete the old corrupt index
	indexPath := dataPath + "/entities.idx"
	if err := os.Remove(indexPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove old index: %v", err)
	} else {
		fmt.Println("Removed old tag index")
	}
	
	// Open repository to force index rebuild
	fmt.Println("Opening repository to rebuild index...")
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()
	
	// List all entities to verify
	entities, err := repo.List()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}
	
	fmt.Printf("Successfully rebuilt index with %d entities\n", len(entities))
	
	// Test a search
	fmt.Println("\nTesting search for identity:username:admin...")
	results, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Printf("Search error: %v", err)
	} else {
		fmt.Printf("Found %d results\n", len(results))
		for _, entity := range results {
			fmt.Printf("  - %s\n", entity.ID)
		}
	}
}