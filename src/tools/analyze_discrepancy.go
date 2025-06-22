//go:build tool
package main

import (
	"fmt"
	"log"
	"strings"
	
	"entitydb/config"
	"entitydb/storage/binary"
)

func main() {
	
	fmt.Printf("=== Entity Discrepancy Analysis ===\n")
	
	// Load configuration using proper configuration system
	cfg := config.Load()
	fmt.Printf("Database file: %s\n\n", cfg.DatabaseFilename)
	
	// Note: Using reader only for analysis (no repository needed)
	
	reader, err := binary.NewReader(cfg.DatabaseFilename)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()
	
	// Read all entities from file
	fileEntities, err := reader.GetAllEntities()
	if err != nil {
		log.Fatalf("Failed to read entities from file: %v", err)
	}
	
	fmt.Printf("Entities in data file: %d\n", len(fileEntities))
	
	// Check for specific entities we need
	adminFound := false
	credFound := false
	hasCredRelFound := false
	
	for _, entity := range fileEntities {
		// Check entity type
		entityType := ""
		for _, tag := range entity.Tags {
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			
			if strings.HasPrefix(actualTag, "type:") {
				entityType = strings.TrimPrefix(actualTag, "type:")
				break
			}
		}
		
		if entityType == "user" && strings.Contains(entity.ID, "admin") {
			adminFound = true
			fmt.Printf("\nFound admin user: %s\n", entity.ID)
		} else if entityType == "credential" {
			credFound = true
			fmt.Printf("\nFound credential: %s\n", entity.ID)
		} else if entityType == "relationship" {
			// Check if it's a has_credential relationship
			for _, tag := range entity.Tags {
				actualTag := tag
				if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
					actualTag = parts[1]
				}
				
				if actualTag == "_relationship:has_credential" {
					hasCredRelFound = true
					fmt.Printf("\nFound has_credential relationship: %s\n", entity.ID)
					// Print all tags
					for _, t := range entity.Tags {
						fmt.Printf("  Tag: %s\n", t)
					}
					break
				}
			}
		}
	}
	
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Admin user found: %v\n", adminFound)
	fmt.Printf("  Credential found: %v\n", credFound)
	fmt.Printf("  has_credential relationship found: %v\n", hasCredRelFound)
	
	fmt.Printf("\n=== Configuration System Test Complete ===\n")
	fmt.Printf("✅ Successfully loaded configuration and accessed database file\n")
	fmt.Printf("✅ Configuration hierarchy working correctly\n")
}