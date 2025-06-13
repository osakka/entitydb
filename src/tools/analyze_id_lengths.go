//go:build tool
package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"fmt"
	"os"
	"strings"
)

func main() {
	logger.Info("=== Analyzing Entity ID Lengths ===")
	
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	allEntities, err := repo.List()
	if err != nil {
		logger.Error("Failed to list entities: %v", err)
		os.Exit(1)
	}
	
	// Analyze ID lengths
	idLengths := make(map[int]int)
	userIDs := []string{}
	credIDs := []string{}
	relIDs := []string{}
	
	for _, entity := range allEntities {
		idLen := len(entity.ID)
		idLengths[idLen]++
		
		if strings.HasPrefix(entity.ID, "user_") {
			userIDs = append(userIDs, entity.ID)
		} else if strings.HasPrefix(entity.ID, "cred_") {
			credIDs = append(credIDs, entity.ID)
		} else if strings.HasPrefix(entity.ID, "rel_") {
			relIDs = append(relIDs, entity.ID)
		}
	}
	
	// Print ID length distribution
	fmt.Println("\nID Length Distribution:")
	for length, count := range idLengths {
		fmt.Printf("  Length %d: %d entities\n", length, count)
	}
	
	// Print some example IDs
	fmt.Println("\nExample User IDs:")
	for i, id := range userIDs {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s (length: %d)\n", id, len(id))
	}
	
	fmt.Println("\nExample Credential IDs:")
	for i, id := range credIDs {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s (length: %d)\n", id, len(id))
	}
	
	fmt.Println("\nExample Relationship IDs:")
	for i, id := range relIDs {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s (length: %d)\n", id, len(id))
		
		// Also check the _source tags for this relationship
		for _, tag := range allEntities[i].Tags {
			if strings.Contains(tag, "_source:") {
				parts := strings.SplitN(tag, "_source:", 2)
				if len(parts) == 2 {
					sourceID := parts[1]
					fmt.Printf("    Source ID in tag: %s (length: %d)\n", sourceID, len(sourceID))
				}
			}
		}
	}
}