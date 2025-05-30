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
	
	// Find and delete all admin-related entities
	adminUserID := "user_aa2c927c2a55d59778673ba23eaa8a89"
	
	// Delete all relationships for admin user
	relationships, err := repo.GetRelationshipsBySource(adminUserID)
	if err == nil {
		for _, rel := range relationships {
			if entityRel, ok := rel.(*models.EntityRelationship); ok {
				fmt.Printf("Deleting relationship: %s\n", entityRel.ID)
				repo.DeleteRelationship(entityRel.ID)
			}
		}
	}
	
	// Delete the admin user entity
	err = repo.Delete(adminUserID)
	if err != nil {
		fmt.Printf("Failed to delete admin user: %v\n", err)
	} else {
		fmt.Printf("Deleted admin user entity\n")
	}
	
	// Find and delete all credential entities that might be related
	credentialEntities, _ := repo.ListByTag("type:credential")
	for _, cred := range credentialEntities {
		fmt.Printf("Deleting credential entity: %s\n", cred.ID)
		repo.Delete(cred.ID)
	}
	
	fmt.Printf("\nAdmin user has been reset. The system will recreate it on next startup.\n")
}