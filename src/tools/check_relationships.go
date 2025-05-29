package main

import (
	"fmt"
	"log"
	"entitydb/storage/binary"
)

func main() {
	fmt.Println("EntityDB Relationship Check")
	fmt.Println("===========================")
	
	// Open relationship repository
	relRepo, err := binary.NewRelationshipRepository("/opt/entitydb/var")
	if err != nil {
		log.Fatalf("Failed to open relationship repository: %v", err)
	}
	defer relRepo.Close()
	
	// Find all relationships
	fmt.Println("\nChecking ALL relationships:")
	
	// Get all entities first to find relationships
	entityRepo, err := binary.NewEntityRepository("/opt/entitydb/var")
	if err != nil {
		log.Fatalf("Failed to open entity repository: %v", err)
	}
	defer entityRepo.Close()
	
	// Get all entities that might be relationships
	entities, err := entityRepo.ListByTag("type:relationship")
	if err != nil {
		fmt.Printf("Error getting relationships by type tag: %v\n", err)
	} else {
		fmt.Printf("Found %d entities with type:relationship tag\n", len(entities))
	}
	
	// Try to get relationships for admin users
	adminUsers := []string{
		"user_ba74552389c637b47ce1d61aba04d4a9",
		"user_9d2077fd9bc38fce1fb75fab03f62dff", 
		"user_cb8a9c9ba5fe2e2e7c904536b95b1b7a",
	}
	
	for _, userID := range adminUsers {
		fmt.Printf("\nChecking relationships for %s:\n", userID)
		
		rels, err := relRepo.GetByFromEntity(userID)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Found %d relationships\n", len(rels))
			for _, rel := range rels {
				fmt.Printf("    - %s -> %s (type: %s)\n", rel.FromEntityID, rel.ToEntityID, rel.Type)
			}
		}
	}
	
	// Check for credential entities
	fmt.Println("\nChecking credential entities:")
	creds, err := entityRepo.ListByTag("type:credential")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d credential entities\n", len(creds))
		for _, cred := range creds {
			fmt.Printf("  - ID: %s\n", cred.ID)
		}
	}
	
	// Check for has_credential relationships stored as entities
	fmt.Println("\nChecking for has_credential relationship entities:")
	hasCredRels, err := entityRepo.ListByTag("relationship:type:has_credential")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d has_credential relationship entities\n", len(hasCredRels))
		for _, rel := range hasCredRels {
			fmt.Printf("  - ID: %s, Tags: %v\n", rel.ID, rel.GetTagsWithoutTimestamp())
		}
	}
	
	// Check for relationship entities with from: tags
	fmt.Println("\nChecking for relationship entities with from: tags:")
	for _, userID := range adminUsers {
		fromTag := fmt.Sprintf("relationship:from:%s", userID)
		fromRels, err := entityRepo.ListByTag(fromTag)
		if err != nil {
			fmt.Printf("Error checking %s: %v\n", fromTag, err)
		} else {
			fmt.Printf("Found %d relationships from %s\n", len(fromRels), userID)
			for _, rel := range fromRels {
				fmt.Printf("  - ID: %s, Tags: %v\n", rel.ID, rel.GetTagsWithoutTimestamp())
			}
		}
	}
}

