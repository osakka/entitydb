package main

import (
	"fmt"
	"log"
	"strings"
	"entitydb/storage/binary"
)

func main() {
	// Create a high performance repository
	baseRepo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to create base repository: %v", err)
	}
	defer baseRepo.Close()

	highPerfRepo, err := binary.NewHighPerformanceRepository("../var")
	if err != nil {
		log.Fatalf("Failed to create high performance repository: %v", err)
	}

	// Find admin user
	fmt.Println("Looking for admin user...")
	adminUsers, err := highPerfRepo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list admin users: %v", err)
	}

	fmt.Printf("Found %d admin users\n", len(adminUsers))
	
	if len(adminUsers) == 0 {
		fmt.Println("No admin user found!")
		return
	}

	adminUser := adminUsers[0]
	fmt.Printf("\nAdmin User ID: %s\n", adminUser.ID)
	fmt.Println("Admin User Tags:")
	
	hasCredTag := false
	credID := ""
	for _, tag := range adminUser.GetTagsWithoutTimestamp() {
		fmt.Printf("  - %s\n", tag)
		if strings.HasPrefix(tag, "relationship:has_credential:") {
			hasCredTag = true
			credID = strings.TrimPrefix(tag, "relationship:has_credential:")
			fmt.Printf("    -> Found credential relationship tag pointing to: %s\n", credID)
		}
	}

	if !hasCredTag {
		fmt.Println("\nNO CREDENTIAL RELATIONSHIP TAG FOUND!")
	}

	// Check for credential entities
	fmt.Println("\nSearching for credential entities...")
	
	// Try to find by user tag
	credsByUser, err := highPerfRepo.ListByTag("user:" + adminUser.ID)
	if err == nil {
		fmt.Printf("Found %d entities with user:%s tag\n", len(credsByUser), adminUser.ID)
		for _, e := range credsByUser {
			tags := e.GetTagsWithoutTimestamp()
			for _, t := range tags {
				if t == "type:credential" {
					fmt.Printf("  -> Found credential entity: %s\n", e.ID)
					break
				}
			}
		}
	}

	// If we have a credential ID from the tag, try to fetch it
	if credID != "" {
		fmt.Printf("\nTrying to fetch credential %s...\n", credID)
		cred, err := highPerfRepo.GetByID(credID)
		if err != nil {
			fmt.Printf("ERROR: Failed to fetch credential: %v\n", err)
		} else {
			fmt.Println("SUCCESS: Credential entity found!")
			fmt.Printf("Content length: %d bytes\n", len(cred.Content))
			fmt.Println("Credential tags:")
			for _, tag := range cred.GetTagsWithoutTimestamp() {
				fmt.Printf("  - %s\n", tag)
			}
		}
	}
}