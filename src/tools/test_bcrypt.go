package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin"
	
	// Generate a hash
	hash1, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash1: %v\n", err)
		return
	}
	fmt.Printf("Hash 1: %s\n", hash1)
	
	// Generate another hash
	hash2, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash2: %v\n", err)
		return
	}
	fmt.Printf("Hash 2: %s\n", hash2)
	
	// Test comparison
	err = bcrypt.CompareHashAndPassword(hash1, []byte(password))
	fmt.Printf("Compare with hash1: %v\n", err)
	
	err = bcrypt.CompareHashAndPassword(hash2, []byte(password))
	fmt.Printf("Compare with hash2: %v\n", err)
	
	// Test with the actual hashes from the database
	testHashes := []string{
		"$2a$10$KGD",
		"$2a$10$QyW",
	}
	
	for i, h := range testHashes {
		fmt.Printf("\nTesting hash %d prefix: %s\n", i+1, h)
	}
}