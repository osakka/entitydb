package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"fmt"
	"log"
)

func main() {
	// Enable debug logging
	logger.SetDebug(true)
	
	// Open repository
	repo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer func() {
		fmt.Println("Closing repository...")
	}()
	
	// Search for admin user
	fmt.Println("\n=== SEARCHING FOR identity:username:admin ===")
	results, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
	} else {
		fmt.Printf("Found %d entities\n", len(results))
		for _, entity := range results {
			fmt.Printf("  - %s\n", entity.ID)
		}
	}
}