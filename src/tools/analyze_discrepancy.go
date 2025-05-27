package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"entitydb/storage/binary"
	"entitydb/models"
)

func main() {
	
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "var"
	}
	
	fmt.Printf("=== Entity Discrepancy Analysis ===\n")
	fmt.Printf("Data path: %s\n\n", dataPath)
	
	// Create repository to get the full picture
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	// Now let's manually check what's in the data file vs memory
	reader, err := binary.NewReader(filepath.Join(dataPath, "entities.ebf"))
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
	
	// Now let's see what GetRelationshipsBySource returns
	fmt.Printf("\n=== Testing GetRelationshipsBySource ===\n")
	adminID := "admin_bc098257eb7fa98e885df720bbaa5f9a"
	
	relationships, err := repo.GetRelationshipsBySource(adminID)
	if err != nil {
		fmt.Printf("Error getting relationships: %v\n", err)
	} else {
		fmt.Printf("GetRelationshipsBySource returned %d relationships\n", len(relationships))
		for i, rel := range relationships {
			if relObj, ok := rel.(*models.EntityRelationship); ok {
				fmt.Printf("  Relationship %d: ID=%s, Source=%s, Target=%s, Type=%s\n", 
					i+1, relObj.ID, relObj.SourceID, relObj.TargetID, relObj.RelationshipType)
			}
		}
	}
}