//go:build tool
package main

import (
	"entitydb/config"
	"entitydb/storage/binary"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	dataPath := cfg.DataPath
	
	// Delete the old corrupt index using configurable index path
	indexPath := cfg.DataPath + "/data/" + cfg.DatabaseFilename + cfg.IndexSuffix
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