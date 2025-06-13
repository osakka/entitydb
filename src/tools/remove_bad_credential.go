//go:build tool
package main

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"os"
)

func main() {
	// Set up logging
	logger.SetLogLevel("INFO")
	
	// Open the repository
	repo, err := binary.NewHighPerformanceRepository("/opt/entitydb/var/")
	if err != nil {
		fmt.Printf("Failed to create repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()
	
	// Get admin user relationships
	adminUserID := "user_aa2c927c2a55d59778673ba23eaa8a89"
	relationships, err := repo.GetRelationshipsBySource(adminUserID)
	if err != nil {
		fmt.Printf("Failed to get relationships: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Found %d relationships for admin user\n", len(relationships))
	
	// Find and remove the relationship to the bad credential
	badCredentialID := "cred_d9d8ede8bf7024d67002dab001af4945"
	
	for _, rel := range relationships {
		if entityRel, ok := rel.(*models.EntityRelationship); ok {
			fmt.Printf("Checking relationship %s: %s -> %s (type: %s)\n", 
				entityRel.ID, entityRel.SourceID, entityRel.TargetID, entityRel.Type)
			
			if entityRel.TargetID == badCredentialID && 
			   (entityRel.Type == "has_credential" || entityRel.RelationshipType == "has_credential") {
				fmt.Printf("Found bad credential relationship: %s\n", entityRel.ID)
				
				// Delete the relationship using just the relationship ID
				err := repo.DeleteRelationship(entityRel.ID)
				if err != nil {
					fmt.Printf("Failed to delete relationship: %v\n", err)
					continue
				}
				fmt.Printf("Successfully deleted bad credential relationship: %s\n", entityRel.ID)
			}
		}
	}
	
	// Verify the change
	updatedRelationships, err := repo.GetRelationshipsBySource(adminUserID)
	if err != nil {
		fmt.Printf("Failed to verify relationships: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("\nAdmin user now has %d relationships:\n", len(updatedRelationships))
	for i, rel := range updatedRelationships {
		if entityRel, ok := rel.(*models.EntityRelationship); ok {
			fmt.Printf("  %d. %s -> %s (type: %s)\n", i+1, entityRel.SourceID, entityRel.TargetID, entityRel.Type)
		}
	}
}