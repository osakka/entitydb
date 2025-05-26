package main

import (
	"fmt"
	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	fmt.Println("Testing authentication system...")
	
	// Create repository using the same factory as the main server
	factory := &binary.RepositoryFactory{}
	entityRepo, err := factory.CreateRepository("/opt/entitydb/var")
	if err != nil {
		fmt.Printf("Failed to create repository: %v\n", err)
		return
	}
	// Note: EntityRepository interface doesn't have Close method
	
	// Create security manager
	securityManager := models.NewSecurityManager(entityRepo)
	
	// Test finding admin user
	fmt.Println("Looking for admin user...")
	userEntities, err := entityRepo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Error looking for admin user: %v\n", err)
	} else {
		fmt.Printf("Found %d entities with username admin\n", len(userEntities))
		for i, entity := range userEntities {
			fmt.Printf("  Entity %d: ID=%s\n", i, entity.ID)
			tags := entity.GetTagsWithoutTimestamp()
			fmt.Printf("    Tags: %v\n", tags)
		}
	}
	
	// Look for credential entities
	fmt.Println("Looking for credential entities...")
	credEntities, err := entityRepo.ListByTag("type:credential")
	if err != nil {
		fmt.Printf("Error looking for credentials: %v\n", err)
	} else {
		fmt.Printf("Found %d credential entities\n", len(credEntities))
		for i, entity := range credEntities {
			fmt.Printf("  Credential %d: ID=%s\n", i, entity.ID)
			tags := entity.GetTagsWithoutTimestamp()
			fmt.Printf("    Tags: %v\n", tags)
		}
	}
	
	// Test relationship lookup if we have a user
	if len(userEntities) > 0 {
		userID := userEntities[0].ID
		fmt.Printf("Looking for relationships for user %s...\n", userID)
		relationships, err := entityRepo.GetRelationshipsBySource(userID)
		if err != nil {
			fmt.Printf("Error getting relationships: %v\n", err)
		} else {
			fmt.Printf("Found %d relationships\n", len(relationships))
			for i, rel := range relationships {
				if relationship, ok := rel.(*models.EntityRelationship); ok {
					fmt.Printf("  Relationship %d: %s -> %s (type: %s)\n", i, relationship.SourceID, relationship.TargetID, relationship.Type)
				} else {
					fmt.Printf("  Relationship %d: %v (type: %T)\n", i, rel, rel)
				}
			}
		}
	}
	
	// Test authentication
	fmt.Println("Testing authentication...")
	user, err := securityManager.AuthenticateUser("admin", "admin")
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
	} else {
		fmt.Printf("Authentication succeeded: User ID=%s, Username=%s\n", user.ID, user.Username)
	}
}