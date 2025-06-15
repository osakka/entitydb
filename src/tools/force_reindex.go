//go:build tool
package main

import (
	"entitydb/config"
	"entitydb/logger"
	"entitydb/storage/binary"
	"flag"
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
	
	logger.Info("[main] === Force Complete Reindex ===")

	// Remove the persistent index file to force a complete rebuild
	indexFile := cfg.DataPath + "/data/" + cfg.DatabaseFilename + cfg.IndexSuffix
	if _, err := os.Stat(indexFile); err == nil {
		logger.Info("[main] Removing existing index file: %s", indexFile)
		if err := os.Remove(indexFile); err != nil {
			logger.Error("[main] Failed to remove index file: %v", err)
		}
	}

	// Open repository - this will trigger a complete index rebuild
	logger.Info("[main] Opening repository to trigger reindex...")
	repo, err := binary.NewEntityRepositoryWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Force a complete WAL replay and index rebuild
	logger.Info("[main] Repository opened. Checking index health...")
	
	// List all entities to verify
	entities, err := repo.List()
	if err != nil {
		logger.Error("[main] Failed to list entities: %v", err)
	} else {
		logger.Info("[main] Total entities in repository: %d", len(entities))
		
		// Count by type
		typeCount := make(map[string]int)
		for _, entity := range entities {
			for _, tag := range entity.Tags {
				if tag == "type:relationship" || strings.Contains(tag, "|type:relationship") {
					typeCount["relationship"]++
					break
				} else if tag == "type:user" || strings.Contains(tag, "|type:user") {
					typeCount["user"]++
					break
				} else if tag == "type:credential" || strings.Contains(tag, "|type:credential") {
					typeCount["credential"]++
					break
				} else if tag == "type:permission" || strings.Contains(tag, "|type:permission") {
					typeCount["permission"]++
					break
				} else if tag == "type:role" || strings.Contains(tag, "|type:role") {
					typeCount["role"]++
					break
				} else if tag == "type:group" || strings.Contains(tag, "|type:group") {
					typeCount["group"]++
					break
				}
			}
		}
		
		logger.Info("[main] Entity breakdown by type:")
		for t, count := range typeCount {
			logger.Info("[main]   %s: %d", t, count)
		}
		
		// Specifically check for admin user and credential relationship
		logger.Info("[main] \nChecking critical authentication entities:")
		
		// Find admin users
		adminUsers, err := repo.ListByTag("identity:username:admin")
		if err != nil {
			logger.Error("[main] Failed to find admin users: %v", err)
		} else {
			logger.Info("[main] Found %d admin users", len(adminUsers))
			for _, user := range adminUsers {
				logger.Info("[main]   Admin user: %s", user.ID)
				
				// Check for credentials
				relationships, err := repo.GetRelationshipsBySource(user.ID)
				if err != nil {
					logger.Error("[main]   Failed to get relationships: %v", err)
				} else {
					logger.Info("[main]   Found %d relationships for %s", len(relationships), user.ID)
					for _, rel := range relationships {
						logger.Info("[main]     Relationship: %+v", rel)
					}
				}
			}
		}
	}
	
	logger.Info("[main] === Reindex Complete ===")
}