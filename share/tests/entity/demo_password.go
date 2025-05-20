package main

import (
	"fmt"
	"log"
	
	"golang.org/x/crypto/bcrypt"
)

// hashPassword securely hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// validatePassword checks if the provided password matches the stored hash
func validatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func main() {
	// Example password
	password := "SecurePassword123"
	
	// Hash the password
	hash, err := hashPassword(password)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}
	
	fmt.Printf("Original password: %s\n", password)
	fmt.Printf("Password hash: %s\n", hash)
	
	// Validate correct password
	isValid := validatePassword(password, hash)
	fmt.Printf("\nValidating correct password: %v\n", isValid)
	
	// Validate incorrect password
	isValid = validatePassword("WrongPassword", hash)
	fmt.Printf("Validating incorrect password: %v\n", isValid)
}