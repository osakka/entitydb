//go:build tool
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"entitydb/config"
	"entitydb/storage/binary"
)

func main() {
	
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "var"
	}
	
	fmt.Printf("=== EntityDB Index Analysis ===\n")
	fmt.Printf("Data path: %s\n\n", dataPath)
	
	// Load configuration to get proper database file path
	cfg := config.Load()
	if dataPath != "." {
		// Override the data path if provided
		cfg.DataPath = dataPath
		// Update database filename to use provided path
		cfg.DatabaseFilename = filepath.Join(dataPath, "entities.edb")
	}
	
	// Create a reader using configured database file
	reader, err := binary.NewReader(cfg.DatabaseFilename)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()
	
	// Read all entities from data file
	entities, err := reader.GetAllEntities()
	if err != nil {
		log.Fatalf("Failed to read entities: %v", err)
	}
	
	fmt.Printf("Entities in data file: %d\n", len(entities))
	
	// Count entities by type
	typeCount := make(map[string]int)
	userEntities := []string{}
	credentialEntities := []string{}
	relationshipEntities := []string{}
	
	for _, entity := range entities {
		entityType := "unknown"
		for _, tag := range entity.Tags {
			// Handle temporal tags
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			
			if strings.HasPrefix(actualTag, "type:") {
				entityType = strings.TrimPrefix(actualTag, "type:")
				break
			}
		}
		
		typeCount[entityType]++
		
		if entityType == "user" {
			userEntities = append(userEntities, entity.ID)
		} else if entityType == "credential" {
			credentialEntities = append(credentialEntities, entity.ID)
		} else if entityType == "relationship" {
			relationshipEntities = append(relationshipEntities, entity.ID)
		}
	}
	
	fmt.Printf("\nEntity types:\n")
	for t, count := range typeCount {
		fmt.Printf("  %s: %d\n", t, count)
	}
	
	fmt.Printf("\nUser entities: %v\n", userEntities)
	fmt.Printf("Credential entities: %v\n", credentialEntities)
	fmt.Printf("Relationship entities (first 5): %v\n", relationshipEntities[:min(5, len(relationshipEntities))])
	
	// Check WAL
	walPath := filepath.Join(dataPath, "entitydb.wal")
	if stat, err := os.Stat(walPath); err == nil {
		fmt.Printf("\nWAL file size: %d bytes\n", stat.Size())
		
		// Create WAL reader
		wal, err := binary.NewWAL(dataPath)
		if err != nil {
			log.Printf("Failed to create WAL: %v", err)
		} else {
			// Count entries in WAL
			walCount := 0
			err = wal.Replay(func(entry binary.WALEntry) error {
				walCount++
				if walCount <= 5 {
					fmt.Printf("WAL entry %d: EntityID=%s\n", walCount, entry.EntityID)
				}
				return nil
			})
			
			if err != nil {
				log.Printf("Failed to replay WAL: %v", err)
			} else {
				fmt.Printf("Total WAL entries: %d\n", walCount)
			}
		}
	} else {
		fmt.Printf("\nNo WAL file found\n")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}