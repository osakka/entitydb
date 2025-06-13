//go:build tool
package main

import (
	"entitydb/config"
	"entitydb/logger"
	"entitydb/storage/binary"
	"flag"
	"log"
	"sort"
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
	
	logger.Info("[main] === Dump Tag Index ===")

	// Open repository using configured path
	repo, err := binary.NewEntityRepository(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Get the tag index (we'll need to access it via reflection or make it accessible)
	// For now, let's use ListByTag with various patterns to probe the index
	
	// First, let's find all _source: tags
	logger.Info("[main] Looking for _source: tags...")
	
	// We'll search for specific user IDs we know exist
	userIDs := []string{
		"user_3b0d8f209787b0baadfd00e555979f8",
		"user_3b0d8f209787b0baadfd00e555979f83",
		"user_340a9f7fae086cfbfb67fb8421a7c16",
		"user_340a9f7fae086cfbfb67fb8421a7c161",
	}
	
	for _, userID := range userIDs {
		searchTag := "_source:" + userID
		entities, err := repo.ListByTag(searchTag)
		if err != nil {
			logger.Error("[main] Failed to search for tag %s: %v", searchTag, err)
			continue
		}
		
		if len(entities) > 0 {
			logger.Info("[main] Found %d entities with tag '%s':", len(entities), searchTag)
			for _, entity := range entities {
				logger.Info("[main]   - Entity ID: %s", entity.ID)
				// Show first few tags
				tagCount := len(entity.Tags)
				showCount := 5
				if tagCount < showCount {
					showCount = tagCount
				}
				for i := 0; i < showCount; i++ {
					logger.Info("[main]     Tag[%d]: %s", i, entity.Tags[i])
				}
				if tagCount > showCount {
					logger.Info("[main]     ... and %d more tags", tagCount-showCount)
				}
			}
		} else {
			logger.Info("[main] No entities found with tag '%s'", searchTag)
		}
	}
	
	// Also check for relationships
	logger.Info("[main] \nLooking for type:relationship entities...")
	relEntities, err := repo.ListByTag("type:relationship")
	if err != nil {
		logger.Error("[main] Failed to search for type:relationship: %v", err)
	} else {
		logger.Info("[main] Found %d entities with tag 'type:relationship'", len(relEntities))
		
		// Group by source
		sourceMap := make(map[string][]string)
		for _, entity := range relEntities {
			// Find _source tag
			sourceID := ""
			for _, tag := range entity.Tags {
				// Handle temporal tags
				parts := strings.SplitN(tag, "|", 2)
				actualTag := tag
				if len(parts) == 2 {
					actualTag = parts[1]
				}
				
				if strings.HasPrefix(actualTag, "_source:") {
					sourceID = strings.TrimPrefix(actualTag, "_source:")
					break
				}
			}
			
			if sourceID != "" {
				sourceMap[sourceID] = append(sourceMap[sourceID], entity.ID)
			}
		}
		
		// Sort and display
		sources := make([]string, 0, len(sourceMap))
		for source := range sourceMap {
			sources = append(sources, source)
		}
		sort.Strings(sources)
		
		logger.Info("[main] \nRelationships by source:")
		for _, source := range sources {
			logger.Info("[main]   Source: %s => %d relationships", source, len(sourceMap[source]))
			for _, relID := range sourceMap[source] {
				logger.Info("[main]     - %s", relID)
			}
		}
	}
	
	// Check specific relationship we know should exist
	logger.Info("[main] \nChecking specific relationship rel_9de03ca57e584fd7a658663cff67f297...")
	rel, err := repo.GetByID("rel_9de03ca57e584fd7a658663cff67f297")
	if err != nil {
		logger.Error("[main] Failed to get relationship: %v", err)
	} else if rel != nil {
		logger.Info("[main] Found relationship with %d tags:", len(rel.Tags))
		for i, tag := range rel.Tags {
			logger.Info("[main]   Tag[%d]: %s", i, tag)
		}
		
		// Check if it has the expected _source tag
		hasSourceTag := false
		for _, tag := range rel.Tags {
			parts := strings.SplitN(tag, "|", 2)
			actualTag := tag
			if len(parts) == 2 {
				actualTag = parts[1]
			}
			if actualTag == "_source:user_3b0d8f209787b0baadfd00e555979f8" {
				hasSourceTag = true
				break
			}
		}
		logger.Info("[main] Has correct _source tag: %v", hasSourceTag)
	}
}