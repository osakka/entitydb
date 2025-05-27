package main

import (
	"entitydb/storage/binary"
	"fmt"
	"log"
	"strings"
)

func main() {
	// Open repository
	repo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()
	
	// List all entities
	entities, err := repo.List()
	if err != nil {
		log.Fatalf("Failed to list entities: %v", err)
	}
	
	// Find all users
	fmt.Println("=== USERS ===")
	users := make(map[string]string)
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			actualTag := tag
			if idx := strings.Index(tag, "|"); idx > 0 {
				actualTag = tag[idx+1:]
			}
			
			if actualTag == "type:user" {
				// Find username
				username := ""
				for _, t := range entity.Tags {
					actualT := t
					if idx := strings.Index(t, "|"); idx > 0 {
						actualT = t[idx+1:]
					}
					if strings.HasPrefix(actualT, "identity:username:") {
						username = strings.TrimPrefix(actualT, "identity:username:")
						break
					}
				}
				users[entity.ID] = username
				fmt.Printf("%s: %s\n", entity.ID, username)
				break
			}
		}
	}
	
	// Find all credentials
	fmt.Println("\n=== CREDENTIALS ===")
	credentials := make(map[string]string)
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			actualTag := tag
			if idx := strings.Index(tag, "|"); idx > 0 {
				actualTag = tag[idx+1:]
			}
			
			if actualTag == "type:credential" {
				// Find user reference
				userRef := ""
				for _, t := range entity.Tags {
					actualT := t
					if idx := strings.Index(t, "|"); idx > 0 {
						actualT = t[idx+1:]
					}
					if strings.HasPrefix(actualT, "user:") {
						userRef = strings.TrimPrefix(actualT, "user:")
						break
					}
				}
				credentials[entity.ID] = userRef
				fmt.Printf("%s -> user: %s\n", entity.ID, userRef)
				break
			}
		}
	}
	
	// Find all relationships
	fmt.Println("\n=== RELATIONSHIPS ===")
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			actualTag := tag
			if idx := strings.Index(tag, "|"); idx > 0 {
				actualTag = tag[idx+1:]
			}
			
			if actualTag == "_relationship" {
				// Find source and target
				source := ""
				target := ""
				for _, t := range entity.Tags {
					actualT := t
					if idx := strings.Index(t, "|"); idx > 0 {
						actualT = t[idx+1:]
					}
					if strings.HasPrefix(actualT, "_source:") {
						source = strings.TrimPrefix(actualT, "_source:")
					} else if strings.HasPrefix(actualT, "_target:") {
						target = strings.TrimPrefix(actualT, "_target:")
					}
				}
				fmt.Printf("%s: %s -> %s\n", entity.ID, source, target)
				
				// Check if it's a user-credential relationship
				if sourceUser, ok := users[source]; ok {
					if _, ok := credentials[target]; ok {
						fmt.Printf("  ^ This is a credential relationship for user '%s'\n", sourceUser)
					}
				}
				break
			}
		}
	}
	
	// Check GetRelationshipsBySource
	fmt.Println("\n=== TESTING GetRelationshipsBySource ===")
	for userID, username := range users {
		fmt.Printf("\nChecking relationships for user %s (%s):\n", username, userID)
		rels, err := repo.GetRelationshipsBySource(userID)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Found %d relationships\n", len(rels))
			for _, rel := range rels {
				fmt.Printf("  - %+v\n", rel)
			}
		}
	}
}