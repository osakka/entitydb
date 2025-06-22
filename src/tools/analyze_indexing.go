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
	
	fmt.Printf("=== EntityDB Index Analysis ===\n")
	
	// Load configuration using proper configuration system
	cfg := config.Load()
	
	// Allow override via environment variable
	if envDataPath := os.Getenv("ENTITYDB_DATA_PATH"); envDataPath != "" {
		cfg.DataPath = envDataPath
		// Reconstruct database filename for the new data path
		cfg.DatabaseFilename = filepath.Join(cfg.DataPath, "entities.edb")
	}
	
	fmt.Printf("Database file: %s\n\n", cfg.DatabaseFilename)
	
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
	
	// Note: WAL is embedded in unified .edb file format
	fmt.Printf("\nUnified database format: WAL embedded in %s\n", cfg.DatabaseFilename)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}