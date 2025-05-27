package main

import (
	"crypto/rand"
	"encoding/hex"
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"log"
	"time"
	
	"golang.org/x/crypto/bcrypt"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	// Open repository
	repo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()
	
	// Create admin user
	username := "admin"
	password := "admin"
	email := "admin@example.com"
	
	// Generate user ID
	userID := "user_" + generateUUID()
	
	// Create user entity with proper tags
	userEntity := &models.Entity{
		ID: userID,
		Tags: []string{
			"type:user",
			"identity:username:" + username,
			"identity:uuid:" + userID,
			"status:active",
			"profile:email:" + email,
			"created:" + time.Now().Format(time.RFC3339),
		},
		Content: nil, // User entities don't have content
	}
	
	// Create credential
	credID := "cred_" + generateUUID()
	saltBytes := make([]byte, 16)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)
	
	// Hash password with salt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	credEntity := &models.Entity{
		ID: credID,
		Tags: []string{
			"type:credential",
			"algorithm:bcrypt",
			"user:" + userID,
			"salt:" + salt,
			"created:" + time.Now().Format(time.RFC3339),
		},
		Content: hashedPassword,
	}
	
	// Create entities
	if err := repo.Create(userEntity); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	
	if err := repo.Create(credEntity); err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}
	
	// Create relationship between user and credential
	relID := "rel_" + generateUUID()
	relEntity := &models.Entity{
		ID: relID,
		Tags: []string{
			"_relationship",
			"_source:" + userID,
			"_target:" + credID,
			"content:type:relationship",
		},
		Content: []byte(fmt.Sprintf(`{"relationship_type":"has_credential","source_id":"%s","target_id":"%s","properties":{"primary":"true"}}`, userID, credID)),
	}
	
	if err := repo.Create(relEntity); err != nil {
		log.Fatalf("Failed to create relationship: %v", err)
	}
	
	// Create admin role assignment
	roleAdminID := "role_admin"
	roleRelID := "rel_" + generateUUID()
	roleRelEntity := &models.Entity{
		ID: roleRelID,
		Tags: []string{
			"_relationship",
			"_source:" + userID,
			"_target:" + roleAdminID,
			"content:type:relationship",
		},
		Content: []byte(fmt.Sprintf(`{"relationship_type":"has_role","source_id":"%s","target_id":"%s"}`, userID, roleAdminID)),
	}
	
	if err := repo.Create(roleRelEntity); err != nil {
		log.Fatalf("Failed to create role relationship: %v", err)
	}
	
	fmt.Printf("Admin user created successfully:\n")
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Credential ID: %s\n", credID)
	fmt.Printf("Role Assignment: admin\n")
}