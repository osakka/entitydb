//go:build tool
package main

import (
	"entitydb/config"
	"entitydb/storage/binary"
	"flag"
	"fmt"
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
	
	fmt.Printf("=== EntityDB Entity Analysis ===\n")

	// Open repository using configured path
	repo, err := binary.NewEntityRepositoryWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}

	// Get all entities
	entities, err := repo.List()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}

	fmt.Printf("Total entities found: %d\n\n", len(entities))

	// Categorize entities and find potential bloated metric entities
	metricEntities := []entityInfo{}
	otherEntities := map[string][]entityInfo{}

	for _, entity := range entities {
		info := entityInfo{
			ID:       entity.ID,
			TagCount: len(entity.Tags),
			Type:     getEntityType(entity.Tags),
		}

		// Check if it's a metric-related entity
		if strings.Contains(info.Type, "metric") || strings.HasPrefix(entity.ID, "metric_") || 
		   strings.HasPrefix(entity.ID, "measurement_") || strings.HasPrefix(entity.ID, "metric_definition_") {
			metricEntities = append(metricEntities, info)
		} else {
			otherEntities[info.Type] = append(otherEntities[info.Type], info)
		}
	}

	// Sort metric entities by tag count (descending)
	sort.Slice(metricEntities, func(i, j int) bool {
		return metricEntities[i].TagCount > metricEntities[j].TagCount
	})

	// Report metric entities
	fmt.Printf("=== METRIC ENTITIES ===\n")
	fmt.Printf("Found %d metric-related entities:\n", len(metricEntities))
	
	bloatedCount := 0
	for i, entity := range metricEntities {
		status := ""
		if entity.TagCount > 100 {
			status = " *** BLOATED ***"
			bloatedCount++
		} else if entity.TagCount > 50 {
			status = " ** HIGH **"
		}
		
		fmt.Printf("%3d. %-50s Type: %-20s Tags: %4d%s\n", 
			i+1, entity.ID, entity.Type, entity.TagCount, status)
	}
	
	if bloatedCount > 0 {
		fmt.Printf("\n*** WARNING: Found %d metric entities with >100 tags that may cause performance issues ***\n", bloatedCount)
	} else {
		fmt.Printf("\n✓ No bloated metric entities found (all have ≤100 tags)\n")
	}

	// Report other entities by type
	fmt.Printf("\n=== OTHER ENTITIES BY TYPE ===\n")
	for entityType, typeEntities := range otherEntities {
		maxTags := 0
		minTags := 999999
		totalTags := 0
		
		for _, entity := range typeEntities {
			if entity.TagCount > maxTags {
				maxTags = entity.TagCount
			}
			if entity.TagCount < minTags {
				minTags = entity.TagCount
			}
			totalTags += entity.TagCount
		}
		
		avgTags := float64(totalTags) / float64(len(typeEntities))
		
		fmt.Printf("%-20s: %3d entities, Tags: min=%d, max=%d, avg=%.1f\n", 
			entityType, len(typeEntities), minTags, maxTags, avgTags)
		
		// Show entities with high tag counts
		highTagEntities := []entityInfo{}
		for _, entity := range typeEntities {
			if entity.TagCount > 50 {
				highTagEntities = append(highTagEntities, entity)
			}
		}
		
		if len(highTagEntities) > 0 {
			sort.Slice(highTagEntities, func(i, j int) bool {
				return highTagEntities[i].TagCount > highTagEntities[j].TagCount
			})
			
			fmt.Printf("  High tag count entities:\n")
			for _, entity := range highTagEntities {
				status := ""
				if entity.TagCount > 100 {
					status = " *** BLOATED ***"
				}
				fmt.Printf("    %-40s Tags: %d%s\n", entity.ID, entity.TagCount, status)
			}
		}
	}

	// Summary
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total entities: %d\n", len(entities))
	fmt.Printf("Metric entities: %d\n", len(metricEntities))
	fmt.Printf("Bloated entities (>100 tags): %d\n", bloatedCount)
	
	if bloatedCount > 0 {
		fmt.Printf("\nRecommendation: Clean up bloated metric entities before re-enabling background metrics collector\n")
	} else {
		fmt.Printf("\nRecommendation: Background metrics collector can likely be safely re-enabled\n")
	}
}

type entityInfo struct {
	ID       string
	TagCount int
	Type     string
}

func getEntityType(tags []string) string {
	for _, tag := range tags {
		// Handle temporal tags
		parts := strings.SplitN(tag, "|", 2)
		actualTag := tag
		if len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "type:") {
			return strings.TrimPrefix(actualTag, "type:")
		}
	}
	return "unknown"
}