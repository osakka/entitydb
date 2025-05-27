package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"entitydb/storage/binary"
)

func main() {
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "var"
	}
	
	dataFile := filepath.Join(dataPath, "entities.ebf")
	
	// Check file info
	stat, err := os.Stat(dataFile)
	if err != nil {
		log.Fatalf("Failed to stat data file: %v", err)
	}
	
	fmt.Printf("Data file: %s\n", dataFile)
	fmt.Printf("File size: %d bytes\n", stat.Size())
	
	// Create a reader
	reader, err := binary.NewReader(dataFile)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()
	
	// Try to read first few entities
	fmt.Printf("\nReading entities...\n")
	count := 0
	entities, err := reader.GetAllEntities()
	if err != nil {
		log.Fatalf("Failed to read entities: %v", err)
	}
	
	fmt.Printf("Total entities found: %d\n", len(entities))
	
	// Show first 5 entities
	for i, entity := range entities {
		if i >= 5 {
			break
		}
		fmt.Printf("\nEntity %d: ID=%s\n", i+1, entity.ID)
		fmt.Printf("  Tags: %v\n", entity.Tags)
		if len(entity.Content) > 0 {
			fmt.Printf("  Content length: %d bytes\n", len(entity.Content))
		}
		count++
	}
	
	if count < len(entities) {
		fmt.Printf("\n... and %d more entities\n", len(entities)-count)
	}
}