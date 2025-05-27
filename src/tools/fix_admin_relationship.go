package main

import (
	"fmt"
	"log"
	"os"
	
	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <data_path>\n", os.Args[0])
		os.Exit(1)
	}
	
	dataPath := os.Args[1]
	
	// Create repository
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	// Find the admin user
	userEntities, err := repo.ListByTag("identity:username:admin")
	if err != nil || len(userEntities) == 0 {
		log.Fatalf("Admin user not found")
	}
	
	adminUser := userEntities[0]
	fmt.Printf("Found admin user: %s\n", adminUser.ID)
	
	// Find the credential for this user
	credEntities, err := repo.ListByTag("user:" + adminUser.ID)
	if err != nil || len(credEntities) == 0 {
		log.Fatalf("Credential not found for admin user")
	}
	
	credential := credEntities[0]
	fmt.Printf("Found credential: %s\n", credential.ID)
	
	// Check if relationship already exists
	relationships, _ := repo.GetRelationshipsBySource(adminUser.ID)
	hasCredentialRel := false
	for _, rel := range relationships {
		if relationship, ok := rel.(*models.EntityRelationship); ok {
			if relationship.TargetID == credential.ID && 
			   (relationship.Type == "has_credential" || relationship.RelationshipType == "has_credential") {
				hasCredentialRel = true
				fmt.Printf("Relationship already exists: %s\n", relationship.ID)
				break
			}
		}
	}
	
	if !hasCredentialRel {
		// Create the missing relationship
		relationship := &models.EntityRelationship{
			ID:               "rel_" + models.GenerateUUID(),
			SourceID:         adminUser.ID,
			TargetID:         credential.ID,
			Type:             "has_credential",
			RelationshipType: "has_credential",
			Properties:       map[string]string{"primary": "true"},
			CreatedAt:        models.Now(),
		}
		
		if err := repo.CreateRelationship(relationship); err != nil {
			log.Fatalf("Failed to create relationship: %v", err)
		}
		
		fmt.Printf("Created has_credential relationship: %s\n", relationship.ID)
	}
	
	fmt.Println("\nAdmin user relationships fixed!")
}