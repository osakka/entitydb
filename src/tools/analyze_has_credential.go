//go:build tool
package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"entitydb/config"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	
	logger.Info("=== Analyzing has_credential Relationships ===")
	
	repo, err := binary.NewEntityRepositoryWithConfig(cfg)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	allEntities, err := repo.List()
	if err != nil {
		logger.Error("Failed to list entities: %v", err)
		os.Exit(1)
	}
	
	// Find all has_credential relationships
	fmt.Println("\nAnalyzing has_credential relationships:")
	fmt.Println("=====================================")
	
	for _, entity := range allEntities {
		// Check if this is a has_credential relationship
		isHasCredential := false
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "_relationship:has_credential") {
				isHasCredential = true
				break
			}
		}
		
		if isHasCredential {
			fmt.Printf("\nRelationship ID: %s\n", entity.ID)
			
			// Extract source and target from tags
			var sourceID, targetID string
			for _, tag := range entity.Tags {
				if strings.Contains(tag, "_source:") {
					parts := strings.SplitN(tag, "_source:", 2)
					if len(parts) == 2 {
						sourceID = parts[1]
					}
				} else if strings.Contains(tag, "_target:") {
					parts := strings.SplitN(tag, "_target:", 2)
					if len(parts) == 2 {
						targetID = parts[1]
					}
				}
			}
			
			fmt.Printf("  Source ID: %s (length: %d)\n", sourceID, len(sourceID))
			fmt.Printf("  Target ID: %s (length: %d)\n", targetID, len(targetID))
			
			// Find the actual user entity with this ID prefix
			fmt.Printf("  Looking for user entities with ID prefix: %s\n", sourceID[:len(sourceID)-1])
			
			for _, userEntity := range allEntities {
				if strings.HasPrefix(userEntity.ID, "user_") && strings.HasPrefix(sourceID, userEntity.ID) {
					fmt.Printf("  FOUND matching user: %s (length: %d)\n", userEntity.ID, len(userEntity.ID))
					fmt.Printf("  Source ID has %d extra chars\n", len(sourceID) - len(userEntity.ID))
				}
			}
		}
	}
}