package main

import (
	"entitydb/storage/binary"
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"strings"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize logger
	logger.SetLogLevel("info")
	
	// Create high-performance repository
	repo, err := binary.NewHighPerformanceRepository("/opt/entitydb/var")
	if err != nil {
		logger.Fatalf("Failed to create repository: %v", err)
	}
	
	// Find all user entities
	users, err := repo.ListByTag("type:user")
	if err != nil {
		logger.Fatalf("Failed to list users: %v", err)
	}
	
	fmt.Printf("Found %d users\n", len(users))
	
	// Find admin user
	var adminUser *models.Entity
	for _, user := range users {
		username := user.GetContentValue("username")
		if username == "admin" {
			adminUser = user
			break
		}
	}
	
	if adminUser == nil {
		fmt.Println("Admin user not found!")
		return
	}
	
	fmt.Printf("Found admin user: %s\n", adminUser.ID)
	
	// Print tags
	fmt.Println("Tags:")
	for _, tag := range adminUser.Tags {
		if strings.Contains(tag, "|") {
			parts := strings.SplitN(tag, "|", 2)
			fmt.Printf("  %s (timestamp: %s)\n", parts[1], parts[0])
		} else {
			fmt.Printf("  %s\n", tag)
		}
	}
	
	// Print content
	fmt.Println("Content:")
	for _, content := range adminUser.Content {
		if content.Type == "password_hash" {
			fmt.Printf("  %s: %s... (length: %d)\n", content.Type, content.Value[:20], len(content.Value))
		} else {
			fmt.Printf("  %s: %s\n", content.Type, content.Value)
		}
	}
	
	// Test password verification
	passwordHash := adminUser.GetContentValue("password_hash")
	fmt.Printf("\nPassword hash: %s\n", passwordHash)
	
	// Test with "admin" password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("admin"))
	if err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
		
		// Let's generate a new hash and compare
		newHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		fmt.Printf("New hash would be: %s\n", newHash)
		
		// Test with new hash
		err = bcrypt.CompareHashAndPassword(newHash, []byte("admin"))
		fmt.Printf("New hash verification: %v\n", err)
	} else {
		fmt.Println("Password verification successful!")
	}
}